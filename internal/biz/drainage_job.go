package biz

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/pojo/dto/response"
	"fission-basic/internal/rest"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/kit/sqlx"
	"fission-basic/util"
	"fission-basic/util/strUtil"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

//var drainageJobPool13 = goroutine_pool.NewGoroutinePool(2)
//var drainageJobPool2 = goroutine_pool.NewGoroutinePool(10)
//var drainageJobPool4 = goroutine_pool.NewGoroutinePool(20)

var pushEvent13JsonMap = map[string]string{
	"60":  "{\"appkey\":\"WC4mRoJ8\",\"business_phone\":\"60166986117\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_60_my\",\"language\":{\"policy\":\"deterministic\",\"code\":\"ml\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"641726371686542\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"62":  "{\"appkey\":\"WC4mRoJ8\",\"business_phone\":\"60166986117\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_62_id\",\"language\":{\"policy\":\"deterministic\",\"code\":\"id\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"3788360078141917\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"63":  "{\"appkey\":\"WC4mRoJ8\",\"business_phone\":\"60166986117\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_63_en\",\"language\":{\"policy\":\"deterministic\",\"code\":\"en\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1353152362254979\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"852": "{\"appkey\":\"WC4mRoJ8\",\"business_phone\":\"60166986117\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_63_en\",\"language\":{\"policy\":\"deterministic\",\"code\":\"en\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1353152362254979\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
}

var pushEvent1ParamMap = map[string]string{
	"60": "events/mlbb25031/promotion/?code=f030100000&lang=03",
	"62": "events/mlbb25031/promotion/?code=f040100000&lang=04",
	"63": "events/mlbb25031/promotion/?code=f020100000&lang=02",
}

var pushEvent2JsonMap2 = map[string]string{
	"60":  "{\"appkey\":\"Af2xgOMY\",\"business_phone\":\"6560390305\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_t2_my\",\"language\":{\"policy\":\"deterministic\",\"code\":\"ml\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"990703102394666\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"62":  "{\"appkey\":\"Af2xgOMY\",\"business_phone\":\"6560390305\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_t2_id\",\"language\":{\"policy\":\"deterministic\",\"code\":\"id\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1243964020665247\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"63":  "{\"appkey\":\"Af2xgOMY\",\"business_phone\":\"6560390305\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_t2_en\",\"language\":{\"policy\":\"deterministic\",\"code\":\"en\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1677826566473784\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"65":  "{\"appkey\":\"Af2xgOMY\",\"business_phone\":\"6560390305\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_t2_en\",\"language\":{\"policy\":\"deterministic\",\"code\":\"en\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1677826566473784\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
	"852": "{\"appkey\":\"Af2xgOMY\",\"business_phone\":\"6560390305\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"type\":\"template\",\"template\":{\"name\":\"mlbb25031_t2_en\",\"language\":{\"policy\":\"deterministic\",\"code\":\"en\"},\"components\":[{\"type\":\"header\",\"parameters\":[{\"type\":\"image\",\"image\":{\"id\":\"1677826566473784\"}}]},{\"type\":\"button\",\"sub_type\":\"url\",\"index\":\"0\",\"parameters\":[{\"type\":\"text\",\"text\":\"%s\"}]}]}}",
}

var recurringProbList = []string{
	"1l5o3du4_1lesx50w?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1l6emzuu_1lnittpt?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1l67zui3_1lfyaqof?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1lhtlnjd_1lrqi8me?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1lhq7h13_1li01wk7?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1ldsb04z_1lpdeal1?deeplink=mobilelegends%3A%2F%2Fappinvites",
	"1lbvaqma_1lcm2se3?deeplink=mobilelegends%3A%2F%2Fappinvites",
}

var socialScoreListMap = map[string][]string{
	"60": {
		"events/mlbb25031/promotion/?code=j030100000&lang=03",
		"events/mlbb25031/promotion/?code=k030100000&lang=03",
		"events/mlbb25031/promotion/?code=l030100000&lang=03",
		"events/mlbb25031/promotion/?code=m030100000&lang=03",
		"events/mlbb25031/promotion/?code=n030100000&lang=03",
		"events/mlbb25031/promotion/?code=o030100000&lang=03",
		"events/mlbb25031/promotion/?code=p030100000&lang=03",
		"events/mlbb25031/promotion/?code=q030100000&lang=03",
	},
	"62": {
		"events/mlbb25031/promotion/?code=j040100000&lang=04",
		"events/mlbb25031/promotion/?code=k040100000&lang=04",
		"events/mlbb25031/promotion/?code=l040100000&lang=04",
		"events/mlbb25031/promotion/?code=m040100000&lang=04",
		"events/mlbb25031/promotion/?code=n040100000&lang=04",
		"events/mlbb25031/promotion/?code=o040100000&lang=04",
		"events/mlbb25031/promotion/?code=p040100000&lang=04",
		"events/mlbb25031/promotion/?code=q040100000&lang=04",
	},
	"852": {
		"events/mlbb25031/promotion/?code=j040100000&lang=04",
		"events/mlbb25031/promotion/?code=k040100000&lang=04",
		"events/mlbb25031/promotion/?code=l040100000&lang=04",
		"events/mlbb25031/promotion/?code=m040100000&lang=04",
		"events/mlbb25031/promotion/?code=n040100000&lang=04",
		"events/mlbb25031/promotion/?code=o040100000&lang=04",
		"events/mlbb25031/promotion/?code=p040100000&lang=04",
		"events/mlbb25031/promotion/?code=q040100000&lang=04",
	},
	//"63": {
	//	"events/mlbb25031/promotion/?code=j020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=k020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=l020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=m020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=n020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=o020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=p020100000&lang=02",
	//	"events/mlbb25031/promotion/?code=q020100000&lang=02",
	//},
}

var pushEvent4JsonList = []string{
	"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1740329666680-PDvMeviTkv.png\"}},\"body\":{\"text\":\"Manalo ng mga Kamangha-manghang gantimpala - MLBB GOLDEN MONTH CREATION CONTEST ay Live na Ngayon! 🥳\\n\\n🎁 Ang unang beses na pagsusumite ay garantisadong makakakuha ng isang eksklusibong Creator Avatar Border at mga mahahalagang item\\n🎁 Masasaganang Gantimpalang Cash\\n🎁 Napakaraming Gantimpalang Diamond\\n\\nI-tap ang button sa ibaba upang makilahok 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"Join Now\",\"url\":\"https://sg-play.mobilelegends.com/events/goldenmonthugc/ph?utm_source=wa\"}}}}",
	"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1740329665793-XrtYUi8cNB.png\"}},\"body\":{\"text\":\"Menangkan Hadiah Menarik - KONTES KREASI GOLDEN MONTH MLBB Telah Hadir! 🥳\\n\\n🎁 Karya pertama dijamin mendapatkan Border Avatar Kreator eksklusif dan item berharga\\n🎁 Hadiah Uang Tunai Melimpah\\n🎁 Hadiah Diamond Melimpah\\n\\nKetuk tombol di bawah untuk berpartisipasi 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"Ikutan Sekarang\",\"url\":\"https://sg-play.mobilelegends.com/events/goldenmonthugc/id?utm_source=wa\"}}}}",
	"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1740329667044-va71SAFFOl.png\"}},\"body\":{\"text\":\"Menangi Ganjaran Menakjubkan - PERADUAN MENCIPTA KANDUNGAN GOLDEN MONTH MLBB Kini Bermula! 🥳\\n\\n🎁 Penyertaan pertama dijamin untuk mendapat Bingkai Avatar Pencipta eksklusif dan item lumayan\\n🎁 Ganjaran Tunai yang Berlimpah\\n🎁 Ganjaran Berlian yang Besar\\n\\nTekan butang di bawah untuk menyertai 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"Sertai Sekarang\",\"url\":\"https://sg-play.mobilelegends.com/events/goldenmonthugc/my?utm_source=wa\"}}}}",
	"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1740329663417-58j2iud9GW.png\"}},\"body\":{\"text\":\"Выиграйте потрясающие награды — ТВОРЧЕСКИЙ КОНКУРС MLBB GOLDEN MONTH уже начался! 🥳\\n\\n🎁 За первую заявку гарантированно получите эксклюзивную рамку аватара создателя и ценные предметы\\n🎁 Щедрые денежные награды\\n🎁 Огромные алмазные награды\\n\\nКоснитесь кнопки ниже, чтобы принять участие 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"УЧАСТВУЙТЕ СЕЙЧАС\",\"url\":\"https://sg-play.mobilelegends.com/events/goldenmonthugc/eeca?utm_source=wa\"}}}}",
	"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1740329665219-RoqS7dczUx.png\"}},\"body\":{\"text\":\"Win Amazing Rewards - MLBB GOLDEN MONTH CREATION CONTEST is Now Live! 🥳\\n\\n🎁 First-time submission is guaranteed to get an exclusive Creator Avatar Border and valuable items\\n🎁 Generous Cash Rewards\\n🎁 Massive Diamond Rewards\\n\\nTap the button below to participate 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"Join Now\",\"url\":\"https://sg-play.mobilelegends.com/events/goldenmonthugc/global?utm_source=wa\"}}}}",
	//"{\"appkey\":\"a5AzGtEr\",\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"%s\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"cta_url\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1741271337159-k8JuRmTpyK.png\"}},\"body\":{\"text\":\"Win Amazing Rewards - MOBA55 GOLDEN MONTH CREATION CONTEST is Now Live! 🥳\\n\\n🎁 First-time submission is guaranteed to get an exclusive Creator Avatar Border and valuable items\\n🎁 Generous Cash Rewards\\n🎁 Massive Diamond Rewards\\n\\nTap the button below to participate 👇\"},\"action\":{\"name\":\"cta_url\",\"parameters\":{\"display_text\":\"Join Now\",\"url\":\"https://play.moba5v5.com/events/goldenmonthugc/?utm_source=wa\"}}}}",
}

type DrainageJob struct {
	waUserScoreRepo      WaUserScoreRepo
	uploadUserInfoRepo   UploadUserInfoRepo
	userInfoRepo         UserInfoRepo
	msgReceivedRepo      WaMsgReceivedRepo
	pushEventSendMsgRepo PushEventSendMessageRepo
	pushEvent4UserRepo   PushEvent4UserRepo
	countLimitRepo       CountLimitRepo
	redisService         *redis.RedisService
	bizConf              *conf.Business
	configInfo           *conf.Bootstrap
	l                    *log.Helper
}

func NewDrainageJob(
	waUserScoreRepo WaUserScoreRepo,
	uploadUserInfoRepo UploadUserInfoRepo,
	userInfoRepo UserInfoRepo,
	msgReceivedRepo WaMsgReceivedRepo,
	pushEventSendMsgRepo PushEventSendMessageRepo,
	pushEvent4UserRepo PushEvent4UserRepo,
	countLimitRepo CountLimitRepo,
	redisService *redis.RedisService,
	bizConf *conf.Business,
	configInfo *conf.Bootstrap,
	l log.Logger,
) *DrainageJob {
	return &DrainageJob{
		waUserScoreRepo:      waUserScoreRepo,
		uploadUserInfoRepo:   uploadUserInfoRepo,
		userInfoRepo:         userInfoRepo,
		msgReceivedRepo:      msgReceivedRepo,
		pushEventSendMsgRepo: pushEventSendMsgRepo,
		pushEvent4UserRepo:   pushEvent4UserRepo,
		countLimitRepo:       countLimitRepo,
		redisService:         redisService,
		bizConf:              bizConf,
		configInfo:           configInfo,
		l:                    log.NewHelper(l),
	}
}

// 活动1推送消息
func (d *DrainageJob) SendPushEvent1Msg(ctx context.Context) {
	taskLock := constants.PushEventJobTaskLockPrefix + "1"
	methodName := util.GetCurrentFuncName()
	getLock, err := d.redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		//d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		//d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,get redis lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := d.redisService.Del(taskLock)
		if !del {
			d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s，del redis lock fail", methodName))
		}
	}()
	d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send push event1 msg", methodName))

	id := 492675
	count := 0
	for {
		d.l.WithContext(ctx).Infof("method:%s current id:%v count:%v", methodName, id, count)
		uploadUserInfos, err := d.uploadUserInfoRepo.ListGtIdWithState(ctx, id, 0, 2000)
		if err != nil {
			d.l.WithContext(ctx).Errorf("mthod%s,PageByField error occurred，waId:%v", util.GetCurrentFuncName(), uploadUserInfos)
			return
		}
		if len(uploadUserInfos) == 0 {
			d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,uploadUserInfos is empty break", methodName))
			return
		}
		waIds := make([]string, 0)
		for _, uploadUserInfo := range uploadUserInfos {
			waIds = append(waIds, uploadUserInfo.PhoneNumber)
		}
		d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,waIds:%v", methodName, len(waIds)))
		userInfoMap, err := d.getUserInfoMap(ctx, waIds)
		if err != nil {
			d.l.Errorf("mthod%s,getUserInfoMap error occurred，waId:%v", util.GetCurrentFuncName(), waIds)
			return
		}
		d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,userInfoMap:%v", methodName, len(userInfoMap)))
		for _, uploadUserInfo := range uploadUserInfos {
			//  裂变活动已存在
			if userInfoMap[uploadUserInfo.PhoneNumber] {
				d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,uploadUserInfo:%v,already exist", methodName, uploadUserInfo))
				continue
			}
			// 缩减30分钟 留有发送余裕
			if uploadUserInfo.LastSendTime.Before(time.Now().Add(-time.Hour*23 - time.Minute*10)) {
				continue
			}
			d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,uploadUserInfo:%v", methodName, uploadUserInfo))

			err = d.sendPushEvent1Msg(ctx, uploadUserInfo)
			if err != nil {
				d.l.Errorf("mthod%s,sendPushEvent1Msg error occurred，waId:%v", methodName, uploadUserInfo)
				return
			}
			d.updatePushEvent1State(ctx, uploadUserInfo)
			//d.sendPushEvent1MsgByCoroutine(ctx, uploadUserInfo)
			count++
		}
		//drainageJobPool.Wait()
		id = uploadUserInfos[len(uploadUserInfos)-1].Id
		d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,nextid:%v", methodName, id))
	}
}

func (d *DrainageJob) sendPushEvent1MsgByCoroutine(ctx context.Context, uploadUserInfo *UploadUserInfoDTO) {
	methodName := util.GetCurrentFuncName()
	err := d.sendPushEvent1Msg(ctx, uploadUserInfo)
	if err != nil {
		d.l.Errorf("mthod%s,sendPushEvent1MsgByCoroutine error occurred，waId:%v", methodName, uploadUserInfo)
		return
	}
	d.updatePushEvent1State(ctx, uploadUserInfo)
	//drainageJobPool13.Execute(func(param interface{}) {
	//	u, ok := param.(*UploadUserInfoDTO) // 断言u是User类型
	//	if !ok {
	//		d.l.Errorf("mthod%s,Assertion error occurred，waId:%v", methodName, u)
	//		return
	//	}
	//	err := d.sendPushEvent1Msg(ctx, u)
	//	if err != nil {
	//		d.l.Errorf("mthod%s,resendGoroutinePool The pool execution task starts error，waId:%v", methodName, u)
	//		return
	//	}
	//	d.updatePushEvent1State(ctx, uploadUserInfo)
	//	d.l.Infof("mthod%s,resendGoroutinePool The pool execution task starts，waId:%v", methodName, u)
	//}, uploadUserInfo)
}

func (d *DrainageJob) sendPushEvent1Msg(ctx context.Context, uploadUserInfo *UploadUserInfoDTO) error {
	countryCode := "63"
	for _, prefix := range d.bizConf.PushEvent1.CountryCodes {
		if strings.HasPrefix(uploadUserInfo.PhoneNumber, prefix) {
			countryCode = prefix
			break
		}
	}

	d.l.WithContext(ctx).Infof("mthod%s,sendPushEvent1Msg，waId:%v countryCode:%v", util.GetCurrentFuncName(), uploadUserInfo.PhoneNumber, countryCode)
	d.send13Msg(ctx, uploadUserInfo.PhoneNumber, countryCode, pushEvent1ParamMap[countryCode], 1)
	return nil
}

func (d *DrainageJob) updatePushEvent1State(ctx context.Context, uploadUserInfo *UploadUserInfoDTO) {
	err := d.uploadUserInfoRepo.UpdateState(ctx, uploadUserInfo.PhoneNumber, 1000)
	if err != nil {
		d.l.Errorf("mthod%s err: %v", util.GetCurrentFuncName(), err)
		return
	}
}

// 活动2推送消息
func (d *DrainageJob) SendPushEvent2Msg(ctx context.Context) {
	taskLock := constants.PushEventJobTaskLockPrefix + "2"
	methodName := util.GetCurrentFuncName()
	getLock, err := d.redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,get redis lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := d.redisService.Del(taskLock)
		if !del {
			d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s，del redis lock fail", methodName))
		}
	}()
	d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send push event2 msg", methodName))

	d.initCount(ctx, constants.PushEvent2CountKey)

	if !d.validCount(ctx, constants.PushEvent2CountKey, constants.PushEvent2CountLimit) {
		d.l.Infof("推送数已达上限，不再推送事件2")
		return
	}

	offset := uint(350000)
	count := 0
	endTime := time.Date(2025, time.February, 25, 23, 59, 59, 0, time.Local)
	countryCodes := d.bizConf.PushEvent2.CountryCodes
	for {
		waUserScores, err := d.waUserScoreRepo.PageByRecurringProb(ctx, offset, 3000)
		d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,offset:%v", methodName, offset))
		if err != nil {
			d.l.Errorf("mthod%s,PageByField error occurred，waId:%v", util.GetCurrentFuncName(), waUserScores)
			return
		}
		if len(waUserScores) == 0 {
			d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,waUserScores is empty break", methodName))
			return
		}
		offset += uint(len(waUserScores))
		waIds := make([]string, 0)
		for _, waUserScore := range waUserScores {
			waIds = append(waIds, waUserScore.WaId)
		}
		if err != nil {
			d.l.Errorf("mthod%s,getUserInfoMap error occurred，waId:%v", util.GetCurrentFuncName(), waIds)
			return
		}
		uploadUserInfoMap, err := d.getUploadUserInfoMap(ctx, waIds)
		if err != nil {
			d.l.Errorf("mthod%s,getUploadUserInfoMap error occurred，waId:%v", util.GetCurrentFuncName(), waIds)
			return
		}
		for _, waUserScore := range waUserScores {
			// 活动1已存在
			if uploadUserInfoMap[waUserScore.WaId] {
				continue
			}
			// 晚于2025年2月25日，不推送
			if waUserScore.LastLoginTime.After(endTime) {
				continue
			}
			// 已发送过的 不推送
			if waUserScore.State != 0 {
				continue
			}
			// 成功率>=30%
			if waUserScore.RecurringProb < 0.3 {
				continue
			}
			// 过滤国码
			if !validCountryCode(countryCodes, waUserScore.WaId) {
				continue
			}
			if !d.validCount(ctx, constants.PushEvent2CountKey, constants.PushEvent2CountLimit) {
				d.l.Infof("推送数已达上限，不再推送事件2")
				return
			}
			d.sendPushEvent2MsgByCoroutine(ctx, waUserScore)
			count++
		}
		//drainageJobPool2.Wait()
	}
}

func (d *DrainageJob) sendPushEvent2MsgByCoroutine(ctx context.Context, waUserScore *WaUserScoreDTO) {
	methodName := util.GetCurrentFuncName()
	if !d.validCount(ctx, constants.PushEvent2CountKey, constants.PushEvent2CountLimit) {
		d.l.Infof("推送数已达上限，不再推送事件2")
		return
	}
	err := d.sendPushEvent2Msg(ctx, waUserScore)
	if err != nil {
		d.l.Errorf("mthod%s,sendPushEvent2MsgByCoroutine error occurred，waId:%v", methodName, waUserScore)
		return
	}
	d.updatePushEventState(ctx, waUserScore, 1)
	//drainageJobPool2.Execute(func(param interface{}) {
	//	u, ok := param.(*WaUserScoreDTO) // 断言u是User类型
	//	if !ok {
	//		d.l.Errorf("mthod%s,Assertion error occurred，waId:%v", methodName, u)
	//		return
	//	}
	//	if !d.validCount(ctx, constants.PushEvent2CountKey, constants.PushEvent2CountLimit) {
	//		d.l.Infof("推送数已达上限，不再推送事件2")
	//		return
	//	}
	//	err := d.sendPushEvent2Msg(ctx, u)
	//	if err != nil {
	//		d.l.Errorf("mthod%s,resendGoroutinePool The pool execution task starts error，waId:%v", methodName, u)
	//		return
	//	}
	//	d.updatePushEventState(ctx, waUserScore, 1)
	//	d.l.Infof("mthod%s,resendGoroutinePool The pool execution task starts，waId:%v", methodName, u)
	//}, waUserScore)
}

func (d *DrainageJob) sendPushEvent2Msg(ctx context.Context, waUserScore *WaUserScoreDTO) error {
	method := util.GetCurrentFuncName()
	err := d.countLimitRepo.AddOne(ctx, constants.PushEvent2CountKey)
	if err != nil {
		d.l.Errorf("mthod%s,AddOne error occurred，waId:%v", method, waUserScore)
		return err
	}
	countryCode := ""
	for _, prefix := range d.bizConf.PushEvent2.CountryCodes {
		if strings.HasPrefix(waUserScore.WaId, prefix) {
			countryCode = prefix
			break
		}
	}
	if countryCode == "" {
		d.l.Errorf("mthod%s,countryCode is empty，waId:%v", method, waUserScore.WaId)
		return nil
	}
	recurringProb := waUserScore.RecurringProb * 100
	param := ""
	if recurringProb >= 90 {
		param = recurringProbList[0]
	} else if recurringProb >= 80 {
		param = recurringProbList[1]
	} else if recurringProb >= 70 {
		param = recurringProbList[2]
	} else if recurringProb >= 60 {
		param = recurringProbList[3]
	} else if recurringProb >= 50 {
		param = recurringProbList[4]
	} else if recurringProb >= 40 {
		param = recurringProbList[5]
	} else if recurringProb >= 30 {
		param = recurringProbList[6]
	} else {
		d.l.Errorf("mthod%s,countryCode is empty，countryCode:%v", method, countryCode)
		return nil
	}
	d.l.WithContext(ctx).Infof("mthod%s,sendPushEvent2Msg，waId:%v countryCode:%v param:%v", method, waUserScore.WaId, countryCode, param)
	d.send2Msg(ctx, waUserScore.WaId, countryCode, param, 2)

	return nil
}

// 活动3推送消息
func (d *DrainageJob) SendPushEvent3Msg(ctx context.Context) {
	taskLock := constants.PushEventJob3TaskLockPrefix + "3"
	methodName := util.GetCurrentFuncName()
	getLock, err := d.redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		//d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,get redis lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := d.redisService.Del(taskLock)
		if !del {
			d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s，del redis lock fail", methodName))
		}
	}()

	d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send pushEvent3Msg", methodName))
	pushEvent3CountKey := constants.PushEvent3OnceCountKey
	d.initCount(ctx, pushEvent3CountKey)
	if !d.validCount(ctx, pushEvent3CountKey, constants.PushEvent3OnceCountLimit) {
		d.l.WithContext(ctx).Infof("单次推送数已达上限，不再推送事件3")
		return
	}

	offset := uint(2000000)
	count := 0
	startTime := time.Date(2025, time.February, 27, 0, 0, 0, 0, time.Local)
	countryCodes := d.bizConf.PushEvent3.CountryCodes
	//ctxTmp1 := context.WithoutCancel(ctx)

	for {
		waUserScores, err := d.waUserScoreRepo.PageBySocialScore(ctx, offset, 5000)
		d.l.WithContext(ctx).Infof("mthod%s,PageBySocialScore，offset:%v", util.GetCurrentFuncName(), offset)
		if err != nil {
			d.l.WithContext(ctx).Errorf("mthod%s,PageBySocialScore error occurred，waId:%v", util.GetCurrentFuncName(), waUserScores)
			return
		}
		if len(waUserScores) == 0 {
			d.l.WithContext(ctx).Infof("mthod%s,PageBySocialScore，offset:%v", util.GetCurrentFuncName(), offset)
			return
		}
		offset += uint(len(waUserScores))
		waIds := make([]string, 0)
		for _, waUserScore := range waUserScores {
			waIds = append(waIds, waUserScore.WaId)
		}
		userInfoMap, err := d.getUserInfoMap(ctx, waIds)
		if err != nil {
			d.l.WithContext(ctx).Errorf("mthod%s,getUserInfoMap error occurred，waId:%v", util.GetCurrentFuncName(), waIds)
			return
		}
		uploadUserInfoMap, err := d.getUploadUserInfoMap(ctx, waIds)
		if err != nil {
			d.l.WithContext(ctx).Errorf("mthod%s,getUploadUserInfoMap error occurred，waId:%v", util.GetCurrentFuncName(), waIds)
			return
		}
		for _, waUserScore := range waUserScores {
			// 裂变活动已存在
			if userInfoMap[waUserScore.WaId] {
				continue
			}
			// 活动1已存在
			if uploadUserInfoMap[waUserScore.WaId] {
				continue
			}
			// 早于2025年2月26日，不推送
			if waUserScore.LastLoginTime.Before(startTime) {
				continue
			}
			// 已发送过的 不推送
			if waUserScore.State != 0 {
				continue
			}
			// 过滤国码
			if !validCountryCode(countryCodes, waUserScore.WaId) {
				continue
			}
			if !d.validCount(ctx, pushEvent3CountKey, constants.PushEvent3OnceCountLimit) {
				d.l.WithContext(ctx).Infof("单次推送数已达上限，不再推送事件3")
				return
			}
			d.sendPushEvent3MsgByCoroutine(ctx, waUserScore, pushEvent3CountKey)
			count++
		}
		//drainageJobPool13.Wait()
	}
}

func (d *DrainageJob) sendPushEvent3MsgByCoroutine(ctx context.Context, waUserScore *WaUserScoreDTO, pushEvent3CountKey string) {
	methodName := util.GetCurrentFuncName()
	err := d.sendPushEvent3Msg(ctx, waUserScore, pushEvent3CountKey)
	if err != nil {
		d.l.Errorf("mthod%s,sendPushEvent3MsgByCoroutine error occurred，waId:%v", methodName, waUserScore)
		return
	}
	d.updatePushEventState(ctx, waUserScore, 3)
	//drainageJobPool13.Execute(func(param interface{}) {
	//	u, ok := param.(*WaUserScoreDTO) // 断言u是User类型
	//	if !ok {
	//		d.l.Errorf("mthod%s,Assertion error occurred，waId:%v", methodName, u)
	//		return
	//	}
	//	err := d.sendPushEvent3Msg(ctx, u, pushEvent3CountKey)
	//	if err != nil {
	//		d.l.Errorf("mthod%s,resendGoroutinePool The pool execution task starts error，waId:%v", methodName, u)
	//		return
	//	}
	//	d.updatePushEventState(ctx, waUserScore, 3)
	//	d.l.Infof("mthod%s,resendGoroutinePool The pool execution task starts，waId:%v", methodName, u)
	//}, waUserScore)
}

func (d *DrainageJob) sendPushEvent3Msg(ctx context.Context, waUserScore *WaUserScoreDTO, pushEvent3CountKey string) error {
	methodName := util.GetCurrentFuncName()
	err2 := d.countLimitRepo.AddOne(ctx, pushEvent3CountKey)
	if err2 != nil {
		d.l.Errorf("mthod%s,AddOne error occurred，waId:%v key:%v", methodName, waUserScore, pushEvent3CountKey)
		return err2
	}
	countryCode := ""
	for _, prefix := range d.bizConf.PushEvent3.CountryCodes {
		if strings.HasPrefix(waUserScore.WaId, prefix) {
			countryCode = prefix
			break
		}
	}
	if countryCode == "" {
		d.l.Errorf("mthod%s,countryCode is empty，waId:%v", methodName, waUserScore.WaId)
		return nil
	}
	score := waUserScore.SocialScore
	param := ""
	if score > 95 {
		param = socialScoreListMap[countryCode][0]
	} else if score > 90 {
		param = socialScoreListMap[countryCode][1]
	} else if score > 85 {
		param = socialScoreListMap[countryCode][2]
	} else if score > 80 {
		param = socialScoreListMap[countryCode][3]
	} else if score > 70 {
		param = socialScoreListMap[countryCode][4]
	} else if score > 60 {
		param = socialScoreListMap[countryCode][5]
	} else if score > 50 {
		param = socialScoreListMap[countryCode][6]
	} else {
		param = socialScoreListMap[countryCode][7]
	}
	d.send13Msg(ctx, waUserScore.WaId, countryCode, param, 3)
	return nil
}

// 活动4推送消息
func (d *DrainageJob) SendPushEvent4Msg(ctx context.Context) {
	taskLock := constants.PushEventJobTaskLockPrefix + "4"
	methodName := util.GetCurrentFuncName()
	getLock, err := d.redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,get redis lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := d.redisService.Del(taskLock)
		if !del {
			d.l.WithContext(ctx).Error(fmt.Sprintf("method:%s，del redis lock fail", methodName))
		}
	}()
	d.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start SendPushEvent4Msg", methodName))

	startTime := time.Now().Add(-time.Hour * 18)
	endTime := time.Now()
	startDate := startTime.Format("2006_01_02")
	endDate := endTime.Format("2006_01_02")
	var pts []string
	pts = append(pts, startDate)
	if startDate != endDate {
		pts = append(pts, endDate)
	}

	id := 0
	for _, pt := range pts {
		for {
			receivedMsgList, err := d.msgReceivedRepo.ListGtIdReceivedTime(ctx, pt, startTime.Unix(), endTime.Unix(), id, 1000)
			if err != nil {
				d.l.Errorf("mthod%s,ListGtIdReceivedTime error occurred，waId:%v", util.GetCurrentFuncName(), receivedMsgList)
				return
			}
			if len(receivedMsgList) == 0 {
				break
			}
			for _, receivedMsg := range receivedMsgList {
				id = receivedMsg.Id
				d.sendPushEvent4MsgByCoroutine(ctx, receivedMsg)
			}
			//drainageJobPool4.Wait()
		}
	}
}

func (d *DrainageJob) sendPushEvent4MsgByCoroutine(ctx context.Context, dto *WaMsgReceivedDTO) {
	methodName := util.GetCurrentFuncName()
	if !d.addUserId(ctx, dto.WaId) {
		d.l.WithContext(ctx).Warnf("mthod%s,waId repeated，waId:%v", methodName, dto)
		return
	}
	err := d.sendPushEvent4Msg(ctx, dto)
	if err != nil {
		d.l.Errorf("mthod%s,sendPushEvent4Msg error occurred, waId:%v", methodName, dto)
		return
	}
	//drainageJobPool4.Execute(func(param interface{}) {
	//	u, ok := param.(*WaMsgReceivedDTO) // 断言u是User类型
	//	if !ok {
	//		d.l.WithContext(ctx).Errorf("mthod%s,Assertion error occurred，waId:%v", methodName, u)
	//		return
	//	}
	//	if !d.addUserId(ctx, u.WaId) {
	//		d.l.WithContext(ctx).Warnf("mthod%s,waId repeated，waId:%v", methodName, u)
	//		return
	//	}
	//	err := d.sendPushEvent4Msg(ctx, u)
	//	if err != nil {
	//		d.l.Errorf("mthod%s,resendGoroutinePool The pool execution task starts error，waId:%v", methodName, u)
	//		return
	//	}
	//	d.l.WithContext(ctx).Infof("mthod%s,resendGoroutinePool The pool execution task starts，waId:%v", methodName, u)
	//}, dto)
}

func (d *DrainageJob) addUserId(ctx context.Context, waId string) bool {
	id, err := d.pushEvent4UserRepo.InsertIgnore(ctx, waId)
	if err != nil {
		d.l.Errorf("mthod%s,InsertIgnore error occurred，waId:%v", util.GetCurrentFuncName(), waId)
		return false
	}
	if id == 0 {
		return false
	}
	return true
}

func (d *DrainageJob) sendPushEvent4Msg(ctx context.Context, dto *WaMsgReceivedDTO) error {
	no := 4
	// no 0 菲律宾
	// no 1 印尼
	// no 2 马来
	// no 3 俄语
	// no 4 英语
	waId := strUtil.RemoveDirectionalFormatting(dto.WaId)
	if strings.HasPrefix(waId, "63") {
		no = 0
	} else if strings.HasPrefix(waId, "62") {
		no = 1
	} else if strings.HasPrefix(waId, "60") {
		no = 2
	} else if strings.HasPrefix(waId, "7") {
		no = 3
	} else if strings.HasPrefix(waId, "375") {
		no = 3
	} else if strings.HasPrefix(waId, "996") {
		no = 3
	} else if strings.HasPrefix(waId, "992") {
		no = 3
	} else if strings.HasPrefix(waId, "91") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "65") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "976") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "380") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "998") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "373") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "994") {
		return errors.New("禁推地区")
	} else if strings.HasPrefix(waId, "374") {
		return errors.New("禁推地区")
	}

	d.l.WithContext(ctx).Infof("mthod%s,sendPushEvent4Msg，waId:%v no:%v", util.GetCurrentFuncName(), waId, no)
	d.send4Msg(ctx, waId, no, 4)
	return nil
}

//func (d *DrainageJob) SAddWaId(waId string) bool {
//	addCount, err := d.redisService.SAddKey(util.GetCurrentFuncName(), constants.PushEvent4WaIdsKey, waId)
//	if err != nil {
//		return false
//	}
//	if addCount == 0 {
//		return false
//	}
//	return true
//}

func (d *DrainageJob) getUserInfoMap(ctx context.Context, waIds []string) (map[string]bool, error) {
	userInfos, err := d.userInfoRepo.FindUserInfos(ctx, waIds)
	if err != nil {
		return nil, err
	}
	userInfoMap := make(map[string]bool)
	for _, userInfo := range userInfos {
		userInfoMap[userInfo.WaID] = true
	}
	return userInfoMap, nil
}

func (d *DrainageJob) getUploadUserInfoMap(ctx context.Context, waIds []string) (map[string]bool, error) {
	uploadUserInfos, err := d.uploadUserInfoRepo.ListInNumber(ctx, waIds)
	if err != nil {
		return nil, err
	}
	uploadUserInfoMap := make(map[string]bool)
	for _, uploadUserInfo := range uploadUserInfos {
		uploadUserInfoMap[uploadUserInfo.PhoneNumber] = true
	}
	return uploadUserInfoMap, nil
}

func (d *DrainageJob) initCount(ctx context.Context, key string) {
	method := util.GetCurrentFuncName()
	count, err := d.countLimitRepo.Get(ctx, key)
	if err != nil && !errors.Is(err, sqlx.ErrNoRows) {
		d.l.Errorf("method[%s] get err: %v", method, err)
		return
	}
	if count > 0 {
		return
	}
	err = d.countLimitRepo.AddKey(ctx, key)
	if err != nil {
		d.l.Errorf("method[%s] addKey err: %v", method, err)
		return
	}
}

func (d *DrainageJob) validCount(ctx context.Context, key string, totalCount int) bool {
	method := util.GetCurrentFuncName()
	d.l.Info(fmt.Sprintf("%s validCount start", key))
	count, err := d.countLimitRepo.Get(ctx, key)
	if err != nil && !errors.Is(err, sqlx.ErrNoRows) {
		d.l.Errorf("method[%s] get err: %v", method, err)
		return false
	}
	d.l.Info(fmt.Sprintf("%s validCount end %v", key, count))
	return count < totalCount
}

func validCountryCode(countryCodes []string, waId string) bool {
	valid := false
	for _, prefix := range countryCodes {
		if strings.HasPrefix(waId, prefix) {
			valid = true
			break
		}
	}
	return valid
}

func (d *DrainageJob) updatePushEventState(ctx context.Context, waUserScore *WaUserScoreDTO, state int) {
	err := d.waUserScoreRepo.UpdateState(ctx, waUserScore.WaId, state)
	if err != nil {
		d.l.Errorf("mthod%s err: %v", util.GetCurrentFuncName(), err)
		return
	}
}

// 代码发送部分
func (d *DrainageJob) send13Msg(ctx context.Context, toNumber string, countryCode string, param string, version int) {
	pushJson := fmt.Sprintf(pushEvent13JsonMap[countryCode], toNumber, param)
	var body map[string]interface{}
	_ = json.NewEncoder().Decode([]byte(pushJson), &body)
	paramsBytes, _ := json.NewEncoder().Encode(body)
	paramsStr := string(paramsBytes)
	headers := d.getRequestHeader(paramsStr)
	d.sendNx(ctx, "send13Msg", body, headers, version)
}

func (d *DrainageJob) send2Msg(ctx context.Context, toNumber string, countryCode string, param string, version int) {
	pushJson := fmt.Sprintf(pushEvent2JsonMap2[countryCode], toNumber, param)
	d.l.WithContext(ctx).Infof(fmt.Sprintf("mthod%s,send2Msg，toNumber:%v bodyMap:%v", util.GetCurrentFuncName(), toNumber, pushJson))
	var body map[string]interface{}
	_ = json.NewEncoder().Decode([]byte(pushJson), &body)
	paramsBytes, _ := json.NewEncoder().Encode(body)
	paramsStr := string(paramsBytes)
	headers := d.getRequestHeader(paramsStr)
	d.sendNx(ctx, "send2Msg", body, headers, version)
}

func (d *DrainageJob) send4Msg(ctx context.Context, toNumber string, no int, version int) {
	pushJson := fmt.Sprintf(pushEvent4JsonList[no], strings.TrimSpace(toNumber))
	var body map[string]interface{}
	_ = json.NewEncoder().Decode([]byte(pushJson), &body)
	paramsBytes, _ := json.NewEncoder().Encode(body)
	paramsStr := string(paramsBytes)
	headers := d.getRequestHeader(paramsStr)
	d.sendNx(ctx, "send4Msg", body, headers, version)
}

func (d *DrainageJob) sendNx(ctx context.Context, origin string, bodyMap any, headers map[string]string, version int) {
	resNx := &response.NXResponse{}
	paramsBytes, _ := json.NewEncoder().Encode(bodyMap)
	paramsStr := string(paramsBytes)
	for i := 1; i < 4; i++ {
		d.l.WithContext(ctx).Infof("%s,sendNx，bodyMap:%v", origin, paramsStr)
		res, nxErr := rest.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", bodyMap, headers, 10*1000*time.Second, 10*1000*time.Second)
		if nxErr != nil {
			continue
		}
		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
		if nxErr != nil {
			continue
		}
		if 0 != resNx.Code {
			continue
		}
		break
	}
	if resNx.Data == nil || len(resNx.Data.Messages) == 0 {
		return
	}
	waMsgID := resNx.Data.Messages[0].Id
	d.l.WithContext(ctx).Infof(fmt.Sprintf("%s,sendNx，waMsgID:%v", origin, waMsgID))
	err := d.pushEventSendMsgRepo.InsertIgnore(ctx, &PushEventSendMessageDTO{
		MessageId: waMsgID,
		Version:   version,
	})
	if err != nil {
		d.l.Errorf("mthod%s err: %v", util.GetCurrentFuncName(), err)
	}
}

func (d *DrainageJob) getRequestHeader(paramsStr string) map[string]string {
	commonHeaders := map[string]string{
		"accessKey": d.configInfo.Data.Nx.Ak,
		"ts":        strconv.FormatInt(time.Now().UnixMilli(), 10),
		"bizType":   "2",
		"action":    "mt",
	}
	sign := util.CallSign(commonHeaders, paramsStr, d.configInfo.Data.Nx.Sk)
	commonHeaders["sign"] = sign
	return commonHeaders
}
