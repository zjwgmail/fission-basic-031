package model

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"fission-basic/kit/sqlx"
)

const tableUserInfo = `user_info`

// UserInfo 对应 user_info 表的结构体
type UserInfo struct {
	ID         int64     `db:"id"`          // 自增主键，唯一标识每条记录
	WaID       string    `db:"wa_id"`       // 用户的唯一标识
	HelpCode   string    `db:"help_code"`   // 用户的助力码
	Channel    string    `db:"channel"`     // 用户来源的渠道
	Language   string    `db:"language"`    // 用户使用的语言
	Generation int       `db:"generation"`  // 用户参与活动的代数
	JoinCount  int       `db:"join_count"`  // 用户的助力人数
	CDKv0      string    `db:"cdk_v0"`      // 类型为 v0 的 CDK 码
	CDKv3      string    `db:"cdk_v3"`      // 类型为 v3 的 CDK 码
	CDKv6      string    `db:"cdk_v6"`      // 类型为 v6 的 CDK 码
	CDKv9      string    `db:"cdk_v9"`      // 类型为 v9 的 CDK 码
	CDKv12     string    `db:"cdk_v12"`     // 类型为 v12 的 CDK 码
	CDKv15     string    `db:"cdk_v15"`     // 类型为 v15 的 CDK 码
	Nickname   string    `db:"nickname"`    // 用户的昵称
	CreateTime time.Time `db:"create_time"` // 记录创建时间，插入时自动记录当前时间
	UpdateTime time.Time `db:"update_time"` // 记录更新时间，更新记录时自动更新为当前时间
	Del        int8      `db:"del"`         // 标记记录是否删除，0 表示未删除，1 表示已删除
}

func InsertUserInfo(
	ctx context.Context, db sqlx.DB,
	userInfo *UserInfo) (int64, error) {
	return sqlx.InsertContext(ctx, db, tableUserInfo, userInfo)
}

func UpdateUserInfoLanguageByWaID(
	ctx context.Context, db sqlx.DB,
	waID, language string,
) error {
	where := map[string]interface{}{
		"wa_id": waID,
	}
	update := map[string]interface{}{
		"language": language,
	}

	return sqlx.UpdateContext(ctx, db, tableUserInfo, where, update)
}

func UpdateUserInfoJoinCount(ctx context.Context, db sqlx.DB,
	waID string, oldCount, newCount int) error {
	where := map[string]interface{}{
		"wa_id":      waID,
		"join_count": oldCount,
	}

	update := map[string]interface{}{
		"join_count": newCount,
	}

	return sqlx.UpdateContext(ctx, db, tableUserInfo, where, update)
}

func UpdateUserInfoJoinCountAndCDK(ctx context.Context, db sqlx.DB,
	waID string, oldJoinCount, newJoinCount int,
	cdkType int, cdk string,
) error {
	where := map[string]interface{}{
		"wa_id":      waID,
		"join_count": oldJoinCount,
	}

	update := map[string]interface{}{
		"join_count":                    newJoinCount,
		fmt.Sprintf("cdk_v%d", cdkType): cdk,
	}

	return sqlx.UpdateContext(ctx, db, tableUserInfo, where, update)
}

func GetUserInfo(ctx context.Context, db sqlx.DB, waID string) (*UserInfo, error) {
	where := map[string]interface{}{
		"wa_id": waID,
	}

	var userInfo UserInfo
	err := sqlx.GetContext(ctx, db, &userInfo, tableUserInfo, where)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

func GetUserInfoByHelpCode(ctx context.Context, db sqlx.DB, helpCode string) (*UserInfo, error) {
	where := map[string]interface{}{
		"help_code": helpCode,
	}

	var userInfo UserInfo
	err := sqlx.GetContext(ctx, db, &userInfo, tableUserInfo, where)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func ListUserInfoByHelpCodes(ctx context.Context, db sqlx.DB, helpCodes []string) ([]*UserInfo, error) {
	where := map[string]interface{}{
		"help_code in": helpCodes,
	}

	var userInfo []*UserInfo
	err := sqlx.SelectContext(ctx, db, &userInfo, tableUserInfo, where)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func UpdateUserInfoCdkByWaId(ctx context.Context, db sqlx.DB, waID string, cdk string, cdkType int) error {
	where := map[string]interface{}{
		"wa_id": waID,
	}
	update := map[string]interface{}{
		"cdk_v" + strconv.Itoa(cdkType): cdk,
	}
	return sqlx.UpdateContext(ctx, db, tableUserInfo, where, update)
}

func SelectUserInfosByWaIDs(ctx context.Context, db sqlx.DB, waIDs []string) ([]*UserInfo, error) {
	if len(waIDs) == 0 {
		return nil, nil
	}

	where := map[string]interface{}{
		"wa_id in": waIDs,
	}

	var userInfos []*UserInfo
	err := sqlx.SelectContext(ctx, db, &userInfos, tableUserInfo, where)
	if err != nil {
		return nil, err
	}

	return userInfos, nil
}

func ListUserInfoGtIdLtEndTime(ctx context.Context, db sqlx.DB, minId int64, endTime time.Time, limit uint) ([]*UserInfo, error) {
	where := map[string]interface{}{
		"id > ":         minId,
		"create_time <": endTime,
		"_orderby":      "id",
		"_limit":        []uint{limit},
	}

	var userInfos []*UserInfo
	err := sqlx.SelectContext(ctx, db, &userInfos, tableUserInfo, where)
	if err != nil {
		return nil, err
	}

	return userInfos, nil
}
