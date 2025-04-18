package biz

import (
	"bytes"
	"context"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/pojo/dto/nx"
	"fission-basic/internal/pojo/dto/response"
	"fission-basic/internal/rest"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/internal/util/encoder/rsa"
	"fission-basic/util"
	"fission-basic/util/strUtil"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type WaMsgService struct {
	l              *log.Helper
	configInfo     *conf.Bootstrap
	waMsgRetryRepo WaMsgRetryRepo
	waMsgSendRepo  WaMsgSendRepo
	redisService   *redis.RedisService
	shortLink      *rest.ShortLink
	imageGenerate  *ImageGenerate
	publicKey      string
}

func NewWaMsgService(d *conf.Data, l log.Logger, configInfo *conf.Bootstrap, waMsgRetryRepo WaMsgRetryRepo, waMsgSendRepo WaMsgSendRepo, redisService *redis.RedisService, imageGenerate *ImageGenerate) *WaMsgService {
	return &WaMsgService{
		l:              log.NewHelper(l),
		configInfo:     configInfo,
		waMsgRetryRepo: waMsgRetryRepo,
		waMsgSendRepo:  waMsgSendRepo,
		redisService:   redisService,
		shortLink:      &rest.ShortLink{},
		imageGenerate:  imageGenerate,
		publicKey:      d.Rsa.PublicKey,
	}
}

// ActivityTask2NX 参与活动消息
func (w *WaMsgService) ActivityTask2NX(ctx context.Context, buildMsgInfo *dto.BuildMsgInfo, param *dto.HelpParam) ([]*dto.SendNxListParamsDto, error) {
	buildMsgInfo.MsgType = constants.ActivityTaskMsg

	sendMsgInfo, err := w.getMsgInfo(ctx, buildMsgInfo)
	if err != nil {
		return nil, err
	}
	paramBytes, err := json.NewEncoder().Encode(param)
	if err != nil {
		w.l.Error(fmt.Sprintf("json encode param error,buildMsgInfo:%v,param:%v;err:%v", buildMsgInfo, param, err))
		return nil, err
	}
	paramStr := string(paramBytes)
	// 要将传给前端的信息拼接好发给前端，要加密成param
	paramStrEncrypt, err := rsa.Encrypt(paramStr, w.configInfo.Data.Rsa.PublicKey)
	if err != nil {
		w.l.Error(fmt.Sprintf("Encrypt params error,buildMsgInfo:%v,param:%v;err:%v", buildMsgInfo, param, err))
		return nil, err
	}
	paramStrEscape := util.QueryEscape(paramStrEncrypt)

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, paramStrEscape, buildMsgInfo.Language, buildMsgInfo.Channel)
	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{buildMsgInfo}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("send message error,buildMsgInfo:%v,param:%v;err:%v", buildMsgInfo, param, err))
		return nil, err
	}
	return sendJson, nil
}

// CannotAttendActivity2NX 不能参与活动消息，非白
func (w *WaMsgService) CannotAttendActivity2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.CannotAttendActivityMsg
	// 非白计数
	// 获取月日，格式：x月x日
	time := time.Now()
	monthDay := fmt.Sprintf("%d月%d日", time.Month(), time.Day())

	phoneSetKey := constants.GetNotWhiteSetKey(w.configInfo.Business.Activity.Id)
	addCount, err := w.redisService.SAddKey(methodName, phoneSetKey, msgInfoEntity.WaId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, phoneSetKey, err))
	}
	if addCount > 0 {
		// 给非白拦截redis增加次数
		notWhite := constants.GetNotWhiteCountKey(w.configInfo.Business.Activity.Id, monthDay, msgInfoEntity.Channel, msgInfoEntity.Language)
		_, err := w.redisService.AddIncrKey(methodName, notWhite)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, notWhite, err))
		}
	}

	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// RepeatHelpMsg2NX 重复助力消息
func (w *WaMsgService) RepeatHelpMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, nickname string) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.RepeatHelpMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}
	originText := sendMsgInfo.Interactive.BodyText
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(originText, nickname)

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// CanNotHelpOneselfMsg2NX 不能助力自己消息
func (w *WaMsgService) CanNotHelpOneselfMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.CanNotHelpOneselfMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// StartGroupMsg2NX 开团消息
func (w *WaMsgService) StartGroupMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, shortLink string, helpNameList []*dto.HelpNickNameInfo, isHelp bool) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	if isHelp {
		msgInfoEntity.MsgType = constants.HelpStartGroupMsg
	} else {
		msgInfoEntity.MsgType = constants.StartGroupMsg
	}

	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称

	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.Id > 0 && helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	if len(nicknameList) > 0 {
		synthesisParam := v1.SynthesisParamRequest{
			BizType:         int64(constants.BizTypeInteractive),
			LangNum:         msgInfoEntity.Language,
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
		}
		imageUrl, err := w.imageGenerate.GetInteractiveImageUrl(ctx, &synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	}

	helpText, err := w.GetHelpText(ctx, msgInfoEntity.Language)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText

	//shortLink := user.RallyCodeShortLink
	//if "" == user.RallyCodeShortLink {
	//	// url中的链接要调用接口活动，并且要用到rallyCode
	//	sendMsgInfo.Interactive.Action.RallyCodeShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.RallyCodeShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
	//	shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.RallyCodeShortLink, msgInfoEntity.WaId)
	//	if err != nil {
	//		return nil, err
	//	}
	//	// 更新user表
	//	session, isExist, err := txUtil.GetTransaction(ctx)
	//	if nil != err {
	//		w.l.Error(fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
	//		return nil, errors.New("database is error")
	//	}
	//	if !isExist {
	//		defer func() {
	//			session.Rollback()
	//			session.Close()
	//		}()
	//	}
	//
	//	userAttendInfoEntity := entity.UserAttendInfoEntityV2{
	//		Id:        user.Id,
	//		RallyCodeShortLink: shortLink,
	//	}
	//	_, err = dao.GetUserAttendInfoMapperV2().UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
	//	if err != nil {
	//		w.l.Error(fmt.Sprintf("方法[%s]，更新短链接失败,WaId:%v", methodName, msgInfoEntity.WaId))
	//		return nil, err
	//	}
	//
	//	if !isExist {
	//		session.Commit()
	//	}
	//}

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = w.configInfo.Business.Activity.WaRedirectListPrefix + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		sendMsgInfo.Params = &conf.Params{
			NicknameList: nicknameList,
			Language:     msgInfoEntity.Language,
		}
		sendMsgInfo.Template.Components[1].Parameters[0].Text = strUtil.ReplacePlaceholders(sendMsgInfo.Template.Components[1].Parameters[0].Text, shortLink)
		sendMsgInfo.Template.Components[1].Parameters[0].Text = util.QueryEscape(sendMsgInfo.Template.Components[1].Parameters[0].Text)
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}

	w.l.Info(fmt.Sprintf("begin build freeCdk message,waId:%v", msgInfoEntity.WaId))

	return sendJson, nil
}

// FounderCanNotStartGroupMsg 缓冲期主态不能开团消息、或者结束期主态不能开团
func (w *WaMsgService) FounderCanNotStartGroupMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// CanNotStartGroupMsg 缓冲期客态不能开团消息
func (w *WaMsgService) CanNotStartGroupMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.CanNotStartGroupMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}
	if sendMsgInfo.Template != nil {
		//  模板消息未定
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// HelpTaskSingleSuccessMsg2NX 单人助力成功信息
func (w *WaMsgService) HelpTaskSingleSuccessMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, shortLink string, helpNameList []*dto.HelpNickNameInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()

	w.l.WithContext(ctx).Infof(fmt.Sprintf("begin build HelpTaskSingleSuccessMsg2NX message,msgInfoEntity:%v shortLink%v", msgInfoEntity, shortLink))
	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}

	stageInfo, err := w.GetStageInfoByAttendStatus(ctx, methodName, msgInfoEntity, helpNameList)
	if err != nil {
		return nil, err
	}
	msgInfoEntity.MsgType = constants.HelpTaskSingleSuccessMsgPrefix + strconv.Itoa(stageInfo.StageNum)

	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	if len(nicknameList) > 0 {
		// 图片
		synthesisParam := v1.SynthesisParamRequest{
			BizType:         int64(constants.BizTypeInteractive),
			LangNum:         msgInfoEntity.Language,
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
		}
		imageUrl, err := w.imageGenerate.GetInteractiveImageUrl(ctx, &synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	} else {
		w.l.Error(fmt.Sprintf("方法[%s]，助力人昵称不存在,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, errors.New("助力人昵称不存在")
	}

	rewardStageDto, err := w.GetStageInfoByAttendStatus(ctx, methodName, msgInfoEntity, helpNameList)
	if err != nil {
		return nil, err
	}
	// 消息修改 恭喜，你的好友{{1}}接受了你的邀请。再邀请{{2}}位好友助力，就能获得{{3}}奖励
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.NextStageMax-len(helpNameList)), rewardStageDto.NextStageName)

	text, err := w.GetHelpText(ctx, msgInfoEntity.Language)
	if err != nil {
		return nil, err
	}

	textWithShortLink := strUtil.ReplacePlaceholders(text, shortLink)
	sendMsgInfo.Interactive.Action.Url = w.configInfo.Business.Activity.WaRedirectListPrefix + util.QueryEscape(textWithShortLink)

	if sendMsgInfo.Template != nil {
		sendMsgInfo.Params = &conf.Params{
			NicknameList: nicknameList,
			Language:     msgInfoEntity.Language,
		}
		sendMsgInfo.Template.Components[1].Parameters[0].Text = helpNameList[len(helpNameList)-1].UserNickname
		sendMsgInfo.Template.Components[1].Parameters[1].Text = strconv.Itoa(rewardStageDto.NextStageMax - len(helpNameList))
		sendMsgInfo.Template.Components[1].Parameters[2].Text = rewardStageDto.NextStageName
		sendMsgInfo.Template.Components[2].Parameters[0].Text = strUtil.ReplacePlaceholders(sendMsgInfo.Template.Components[1].Parameters[0].Text, shortLink)
		sendMsgInfo.Template.Components[2].Parameters[0].Text = util.QueryEscape(sendMsgInfo.Template.Components[1].Parameters[0].Text)
	}

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	//if redPacketCode != "" {
	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
	//	if err != nil {
	//		return "", err
	//	}
	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
	//}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// HelpOverMsg2NX 助力完成信息即奖励消息  sendNxMsgType 传 constants.BizTypeInteractive
func (w *WaMsgService) HelpOverMsg2NX(ctx context.Context,
	msgInfoEntity *dto.BuildMsgInfo, cdkShortLink string, helpNameList []*dto.HelpNickNameInfo,
	sendNxMsgType int, inviteShortLink string) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()

	rewardStageDto, err := w.GetStageInfoByAttendStatus(ctx, methodName, msgInfoEntity, helpNameList)
	if err != nil {
		return nil, err
	}
	msgInfoEntity.MsgType = constants.HelpOverMsgPrefix + strconv.Itoa(rewardStageDto.StageNum)

	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	if len(nicknameList) > 0 {
		//  图片
		synthesisParam := v1.SynthesisParamRequest{
			BizType:         int64(constants.BizTypeInteractive),
			LangNum:         msgInfoEntity.Language,
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
		}
		imageUrl, err := w.imageGenerate.GetInteractiveImageUrl(ctx, &synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	} else {
		w.l.Error(fmt.Sprintf("方法[%s]，助力人昵称不存在,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, errors.New("助力人昵称不存在")
	}
	// 奖励短链接
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.CurrentStageMax), cdkShortLink, strconv.Itoa(rewardStageDto.NextStageMax-len(helpNameList)))

	helpText, err := w.GetHelpText(ctx, msgInfoEntity.Language)
	if err != nil {
		return nil, err
	}

	helpTextWithUrl := strUtil.ReplacePlaceholders(helpText, inviteShortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(w.configInfo.Business.Activity.WaRedirectListPrefix) + util.PathEscape(helpTextWithUrl)

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendNxListParamsDto []*dto.SendNxListParamsDto
	if constants.BizTypeInteractive == sendNxMsgType {
		sendNxListParamsDto, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendNxListParamsDto, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}

	return sendNxListParamsDto, nil
}

//
//// HelpFiveOverMsg2NX 5人助力完成信息
//func (w *WaMsgService) HelpFiveOverMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpNickNameInfo, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
//	methodName := util.GetCurrentFuncName()
//
//	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity, constant.HelpFiveOverMsg, user.Language)
//	if err != nil {
//		return nil, err
//	}
//
//	queryUser := entity.UserAttendInfoEntityV2{
//		WaId:         user.WaId,
//		Language:     user.Language,
//		AttendStatus: user.AttendStatus,
//		IsThreeStage: constant.IsStage,
//		IsFiveStage:  constant.IsStage,
//	}
//	rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, queryUser, helpNameList)
//	if err != nil {
//		return nil, err
//	}
//
//	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
//	var nicknameList []string
//	for _, helpNameEntity := range helpNameList {
//		if helpNameEntity.UserNickname != "" {
//			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
//		}
//	}
//	if len(nicknameList) > 0 {
//		synthesisParam := &request.SynthesisParam{
//			NicknameList:    nicknameList,
//			CurrentProgress: int64(len(helpNameList)),
//			LangNum:         user.Language,
//			BizType:         constant.BizTypeInteractive,
//		}
//		imageUrl, err := GetImageService().GetInteractiveImageUrl(ctx, synthesisParam, msgInfoEntity.WaId)
//		if err != nil {
//			return nil, err
//		}
//		sendMsgInfo.Interactive.ImageLink = imageUrl
//	} else {
//		w.l.Error(fmt.Sprintf("方法[%s]，助力人昵称不存在,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//		return nil, errors.New("助力人昵称不存在")
//	}
//
//	// 要将传给前端的信息拼接好发给前端，要加密成param
//	cdkEncrypt, err := rsa.Encrypt(cdk, conf.ApplicationConfig.Rsa.PublicKey)
//	if err != nil {
//		w.l.Error(fmt.Sprintf("方法[%s]，加密cdk报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//		return nil, err
//	}
//	// 奖励短链接
//	awardLink := rewardStageDto.CurrentAwardLink
//	awardLink = strUtil.ReplacePlaceholders(awardLink, user.RallyCode, cdkEncrypt, user.Language, user.Channel)
//	awardShortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, awardLink, msgInfoEntity.WaId)
//	if err != nil {
//		return nil, err
//	}
//	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.CurrentStageMax), awardShortLink, strconv.Itoa(rewardStageDto.NextStageMax-len(helpNameList)))
//
//	helpText, err := GetHelpTextWeight(ctx)
//	if err != nil {
//		return nil, err
//	}
//	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[conf.ApplicationConfig.Activity.Scheme][user.Language]
//	// url中的链接要调用接口活动，并且要用到rallyCode
//	shortLink := user.RallyCodeShortLink
//	if "" == user.RallyCodeShortLink {
//		// url中的链接要调用接口活动，并且要用到rallyCode
//		sendMsgInfo.Interactive.Action.RallyCodeShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.RallyCodeShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
//		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.RallyCodeShortLink, msgInfoEntity.WaId)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
//	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(conf.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.PathEscape(shortLink)
//
//	if sendMsgInfo.Template != nil {
//		//  模板消息未定
//	}
//
//	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
//	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}
//	//if redPacketCode != "" {
//	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
//	//	if err != nil {
//	//		return "", err
//	//	}
//	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
//	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
//	//}
//
//	var sendJson []*dto.SendNxListParamsDto
//	if constant.BizTypeInteractive == sendNxMsgType {
//		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
//		if err != nil {
//			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//			return nil, err
//		}
//	} else {
//		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
//		if err != nil {
//			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//			return nil, err
//		}
//	}
//	return sendJson, nil
//}
//
//// HelpEightOverMsg2NX 8人助力完成信息
//func (w *WaMsgService) HelpEightOverMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpNickNameInfo, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
//	methodName := util.GetCurrentFuncName()
//
//	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity, constant.HelpEightOverMsg, user.Language)
//	if err != nil {
//		return nil, err
//	}
//
//	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
//	var nicknameList []string
//	for _, helpNameEntity := range helpNameList {
//		if helpNameEntity.UserNickname != "" {
//			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
//		}
//	}
//	if len(nicknameList) > 0 {
//		synthesisParam := &request.SynthesisParam{
//			NicknameList:    nicknameList,
//			CurrentProgress: int64(len(helpNameList)),
//			LangNum:         user.Language,
//			BizType:         constant.BizTypeInteractive,
//		}
//		imageUrl, err := GetImageService().GetInteractiveImageUrl(ctx, synthesisParam, msgInfoEntity.WaId)
//		if err != nil {
//			return nil, err
//		}
//		sendMsgInfo.Interactive.ImageLink = imageUrl
//	} else {
//		w.l.Error(fmt.Sprintf("方法[%s]，助力人昵称不存在,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//		return nil, errors.New("助力人昵称不存在")
//	}
//
//	// url中的链接要调用接口活动，可能会用到rallyCode，cdk； 还需要url加密吗
//	// rallyCodeEscape := util.QueryEscape(rallyCode)
//	// sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, rallyCodeEscape)
//	queryUser := entity.UserAttendInfoEntityV2{
//		WaId:         user.WaId,
//		Language:     user.Language,
//		AttendStatus: constant.AttendStatusEightOver,
//		IsThreeStage: constant.IsStage,
//		IsFiveStage:  constant.IsStage,
//	}
//	rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, queryUser, helpNameList)
//	if err != nil {
//		return nil, err
//	}
//
//	// 要将传给前端的信息拼接好发给前端，要加密成param
//	cdkEncrypt, err := rsa.Encrypt(cdk, conf.ApplicationConfig.Rsa.PublicKey)
//	if err != nil {
//		w.l.Error(fmt.Sprintf("方法[%s]，加密cdk报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//		return nil, err
//	}
//	// 奖励短链接
//	awardLink := rewardStageDto.CurrentAwardLink
//	awardLink = strUtil.ReplacePlaceholders(awardLink, user.RallyCode, cdkEncrypt, user.Language, user.Channel)
//	awardShortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, awardLink, msgInfoEntity.WaId)
//	if err != nil {
//		return nil, err
//	}
//	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.CurrentStageMax), awardShortLink)
//
//	// 八人的行动点，改为邀请好友
//	helpText, err := GetHelpTextWeight(ctx)
//	if err != nil {
//		return nil, err
//	}
//	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[conf.ApplicationConfig.Activity.Scheme][user.Language]
//	// url中的链接要调用接口活动，并且要用到rallyCode
//	shortLink := user.RallyCodeShortLink
//	if "" == user.RallyCodeShortLink {
//		// url中的链接要调用接口活动，并且要用到rallyCode
//		sendMsgInfo.Interactive.Action.RallyCodeShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.RallyCodeShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
//		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.RallyCodeShortLink, msgInfoEntity.WaId)
//		if err != nil {
//			return nil, err
//		}
//	}
//	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
//	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(conf.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.PathEscape(shortLink)
//
//	if sendMsgInfo.Template != nil {
//		//  模板消息未定
//	}
//
//	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
//	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}
//	//if redPacketCode != "" {
//	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
//	//	if err != nil {
//	//		return "", err
//	//	}
//	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
//	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
//	//}
//
//	var sendJson []*dto.SendNxListParamsDto
//	if constant.BizTypeInteractive == sendNxMsgType {
//		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
//		if err != nil {
//			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//			return nil, err
//		}
//	} else {
//		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
//		if err != nil {
//			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, user.Language, err))
//			return nil, err
//		}
//	}
//	return sendJson, nil
//}

// FreeCdkMsg2NX 免费cdk红包
func (w *WaMsgService) FreeCdkMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, cdk string, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.FreeCdkMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	encryptedCode, err := rsa.Encrypt(msgInfoEntity.RallyCode, w.publicKey)
	if err != nil {
		w.l.WithContext(ctx).Errorf(fmt.Sprintf("method[%s]，Encrypt vo link error,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	urlCode := url.QueryEscape(encryptedCode)
	v0Link := strings.ReplaceAll("https://sg-play.mobilelegends.com/events/mlbb25031/promotion/?code={code}&gpt=1&lp=1&mode=1", "{code}", urlCode)
	sendMsgInfo.Interactive.Action.Url = v0Link

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendNxListParamsDto []*dto.SendNxListParamsDto
	if constants.BizTypeInteractive == sendNxMsgType {
		sendNxListParamsDto, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendNxListParamsDto, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}

	return sendNxListParamsDto, nil
}

// RenewFreeMsg 续免费信息
func (w *WaMsgService) RenewFreeMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.RenewFreeMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constants.BizTypeInteractive == sendNxMsgType {
		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// PayRenewFreeMsg 付费-续免费信息
func (w *WaMsgService) PayRenewFreeMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.PayRenewFreeMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constants.BizTypeInteractive == sendNxMsgType {
		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// PromoteClusteringMsg2NX 催促成团
func (w *WaMsgService) PromoteClusteringMsg2NX(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, shortLink string, sendNxMsgType int, helpNameList []dto.HelpNickNameInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.PromoteClusteringMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.Id > 0 && helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	// 没有助力人就保持不变，用活动图
	if len(nicknameList) > 0 {
		// 图片
		synthesisParam := v1.SynthesisParamRequest{
			BizType:         int64(constants.BizTypeInteractive),
			LangNum:         msgInfoEntity.Language,
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
		}
		imageUrl, err := w.imageGenerate.GetInteractiveImageUrl(ctx, &synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	}

	helpText, err := w.GetHelpText(ctx, msgInfoEntity.Language)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText
	// url中的链接要调用接口活动，并且要用到rallyCode
	//shortLink := user.RallyCodeShortLink
	//if "" == user.RallyCodeShortLink {
	//	// url中的链接要调用接口活动，并且要用到rallyCode
	//	sendMsgInfo.Interactive.Action.RallyCodeShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.RallyCodeShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
	//	shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.RallyCodeShortLink, msgInfoEntity.WaId)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = w.configInfo.Business.Activity.WaRedirectListPrefix + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constants.BizTypeInteractive == sendNxMsgType {
		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// EndCanNotStartGroupMsg 结束期-不能开团消息
func (w *WaMsgService) EndCanNotStartGroupMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.EndCanNotStartGroupMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// SwitchLangMsg 切换语言消息
func (w *WaMsgService) SwitchLangMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.SwitchLangMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}
	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// EndCanNotHelpMsg 结束期-不能助力消息
func (w *WaMsgService) EndCanNotHelpMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.EndCanNotHelpMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	sendJson, err := w.BuildInteractionMessage2NX(ctx, []*dto.BuildMsgInfo{msgInfoEntity}, []*conf.MsgInfo{sendMsgInfo})
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// RenewFreeReplyMsg 续订回复信息
func (w *WaMsgService) RenewFreeReplyMsg(ctx context.Context, msgInfoEntity *dto.BuildMsgInfo, shortLink string, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	msgInfoEntity.MsgType = constants.RenewFreeReplyMsg
	sendMsgInfo, err := w.getMsgInfo(ctx, msgInfoEntity)
	if err != nil {
		return nil, err
	}

	helpText, err := w.GetHelpText(ctx, msgInfoEntity.Language)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText
	// url中的链接要调用接口活动，并且要用到rallyCode
	//shortLink := user.RallyCodeShortLink
	//if "" == user.RallyCodeShortLink {
	//	// url中的链接要调用接口活动，并且要用到rallyCode
	//	sendMsgInfo.Interactive.Action.RallyCodeShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.RallyCodeShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
	//	shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.RallyCodeShortLink, msgInfoEntity.WaId)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = w.configInfo.Business.Activity.WaRedirectListPrefix + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	msgInfoEntityList := []*dto.BuildMsgInfo{msgInfoEntity}
	sendMsgInfoList := []*conf.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto
	if constants.BizTypeInteractive == sendNxMsgType {
		sendJson, err = w.BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = w.BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.WaId, msgInfoEntity.WaId, msgInfoEntity.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

func (w *WaMsgService) getMsgInfo(ctx context.Context, buildMsgInfo *dto.BuildMsgInfo) (*conf.MsgInfo, error) {
	allMsgMap := w.configInfo.MsgMap

	if singleMsgInfo, exists := allMsgMap[buildMsgInfo.MsgType]; exists {
		var msgInfoMap map[string]*conf.MsgInfo
		switch buildMsgInfo.Language {
		case "01":
			msgInfoMap = singleMsgInfo.L01
		case "02":
			msgInfoMap = singleMsgInfo.L02
		case "03":
			msgInfoMap = singleMsgInfo.L03
		case "04":
			msgInfoMap = singleMsgInfo.L04
		case "05":
			msgInfoMap = singleMsgInfo.L05
		default:
			w.l.Error(fmt.Sprintf("not supported language %v", buildMsgInfo.Language))
			return nil, errors.New("not supported language configuration")
		}
		if msgInfo, exists := msgInfoMap[w.configInfo.Business.Activity.Scheme]; exists {
			sendMsgInfo := &conf.MsgInfo{}
			err := util.CopyFieldsByJson(*msgInfo, sendMsgInfo)
			if err != nil {
				w.l.Error(fmt.Sprintf("copy error,SourceWaId:%v,toWaId: %v,language:%v，err:%v", buildMsgInfo.WaId, buildMsgInfo.WaId, buildMsgInfo.Language, err))
				return nil, errors.New("copy error")
			}
			return sendMsgInfo, nil
		} else {
			w.l.Error(fmt.Sprintf("Unsupported Scheme configuration ,SourceWaId:%v,toWaId: %v,language:%v", buildMsgInfo.WaId, buildMsgInfo.WaId, buildMsgInfo.Language))
			return nil, errors.New("Unsupported Scheme configuration")
		}
	} else {
		w.l.Error(fmt.Sprintf("Unsupported msgType configuration ,SourceWaId:%v,toWaId: %v,language:%v", buildMsgInfo.WaId, buildMsgInfo.WaId, buildMsgInfo.Language))
		return nil, errors.New("Unsupported msgType configuration")
	}
}

func (w *WaMsgService) GetHelpText(ctx context.Context, language string) (string, error) {
	//helpTextMap, err := w.redisService.GetHelpTextWeight(ctx)
	//if err != nil {
	//	w.l.Error(fmt.Sprintf("get helpText fail,err:%v", err))
	//	return "", err
	//}
	//todo 暂时用不到权重
	helpTextMap := w.configInfo.Business.Activity.HelpTextList[0]
	helpTextInfo := helpTextMap.BodyText[w.configInfo.Business.Activity.Scheme]
	helpText := ""
	if helpTextInfo != nil {
		switch language {
		case "01":
			helpText = helpTextInfo.L01
		case "02":
			helpText = helpTextInfo.L02
		case "03":
			helpText = helpTextInfo.L03
		case "04":
			helpText = helpTextInfo.L04
		case "05":
			helpText = helpTextInfo.L03
		default:
			w.l.Error(fmt.Sprintf("not supported language %v", language))
			return "", errors.New("not supported language configuration")
		}
	}
	return helpText, nil
}

func (w *WaMsgService) BuildInteractionMessage2NX(ctx context.Context, msgInfoEntityList []*dto.BuildMsgInfo, buildMsgParamsList []*conf.MsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	// msgInfoService :=w.getMsgInfoService()
	var sendMsgList []nx.NxReq
	var entityList []*dto.WaMsgSend
	for index, msgInfoEntity := range msgInfoEntityList {
		buildMsgParams := buildMsgParamsList[index]

		var interactive *nx.Interactive

		if buildMsgParams.Interactive.Type == "cta_url" {
			interactive = &nx.Interactive{
				Type: buildMsgParams.Interactive.Type,
				Body: &nx.NxReqInteractiveBody{
					Text: buildMsgParams.Interactive.BodyText,
				},
				//Footer: &nx.NxReqInteractiveFooter{
				//	Text: buildMsgParams.Interactive.FooterText,
				//},
				//Action: &nx.NxReqInteractiveAction{
				//	Name: "cta_url",
				//	Parameters: &nx.NxReqActionParameter{
				//		DisplayText: buildMsgParams.Interactive.Action.DisplayText,
				//		Url:         buildMsgParams.Interactive.Action.Url,
				//	},
				//},
			}
			if buildMsgParams.Interactive.ImageLink != "" {
				interactive.Header = &nx.NxReqInteractiveHeader{
					Type: "image",
					Image: &nx.NxReqInteractiveImage{
						Link: buildMsgParams.Interactive.ImageLink,
					},
				}
			}

			if buildMsgParams.Interactive.Action != nil {
				interactive.Action = &nx.NxReqInteractiveAction{
					Name: "cta_url",
					Parameters: &nx.NxReqActionParameter{
						DisplayText: buildMsgParams.Interactive.Action.DisplayText,
						Url:         buildMsgParams.Interactive.Action.Url,
					},
				}
			}
		} else if buildMsgParams.Interactive.Type == "button" {
			interactive = &nx.Interactive{
				Type: buildMsgParams.Interactive.Type,
				Body: &nx.NxReqInteractiveBody{
					Text: buildMsgParams.Interactive.BodyText,
				},
				//Footer: &nx.NxReqInteractiveFooter{
				//	Text: buildMsgParams.Interactive.FooterText,
				//},
				//Action: &nx.NxReqInteractiveAction{
				//	Buttons: buildMsgParams.Interactive.Action.Buttons,
				//},
			}
			if buildMsgParams.Interactive.ImageLink != "" {
				interactive.Header = &nx.NxReqInteractiveHeader{
					Type: "image",
					Image: &nx.NxReqInteractiveImage{
						Link: buildMsgParams.Interactive.ImageLink,
					},
				}
			}
			if buildMsgParams.Interactive.Action != nil {
				interactive.Action = &nx.NxReqInteractiveAction{
					Buttons: buildMsgParams.Interactive.Action.Buttons,
				}
			}
		} else {
			w.l.Error(fmt.Sprintf("方法[%s]，Interactive message types that are not supported,buildMsgParams.Interactive.Type:%v", methodName, buildMsgParams.Interactive.Type))
			return nil, errors.New("unsupported message type")
		}

		params := &nx.NxReqParam{
			Appkey:           w.configInfo.Data.Nx.AppKey,
			BusinessPhone:    w.configInfo.Data.Nx.BusinessPhone,
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               msgInfoEntity.WaId,
			// CusMessageId:     msgInfoEntity.Id,
			// "dr_webhook":        conf.ApplicationConfig.Nx.CallBackUrl,
			Type:        "interactive",
			Interactive: interactive,
		}

		paramsBytes, err := json.NewEncoder().Encode(params)
		if err != nil {
			w.l.Error(fmt.Sprintf("params convert json fail,params:%v,err:%v", params, err))
			return nil, err
		}
		paramsStr := string(paramsBytes)
		commonHeaders := w.getRequestHeader("mt", paramsStr, false)

		nxReq := nx.NxReq{
			Params:        params,
			CommonHeaders: commonHeaders,
		}
		sendMsgList = append(sendMsgList, nxReq)

		// 存储请求
		sendJsonBytes, err := json.NewEncoder().Encode(nxReq)
		if err != nil {
			w.l.Error(fmt.Sprintf("nxReq convert json fail,params:%v,err:%v", params, err))
			return nil, err
		}
		sendJson := string(sendJsonBytes)

		buildMsgParamsBytes, err := json.NewEncoder().Encode(buildMsgParams)
		if err != nil {
			w.l.Error(fmt.Sprintf("buildMsgParams convert json fail,params:%v,err:%v", params, err))
			return nil, err
		}
		buildMsgParamsJson := string(buildMsgParamsBytes)

		// 新增msgInfo
		waMsgSend := &dto.WaMsgSend{
			WaID:          msgInfoEntity.WaId,
			MsgType:       msgInfoEntity.MsgType,
			State:         constants.MsgSendStateUnSend,
			Content:       sendJson,
			BuildMsgParam: buildMsgParamsJson,
		}
		entityList = append(entityList, waMsgSend)
		//id, err := w.waMsgSendRepo.AddWaMsgSend(ctx, waMsgSend)
		//if err != nil {
		//	return nil, err
		//}
		//msgInfoEntity.Id = id
	}

	var res []*dto.SendNxListParamsDto
	for index, sendMsg := range sendMsgList {
		msgInfoEntity := msgInfoEntityList[index]
		entity := entityList[index]
		dto := &dto.SendNxListParamsDto{
			SendMsg:       sendMsg,
			MsgInfoEntity: msgInfoEntity,
			WaMsgSend:     entity,
		}
		res = append(res, dto)
	}
	return res, nil
}

func (w *WaMsgService) BuildTemplateMessage2NX(ctx context.Context, msgInfoEntityList []*dto.BuildMsgInfo, buildMsgParamsList []*conf.MsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := util.GetCurrentFuncName()
	//msgInfoService :=w.getMsgInfoService()
	var sendMsgList []nx.NxReq
	var entityList []*dto.WaMsgSend
	for index, msgInfoEntity := range msgInfoEntityList {
		buildMsgParams := buildMsgParamsList[index]

		if buildMsgParams.Template == nil {
			w.l.Error(fmt.Sprintf("方法[%s]，template is null,buildMsgParams.Template:%v", methodName, buildMsgParams.Template))
			return nil, errors.New("template is null")
		}

		buildMsgParams.Template.Language.Policy = "deterministic"

		params := &nx.NxReqParam{
			Appkey:           w.configInfo.Data.Nx.AppKey,
			BusinessPhone:    w.configInfo.Data.Nx.BusinessPhone,
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               msgInfoEntity.WaId,
			//CusMessageId:     msgInfoEntity.Id,
			// "dr_webhook":        conf.ApplicationConfig.Nx.CallBackUrl,
			Type:     "template",
			Template: buildMsgParams.Template,
		}

		paramsBytes, err := json.NewEncoder().Encode(params)
		if err != nil {
			w.l.Error(fmt.Sprintf("parma convert json fail,params:%v,err:%v", params, err))
			return nil, err
		}
		paramsStr := string(paramsBytes)
		commonHeaders := w.getRequestHeader("mt", paramsStr, false)

		nxReq := nx.NxReq{
			Params:        params,
			CommonHeaders: commonHeaders,
		}
		sendMsgList = append(sendMsgList, nxReq)

		// 存储请求
		sendJsonBytes, err := json.NewEncoder().Encode(nxReq)
		if err != nil {
			w.l.Error(fmt.Sprintf("nxReq convert json fail,params:%v,err:%v", params, err))
			return nil, err
		}
		sendJson := string(sendJsonBytes)

		//buildMsgParamsBytes, err := json.NewEncoder().Encode(buildMsgParams)
		//if err != nil {
		//	w.l.Error(fmt.Sprintf("方法[%s]，buildMsgParams转换json失败,params:%v,err:%v", methodName, params, err))
		//	return "", err
		//}
		//buildMsgParamsJson := string(buildMsgParamsBytes)

		waMsgSend := &dto.WaMsgSend{
			WaID:    msgInfoEntity.WaId,
			MsgType: msgInfoEntity.MsgType,
			State:   constants.MsgSendStateUnSend,
			Content: sendJson,
			//BuildMsgParam: buildMsgParamsJson,
		}
		entityList = append(entityList, waMsgSend)
		//id, err := w.waMsgSendRepo.AddWaMsgSend(ctx, waMsgSend)
		//if err != nil {
		//	return nil, err
		//}
		//msgInfoEntity.Id = id
	}

	var res []*dto.SendNxListParamsDto
	for index, sendMsg := range sendMsgList {
		msgInfoEntity := msgInfoEntityList[index]
		send := entityList[index]
		dto := &dto.SendNxListParamsDto{
			SendMsg:       sendMsg,
			MsgInfoEntity: msgInfoEntity,
			WaMsgSend:     send,
		}
		res = append(res, dto)
	}
	return res, nil
}

func (w *WaMsgService) SendMsgList2NX(ctx context.Context, sendNxListParamsDtoList []*dto.SendNxListParamsDto) (string, error) {
	if sendNxListParamsDtoList != nil && len(sendNxListParamsDtoList) > 0 {
		for _, sendNxListParamsDto := range sendNxListParamsDtoList {
			msgInfoEntity := sendNxListParamsDto.MsgInfoEntity
			sendMsg := sendNxListParamsDto.SendMsg
			waMsgSend := sendNxListParamsDto.WaMsgSend
			// 发送模板消息
			_, nxErr := w.SendNx(ctx, msgInfoEntity, sendMsg, waMsgSend)
			if nxErr != nil {
				return "", nxErr
			}
		}
	}

	return "", nil
}

func (w *WaMsgService) SendMsgList2NXHelpTimeOut(ctx context.Context, sendNxListParamsDtoList []*dto.SendNxListParamsDto) (string, error) {
	if sendNxListParamsDtoList != nil && len(sendNxListParamsDtoList) > 0 {
		for _, sendNxListParamsDto := range sendNxListParamsDtoList {
			msgInfoEntity := sendNxListParamsDto.MsgInfoEntity
			sendMsg := sendNxListParamsDto.SendMsg
			waMsgSend := sendNxListParamsDto.WaMsgSend
			// 发送模板消息
			_, nxErr := w.SendNx(ctx, msgInfoEntity, sendMsg, waMsgSend)
			if nxErr != nil {
				return "", nxErr
			}
		}
	}

	return "", nil
}

func (w *WaMsgService) SendMsgList2NXOfRetry(ctx context.Context, sendNxListParamsDtoList []*dto.SendNxListParamsDto) (string, error) {
	if sendNxListParamsDtoList != nil && len(sendNxListParamsDtoList) > 0 {
		for _, sendNxListParamsDto := range sendNxListParamsDtoList {
			msgInfoEntity := sendNxListParamsDto.MsgInfoEntity
			sendMsg := sendNxListParamsDto.SendMsg
			waMsgSend := sendNxListParamsDto.WaMsgSend
			// 发送模板消息
			_, nxErr := w.SendNxRetry(ctx, msgInfoEntity, sendMsg, waMsgSend)
			if nxErr != nil {
				return "", nxErr
			}
		}
	}

	return "", nil
}

func (w *WaMsgService) getRequestHeader(action string, paramsStr string, formData bool) map[string]string {
	commonHeaders := map[string]string{
		"accessKey": w.configInfo.Data.Nx.Ak,
		"ts":        strconv.FormatInt(time.Now().UnixMilli(), 10),
		"bizType":   "2",
		"action":    action,
	}
	var sign string
	if formData {
		sign = util.CallSignFormData(commonHeaders, w.configInfo.Data.Nx.Sk)
	} else {
		sign = util.CallSign(commonHeaders, paramsStr, w.configInfo.Data.Nx.Sk)
	}
	commonHeaders["sign"] = sign
	return commonHeaders
}

func (w *WaMsgService) SendNx(ctx context.Context, buildMsgInfo *dto.BuildMsgInfo, sendMsg nx.NxReq, waMsgSend *dto.WaMsgSend) (*response.NXResponse, error) {
	methodName := util.GetCurrentFuncName()

	resNx := &response.NXResponse{}
	var sendErr error
	for i := 1; i < 4; i++ {
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，start request nx,req：SourceWaId:%v, toWaId: %v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId))
		res, nxErr := rest.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", sendMsg.Params, sendMsg.CommonHeaders, 10*1000*time.Second, 10*1000*time.Second)

		if nxErr != nil {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			continue
		}
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，stop request nx,req：SourceWaId:%v, toWaId: %v,res:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, res))

		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
		if nxErr != nil {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，nx res convert struct error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，nx res convert struct error,res:%v,err：%v", methodName, i, res, nxErr))
			continue
		}
		if 0 != resNx.Code {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，request nx error,SourceWaId,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, resNx))
			continue
		}
		sendErr = nil
		break
	}

	w.l.Info(fmt.Sprintf("method[%s]， request nx success,req：SourceWaId:%v, toWaId: %v,res:%v", methodName, buildMsgInfo.WaId, buildMsgInfo.WaId, resNx))

	nxSendRes := &response.NXSendRes{
		NXResponse: resNx,
	}
	nxSendResBytes, err := json.NewEncoder().Encode(nxSendRes)
	if err != nil {
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx success nxSendResList convert json fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	nxSendResJson := string(nxSendResBytes)

	if sendErr != nil {
		// 更新
		waMsgSendEntity := &dto.WaMsgSend{
			ID:      waMsgSend.ID,
			State:   constants.MsgSendStateFail,
			SendRes: nxSendResJson,
		}
		if resNx.Data != nil && len(resNx.Data.Messages) > 0 {
			waMsgSendEntity.WaMsgID = resNx.Data.Messages[0].Id
		}
		err = w.waMsgSendRepo.UpdateWaMsg(ctx, waMsgSendEntity)
		if err != nil {
			w.l.WithContext(ctx).Error(fmt.Sprintf("方法[%s]，send msg fail，update waMsgSend fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
			return resNx, err
		}
		return nil, sendErr
	}

	// 判断waMsgId是否未空，不为空存储redis
	if resNx.Data.Messages[0].Id != "" {
		redisKey := fmt.Sprintf(constants.MsgSignKey, w.configInfo.Business.Activity.Id, resNx.Data.Messages[0].Id)
		setStatus, err := w.redisService.SetNX("UpdateWaMsg", redisKey, "1", 24*time.Hour)
		if err != nil {
			w.l.Error(fmt.Sprintf("set redis error key: %v", redisKey))
			return nil, err
		}
		if !setStatus {
			w.l.Error(fmt.Sprintf("set redis fail key: %v", redisKey))
			return nil, err
		}
	}

	// 更新
	waMsgSendEntity := &dto.WaMsgSend{
		ID:      waMsgSend.ID,
		WaMsgID: resNx.Data.Messages[0].Id,
		State:   constants.MsgSendStateSuccess,
		SendRes: nxSendResJson,
	}
	err = w.waMsgSendRepo.UpdateWaMsg(ctx, waMsgSendEntity)
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，send msg success，update waMsgSend fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	w.l.Info(fmt.Sprintf("update waMsgSend success,waMsgSend.ID:%v", waMsgSend.ID))
	return resNx, nil
}

// SendNxRetry 重试表发送到牛信云
func (w *WaMsgService) SendNxRetry(ctx context.Context, buildMsgInfo *dto.BuildMsgInfo, sendMsg nx.NxReq, waMsgSend *dto.WaMsgSend) (*response.NXResponse, error) {
	methodName := util.GetCurrentFuncName()

	resNx := &response.NXResponse{}
	var sendErr error
	for i := 1; i < 4; i++ {
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，start request nx,req：SourceWaId:%v, toWaId: %v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId))
		res, nxErr := rest.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", sendMsg.Params, sendMsg.CommonHeaders, 10*1000*time.Second, 10*1000*time.Second)

		if nxErr != nil {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			continue
		}
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，stop request nx,req：SourceWaId:%v, toWaId: %v,res:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, res))

		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
		if nxErr != nil {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，nx res convert struct error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，nx res convert struct error,res:%v,err：%v", methodName, i, res, nxErr))
			continue
		}
		if 0 != resNx.Code {
			w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx error,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，request nx error,SourceWaId,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, buildMsgInfo.WaId, buildMsgInfo.WaId, sendMsg.Params, sendMsg.CommonHeaders, resNx))
			continue
		}
		sendErr = nil
		break
	}

	w.l.Info(fmt.Sprintf("method[%s]， request nx success,req：SourceWaId:%v, toWaId: %v,res:%v", methodName, buildMsgInfo.WaId, buildMsgInfo.WaId, resNx))

	nxSendRes := &response.NXSendRes{
		NXResponse: resNx,
	}
	nxSendResBytes, err := json.NewEncoder().Encode(nxSendRes)
	if err != nil {
		w.l.Info(fmt.Sprintf("method[%s]，第[%v]次，request nx success nxSendResList convert json fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	nxSendResJson := string(nxSendResBytes)

	if sendErr != nil {
		// 更新
		waMsgSendEntity := &dto.WaMsgRetryDto{
			ID:      waMsgSend.ID,
			WaMsgID: resNx.Data.Messages[0].Id,
			State:   constants.MsgSendStateFail,
			SendRes: nxSendResJson,
		}
		err = w.waMsgRetryRepo.UpdateWaRetryMsg(ctx, waMsgSendEntity)
		if err != nil {
			w.l.Error(fmt.Sprintf("方法[%s]，send msg fail，update waMsgSend fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
			return resNx, err
		}
		return nil, sendErr
	}

	// 更新

	// 判断waMsgId是否未空，不为空存储redis
	if resNx.Data.Messages[0].Id != "" {
		redisKey := fmt.Sprintf(constants.MsgSignKey, w.configInfo.Business.Activity.Id, resNx.Data.Messages[0].Id)
		setStatus, err := w.redisService.SetNX("UpdateWaMsg", redisKey, "1", 24*time.Hour)
		if err != nil {
			w.l.Error(fmt.Sprintf("set redis error key: %v", redisKey))
			return nil, err
		}
		if !setStatus {
			w.l.Error(fmt.Sprintf("set redis fail key: %v", redisKey))
			return nil, err
		}
	}

	waMsgSendEntity := &dto.WaMsgRetryDto{
		ID:      waMsgSend.ID,
		WaMsgID: resNx.Data.Messages[0].Id,
		State:   constants.MsgSendStateSuccess,
		SendRes: nxSendResJson,
	}
	err = w.waMsgRetryRepo.UpdateWaRetryMsg(ctx, waMsgSendEntity)
	if err != nil {
		w.l.Error(fmt.Sprintf("方法[%s]，send msg success，update waMsgSend fail,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	w.l.Info(fmt.Sprintf("update waMsgSend success,waMsgSend.ID:%v", waMsgSend.ID))
	return resNx, nil
}

// CheckCanSendMsg2NX 检查是否可以发消息给牛信云
//func (w *WaMsgService) CheckCanSendMsg2NX(ctx context.Context, waId string) (bool, bool, error) {
//	methodName := util.GetCurrentFuncName()
//	isFree, err := CheckIsFreeByWaId(ctx, waId)
//	if err != nil {
//		w.l.Error(fmt.Sprintf("方法[%s]，根据waId查询是否是免费期报错,活动id:%v,waId:%v,err：%v", methodName, conf.ApplicationConfig.Activity.Id, waId, err))
//		return false, false, errors.New("database is error")
//	}
//	if !isFree {
//		// 不在免费时间内，查询是否超额
//		isUltraLimit, err := CostIsUltraLimit(ctx)
//		if err != nil {
//			return false, isFree, err
//		}
//		if isUltraLimit {
//			return false, isFree, nil
//		}
//	}
//	return true, isFree, nil
//}

// CheckIsFreeByWaId 检查是否在免费期
//func (w *WaMsgService) CheckIsFreeByWaId(ctx context.Context, waId string) (bool, error) {
//	methodName := util.GetCurrentFuncName()
//	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
//	session, isExist, err := txUtil.GetTransaction(ctx)
//	if nil != err {
//		w.l.Error(fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
//		return false, errors.New("database is error")
//	}
//	if !isExist {
//		defer func() {
//			session.Rollback()
//			session.Close()
//		}()
//	}
//
//	userAttendInfo, err := userAttendInfoMapper.SelectByWaIdBySession(&session, waId)
//	if err != nil {
//		w.l.Error(fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo报错,活动id:%v,waId:%v,err：%v", methodName, conf.ApplicationConfig.Activity.Id, waId, err))
//		return false, errors.New("database is error")
//	}
//	if userAttendInfo.Id <= 0 {
//		w.l.Error(fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo不存在,活动id:%v,waId:%v,err：%v", methodName, conf.ApplicationConfig.Activity.Id, waId, err))
//		return false, errors.New("database is error")
//	}
//	now := util.GetNowCustomTime()
//	if userAttendInfo.NewestFreeEndAt.Before(now.Time) || userAttendInfo.NewestFreeEndAt.Equal(now.Time) {
//		return false, nil
//	}
//	if !isExist {
//		session.Commit()
//	}
//	return true, nil
//}

func (w WaMsgService) UploadMedia2NX(ctx context.Context, path string) (string, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		w.l.Error(fmt.Sprintf("打开文件失败,path:%v,err:%v", path, err))
		return "", err
	}
	defer file.Close()

	// 创建一个buffer
	var buffer bytes.Buffer

	// 创建multipart writer
	writer := multipart.NewWriter(&buffer)

	// 添加公共参数
	writer.WriteField("appkey", w.configInfo.Data.Nx.AppKey)
	writer.WriteField("business_phone", w.configInfo.Data.Nx.BusinessPhone)
	writer.WriteField("messaging_product", "whatsapp")
	writer.WriteField("type", "image/png")

	// 添加文件
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		w.l.Error(fmt.Sprintf("创建form文件失败,err:%v", err))
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		w.l.Error(fmt.Sprintf("复制文件失败,err:%v", err))
		return "", err
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		w.l.Error(fmt.Sprintf("关闭writer失败,err:%v", err))
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://api2.nxcloud.com/api/wa/uploadMedia", &buffer)
	if err != nil {
		w.l.Error(fmt.Sprintf("创建请求失败,err:%v", err))
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	headerMap := w.getRequestHeader("uploadTemplateFile", "", true)
	for s2 := range headerMap {
		req.Header.Set(s2, headerMap[s2])
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	resNx := &response.NXResponse{}
	err = json.NewEncoder().Decode(respBytes, resNx)
	if err != nil {
		w.l.Error(fmt.Sprintf("转换json失败,err:%v", err))
		return "", err
	}
	if 0 != resNx.Code {
		w.l.Error(fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
		return "", errors.New(fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
	}
	return resNx.Data.Id, nil
}

func (w *WaMsgService) GetStageInfoByAttendStatus(ctx context.Context, methodName string, user *dto.BuildMsgInfo, helpNameList []*dto.HelpNickNameInfo) (*dto.RewardStageDto, error) {
	stageInfoList := w.configInfo.Business.Activity.StageAwardList
	currentStageMax := stageInfoList[0].HelpNum
	currentStageName := stageInfoList[0].AwardName
	currentAwardLink := stageInfoList[0].AwardLink

	nextStageMax := stageInfoList[0].HelpNum
	nextStageName := stageInfoList[0].AwardName
	nextAwardLink := stageInfoList[0].AwardLink
	helpCount := int32(len(helpNameList))
	stageNum := 1

	for index, stageInfo := range stageInfoList {
		if index == 0 {
			if helpCount <= stageInfo.HelpNum {
				stageNum = 1
				nextStageMax = stageInfo.HelpNum
				nextStageName = stageInfo.AwardName
				nextAwardLink = stageInfo.AwardLink
				break
			}
		} else if index == len(stageInfoList)-1 {
			if helpCount <= stageInfo.HelpNum {
				stageNum = len(stageInfoList)
				currentStageMax = stageInfo.HelpNum
				currentStageName = stageInfo.AwardName
				currentAwardLink = stageInfo.AwardLink
				nextStageMax = stageInfo.HelpNum
				nextStageName = stageInfo.AwardName
				nextAwardLink = stageInfo.AwardLink
				break
			}
		} else {
			preStage := stageInfoList[index-1]
			if helpCount > preStage.HelpNum && helpCount <= stageInfo.HelpNum {
				stageNum = index + 1
				currentStageMax = preStage.HelpNum
				currentStageName = preStage.AwardName
				currentAwardLink = preStage.AwardLink
				nextStageMax = stageInfo.HelpNum
				nextStageName = stageInfo.AwardName
				nextAwardLink = stageInfo.AwardLink
				break
			}
		}
	}
	//
	//if helpCount < activity.Stage1Award.HelpNum {
	//	nextStageMax = activity.Stage1Award.HelpNum
	//	nextStageName = activity.Stage1Award.AwardName
	//	nextAwardLink = activity.Stage1Award.AwardLink
	//
	//} else if helpCount >= activity.Stage1Award.HelpNum && helpCount < activity.Stage2Award.HelpNum {
	//	currentStageMax = activity.Stage1Award.HelpNum
	//	currentStageName = activity.Stage1Award.AwardName
	//	currentAwardLink = activity.Stage1Award.AwardLink
	//	nextStageMax = activity.Stage2Award.HelpNum
	//	nextStageName = activity.Stage2Award.AwardName
	//	nextAwardLink = activity.Stage2Award.AwardLink
	//
	//} else if helpCount >= activity.Stage2Award.HelpNum && helpCount < activity.Stage3Award.HelpNum {
	//	currentStageMax = activity.Stage2Award.HelpNum
	//	currentStageName = activity.Stage2Award.AwardName
	//	currentAwardLink = activity.Stage2Award.AwardLink
	//	nextStageMax = activity.Stage3Award.HelpNum
	//	nextStageName = activity.Stage3Award.AwardName
	//	nextAwardLink = activity.Stage3Award.AwardLink
	//
	//} else if helpCount >= activity.Stage3Award.HelpNum {
	//	currentStageMax = activity.Stage3Award.HelpNum
	//	currentStageName = activity.Stage3Award.AwardName
	//	currentAwardLink = activity.Stage3Award.AwardLink
	//	nextStageMax = activity.Stage3Award.HelpNum
	//	nextStageName = activity.Stage3Award.AwardName
	//	nextAwardLink = activity.Stage3Award.AwardLink
	//} else {
	//	w.l.Error(fmt.Sprintf("方法[%s]，不支持的状态,WaId: %v", methodName, user.WaId))
	//	return nil, errors.New("不支持的状态")
	//}
	return &dto.RewardStageDto{
		StageNum:         stageNum,
		CurrentStageMax:  int(currentStageMax),
		CurrentStageName: currentStageName[user.Language],
		CurrentAwardLink: currentAwardLink[user.Language],
		NextStageMax:     int(nextStageMax),
		NextStageName:    nextStageName[user.Language],
		NextAwardLink:    nextAwardLink[user.Language],
	}, nil
}

//
//// 电话号码列表
//var phoneNumbers = []string{
//	"8618758081695",
//	"60177761865",
//	"601126703621",
//	"60109489084",
//	"60129708408",
//	"60149619180",
//	"6589410609",
//	"8618321868434",
//	"60126714138",
//	"85257481920",
//	"85296745569",
//	"85266831314",
//}
//
//func GetRanDomMessage(ctx context.Context, interactive *nx.Interactive, batch int, j int) {
//	// 初始化随机数种子
//	rand.Seed(time.Now().UnixNano())
//
//	toWaId := phoneNumbers[j]
//	w.l.Error( fmt.Sprintf("nickename index %v toWaId %v", j, toWaId))
//
//	params := &nx.NxReqParam{
//		Appkey:           conf.ApplicationConfig.Nx.AppKey,
//		BusinessPhone:    conf.ApplicationConfig.Nx.BusinessPhone,
//		MessagingProduct: "whatsapp",
//		RecipientType:    "individual",
//		To:               toWaId,
//		Type:             "interactive",
//		Interactive:      interactive,
//	}
//
//	paramsBytes, err := json.NewEncoder().Encode(params)
//	if err != nil {
//		//w.l.Error( fmt.Sprintf("方法[%s]，params转换json失败,params:%v,err:%v", GetRanDomMessage, params, err))
//		return
//	}
//	var sendMsgList []nx.NxReq
//
//	paramsStr := string(paramsBytes)
//	commonHeaders := w.getRequestHeader("mt", paramsStr, false)
//
//	nxReq := nx.NxReq{
//		Params:        params,
//		CommonHeaders: commonHeaders,
//	}
//	resNx := &response.NXResponse{}
//
//	methodName := util.GetCurrentFuncName()
//	sendMsgList = append(sendMsgList, nxReq)
//	for i, sendMsg := range sendMsgList {
//		res, nxErr := http_client.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", sendMsg.Params, sendMsg.CommonHeaders, 10*1000*time.Second, 10*1000*time.Second)
//		if nxErr != nil {
//			w.l.Error( fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口发生错误, %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
//			continue
//		}
//		//logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，结束调用牛信云接口,请求：SourceWaId:%v, toWaId: %v, 返回: %v", methodName, i, res))
//
//		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
//		if nxErr != nil {
//			w.l.Error( fmt.Sprintf("方法[%s]，第[%v]次，牛信云接口返回转实体报错,res:%v,err：%v", methodName, i, res, nxErr))
//			continue
//		}
//		if 0 != resNx.Code {
//			w.l.Error( fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口失败, %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, sendMsg.Params, sendMsg.CommonHeaders, resNx))
//			continue
//		}
//		//fmt.Println("随机推送成功 # # #", toWaId, i, sendMsg)
//		json_str, _ := json.NewEncoder().Encode(sendMsg)
//
//		w.l.Error( fmt.Sprintf("send success waid %v  sendMsg %v ", toWaId, string(json_str)))
//
//	}
//
//}
