package service

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/feishu"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/rest"
	iutil "fission-basic/internal/util"
	"fission-basic/internal/util/encoder/rsa"
	"fission-basic/kit/sqlx"
	"fission-basic/util"
	"net/url"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

type HelpCodeService struct {
	feishuReportJob *biz.FeishuReportJob
	helpCodeUsecase *biz.HelpCodeUsecase
	feishuRest      *feishu.Feishu
	bizConf         *conf.Business
	redisService    *redis.RedisService
	shortLink       *rest.ShortLink
	publicKey       string
	l               *log.Helper

	aiUc *biz.ActivityInfoUsecase
}

func NewHelpCodeService(
	d *conf.Data,
	c *conf.Business,
	l log.Logger,
	feishuReportJob *biz.FeishuReportJob,
	hcBizService *biz.HelpCodeUsecase,
	feishuRest *feishu.Feishu,
	aiUc *biz.ActivityInfoUsecase,
	redisService *redis.RedisService) *HelpCodeService {
	return &HelpCodeService{
		feishuReportJob: feishuReportJob,
		helpCodeUsecase: hcBizService,
		feishuRest:      feishuRest,
		bizConf:         c,
		redisService:    redisService,
		shortLink:       &rest.ShortLink{},
		publicKey:       d.Rsa.PublicKey,
		l:               log.NewHelper(l),
		aiUc:            aiUc,
	}
}

func (n *HelpCodeService) GetActivityInfo(ctx context.Context, req *v1.GetActivityInfoRequest) (*v1.GetActivityInfoResponse, error) {
	cost := iutil.MethodCost(ctx, n.l, "NxCloudService.GetActivityInfo")
	defer cost()

	dto, err := n.aiUc.GetActivityInfo(ctx)
	if err != nil {
		return nil, err
	}

	n.l.WithContext(ctx).Infof("GetActivityInfo-gw req=%+v, dto=%+v", req, dto)
	return &v1.GetActivityInfoResponse{}, nil
}

// PreheatHelpCode 预热助力码
func (hc *HelpCodeService) PreheatHelpCode(ctx context.Context, req *v1.PreheatHelpCodeRequest) (*v1.PreheatHelpCodeResponse, error) {
	//count := req.Count
	//methodName := util.GetCurrentFuncName()
	//if count == 0 {
	//	return nil, fmt.Errorf("count is empty")
	//}
	//if req.CleanRedis {
	//	hc.redisService.Del(constants.HelpCodeKey)
	//}
	//ctxTmp := context.WithoutCancel(ctx)
	//// defer 异常处理
	//defer func() {
	//	if e := recover(); e != nil {
	//		return
	//	}
	//}()
	//var shortLinkVersions []int
	//for i := 0; i < int(hc.bizConf.ShortLink.Count); i++ {
	//	shortLinkVersions = append(shortLinkVersions, i)
	//}
	//_ = hc.feishuRest.SendTextMsg(ctx, fmt.Sprintf("预热开始 %d条", count))
	//startTime := time.Now()
	//// 第一层是为了不阻断
	//go func() {
	//	coroutineCount := int(hc.bizConf.HelpCode.CoroutineCount)
	//	tasks := make(chan func(ctx context.Context) error, coroutineCount)
	//	go func() {
	//		defer close(tasks)
	//		for i := 0; i < int(count); i++ {
	//			tasks <- func(ctx context.Context) error {
	//				id, err := hc.addHelpCodeOrShortLink(ctxTmp, "", shortLinkVersions)
	//				if err != nil {
	//					hc.l.WithContext(ctxTmp).Errorf("methodName:%s addHelpCodeOrShortLink i %d err:%v", methodName, i, err)
	//					_ = hc.feishuRest.SendTextMsg(ctxTmp, fmt.Sprintf("助力码预热出错, id: %d", id))
	//				}
	//				return nil
	//			}
	//		}
	//	}()
	//
	//	err := goroutine_pool.ParallN(ctxTmp, coroutineCount, tasks)
	//	if err != nil {
	//		hc.l.WithContext(ctxTmp).Errorf("methodName:%s ParallN failed %v", methodName, err)
	//	}
	//	_ = hc.feishuRest.SendTextMsg(ctx, fmt.Sprintf("预热完成 %d条, 任务耗时: %d秒", count, int(time.Now().Sub(startTime).Seconds())))
	//}()

	return &v1.PreheatHelpCodeResponse{}, nil
}

func (hc *HelpCodeService) RepairHelpCode(ctx1 context.Context, req *v1.RepairHelpCodeRequest) (*v1.RepairHelpCodeResponse, error) {
	_ = hc.feishuRest.SendTextMsg(ctx1, "助力码开始放入缓存")
	//methodName := util.GetCurrentFuncName()
	//ctx := context.WithoutCancel(ctx1)

	go func() {

		//coroutineCount := 1
		//limit := uint(10000)
		//tasks := make(chan func(ctx context.Context) error, coroutineCount)
		//go func() {
		//	minId := int64(0)
		//	// 分页从数据库中查询所有助力码一次一万条
		//	for {
		//		entityList, err2 := hc.helpCodeUsecase.ListGtIdLtEndTime(ctxTmp, minId, limit)
		//		if err2 != nil {
		//			hc.l.WithContext(ctxTmp).Errorf("methodName:%ss PageByField failed %v", methodName, err2)
		//			return
		//		}
		//		if len(entityList) == 0 {
		//			break
		//		}
		//		for _, codeModel := range entityList {
		//			if codeModel.HelpCode != "" && codeModel.ShortLinkV0 != "" && codeModel.ShortLinkV1 != "" &&
		//				codeModel.ShortLinkV2 != "" && codeModel.ShortLinkV3 != "" && codeModel.ShortLinkV4 != "" &&
		//				codeModel.ShortLinkV5 != "" {
		//				continue
		//			} else {
		//				hc.l.WithContext(ctxTmp).Infof("methodName:%s need repairData %v", methodName, codeModel)
		//				err := hc.repairData2(ctxTmp, *codeModel)
		//				if err != nil {
		//					hc.l.WithContext(ctxTmp).Errorf("methodName:%s repairData failed %v", methodName, err)
		//				}
		//			}
		//		}
		//		minId = entityList[len(entityList)-1].Id
		//	}
		//
		//	hc.redisService.Del(constants.HelpCodeKey)
		//	hc.l.WithContext(ctxTmp).Infof("methodName:%s del helpCodeKey", methodName)
		//	minId2 := int64(0)
		//	// 分页从数据库中查询所有助力码一次一万条
		//	for {
		//		entitys, err2 := hc.helpCodeUsecase.ListGtIdLtEndTime(ctxTmp, minId2, limit)
		//		if err2 != nil {
		//			hc.l.WithContext(ctxTmp).Errorf("methodName:%s PageByField failed %v", methodName, err2)
		//			return
		//		}
		//		if len(entitys) == 0 {
		//			size, err2 := hc.redisService.GetQueueSize(ctxTmp, constants.HelpCodeKey)
		//			if err2 != nil {
		//				hc.l.WithContext(ctxTmp).Errorf("methodName:%s GetQueueSize failed %v", methodName, err2)
		//				return
		//			}
		//			hc.l.WithContext(ctxTmp).Infof("methodName:%s repairData2 end", methodName)
		//			_ = hc.feishuRest.SendTextMsg(ctxTmp, fmt.Sprintf("修复完成, 任务耗时: %d秒 预热助力码队列总长度 %d", int(time.Now().Sub(startTime).Seconds()), size))
		//			break
		//		}
		//		//渠道modes里面所有的helpCode
		//		var helpCodes []string
		//		for _, codeModel := range entitys {
		//			if codeModel.HelpCode != "" && codeModel.ShortLinkV0 != "" && codeModel.ShortLinkV1 != "" &&
		//				codeModel.ShortLinkV2 != "" && codeModel.ShortLinkV3 != "" && codeModel.ShortLinkV4 != "" &&
		//				codeModel.ShortLinkV5 != "" {
		//				helpCodes = append(helpCodes, codeModel.HelpCode)
		//			} else {
		//				hc.l.WithContext(ctxTmp).Errorf("methodName:%s  need repairData2  %v", methodName, codeModel)
		//			}
		//		}
		//		if len(helpCodes) == 0 {
		//			hc.l.WithContext(ctxTmp).Infof("methodName:%s current not need add redis")
		//			continue
		//		}
		//		err := hc.redisService.CreateQueueAndBatchPut(ctxTmp, constants.HelpCodeKey, helpCodes...)
		//		if err != nil {
		//			hc.l.WithContext(ctxTmp).Errorf("methodName:%s CreateQueueAndBatchPut failed %v", methodName, err)
		//		} else {
		//			hc.l.WithContext(ctxTmp).Infof("methodName:%s CreateQueueAndBatchPut success size %v", methodName, len(helpCodes))
		//		}
		//		minId2 = entitys[len(entitys)-1].Id
		//	}
		//
		//}()

		//minId2 := int64(0)
		//maxLimit := int64(20000000) // 最多放入 Redis 的数据量
		//currentCount := int64(0)    // 当前已放入 Redis 的数据量
		//
		//// 分页从数据库中查询所有助力码一次一万条
		//startTime := time.Now()
		//for {
		//	entitys, err2 := hc.helpCodeUsecase.ListGtIdLtEndTime(ctx, minId2, 10000)
		//	if err2 != nil {
		//		hc.l.WithContext(ctx).Errorf("methodName:%s ListGtIdLtEndTime failed %v", methodName, err2)
		//		break
		//	}
		//	if len(entitys) == 0 {
		//		size, err2 := hc.redisService.GetQueueSize(ctx, constants.HelpCodeKey)
		//		if err2 != nil {
		//			hc.l.WithContext(ctx).Errorf("methodName:%s GetQueueSize failed %v", methodName, err2)
		//			break
		//		}
		//		hc.l.WithContext(ctx).Infof("methodName:%s repairData2 end")
		//		_ = hc.feishuRest.SendTextMsg(ctx, fmt.Sprintf("初始化helpCode, 任务耗时: %d秒 预热助力码队列总长度 %d", int(time.Now().Sub(startTime).Seconds()), size))
		//		break
		//	}
		//
		//	// 提取满足条件的 helpCode
		//	var helpCodes []string
		//	for _, codeModel := range entitys {
		//		if codeModel.HelpCode != "" && codeModel.ShortLinkV0 != "" && codeModel.ShortLinkV1 != "" &&
		//			codeModel.ShortLinkV2 != "" && codeModel.ShortLinkV3 != "" && codeModel.ShortLinkV4 != "" &&
		//			codeModel.ShortLinkV5 != "" {
		//			helpCodes = append(helpCodes, codeModel.HelpCode)
		//		} else {
		//			hc.l.WithContext(ctx).Errorf("methodName:%s need repairData2 %v", methodName, codeModel)
		//		}
		//	}
		//
		//	if len(helpCodes) == 0 {
		//		hc.l.WithContext(ctx).Infof("methodName:%s current not need add redis", methodName)
		//		continue
		//	}
		//
		//	// 检查是否达到最大限制
		//	if currentCount+int64(len(helpCodes)) > maxLimit {
		//		helpCodes = helpCodes[:maxLimit-currentCount] // 裁剪超出部分
		//	}
		//
		//	err := hc.redisService.CreateQueueAndBatchPut(ctx, constants.HelpCodeKey, helpCodes...)
		//	if err != nil {
		//		hc.l.WithContext(ctx).Errorf("methodName:%s CreateQueueAndBatchPut failed %v", methodName, err)
		//	} else {
		//		hc.l.WithContext(ctx).Infof("methodName:%s CreateQueueAndBatchPut  process size %v", methodName, currentCount)
		//		currentCount += int64(len(helpCodes)) // 更新已放入 Redis 的数据量
		//	}
		//
		//	// 检查是否达到最大限制
		//	if currentCount >= maxLimit {
		//		hc.l.WithContext(ctx).Infof("methodName:%s 已达到最大限制，停止处理", methodName)
		//		break
		//	}
		//
		//	minId2 = entitys[len(entitys)-1].Id
		//}
		//
		//// 检查队列的第一个和最后一个元素
		//if currentCount >= maxLimit {
		//	firstElement, err := hc.redisService.LIndex(ctx, constants.HelpCodeKey, 0)
		//	lastElement, err := hc.redisService.LIndex(ctx, constants.HelpCodeKey, -1)
		//	if err != nil {
		//		hc.l.WithContext(ctx).Errorf("methodName:%s 获取队列首尾元素失败: %v", methodName, err)
		//	} else {
		//		message := fmt.Sprintf("初始化helpCode完成，任务耗时: %d秒，队列总长度: %d\n队列第一个元素: %s\n队列最后一个元素: %s",
		//			int(time.Now().Sub(startTime).Seconds()), currentCount, firstElement, lastElement)
		//		_ = hc.feishuRest.SendTextMsg(ctx, message)
		//	}
		//}
		//err := goroutine_pool.ParallN(ctxTmp, coroutineCount, tasks)
		//if err != nil {
		//	hc.l.WithContext(ctxTmp).Errorf("methodName:%s ParallN failed %v", methodName, err)
		//}
	}()

	return &v1.RepairHelpCodeResponse{}, nil
}

// GetHelpCode 获取助力码 如果获取不到则实时生成
func (hc *HelpCodeService) GetHelpCode(ctx context.Context) (string, error) {
	methodName := util.GetCurrentFuncName()
	helpCode, _ := hc.redisService.PopQueueData(ctx, constants.HelpCodeKey)
	if helpCode != "" {
		hc.l.WithContext(ctx).Infof("methodName:%s PopQueueData success queueName:%v helpCode: %v", methodName, constants.HelpCodeKey, helpCode)
		return helpCode, nil
	}
	//无法从redis获取，立刻生成并返回
	helpCode, _, err := hc.helpCodeUsecase.Add(ctx)
	if err != nil {
		hc.l.WithContext(ctx).Errorf("methodName:%s add help code failed %v", methodName, err)
		return "", err
	}
	return helpCode, nil
}

// GetShortLinkByHelpCode 根据助力码获取短链 如果获取不到则实时生成
func (hc *HelpCodeService) GetShortLinkByHelpCode(ctx context.Context, helpCode string, shortLinkVersion int) (string, error) {
	methodName := util.GetCurrentFuncName()
	shortLink, err := hc.helpCodeUsecase.GetShortLinkByHelpCodeAndVersion(ctx, helpCode, shortLinkVersion)
	if err != nil {
		if !errors.Is(err, sqlx.ErrNoRows) {
			hc.l.WithContext(ctx).Errorf("methodName:%s get short link by help code helpCode %v shortLinkVersion %v err %v", methodName, helpCode, shortLinkVersion, err)
			return "", err
		}
	}
	if shortLink != "" {
		hc.l.WithContext(ctx).Infof("methodName:%s get short link from db success helpCode %v shortLinkVersion %v shortLink %v", methodName, helpCode, shortLinkVersion, shortLink)
		return shortLink, nil
	}

	// 没有获取到shortLink 直接生成
	longUrl, err := hc.getLongUrlByHelpCode(helpCode, shortLinkVersion)
	shortLink, err = hc.addShortLink(ctx, longUrl, shortLinkVersion, helpCode)
	if err != nil {
		hc.l.WithContext(ctx).Errorf("methodName:%s add short link failed helpCode %v shortLinkVersion %v err %v", methodName, helpCode, shortLinkVersion, err)
		return "", err
	}
	return shortLink, nil
}

// 新增助力码/短链
func (hc *HelpCodeService) addHelpCodeOrShortLink(ctx context.Context, helpCode string, shortLinkVersions []int) (int64, error) {
	/*创建助力码*/
	methodName := util.GetCurrentFuncName()
	id := int64(0)
	if helpCode == "" {
		helpCodeTmp, idTmp, err := hc.helpCodeUsecase.Add(ctx)
		if err != nil {
			hc.l.WithContext(ctx).Errorf("methodName:%s add help code failed %v", methodName, err)
			return idTmp, err
		}
		helpCode = helpCodeTmp
		id = idTmp
	}
	err := hc.redisService.CreateQueueAndPut(ctx, constants.HelpCodeKey, helpCode)
	if err != nil {
		hc.l.WithContext(ctx).Errorf("methodName:%s CreateQueueAndPut add helpCode %v failed %v", methodName, helpCode, err)
		return id, err
	}
	/*生成短链*/
	for _, shortLinkVersion := range shortLinkVersions {
		longUrl, err := hc.getLongUrlByHelpCode(helpCode, shortLinkVersion)
		if err != nil {
			hc.l.WithContext(ctx).Errorf("methodName:%s get long url by help code helpCode %v shortLinkVersion %v err %v", methodName, helpCode, shortLinkVersion, err)
			return id, err
		}
		_, err = hc.addShortLink(ctx, longUrl, shortLinkVersion, helpCode)
		if err != nil {
			hc.l.WithContext(ctx).Errorf("methodName:%s add short link failed helpCode %v shortLinkVersion %v err %v", methodName, helpCode, shortLinkVersion, err)
			return id, err
		}
	}
	hc.l.WithContext(ctx).Infof("methodName:%s add help code success helpCode %v shortLinkVersion %v", methodName, helpCode, shortLinkVersions)
	return id, nil
}

// 远程生成短链并更新至db
func (hc *HelpCodeService) addShortLink(ctx context.Context, longUrl string, shortLinkVersion int, helpCode string) (string, error) {
	shortLink, err := hc.shortLink.GetShortUrlByUrl(ctx, longUrl, hc.bizConf)
	if err != nil {
		log.Errorf("[addShortLink] GetShortUrlByUrl failed %v", err)
		return "", err
	}
	err = hc.helpCodeUsecase.UpdateShortLinkByHelpCode(ctx, shortLinkVersion, shortLink, helpCode)
	if err != nil {
		log.Errorf("[addShortLink] UpdateShortLinkByHelpCode err helpCode %v shortLinkVersion %v err %v", helpCode, shortLinkVersion, err)
		return "", err
	}
	return shortLink, nil
}

// 拼接短链
func (hc *HelpCodeService) getLongUrlByHelpCode(helpCode string, shortLinkVersion int) (string, error) {
	encryptedCode, err := rsa.Encrypt(helpCode, hc.publicKey)
	if err != nil {
		return "", err
	}
	urlCode := url.QueryEscape(encryptedCode)
	return strings.ReplaceAll(hc.bizConf.ShortLink.BaseUrls[shortLinkVersion], "{code}", urlCode), nil
}

func (hc *HelpCodeService) repairData(ctx context.Context, id int64) error {
	helpCode, shortLinkMap, _ := hc.helpCodeUsecase.GetDataById(ctx, id)
	if helpCode == "" {
		//删除记录
		_ = hc.helpCodeUsecase.DeleteById(ctx, id)
		return nil
	}
	var shortLinkVersions []int
	for i := 0; i < int(hc.bizConf.ShortLink.Count); i++ {
		if shortLinkMap[i] != "" {
			continue
		}
		shortLinkVersions = append(shortLinkVersions, i)
	}
	_, _ = hc.addHelpCodeOrShortLink(ctx, helpCode, shortLinkVersions)
	return nil
}

func (hc *HelpCodeService) repairData2(ctx context.Context, entity biz.HelpCode) error {
	if entity.HelpCode == "" {
		//删除记录
		hc.l.Errorf("repairData2 entity  %v is empty", entity)
		_ = hc.helpCodeUsecase.DeleteById(ctx, entity.Id)
		return nil
	}
	var shortLinkVersions []int

	shortLinkMap := map[int]string{
		0: entity.ShortLinkV0,
		1: entity.ShortLinkV1,
		2: entity.ShortLinkV2,
		3: entity.ShortLinkV3,
		4: entity.ShortLinkV4,
		5: entity.ShortLinkV5,
	}
	for i := 0; i < int(hc.bizConf.ShortLink.Count); i++ {
		if shortLinkMap[i] != "" {
			continue
		}
		hc.l.Infof("repairData2 entity %v shortLinkVersions %v", entity, i)
		shortLinkVersions = append(shortLinkVersions, i)
	}
	_, _ = hc.addHelpCodeOrShortLink(ctx, entity.HelpCode, shortLinkVersions)
	return nil
}

func (hc *HelpCodeService) HCTest(ctx context.Context, req *v1.HCTestRequest) (*v1.HCTestResponse, error) {
	return &v1.HCTestResponse{}, nil
}
