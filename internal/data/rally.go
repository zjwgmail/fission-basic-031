package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pojo/dto"
	"fission-basic/kit/sqlx"
)

type Rally struct {
	data *Data
	l    *log.Helper
}

func NewRally(
	d *Data,
	l log.Logger,
) *Rally {
	return &Rally{
		data: d,
		l:    log.NewHelper(l),
	}
}

func (o *Rally) saveMsgSend(ctx context.Context, t time.Time, db sqlx.DB, waMsgSends []*dto.WaMsgSend) error {
	// 发送消息
	for i := range waMsgSends {
		id, err := model.InsertWaMsgSend(ctx, db,
			&model.WaMsgSend{
				WaMsgID:       waMsgSends[i].WaMsgID,
				WaID:          waMsgSends[i].WaID,
				MsgType:       waMsgSends[i].MsgType,
				State:         waMsgSends[i].State,
				Content:       waMsgSends[i].Content,
				BuildMsgParam: waMsgSends[i].BuildMsgParam,
				SendRes:       waMsgSends[i].SendRes,
				CreateTime:    t,
				UpdateTime:    t,
				Del:           biz.NotDeleted,
			},
		)
		if err != nil {
			o.l.WithContext(ctx).
				Errorf("insert wa msg send failed, err=%v", err)
			return err
		}
		waMsgSends[i].ID = id
	}
	return nil
}

func (o *Rally) createGroup(ctx context.Context,
	userInfo *biz.UserInfo, lastSendTime int64,
	waID, helpCode string,
	waMsgSends []*dto.WaMsgSend,
	extend func(ctx context.Context, db sqlx.DB) error,
) error {
	now := time.Now()
	userInfoDO := model.UserInfo{
		WaID:       waID,
		HelpCode:   helpCode,
		Channel:    userInfo.Channel,
		Language:   userInfo.Language,
		Generation: userInfo.Generation,
		JoinCount:  userInfo.JoinCount,
		Nickname:   userInfo.Nickname,
		CreateTime: now,
		UpdateTime: now,
		Del:        biz.NotDeleted,
	}

	userCreateGroupDO := model.UserCreateGroup{
		CreateWaID:      waID,
		HelpCode:        helpCode,
		Generation:      userInfo.Generation,
		CreateGroupTime: lastSendTime,
		CreateTime:      now,
		UpdateTime:      now,
		Del:             biz.NotDeleted,
	}

	return sqlx.TxContext(ctx, o.data.db, func(ctx context.Context, tx sqlx.DB) error {
		// 新增用户
		_, err := model.InsertUserInfo(ctx, tx, &userInfoDO)
		if err != nil {
			o.l.WithContext(ctx).
				Errorf("insert user info failed, err=%v", err)
			return err
		}

		// 保存要发送的消息
		err = o.saveMsgSend(ctx, now, tx, waMsgSends)
		if err != nil {
			return err
		}

		_, err = model.InsertUserCreateGroup(ctx, tx, &userCreateGroupDO)
		if err != nil {
			o.l.WithContext(ctx).
				Errorf("insert user create group failed, err=%v", err)
			return err
		}

		if extend != nil {
			err = extend(ctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateUserInfoLanguageByWaID
func (o *Rally) updateUserInfoLanguageByWaID(ctx context.Context,
	waID string, rallyCode string, language string,
	waMsgSends []*dto.WaMsgSend, t time.Time,
	extend func(ctx context.Context, db sqlx.DB) error,
) error {
	return sqlx.TxContext(ctx, o.data.db, func(ctx context.Context, tx sqlx.DB) error {
		err := model.UpdateUserInfoLanguageByWaID(ctx, tx, waID, language)
		if err != nil {
			o.l.WithContext(ctx).
				Errorf("update user info language failed, err=%v, waID=%s, rallyCode=%s, language=%s",
					err, waID, rallyCode, language)
			return err
		}

		// 保存要发送的消息
		err = o.saveMsgSend(ctx, t, tx, waMsgSends)
		if err != nil {
			return err
		}

		if extend != nil {
			err = extend(ctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// FindUserCreateGroup
func (o *Rally) FindUserCreateGroup(ctx context.Context, waID string) (
	*biz.UserCreateGroup, error) {
	userCreateGroup, err := model.GetUserCreateGroup(ctx, o.data.db, waID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			o.l.WithContext(ctx).
				Errorf("get user create group failed, err=%v, waID=%s", err, waID)
		}
		return nil, err
	}

	return ConvertUserCreateGroup2Biz(userCreateGroup), nil
}
