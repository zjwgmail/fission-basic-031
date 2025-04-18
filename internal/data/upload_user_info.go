package data

import (
	"context"
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"github.com/go-kratos/kratos/v2/log"
)

var _ biz.UploadUserInfoRepo = (*UploadUserInfo)(nil)

type UploadUserInfo struct {
	data *Data
	l    *log.Helper
}

func NewUploadUserInfo(d *Data, logger log.Logger) biz.UploadUserInfoRepo {
	return &UploadUserInfo{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (u *UploadUserInfo) InsertBatch(ctx context.Context, list []*biz.UploadUserInfoDTO) error {
	return model.UploadUserInfoInsertBatch(ctx, u.data.db, ConvertUploadUserInfo2EntityList(list))
}

func (u *UploadUserInfo) UpdateState(ctx context.Context, phoneNumber string, state int) error {
	return model.UploadUserInfoUpdateState(ctx, u.data.db, phoneNumber, state)
}
func (u *UploadUserInfo) ListInNumber(ctx context.Context, phoneNumberList []string) ([]*biz.UploadUserInfoDTO, error) {
	list, err := model.UploadUserInfoListInNumber(ctx, u.data.db, phoneNumberList)
	if err != nil {
		return nil, err
	}
	return ConvertUploadUserInfo2BizList(list), nil
}
func (u *UploadUserInfo) ListGtIdWithState(ctx context.Context, id int, state int, limit uint) ([]*biz.UploadUserInfoDTO, error) {
	list, err := model.UploadUserInfoListGtIdWithState(ctx, u.data.db, id, state, limit)
	if err != nil {
		return nil, err
	}
	return ConvertUploadUserInfo2BizList(list), nil
}
