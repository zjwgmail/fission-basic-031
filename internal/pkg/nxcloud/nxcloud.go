package nxcloud

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/conf"
	iutil "fission-basic/internal/util"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/util"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type MsgType string

const (
	MsgTypeUnknown       MsgType = "unknown"         // 未知消息
	MsgTypeAttend        MsgType = "attend_msg"      // 初代参团消息
	MsgTypeRallyCode     MsgType = "rally_msg"       // 助力消息
	MsgTypeRenewMsgReply MsgType = "renew_msg_reply" // 续免费的回复消息
	MsgTypeCallback      MsgType = "callback"        // 回执消息
	MsgTypeNotWhite      MsgType = "notWhite"        // 非白消息
)

type RecallbackStatus int

const (
	CallbackStatusUnknown RecallbackStatus = 0
	CallbackStatusSuccess RecallbackStatus = 1 // FIXME: @qiankun确认
	CallbackStatusFailed  RecallbackStatus = 2 // FIXME: @qiankun确认
	CallbackStatusTimeout RecallbackStatus = 3 // FIXME: @qiankun确认
)

type ReceiptMsgQueueDTO struct {
	WaID    string
	MsgID   string
	MsgType MsgType
	// 回执消息有下面的数据
	MsgState int
	Costs    []*v1.Cost
}

type NXCloudInfo struct {
	MsgID   string
	MsgType MsgType
	Content string // 消息体
	// 助力、参团、续免费回复消息有下面的数据
	WaID               string
	UserNickName       string // 消息体
	RallyCode          string // 助力码
	Channel            string // 渠道
	Language           string // 语言
	Generation         string // 代次
	IdentificationCode string // 生成的唯一码
	SendTime           int64  // 发送时间
	// 回执消息有下面的数据
	Status string // 消息的状态，sent（已发送），delivered（已送达)，read（已读），failed（发送失败）,timeOut 超时
	Costs  []*v1.Cost
}

// IsOfficialRallyCode 是否是官方助力码
func (nx *NXCloudInfo) IsOfficialRallyCode() bool {
	return strings.HasSuffix(nx.RallyCode, "00000")
}

// ParseNXCloud 校验消息
func ParseNXCloud(ctx context.Context, log *log.Helper, confData *conf.Data, confBusiness *conf.Business, req *v1.UserAttendInfoRequest) (*NXCloudInfo, error) {
	cost := iutil.MethodCost(ctx, log, "message.ParseNXCloud")
	defer cost()

	methodName := util.GetCurrentFuncName()

	if req.BusinessPhone == "" {
		log.WithContext(ctx).Errorf("方法[%s]，Business_phone为空,req：%v", methodName, req)
		return nil, errors.New("Business_phone为空")
	}

	if confData.Nx.BusinessPhone != req.BusinessPhone {
		log.WithContext(ctx).Errorf("方法[%s]，Business_phone与配置不匹配,req：%v", methodName, req)
		return nil, errors.New("Business_phone与配置不匹配")
	}

	reqBytes, err := json.NewEncoder().Encode(req)
	if err != nil {
		log.WithContext(ctx).Errorf("方法[%s]，json解析报错,req：%v", methodName, req)
		return nil, errors.New("json解析报错")
	}

	if len(req.Contacts) > 0 && len(req.Messages) > 0 {
		// 用户消息解析
		if len(req.Messages) <= 0 {
			log.WithContext(ctx).Errorf("方法[%s]，Messages为空,req：%v", methodName, req)
			return nil, errors.New("messages为空")
		}

		if len(req.Contacts) <= 0 {
			log.WithContext(ctx).Errorf("方法[%s]，Contacts为空,req：%v", methodName, req)
			return nil, errors.New("Contacts为空")
		}
		waId := req.Contacts[0].WaId
		if waId == "" {
			log.WithContext(ctx).Errorf("方法[%s]，wa_id为空,req：%v", methodName, req)
			return nil, errors.New("wa_id为空")
		}

		profile := req.Contacts[0].Profile
		if profile == nil {
			log.WithContext(ctx).Errorf("方法[%s]，profile为空,req：%v", methodName, req)
			return nil, errors.New("profile为空")
		}
		userNickName := profile.Name
		if userNickName == "" {
			log.WithContext(ctx).Errorf("方法[%s]，profile.Name为空,req：%v", methodName, req)
			return nil, errors.New("profile.Name为空")
		}

		var err error
		timestampDB := time.Now().Unix()
		//if req.Messages[0].Timestamp == "" {
		//	log.WithContext(ctx).Errorf("方法[%s]，message timestamp is null", methodName)
		//	return nil, errors.New(fmt.Sprintf("message timestamp is null"))
		//} else {
		//	timestampStr := req.Messages[0].Timestamp
		//	// 将字符串转换为 int64
		//	timestampDB, err = strconv.ParseInt(timestampStr, 10, 64)
		//	if err != nil {
		//		log.WithContext(ctx).Errorf("方法[%s]，Error parsing timestamp:,err：%v", methodName, err)
		//		return nil, errors.New(fmt.Sprintf("Error parsing timestamp:%v", err))
		//	}
		//}

		var textBody string
		if req.Messages[0].Type == "text" {
			if req.Messages[0].Text == nil {
				log.WithContext(ctx).Errorf("方法[%s]，messages中Text为空,req：%v", methodName, req)
				return nil, errors.New("messages中Text为空")
			} else {
				textBody = req.Messages[0].Text.Body
			}
		} else if req.Messages[0].Type == "button" {
			if req.Messages[0].Button == nil {
				log.WithContext(ctx).Errorf("方法[%s]，messages中Button为空,req：%v", methodName, req)
				return nil, errors.New("messages中Button为空")
			} else {
				textBody = req.Messages[0].Button.Text
			}
		} else if req.Messages[0].Type == "interactive" {
			if req.Messages[0].Interactive == nil {
				log.WithContext(ctx).Errorf("方法[%s]，messages中interactive为空,req：%v", methodName, req)
				return nil, errors.New("messages中interactive为空")
			} else {
				if req.Messages[0].Interactive.Type == "button_reply" {
					if req.Messages[0].Interactive.ButtonReply == nil {
						log.WithContext(ctx).Errorf("方法[%s]，messages中interactive中的button_reply为空,req：%v", methodName, req)
						return nil, errors.New("messages中interactive中的button_reply为空")
					} else {
						textBody = req.Messages[0].Interactive.ButtonReply.Title
					}
				}
			}
		}

		sendRallyCode := ""
		containsPrefix := false
		msgType := MsgTypeUnknown
		for _, prefix := range confData.MethodInsertMsgInfo[confBusiness.Activity.Scheme].UserAttendPrefixList {
			if strings.Contains(textBody, prefix) {
				containsPrefix = true
				msgType = MsgTypeAttend
				codeStr := strings.Split(textBody, prefix)
				sendRallyCode = strings.TrimSpace(codeStr[len(codeStr)-1])
				break
			}
		}
		if !containsPrefix {
			for _, prefix := range confData.MethodInsertMsgInfo[confBusiness.Activity.Scheme].UserAttendOfHelpPrefixList {
				if strings.Contains(textBody, prefix) {
					containsPrefix = true
					msgType = MsgTypeRallyCode
					codeStr := strings.Split(textBody, prefix)
					sendRallyCode = strings.TrimSpace(codeStr[len(codeStr)-1])
					break
				}
			}
		}
		if !containsPrefix {
			// 判断是否是点击续费消息的回复
			for _, prefix := range confData.MethodInsertMsgInfo[confBusiness.Activity.Scheme].RenewFreePrefixList {
				if strings.Contains(textBody, prefix) {
					containsPrefix = true
					msgType = MsgTypeRenewMsgReply
					nXCloudInfo := &NXCloudInfo{
						WaID:         waId,
						Content:      string(reqBytes),
						SendTime:     timestampDB,
						RallyCode:    sendRallyCode,
						MsgType:      msgType,
						UserNickName: userNickName,
						MsgID:        req.Messages[0].Id,
					}
					return nXCloudInfo, nil
				}
			}
		}

		if !containsPrefix {
			//返回未知消息
			nXCloudInfo := &NXCloudInfo{
				WaID:         waId,
				Content:      string(reqBytes),
				SendTime:     timestampDB,
				RallyCode:    sendRallyCode,
				MsgType:      msgType,
				UserNickName: userNickName,
				MsgID:        req.Messages[0].Id,
			}
			return nXCloudInfo, nil
		}

		sendChannel := sendRallyCode[0:1]
		sendLanguage := sendRallyCode[1:3]
		sendGeneration := sendRallyCode[3:5]
		sendIdentificationCode := sendRallyCode[5:]

		if MsgTypeRallyCode != msgType && (sendGeneration != constants.Generation01 || sendIdentificationCode != constants.FirstIdentificationCode) {
			log.WithContext(ctx).Errorf("messages's Text's body sendIdentificationCode is not compliance,req：%v", req)
			return nil, errors.New("messages's Text's body sendIdentificationCode is not compliance")
		}

		generation := sendGeneration
		if sendIdentificationCode != constants.FirstIdentificationCode {
			// 非初代，迭代
			generation, err = util.GetNewGeneration(generation)
			if err != nil {
				log.WithContext(ctx).Errorf("Upgrade propagation algebra failed,req：%v,err:%v", req, err)
				return nil, errors.New("Upgrade propagation algebra failed")
			}
		}

		// 校验waId是否正确
		if !util.StartsWithPrefix(waId, confBusiness.Activity.WaIdPrefixList) {
			log.WithContext(ctx).Info("wa_id is not in WaIdPrefixList,req：%v", req)
			msgType = MsgTypeNotWhite
		}

		nXCloudInfo := &NXCloudInfo{
			WaID:               waId,
			Content:            string(reqBytes),
			SendTime:           timestampDB,
			RallyCode:          sendRallyCode,
			MsgType:            msgType,
			UserNickName:       userNickName,
			MsgID:              req.Messages[0].Id,
			Channel:            sendChannel,
			Language:           sendLanguage,
			Generation:         generation,
			IdentificationCode: sendIdentificationCode,
		}
		return nXCloudInfo, nil
	} else {
		if len(req.Statuses) <= 0 {
			log.WithContext(ctx).Errorf("方法[%s]，Statuses为空,req：%v", methodName, req)
			return nil, errors.New(fmt.Sprintf("Statuses为空,req：%v", req))
		}
		status := req.Statuses[0].Status
		if constants.NxStatusFailed == req.Statuses[0].Status {
			if len(req.Statuses[0].Errors) > 0 {
				if 10002 == req.Statuses[0].Errors[0].Code {
					status = constants.NxStatusTimeout
				}
			}
		}
		nXCloudInfo := &NXCloudInfo{
			WaID:    req.Statuses[0].RecipientId,
			MsgType: MsgTypeCallback,
			MsgID:   req.Statuses[0].Id,
			Costs:   req.Statuses[0].Costs,
			Status:  status,
			Content: string(reqBytes),
		}
		return nXCloudInfo, nil
	}

	//log.WithContext(ctx).Errorf("方法[%s]，消息不能解析,req：%v", methodName, req)
	//nXCloudInfo := &NXCloudInfo{
	//	MsgType: MsgTypeUnknown,
	//	Content: string(reqBytes),
	//}
	//return nXCloudInfo, nil
}
