package data

import (
	"context"
	"errors"
	"time"

	"github.com/samber/lo"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/internal/pojo/dto"
	"fission-basic/kit/sqlx"
)

var _ biz.UnOfficialRallyRepo = (*UnOfficialRally)(nil)

type UnOfficialRally struct {
	*Rally
}

func NewUnOfficialRally(r *Rally) biz.UnOfficialRallyRepo {
	return &UnOfficialRally{
		Rally: r,
	}
}

type dbFunc func(context.Context, sqlx.DB) error

func (u *UnOfficialRally) buildRallyDBFunc(now time.Time, rallyInfo *biz.BaseInfo, helpCode string) []dbFunc {
	var funcs []dbFunc

	funcs = append(funcs, func(ctx context.Context, d sqlx.DB) error {
		_, err := model.InsertUserJoinGroup(ctx, d, &model.UserJoinGroup{
			JoinWaID:      rallyInfo.WaID,
			HelpCode:      helpCode, // 被助力人的助力码
			JoinGroupTime: rallyInfo.SendTime,
			CreateTime:    now,
			UpdateTime:    now,
			Del:           biz.NotDeleted,
		})
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("insert user join group failed, err=%v, helpCode=%s", err, rallyInfo.RallyCode)
			return err
		}
		// u.l.WithContext(ctx).Debugf("insert user join group success, rallyInfo=%+v, helpInfo=%+v", rallyInfo, helpCode)

		return nil
	})

	// FIXME: 不需要更新
	// funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
	// 	err := model.UpdateUserRemindLastSentTime(ctx, tx, rallyInfo.WaID, rallyInfo.SendTime)
	// 	if err != nil {
	// 		u.l.WithContext(ctx).
	// 			Errorf("update user remind last sent time failed, err=%v, waID=%s", err, rallyInfo.WaID)
	// 		return err
	// 	}

	// 	return nil
	// })

	if rallyInfo.NeedCreateGroup {
		funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
			_, err := model.InsertUserCreateGroup(ctx, tx, &model.UserCreateGroup{
				CreateWaID:      rallyInfo.WaID,
				HelpCode:        rallyInfo.RallyCode,
				Generation:      rallyInfo.Generation,
				CreateGroupTime: rallyInfo.SendTime,
				CreateTime:      now,
				UpdateTime:      now,
				Del:             biz.NotDeleted,
			})
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("insert user create group failed, err=%v, helpCode=%s", err, rallyInfo.RallyCode)
				return err
			}

			_, err = model.InsertUserInfo(ctx, tx, &model.UserInfo{
				WaID:       rallyInfo.WaID,
				Language:   rallyInfo.Language,
				Channel:    rallyInfo.Channel,
				Generation: rallyInfo.Generation,
				HelpCode:   rallyInfo.RallyCode,
				JoinCount:  0,
				CDKv0:      rallyInfo.CDK,
				Nickname:   rallyInfo.NickName,
				CreateTime: now,
				UpdateTime: now,
				Del:        biz.NotDeleted,
			})
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("insert user info failed, err=%v, waID=%s", err, rallyInfo.WaID)
				return err
			}

			return nil
		})
	}

	return funcs
}

// 被助力人相关
func (u *UnOfficialRally) buildHelpDBFunc(helpInfo *biz.BaseInfo, newHelpNum int) []dbFunc {
	var funcs []dbFunc

	u.l.WithContext(context.Background()).Infof("helpInfo=%+v, newHelpNum=%d", helpInfo, newHelpNum)

	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		if helpInfo.CDKType >= 0 && helpInfo.CDK != "" {
			err := model.UpdateUserInfoJoinCountAndCDK(ctx, tx,
				helpInfo.WaID, newHelpNum-1, newHelpNum, helpInfo.CDKType, helpInfo.CDK)
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("update user info join count and cdk failed, err=%v, waID=%s, newHelpNum=%d", err, helpInfo.WaID, newHelpNum)
				return err
			}
		} else {
			err := model.UpdateUserInfoJoinCount(ctx, tx,
				helpInfo.WaID, newHelpNum-1, newHelpNum)
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("update user info join count failed, err=%v, waID=%s, newHelpNum=%d", err, helpInfo.WaID, newHelpNum)
				return err
			}
		}

		return nil
	})

	return funcs
}

func (u *UnOfficialRally) buildMsgSendDBFunc(now time.Time, msgSends []*dto.WaMsgSend) []dbFunc {
	var funcs []dbFunc
	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		// 发送消息
		for i := range msgSends {
			id, err := model.InsertWaMsgSend(ctx, tx, &model.WaMsgSend{
				WaMsgID:       msgSends[i].WaMsgID,
				WaID:          msgSends[i].WaID,
				MsgType:       msgSends[i].MsgType,
				State:         msgSends[i].State,
				Content:       msgSends[i].Content,
				BuildMsgParam: msgSends[i].BuildMsgParam,
				SendRes:       msgSends[i].SendRes,
				CreateTime:    now,
				UpdateTime:    now,
				Del:           biz.NotDeleted,
			})
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("insert wa msg send failed, err=%v", err)
				return err
			}

			msgSends[i].ID = id
		}

		return nil
	})

	return funcs
}

func (u *UnOfficialRally) CreateJoinGroup2(ctx context.Context,
	rallyInfo, helpInfo *biz.BaseInfo, msgSends []*dto.WaMsgSend, newHelpNum int,
) error {
	now := time.Now()
	var funcs []dbFunc

	funcs = append(funcs, u.buildRallyDBFunc(now, rallyInfo, helpInfo.RallyCode)...)
	funcs = append(funcs, u.buildHelpDBFunc(helpInfo, newHelpNum)...)
	funcs = append(funcs, u.buildMsgSendDBFunc(now, msgSends)...)
	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		return u.completeRally(ctx, tx, rallyInfo.WaID, helpInfo.RallyCode)
	})

	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, tx sqlx.DB) error {
		for i := range funcs {
			err := funcs[i](ctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *UnOfficialRally) buildBufferRallyDBFunc(ctx context.Context, now time.Time,
	rallyInfo *biz.BaseInfo, helpCode string) ([]dbFunc, error) {
	var funcs []dbFunc

	userInfo, err := model.GetUserInfo(ctx, u.data.db, rallyInfo.WaID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).Errorf("GetUserInfo failed, err=%v, waID=%s", err, rallyInfo.WaID)
			return nil, err
		}
		// 不存在用户
	}

	// 新增助力人用户
	if userInfo == nil {
		funcs = append(funcs, func(ctx context.Context, db sqlx.DB) error {
			_, err = model.InsertUserInfo(ctx, db, &model.UserInfo{
				WaID:       rallyInfo.WaID,
				Language:   rallyInfo.Language,
				Channel:    rallyInfo.Channel,
				Generation: rallyInfo.Generation,
				// HelpCode:   rallyInfo.RallyCode,
				JoinCount: 0,
				// CDKv0:      rallyInfo.CDK,
				Nickname:   rallyInfo.NickName,
				CreateTime: now,
				UpdateTime: now,
				Del:        biz.NotDeleted,
			})
			if err != nil {
				u.l.WithContext(ctx).
					Errorf("insert user info failed, err=%v, waID=%s", err, rallyInfo.WaID)
				return err
			}
			return nil
		})
	}

	// 到这里肯定没有助力过
	funcs = append(funcs, func(ctx context.Context, db sqlx.DB) error {
		_, err := model.InsertUserJoinGroup(ctx, db, &model.UserJoinGroup{
			JoinWaID:      rallyInfo.WaID,
			HelpCode:      helpCode, // 被助力人的助力码
			JoinGroupTime: rallyInfo.SendTime,
			CreateTime:    now,
			UpdateTime:    now,
			Del:           biz.NotDeleted,
		})
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("insert user join group failed, err=%v, helpCode=%s", err, rallyInfo.RallyCode)
			return err
		}

		return nil
	})

	return funcs, nil
}

// CreateBufferJoinGroup implements biz.UnOfficialRallyRepo.
func (u *UnOfficialRally) CreateBufferJoinGroup(ctx context.Context,
	rallyInfo, helpInfo *biz.BaseInfo,
	msgSends []*dto.WaMsgSend, newHelpNum int,
) error {
	now := time.Now()
	var funcs []dbFunc

	bufferRallyDbFuncs, err := u.buildBufferRallyDBFunc(ctx, now, rallyInfo, helpInfo.RallyCode)
	if err != nil {
		return err
	}

	funcs = append(funcs, bufferRallyDbFuncs...)
	funcs = append(funcs, u.buildHelpDBFunc(helpInfo, newHelpNum)...)
	funcs = append(funcs, u.buildMsgSendDBFunc(now, msgSends)...)
	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		return u.completeRally(ctx, tx, rallyInfo.WaID, helpInfo.RallyCode)
	})

	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, tx sqlx.DB) error {
		for i := range funcs {
			err := funcs[i](ctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *UnOfficialRally) completeRally(ctx context.Context, db sqlx.DB, waID, rallyCode string) error {
	err := model.UpdateUnOfficialMsgRecordState(ctx, db, waID, rallyCode,
		biz.MsgStateDoing, biz.MsgStateComplete)
	if err != nil {
		u.l.WithContext(ctx).
			Errorf("update unofficial msg record state failed, err=%v", err)
		return err
	}

	return nil
}

// CompleteRally implements biz.UnOfficialRallyRepo.
func (u *UnOfficialRally) CompleteRally(ctx context.Context, waID, rallyCode string,
	msgSends []*dto.WaMsgSend, withMsgDB bool) error {
	now := time.Now()

	err := sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, db sqlx.DB) error {
		if withMsgDB {
			err := u.completeRally(ctx, u.data.db, waID, rallyCode)
			if err != nil {
				return err
			}
		}

		err := u.saveMsgSend(ctx, now, db, msgSends)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		u.l.WithContext(ctx).Errorf("completeRally failed, err=%v, waID=%s, rallyCode=%s", err, waID, rallyCode)
		return err
	}
	return nil
}

// FindMsg implements biz.UnOfficialRallyRepo.
func (u *UnOfficialRally) FindMsg(ctx context.Context, waID, rallyCode string) (*biz.UnOfficialMsgRecord, error) {
	msg, err := model.GetUnOfficialMsgRecord(ctx, u.data.db, waID, rallyCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).
				Errorf("get unofficial msg record failed, err=%v, waID=%s, rallyCode=%s", err, waID, rallyCode)
		}
		return nil, err
	}

	return convertUnOfficialMsgRecord2Biz(msg), nil
}

// FindUserCreateGroupByHelpCode implements biz.UnOfficialRallyRepo.
func (u *UnOfficialRally) FindUserCreateGroupByHelpCode(ctx context.Context, helpCode string) (*biz.UserCreateGroup, error) {
	userCreateGroup, err := model.GetUserCreateGroupByHelpCode(ctx, u.data.db, helpCode)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).
				Errorf("get user create group failed, err=%v, helpCode=%s", err, helpCode)
		}
		return nil, err
	}

	return ConvertUserCreateGroup2Biz(userCreateGroup), nil

}

// ListUserJoinGroups implements biz.UnOfficialRallyRepo.
// 查询所有助力信息
func (u *UnOfficialRally) ListUserJoinGroups(ctx context.Context, helpCode string) ([]*biz.UserJoinGroup, error) {
	userJoinGroups, err := model.SelectUserJoinGroupsByHelpCode(ctx, u.data.db, helpCode)
	if err != nil {
		u.l.WithContext(ctx).
			Errorf("get user join groups failed, err=%v, helpCode=%s", err, helpCode)
		return nil, err
	}

	return lo.Map(userJoinGroups, func(userJoinGroup *model.UserJoinGroup, _ int) *biz.UserJoinGroup {
		return ConvertUserJoinGroup2Biz(userJoinGroup)
	}), nil
}

func (u *UnOfficialRally) ListDoingMsgs(ctx context.Context,
	minID int, offset, length uint,
	maxTime time.Time,
) ([]*biz.UnOfficialMsgRecord, error) {
	msgs, err := model.SelectUnOfficialMsgRecords(ctx, u.data.db, minID, offset, length, biz.MsgStateDoing, maxTime)
	if err != nil {
		u.l.WithContext(ctx).Errorf("select unofficial msg records failed, err=%v, minID=%d, offset=%d, length=%d, maxTime=%v",
			err, minID, offset, length, maxTime)
		return nil, err
	}

	return lo.Map(
		msgs,
		func(msg *model.UnOfficialMsgRecord, _ int) *biz.UnOfficialMsgRecord {
			return convertUnOfficialMsgRecord2Biz(msg)
		},
	), nil
}

func (u *UnOfficialRally) FindUserJoinGroupByWaID(ctx context.Context, waID string) (*biz.UserJoinGroup, error) {
	userJoinGroup, err := model.FindUserJoinGroupByWaID(ctx, u.data.db, waID)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			u.l.WithContext(ctx).
				Errorf("get user join group failed, err=%v, waID=%s", err, waID)
		}

		return nil, err
	}

	return ConvertUserJoinGroup2Biz(userJoinGroup), nil
}

func (u *UnOfficialRally) CreateBufferMaxJoinGroup(ctx context.Context,
	rallyInfo *biz.BaseInfo,
	helpCode, helpWaID string,
	newHelpNum int,
	msgSends []*dto.WaMsgSend,
) error {
	now := time.Now()
	var funcs []dbFunc

	bufferRallyDbFuncs, err := u.buildBufferRallyDBFunc(ctx, now, rallyInfo, helpCode)
	if err != nil {
		return err
	}

	funcs = append(funcs, bufferRallyDbFuncs...)
	funcs = append(funcs, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateUserInfoJoinCount(ctx, db, helpWaID, newHelpNum-1, newHelpNum)
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("update user info join count failed, err=%v, waID=%s, newHelpNum=%d", err, helpWaID, newHelpNum)
			return err
		}
		return nil
	})
	funcs = append(funcs, u.buildMsgSendDBFunc(now, msgSends)...)
	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		return u.completeRally(ctx, tx, rallyInfo.WaID, helpCode)
	})

	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, tx sqlx.DB) error {
		for i := range funcs {
			err := funcs[i](ctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *UnOfficialRally) CreateStartedMaxJoinGroup(ctx context.Context,
	rallyInfo *biz.BaseInfo,
	helpCode, helpWaID string,
	newHelpNum int,
	msgSends []*dto.WaMsgSend,
) error {
	now := time.Now()
	var funcs []dbFunc

	funcs = append(funcs, u.buildRallyDBFunc(now, rallyInfo, helpCode)...)
	funcs = append(funcs, func(ctx context.Context, db sqlx.DB) error {
		err := model.UpdateUserInfoJoinCount(ctx, db,
			helpWaID, newHelpNum-1, newHelpNum)
		if err != nil {
			u.l.WithContext(ctx).
				Errorf("update user info join count failed, err=%v, waID=%s, newHelpNum=%d", err, helpWaID, newHelpNum)
			return err
		}
		return nil
	})
	funcs = append(funcs, u.buildMsgSendDBFunc(now, msgSends)...)
	funcs = append(funcs, func(ctx context.Context, tx sqlx.DB) error {
		return u.completeRally(ctx, tx, rallyInfo.WaID, helpCode)
	})

	return sqlx.TxContext(ctx, u.data.db, func(ctx context.Context, db sqlx.DB) error {
		for i := range funcs {
			err := funcs[i](ctx, db)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
