package data

import (
	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"strconv"
	"strings"
	"time"
)

func ConvertUserCreateGroup2Biz(u *model.UserCreateGroup) *biz.UserCreateGroup {
	return &biz.UserCreateGroup{
		ID:              u.ID,
		CreateWAID:      u.CreateWaID,
		HelpCode:        u.HelpCode,
		CreateTime:      u.CreateTime,
		CreateGroupTime: u.CreateGroupTime,
		UpdateTime:      u.UpdateTime,
		Generation:      u.Generation,
		Del:             u.Del,
	}
}

func ConvertUserJoinGroup2Biz(u *model.UserJoinGroup) *biz.UserJoinGroup {
	return &biz.UserJoinGroup{
		ID:            u.ID,
		Del:           u.Del,
		CreateTime:    u.CreateTime,
		UpdateTime:    u.UpdateTime,
		JoinWaID:      u.JoinWaID,
		HelpCode:      u.HelpCode,
		JoinGroupTime: u.JoinGroupTime,
	}
}

func convertUserInfo2Biz(u *model.UserInfo) *biz.UserInfo {
	return &biz.UserInfo{
		ID:         u.ID,
		Del:        u.Del,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
		WaID:       u.WaID,
		HelpCode:   u.HelpCode,
		Channel:    u.Channel,
		Language:   u.Language,
		Generation: u.Generation,
		JoinCount:  u.JoinCount,
		CDKv0:      u.CDKv0,
		CDKv3:      u.CDKv3,
		CDKv6:      u.CDKv6,
		CDKv9:      u.CDKv9,
		CDKv12:     u.CDKv12,
		CDKv15:     u.CDKv15,
		Nickname:   u.Nickname,
	}
}

func convertUserRemind2Biz(u *model.UserRemind) *biz.UserRemind {
	return &biz.UserRemind{
		ID:           u.ID,
		WaID:         u.WaID,
		LastSendTime: u.LastSendTime,
		SendTimeV0:   u.SendTimeV0,
		StatusV0:     u.StatusV0,
		SendTimeV22:  u.SendTimeV22,
		StatusV22:    u.StatusV22,
		SendTimeV3:   u.SendTimeV3,
		StatusV3:     u.StatusV3,
		SendTimeV36:  u.SendTimeV36,
		StatusV36:    u.StatusV36,
		CreateTime:   u.CreateTime,
		UpdateTime:   u.UpdateTime,
		Del:          u.Del,
	}
}

func convertWaMsgSend2SimpleBiz(u *model.WaMsgSend) *biz.WaMsgSend {
	return &biz.WaMsgSend{
		ID:      u.ID,
		WaID:    u.WaID,
		MsgType: u.MsgType,
		State:   u.State,
	}
}

func ConvertFeishuReportParam2Entity(param *biz.FeishuReportParam) *model.FeishuReportEntity {
	return &model.FeishuReportEntity{
		Date:           param.Date,
		Time:           param.Time,
		CdkCount:       param.CdkCount,
		CoverCount:     param.CoverCount,
		FailedCount:    param.FailedCount,
		FirstCount:     param.FirstCount,
		InterceptCount: param.InterceptCount,
		FissionCount:   param.FissionCount,
		TimeoutCount:   param.TimeoutCount,
	}
}

func ConvertHelpCode2Biz(u *model.HelpCodeEntity) *biz.HelpCode {
	return &biz.HelpCode{
		Id:          u.Id,
		HelpCode:    u.HelpCode,
		ShortLinkV0: u.ShortLinkV0,
		ShortLinkV1: u.ShortLinkV1,
		ShortLinkV2: u.ShortLinkV2,
		ShortLinkV3: u.ShortLinkV3,
		ShortLinkV4: u.ShortLinkV4,
		ShortLinkV5: u.ShortLinkV5,
		CreateTime:  u.CreateTime,
		UpdateTime:  u.UpdateTime,
		Del:         u.Del,
	}
}

func convertOfficialMsgRecord2Biz(msg *model.OfficialMsgRecord) *biz.OfficialMsgRecord {
	return &biz.OfficialMsgRecord{
		ID:         msg.ID,
		WaID:       msg.WaID,
		RallyCode:  msg.RallyCode,
		State:      msg.State,
		Channel:    msg.Channel,
		Language:   msg.Language,
		Generation: msg.Generation,
		NickName:   msg.Nickname,
		SendTime:   msg.SendTime,
		CreateTime: msg.CreateTime,
		UpdateTime: msg.UpdateTime,
		Del:        msg.Del,
	}
}

func convertUnOfficialMsgRecord2Biz(msg *model.UnOfficialMsgRecord) *biz.UnOfficialMsgRecord {
	return &biz.UnOfficialMsgRecord{
		ID:         msg.ID,
		WaID:       msg.WaID,
		RallyCode:  msg.RallyCode,
		State:      msg.State,
		Channel:    msg.Channel,
		Language:   msg.Language,
		Generation: msg.Generation,
		NickName:   msg.Nickname,
		SendTime:   msg.SendTime,
		CreateTime: msg.CreateTime,
		UpdateTime: msg.UpdateTime,
		Del:        msg.Del,
	}
}

func convertReceiptMsgRecord2Biz(record *model.ReceiptMsgRecord) *biz.ReceiptMsgRecord {
	return &biz.ReceiptMsgRecord{
		ID:         record.ID,
		WaID:       record.WaId,
		MsgID:      record.MsgID,
		MsgState:   record.MsgState,
		State:      record.State,
		CreateTime: record.CreateTime,
		UpdateTime: record.UpdateTime,
		Del:        record.Del,
		CostInfo:   record.CostInfo,
	}
}

func ConvertEmailReport2Entity(u *biz.EmailReportDTO, utc int) *model.EmailReportEntity {
	generationCount := make([]string, 0)
	for _, count := range u.GenerationCount {
		generationCount = append(generationCount, strconv.Itoa(count))
	}
	totalJoinCount := make([]string, 0)
	for _, count := range u.TotalJoinCount {
		totalJoinCount = append(totalJoinCount, strconv.Itoa(count))
	}
	dailyJoinCount := make([]string, 0)
	for _, count := range u.DailyJoinCount {
		dailyJoinCount = append(dailyJoinCount, strconv.Itoa(count))
	}
	return &model.EmailReportEntity{
		Date:            u.Date,
		Utc:             utc,
		Channel:         u.Channel,
		Language:        u.Language,
		CountryCode:     u.CountryCode,
		SuccessCount:    u.SuccessCount,
		FailedCount:     u.FailedCount,
		TimeoutCount:    u.TimeoutCount,
		InterceptCount:  u.InterceptCount,
		GenerationCount: strings.Join(generationCount, ","),
		TotalJoinCount:  strings.Join(totalJoinCount, ","),
		DailyJoinCount:  strings.Join(dailyJoinCount, ","),
		CountV22:        u.CountV22,
		CountV3:         u.CountV3,
		CountV36:        u.CountV36,
	}
}

func ConvertEmailReport2DTO(u *model.EmailReportEntity) *biz.EmailReportDTO {
	generationCount := make([]int, 0)
	for _, v := range strings.Split(u.GenerationCount, ",") {
		if num, err := strconv.Atoi(v); err == nil {
			generationCount = append(generationCount, num)
		}
	}
	totalJoinCount := make([]int, 0)
	for _, v := range strings.Split(u.TotalJoinCount, ",") {
		if num, err := strconv.Atoi(v); err == nil {
			totalJoinCount = append(totalJoinCount, num)
		}
	}
	dailyJoinCount := make([]int, 0)
	for _, v := range strings.Split(u.DailyJoinCount, ",") {
		if num, err := strconv.Atoi(v); err == nil {
			dailyJoinCount = append(dailyJoinCount, num)
		}
	}
	return &biz.EmailReportDTO{
		Date:            u.Date,
		Channel:         u.Channel,
		Language:        u.Language,
		CountryCode:     u.CountryCode,
		SuccessCount:    u.SuccessCount,
		FailedCount:     u.FailedCount,
		TimeoutCount:    u.TimeoutCount,
		InterceptCount:  u.InterceptCount,
		GenerationCount: [7]int(generationCount),
		TotalJoinCount:  [16]int(totalJoinCount),
		DailyJoinCount:  [16]int(dailyJoinCount),
		CountV22:        u.CountV22,
		CountV3:         u.CountV3,
		CountV36:        u.CountV36,
	}
}

func ConvertUploadUserInfo2EntityList(list []*biz.UploadUserInfoDTO) []*model.UploadUserInfoEntity {
	var uuis []*model.UploadUserInfoEntity
	for _, u := range list {
		uuis = append(uuis, &model.UploadUserInfoEntity{
			PhoneNumber:  u.PhoneNumber,
			LastSendTime: u.LastSendTime,
		})
	}
	return uuis
}

func ConvertUploadUserInfo2BizList(list []*model.UploadUserInfoEntity) []*biz.UploadUserInfoDTO {
	var uuis []*biz.UploadUserInfoDTO
	for _, u := range list {
		uuis = append(uuis, &biz.UploadUserInfoDTO{
			Id:           u.Id,
			PhoneNumber:  u.PhoneNumber,
			LastSendTime: u.LastSendTime,
			State:        u.State,
		})
	}
	return uuis
}

func ConvertWaUserScore2BizList(list []*model.WaUserScore) []*biz.WaUserScoreDTO {
	var uuis []*biz.WaUserScoreDTO
	for _, u := range list {
		recurringProb, _ := strconv.ParseFloat(u.RecurringProb, 64)
		date, _ := time.Parse("2006-01-02", u.LastLoginTime)
		uuis = append(uuis, &biz.WaUserScoreDTO{
			Id:            u.Id,
			LastLoginTime: date,
			RecurringProb: recurringProb,
			SocialScore:   u.SocialScore,
			State:         u.State,
			WaId:          u.WaId,
		})
	}
	return uuis
}

func waMsgReceivedList2DTO(list []*model.WaMsgReceived) []*biz.WaMsgReceivedDTO {
	var uuis []*biz.WaMsgReceivedDTO
	for _, u := range list {
		uuis = append(uuis, &biz.WaMsgReceivedDTO{
			Id:              u.ID,
			MsgReceivedTime: u.MsgReceivedTime,
			WaId:            u.WaID,
		})
	}
	return uuis
}
