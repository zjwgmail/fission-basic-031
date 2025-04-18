package biz

import (
	"context"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/feishu"
	"fission-basic/internal/pkg/redis"
	"fission-basic/util"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"math"
	"strconv"
	"strings"
	"time"
)

type FeishuReportJob struct {
	systemConfigRepo            SystemConfigRepo
	userInfoRepo                UserInfoRepo
	msgSendRepo                 WaMsgSendRepo
	feishuReportRepo            FeishuReportRepo
	helpCodeRepo                HelpCodeRepo
	feishuRest                  *feishu.Feishu
	redisService                *redis.RedisService
	bizConf                     *conf.Business
	l                           *log.Helper
	taskLock                    string
	feishuUserInfoId            string
	feishuFirstGenerationCount  string
	feishuOtherGenerationCount  string
	feishuMsgSendId             string
	feishuMsgSendTotalCount     string
	feishuMsgSendFailedCount    string
	feishuMsgSendTimeoutCount   string
	feishuMsgSendInterceptCount string
}

func NewFeishuReportJob(
	systemConfigRepo SystemConfigRepo,
	userInfoRepo UserInfoRepo,
	msgSendRepo WaMsgSendRepo,
	feishuReportRepo FeishuReportRepo,
	helpCodeRepo HelpCodeRepo,
	feishuRest *feishu.Feishu,
	redisService *redis.RedisService,
	l log.Logger,
	bizConf *conf.Business) *FeishuReportJob {
	return &FeishuReportJob{
		systemConfigRepo:            systemConfigRepo,
		userInfoRepo:                userInfoRepo,
		msgSendRepo:                 msgSendRepo,
		feishuReportRepo:            feishuReportRepo,
		helpCodeRepo:                helpCodeRepo,
		feishuRest:                  feishuRest,
		redisService:                redisService,
		bizConf:                     bizConf,
		l:                           log.NewHelper(l),
		taskLock:                    "feishu_lock",
		feishuUserInfoId:            "feishuUserInfoId",
		feishuFirstGenerationCount:  "feishuFirstGenerationCount",
		feishuOtherGenerationCount:  "feishuOtherGenerationCount",
		feishuMsgSendId:             "feishuMsgSendId",
		feishuMsgSendFailedCount:    "feishuMsgSendFailedCount",
		feishuMsgSendTimeoutCount:   "feishuMsgSendTimeoutCount",
		feishuMsgSendInterceptCount: "feishuMsgSendInterceptCount",
	}
}

func (f *FeishuReportJob) SendReport(ctx context.Context) {
	methodName := util.GetCurrentFuncName()
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			f.l.Errorf("method[%s]，panic", methodName)
			return
		}
	}()

	f.l.Infof("method[%s], start send feishuRest report", methodName)

	redisService := f.redisService
	taskLock := f.taskLock

	getLock, err := redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		f.l.Error(fmt.Sprintf("method[%s],call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		f.l.Error(fmt.Sprintf("method[%s],get reids lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := redisService.Del(taskLock)
		if !del {
			f.l.Error(fmt.Sprintf("method[%s]，del redis lock fail", methodName))
		}
	}()

	message, feishuReportParam, err := f.buildMessage(ctx)
	if err != nil {
		f.l.Errorf("method[%s],build feishuRest report failed，err:%v", methodName, err)
		return
	}

	_, err = f.feishuReportRepo.AddFeishuReport(ctx, feishuReportParam)
	if err != nil {
		f.l.Warnf("method[%s],insert feishuRest report failed，err:%v", methodName, err)
		return
	}

	var res string
	for i := 1; i < 4; i++ {
		err = f.feishuRest.SendTextMsg(ctx, message)
		if err != nil {
			f.l.Warnf("方法[%s],调用飞书接口失败，message:%v,err:%v", methodName, message, err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		res = fmt.Sprintf("发送飞书失败，err：%v", err)
		return
	}

	f.l.Infof("方法[%s],发送飞书消息执行完成，res:%v，message:%v", methodName, res, message)
	return
}

func (f *FeishuReportJob) buildMessage(ctx context.Context) (string, *FeishuReportParam, error) {
	methodName := util.GetCurrentFuncName()
	now := time.Now()
	// 获取月日，格式：x月x日
	monthDay := fmt.Sprintf("%d月%d日", now.Month(), now.Day())
	// 获取小时，格式：xx:00
	hour := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	fmt.Println(monthDay, hour)

	firstCount, fissionCount, err := f.getGenerationCount(ctx)
	if err != nil {
		f.l.Errorf("方法[%s],获取用户总数失败，err:%v", methodName, err)
		return "", nil, err
	}

	// 飞书
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s 更新：\n", monthDay))
	coverCount := firstCount + fissionCount*3

	builder.WriteString(fmt.Sprintf("初代种子引入：%v\n总裂变人数：%v\n预计覆盖人数：%v\n",
		util.AddThousandSeparators(firstCount),
		util.AddThousandSeparators(fissionCount),
		util.AddThousandSeparators(coverCount)))

	cdkTypeList := []int{0, 3, 6, 9, 12, 15}
	cdkCountMap := make(map[int]int)
	for _, cdkType := range cdkTypeList {
		totalCount, currentCount, err := f.getGrantCount(ctx, cdkType)
		if err != nil {
			f.l.Errorf("方法[%s],获取%v发放量失败，err:%v", methodName, cdkType, err)
			return "", nil, err
		}
		cdkCount := int64(totalCount)
		cdkNotUsedLen := int64(currentCount)
		percent := math.Round(float64(cdkCount-cdkNotUsedLen)/float64(cdkCount)*10000) / 100
		switch cdkType {
		case 0:
			builder.WriteString(fmt.Sprintf("免费奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case 3:
			builder.WriteString(fmt.Sprintf("第一档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case 6:
			builder.WriteString(fmt.Sprintf("第二档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case 9:
			builder.WriteString(fmt.Sprintf("第三档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case 12:
			builder.WriteString(fmt.Sprintf("第四档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case 15:
			builder.WriteString(fmt.Sprintf("第五档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		}
		cdkCountMap[cdkType] = int(cdkCount - cdkNotUsedLen)
	}

	builder.WriteString(fmt.Sprintf("缓冲期阈值：%v%%\n", (1-f.bizConf.Cdk.AlarmThreshold)*100))

	failedCount, timeoutCount, interceptCount, totalCount, err := f.getLastMsgSendCount(ctx)
	if err != nil {
		f.l.Errorf("方法[%s],获取发送失败，发送超时，非白拦截失败，err:%v", methodName, err)
		return "", nil, err
	}

	msgCount := int(totalCount)
	sendFailMsgCount := int64(failedCount)
	sendTimeOutMsgCount := int64(timeoutCount)
	notWhiteCount := int64(interceptCount)
	percent := 0.0
	if msgCount > 0 {
		percent = math.Round(float64(sendFailMsgCount)/float64(msgCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("发送失败：%v条;%v%%\n", util.AddThousandSeparators64(sendFailMsgCount), percent))

	percent = 0
	if msgCount > 0 {
		percent = math.Round(float64(sendTimeOutMsgCount)/float64(msgCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("发送超时：%v条;%v%%\n", util.AddThousandSeparators64(sendTimeOutMsgCount), percent))

	//查询已经参与活动的人
	userCount := firstCount + fissionCount
	percent = 0

	if msgCount > 0 {
		percent = math.Round(float64(notWhiteCount)/float64(userCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("非白拦截：%v条;%v%%", util.AddThousandSeparators64(notWhiteCount), percent))

	cdkCountList := make([]string, 0, len(cdkCountMap))
	for _, cdkCount := range cdkCountMap {
		cdkCountList = append(cdkCountList, strconv.Itoa(cdkCount))
	}

	cdkCount := strings.Join(cdkCountList, ",")

	return builder.String(), &FeishuReportParam{
		Date:           monthDay,
		Time:           hour,
		FirstCount:     firstCount,
		FissionCount:   fissionCount,
		CdkCount:       cdkCount,
		CoverCount:     coverCount,
		FailedCount:    failedCount,
		TimeoutCount:   timeoutCount,
		InterceptCount: interceptCount,
	}, nil
}

func (f *FeishuReportJob) getGenerationCount(ctx context.Context) (int, int, error) {
	minId, err := f.getIntValueByKey(ctx, f.feishuUserInfoId)
	if err != nil {
		return 0, 0, err
	}
	id := int64(minId)
	firstCount, fissionCount, err := f.getLastGenerationCount(ctx)
	// 循环查询统计数量
	for {
		userInfos, _ := f.userInfoRepo.ListGtIdLtEndTime(ctx, id, time.Now(), 1000)
		if len(userInfos) == 0 {
			break
		}
		for _, userInfo := range userInfos {
			id++
			if userInfo.Generation == 1 {
				firstCount++
			} else {
				fissionCount++
			}
		}
	}
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuUserInfoId,
		Value: strconv.Itoa(int(id)),
	})
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuFirstGenerationCount,
		Value: strconv.Itoa(firstCount),
	})
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuOtherGenerationCount,
		Value: strconv.Itoa(fissionCount),
	})
	return firstCount, fissionCount, nil
}

// 获取值
func (f *FeishuReportJob) getValueByKey(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := f.systemConfigRepo.GetByKey(ctx, key)
	if err != nil {
		f.l.Errorf("方法[%s],获取%s失败，err:%v", "getValueByKey", key, err)
	}
	if value == "" {
		value = defaultValue
		err := f.systemConfigRepo.AddOne(ctx, &SystemConfigParam{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return "", err
		}
	}
	return defaultValue, nil
}

// 获取int值
func (f *FeishuReportJob) getIntValueByKey(ctx context.Context, key string) (int, error) {
	value, err := f.getValueByKey(ctx, key, "0")
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return intValue, nil
}

func (f *FeishuReportJob) getHelpCodeCount(ctx context.Context) (int, int, error) {
	maxId, err := f.helpCodeRepo.GetMaxId(ctx)
	if err != nil {
		return 0, 0, err
	}
	totalCount := int(maxId)
	currentCount, err := f.redisService.GetQueueSize(ctx, constants.HelpCodeKey)
	if err != nil {
		return 0, 0, err
	}
	return totalCount, currentCount, nil
}

func (f *FeishuReportJob) getLastGenerationCount(ctx context.Context) (int, int, error) {
	firstCount, err := f.getIntValueByKey(ctx, f.feishuFirstGenerationCount)
	if err != nil {
		return 0, 0, err
	}
	otherCount, err := f.getIntValueByKey(ctx, f.feishuOtherGenerationCount)
	if err != nil {
		return 0, 0, err
	}
	return firstCount, otherCount, nil
}

func (f *FeishuReportJob) getGrantCount(ctx context.Context, cdkType int) (int, int, error) {
	countKey := constants.CdkQueueKeyPrefix + strconv.Itoa(cdkType) + constants.CdkTotalCountKeySuffix
	totalCountStr := f.redisService.Get(countKey)
	totalCount, err := strconv.Atoi(totalCountStr)
	if err != nil {
		return 0, 0, err
	}
	currentCount, err := f.redisService.GetQueueSize(ctx, constants.CdkQueueKeyPrefix+strconv.Itoa(cdkType))
	if err != nil {
		return 0, 0, err
	}
	return totalCount, currentCount, nil
}

func (f *FeishuReportJob) getLastMsgSendCount(ctx context.Context) (int, int, int, int64, error) {
	minId, err := f.getIntValueByKey(ctx, f.feishuMsgSendId)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	id := int64(minId)
	failedCount, err := f.getIntValueByKey(ctx, f.feishuMsgSendFailedCount)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	timeoutCount, err := f.getIntValueByKey(ctx, f.feishuMsgSendTimeoutCount)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	interceptCount, err := f.getIntValueByKey(ctx, f.feishuMsgSendInterceptCount)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	// 循环查询统计数量
	for {
		msgSendList, err := f.msgSendRepo.ListGtId(ctx, id, 1000)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		if len(msgSendList) == 0 {
			break
		}
		for _, msgSend := range msgSendList {
			id++
			if msgSend.MsgType == constants.CannotAttendActivityMsg {
				interceptCount++
			} else if msgSend.State == constants.MsgSendStateFail {
				failedCount++
			} else if msgSend.State == constants.MsgSendStateNxFail {
				failedCount++
			} else if msgSend.State == constants.MsgSendStateNxTimeout {
				timeoutCount++
			}
		}
	}
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuMsgSendId,
		Value: strconv.Itoa(int(id)),
	})
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuMsgSendFailedCount,
		Value: strconv.Itoa(failedCount),
	})
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuMsgSendTimeoutCount,
		Value: strconv.Itoa(timeoutCount),
	})
	_ = f.systemConfigRepo.UpdateByKey(ctx, &SystemConfigParam{
		Key:   f.feishuMsgSendInterceptCount,
		Value: strconv.Itoa(interceptCount),
	})
	return failedCount, timeoutCount, interceptCount, id, nil
}

func (f *FeishuReportJob) Test(ctx context.Context) {

}
