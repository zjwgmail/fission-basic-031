package biz

import (
	"bytes"
	"context"
	"fission-basic/api/constants"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/util"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/xuri/excelize/v2"
	"gopkg.in/gomail.v2"
)

type BasicInfo struct {
	Language    string
	Channel     string
	CountryCode string
}
type DataByUserInfo struct {
	GenerationCount [7]int
	TotalJoinCount  [16]int
}
type DataByUserJoinGroup struct {
	DailyJoinCount [16]int
}
type MsgData struct {
	CountV3        int
	CountV22       int
	CountV36       int
	SuccessCount   int
	FailedCount    int
	TimeoutCount   int
	InterceptCount int
}
type TimeDTO struct {
	StartTime time.Time
	EndTime   time.Time
	Date      string
}

type EmailReportJob struct {
	systemConfigRepo  SystemConfigRepo
	msgSendRepo       WaMsgSendRepo
	userInfoRepo      UserInfoRepo
	userJoinGroupRepo UserJoinGroupRepo
	emailReportRepo   EmailReportRepo
	redisService      *redis.RedisService
	dataConf          *conf.Data
	bizConf           *conf.Business
	l                 *log.Helper
	limit             int
}

type ReportJsonDTO struct {
	Date            string
	Language        string
	Channel         string
	CountryCode     string
	GenerationCount [7]int
	DailyJoinCount  [16]int
	TotalJoinCount  [16]int
	CountV3         int
	CountV22        int
	CountV36        int
	SuccessCount    int
	FailedCount     int
	TimeoutCount    int
	InterceptCount  int
}

func NewEmailReportJob(
	systemConfigRepo SystemConfigRepo,
	msgSendRepo WaMsgSendRepo,
	userInfoRepo UserInfoRepo,
	userJoinGroupRepo UserJoinGroupRepo,
	emailReportRepo EmailReportRepo,
	redisService *redis.RedisService,
	l log.Logger,
	bizConf *conf.Business,
	dataConf *conf.Data) *EmailReportJob {
	return &EmailReportJob{
		systemConfigRepo:  systemConfigRepo,
		msgSendRepo:       msgSendRepo,
		userInfoRepo:      userInfoRepo,
		userJoinGroupRepo: userJoinGroupRepo,
		emailReportRepo:   emailReportRepo,
		redisService:      redisService,
		bizConf:           bizConf,
		dataConf:          dataConf,
		l:                 log.NewHelper(l),
		limit:             1000,
	}
}

func (e *EmailReportJob) SendReport(ctx context.Context, utc int) {
	taskLock := constants.EmailReportJobTaskLockPrefix + strconv.Itoa(utc)
	methodName := util.GetCurrentFuncName()
	getLock, err := e.redisService.SetNX(methodName, taskLock, "1", lockTimeout)
	if err != nil {
		e.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,call redis nx fail，this server not run this job", methodName))
		return
	}
	if !getLock {
		e.l.WithContext(ctx).Error(fmt.Sprintf("method:%s,get redis lock fail，this server not run this job", methodName))
		return
	}
	defer func() {
		del := e.redisService.Del(taskLock)
		if !del {
			e.l.WithContext(ctx).Error(fmt.Sprintf("method:%s，del redis lock fail", methodName))
		}
	}()
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email report utc%v", methodName, utc))
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email getTimeDTO utc%v", methodName, utc))
	timeDTO := getTimeDTO(utc, time.Now().Unix())
	e.sendJob(ctx, utc, timeDTO, true)
}

func (e *EmailReportJob) ManualSend(ctx context.Context, utc int, timestamp int64, sendEmail bool) {
	timeDTO := getTimeDTO(utc, timestamp)
	e.sendJob(ctx, utc, timeDTO, sendEmail)
}

func (e *EmailReportJob) sendJob(ctx context.Context, utc int, timeDTO *TimeDTO, sendEmail bool) {
	methodName := util.GetCurrentFuncName()
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email getByUserInfo utc%v", methodName, utc))
	userInfoDataMap, userInfoTotalData, err := e.getByUserInfo(ctx, timeDTO)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s get userInfo data error err:%v", methodName, err)
		return
	}
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email getByUserJoinGroup utc%v", methodName, utc))
	userJoinGroupDataMap, userJoinGroupTotalData, err := e.getByUserJoinGroup(ctx, timeDTO)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s  getByUserJoinGroup error:%v", methodName, err)
		return
	}
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email getMsgData utc%v", methodName, utc))
	msgSendDataMap, msgSendTotalData, err := e.getMsgData(ctx, timeDTO)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s  getMsgData error:%v", methodName, err)
		return
	}
	e.l.WithContext(ctx).Infof(fmt.Sprintf("method:%s,start send email sendReport utc%v", methodName, utc))
	e.sendReport(ctx, timeDTO, userInfoDataMap, userInfoTotalData, userJoinGroupDataMap, userJoinGroupTotalData, msgSendDataMap, msgSendTotalData, timeDTO.Date, utc, sendEmail)
}

// 获取国码
func (e *EmailReportJob) getCountryCode(waId string) string {
	for _, prefix := range e.bizConf.Activity.WaIdPrefixList {
		if strings.HasPrefix(waId, prefix) {
			return prefix
		}
	}
	return ""
}

// 获取用户信息相关数据
func (e *EmailReportJob) getByUserInfo(ctx context.Context, timeDTO *TimeDTO) (map[BasicInfo]*DataByUserInfo, *DataByUserInfo, error) {
	id := int64(0)
	dataMap := make(map[BasicInfo]*DataByUserInfo)
	totalData := &DataByUserInfo{}
	languageMap := e.bizConf.Activity.LanguageMap
	channelMap := e.bizConf.Activity.ChannelMap
	for {
		userInfos, err := e.userInfoRepo.ListGtIdLtEndTime(ctx, id, timeDTO.EndTime, e.limit)
		if err != nil {
			return dataMap, totalData, err
		}
		if len(userInfos) == 0 {
			break
		}
		for _, userInfo := range userInfos {
			id = userInfo.ID
			countryCode := e.getCountryCode(userInfo.WaID)
			if countryCode == "" {
				continue
			}
			basicInfo := BasicInfo{
				Language:    languageMap[userInfo.Language],
				Channel:     channelMap[userInfo.Channel],
				CountryCode: countryCode,
			}
			userInfoData := dataMap[basicInfo]
			if userInfoData == nil {
				userInfoData = &DataByUserInfo{}
				dataMap[basicInfo] = userInfoData
			}
			generation := userInfo.Generation
			if generation > 6 {
				generation = 6
			}
			if generation < 1 {
				continue
			}
			// 当日数据
			if userInfo.CreateTime.After(timeDTO.StartTime) && userInfo.CreateTime.Before(timeDTO.EndTime) {
				userInfoData.GenerationCount[generation]++
			}
			// 累计数据
			totalData.GenerationCount[generation]++
			joinCount := userInfo.JoinCount
			if joinCount == 0 {
				continue
			}
			if joinCount > 15 {
				joinCount = 15
			}
			userInfoData.TotalJoinCount[joinCount]++
			totalData.TotalJoinCount[joinCount]++
		}
	}
	return dataMap, totalData, nil
}

// 获取参团相关数据
func (e *EmailReportJob) getByUserJoinGroup(ctx context.Context, timeDTO *TimeDTO) (map[BasicInfo]*DataByUserJoinGroup, *DataByUserJoinGroup, error) {
	// 拼接统计数据
	dataMap := make(map[BasicInfo]*DataByUserJoinGroup)
	totalData := &DataByUserJoinGroup{}
	groupTime, _ := e.userJoinGroupRepo.GetFirstLeJoinGroupTime(ctx, timeDTO.StartTime.Unix())
	if groupTime == nil {
		return dataMap, totalData, nil
	}
	id := groupTime.ID - 1
	codeCountMap := make(map[string]int)
	codeInfoMap := make(map[string]*UserInfo)
	languageMap := e.bizConf.Activity.LanguageMap
	channelMap := e.bizConf.Activity.ChannelMap
	for {
		// 统计区间内数据
		userJoinGroups, _ := e.userJoinGroupRepo.ListGtIdGtJoinGroupTime(ctx, id, timeDTO.StartTime.Unix(), timeDTO.EndTime.Unix(), e.limit)
		if len(userJoinGroups) == 0 {
			break
		}
		// 统计拉团人数 并获取用户信息
		helpCodes := make([]string, 0)
		for _, userJoinGroup := range userJoinGroups {
			id = userJoinGroup.ID
			codeCountMap[userJoinGroup.HelpCode]++
			helpCodes = append(helpCodes, userJoinGroup.HelpCode)
		}
		userInfos, err := e.userInfoRepo.ListUserInfoByHelpCodes(ctx, helpCodes)
		if err != nil {
			return dataMap, totalData, err
		}
		for _, userInfo := range userInfos {
			codeInfoMap[userInfo.HelpCode] = userInfo
		}
	}

	for helpCode, userInfo := range codeInfoMap {
		joinCount := codeCountMap[helpCode]
		if joinCount == 0 {
			continue
		}
		if joinCount > 15 {
			joinCount = 15
		}
		countryCode := e.getCountryCode(userInfo.WaID)
		if countryCode == "" {
			continue
		}
		basicInfo := BasicInfo{
			Language:    languageMap[userInfo.Language],
			Channel:     channelMap[userInfo.Channel],
			CountryCode: countryCode,
		}
		userJoinGroupData := dataMap[basicInfo]
		if userJoinGroupData == nil {
			userJoinGroupData = &DataByUserJoinGroup{}
			dataMap[basicInfo] = userJoinGroupData
		}
		userJoinGroupData.DailyJoinCount[joinCount]++
		totalData.DailyJoinCount[joinCount]++
	}
	return dataMap, totalData, nil
}

// 获取消息相关数据
func (e *EmailReportJob) getMsgData(ctx context.Context, timeDTO *TimeDTO) (map[BasicInfo]*MsgData, *MsgData, error) {
	startDate := timeDTO.StartTime.Format("20060102")
	endDate := timeDTO.EndTime.Format("20060102")

	var pts []string
	pts = append(pts, startDate)
	if startDate != endDate {
		pts = append(pts, endDate)
	}

	id := int64(0)
	dataMap := make(map[BasicInfo]*MsgData)
	totalData := &MsgData{}
	languageMap := e.bizConf.Activity.LanguageMap
	channelMap := e.bizConf.Activity.ChannelMap
	for {
		sendMsgList, _ := e.msgSendRepo.ListGtIdInPts(ctx, id, pts, e.limit)
		if len(sendMsgList) == 0 {
			break
		}
		// 构建用户信息map
		waIds := make([]string, 0)
		for _, sendMsg := range sendMsgList {
			id = sendMsg.ID
			waIds = append(waIds, sendMsg.WaID)
		}
		userInfos, err := e.userInfoRepo.FindUserInfos(ctx, waIds)
		if err != nil {
			return dataMap, totalData, err
		}
		userInfoMap := make(map[string]*UserInfo)
		for _, userInfo := range userInfos {
			userInfoMap[userInfo.WaID] = userInfo
		}
		// 继续统计流程
		for _, sendMsg := range sendMsgList {
			userInfo := userInfoMap[sendMsg.WaID]
			if userInfo == nil {
				continue
			}
			countryCode := e.getCountryCode(userInfo.WaID)
			if countryCode == "" {
				continue
			}
			basicInfo := BasicInfo{
				Language:    languageMap[userInfo.Language],
				Channel:     channelMap[userInfo.Channel],
				CountryCode: countryCode,
			}
			msgData := dataMap[basicInfo]
			if msgData == nil {
				msgData = &MsgData{}
				dataMap[basicInfo] = msgData
			}
			switch sendMsg.MsgType {
			case constants.PromoteClusteringMsg:
				msgData.CountV3++
				totalData.CountV3++
			case constants.RenewFreeMsg:
				msgData.CountV22++
				totalData.CountV22++
			case constants.PayRenewFreeMsg:
				msgData.CountV36++
				totalData.CountV36++
			}
			if sendMsg.MsgType == constants.CannotAttendActivityMsg {
				// 非白拦截
				msgData.InterceptCount++
				totalData.InterceptCount++
			} else if sendMsg.State == constants.MsgSendStateSuccess {
				msgData.SuccessCount++
				totalData.SuccessCount++
			} else if sendMsg.State == constants.MsgSendStateFail {
				msgData.FailedCount++
				totalData.FailedCount++
			} else if sendMsg.State == constants.MsgSendStateNxFail {
				msgData.FailedCount++
				totalData.FailedCount++
			} else if sendMsg.State == constants.MsgSendStateNxTimeout {
				msgData.TimeoutCount++
				totalData.TimeoutCount++
			}
		}
	}
	return dataMap, totalData, nil
}

func getTimeDTO(utc int, timestamp int64) *TimeDTO {
	utcZone := time.FixedZone("UTC "+strconv.Itoa(utc), utc*60*60)
	utcTime := time.Unix(timestamp, 0).In(utcZone)
	utcDate := utcTime.Format("20060102")
	startTime := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day()-1, 0, 0, 0, 0, utcTime.Location())
	endTime := time.Date(utcTime.Year(), utcTime.Month(), utcTime.Day(), 0, 0, 0, 0, utcTime.Location())
	return &TimeDTO{
		Date:      utcDate,
		StartTime: time.Unix(startTime.Unix(), 0),
		EndTime:   time.Unix(endTime.Unix(), 0),
	}
}

// 构建excel部分
func (e *EmailReportJob) sendReport(ctx context.Context, timeDTO *TimeDTO, userInfoDataMap map[BasicInfo]*DataByUserInfo, totalUserInfoData *DataByUserInfo,
	userJoinGroupDataMap map[BasicInfo]*DataByUserJoinGroup, totalUserJoinGroupData *DataByUserJoinGroup,
	msgSendDataMap map[BasicInfo]*MsgData, totalMsgData *MsgData,
	date string, utc int, sendEmail bool) {
	e.l.WithContext(ctx).Infof("method:%s,start getHistorySendList，utc:%d,date:%s", "sendReport", utc, date)
	// 获取历史数据
	sendDataList := e.getHistorySendList(ctx, utc, timeDTO)
	// 统计今日数据
	currentSendDataList, err := e.buildCurrentSendList(userInfoDataMap, userJoinGroupDataMap, msgSendDataMap, date)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,buildCurrentSendList，err:%v", "sendReport", err)
		return
	}
	e.l.WithContext(ctx).Infof("method:%s,start saveSendList，utc:%d,date:%s", "sendReport", utc, date)
	// 存储今日统计数据
	err = e.saveSendList(ctx, currentSendDataList, utc)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,保存今日统计数据失败，err:%v", "sendReport", err)
		return
	}
	e.l.WithContext(ctx).Infof("method:%s,start buildTotalSendList，utc:%d,date:%s", "sendReport", utc, date)
	if !sendEmail {
		return
	}
	// 统计累计数据
	totalSendData, err := e.buildTotalSendList(totalUserInfoData, totalUserJoinGroupData, totalMsgData, date)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,统计累计数据失败，err:%v", "sendReport", err)
		return
	}
	// 合并数据
	for _, sendData := range currentSendDataList {
		sendDataList = append(sendDataList, sendData)
	}
	sendDataList = append(sendDataList, totalSendData)
	e.l.WithContext(ctx).Infof("method:%s,start generateExcelFile，utc:%d,date:%s", "sendReport", utc, date)
	fileBytes, err := e.generateExcelFile(sendDataList)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,生成excel文件失败，err:%v", "sendReport", err)
		return
	}
	e.l.WithContext(ctx).Infof("method:%s,start sendEmail，utc:%d,date:%s", "sendReport", utc, date)
	err = e.sendEmail(fileBytes, strconv.Itoa(utc), date)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,发送邮件失败，err:%v", "sendReport", err)
		return
	}
}

func (e *EmailReportJob) getHistorySendList(ctx context.Context, utc int, timeDTO *TimeDTO) []*ReportJsonDTO {
	historySendList, err := e.emailReportRepo.ListAllEmailReport(ctx, utc)
	if err != nil {
		e.l.WithContext(ctx).Errorf("method:%s,获取历史邮件发送记录失败，err:%v", "getHistorySendList", err)
		return []*ReportJsonDTO{}
	}
	sendList := make([]*ReportJsonDTO, 0)
	for _, historyDB := range historySendList {
		if historyDB.Date == timeDTO.Date {
			continue
		}
		send := &ReportJsonDTO{
			Date:            historyDB.Date,
			Language:        historyDB.Language,
			Channel:         historyDB.Channel,
			CountryCode:     historyDB.CountryCode,
			SuccessCount:    historyDB.SuccessCount,
			FailedCount:     historyDB.FailedCount,
			TimeoutCount:    historyDB.TimeoutCount,
			InterceptCount:  historyDB.InterceptCount,
			GenerationCount: historyDB.GenerationCount,
			DailyJoinCount:  historyDB.DailyJoinCount,
			TotalJoinCount:  historyDB.TotalJoinCount,
			CountV3:         historyDB.CountV3,
			CountV22:        historyDB.CountV22,
			CountV36:        historyDB.CountV36,
		}
		sendList = append(sendList, send)
	}
	return sendList
}

func (e *EmailReportJob) saveSendList(ctx context.Context, sendList []*ReportJsonDTO, utc int) error {
	dbSendList := make([]*EmailReportDTO, 0)
	for _, send := range sendList {
		dbSend := &EmailReportDTO{
			Date:            send.Date,
			Utc:             strconv.Itoa(utc),
			Language:        send.Language,
			Channel:         send.Channel,
			CountryCode:     send.CountryCode,
			SuccessCount:    send.SuccessCount,
			FailedCount:     send.FailedCount,
			TimeoutCount:    send.TimeoutCount,
			InterceptCount:  send.InterceptCount,
			GenerationCount: send.GenerationCount,
			TotalJoinCount:  send.TotalJoinCount,
			DailyJoinCount:  send.DailyJoinCount,
			CountV3:         send.CountV3,
			CountV22:        send.CountV22,
			CountV36:        send.CountV36,
		}
		dbSendList = append(dbSendList, dbSend)
	}
	_, err := e.emailReportRepo.AddBatchEmailReport(ctx, dbSendList, utc)
	if err != nil {
		return err
	}
	return nil
}
func (e *EmailReportJob) buildCurrentSendList(userInfoDataMap map[BasicInfo]*DataByUserInfo,
	userJoinGroupDataMap map[BasicInfo]*DataByUserJoinGroup,
	msgSendDataMap map[BasicInfo]*MsgData, date string) ([]*ReportJsonDTO, error) {
	channelList := e.bizConf.Activity.ChannelList
	languageList := e.bizConf.Activity.LanguageList
	countryCodeList := e.bizConf.Activity.WaIdPrefixList
	reportJsonDTOList := make([]*ReportJsonDTO, 0)
	languageMap := e.bizConf.Activity.LanguageMap
	channelMap := e.bizConf.Activity.ChannelMap
	for _, channel := range channelList {
		for _, language := range languageList {
			for _, countryCode := range countryCodeList {
				baseInfo := BasicInfo{
					Channel:     channelMap[channel],
					CountryCode: countryCode,
					Language:    languageMap[language],
				}
				userInfoData := userInfoDataMap[baseInfo]
				if userInfoData == nil {
					userInfoData = &DataByUserInfo{}
				}
				userJoinGroupData := userJoinGroupDataMap[baseInfo]
				if userJoinGroupData == nil {
					userJoinGroupData = &DataByUserJoinGroup{}
				}
				msgSendData := msgSendDataMap[baseInfo]
				if msgSendData == nil {
					msgSendData = &MsgData{}
				}
				reportJsonDTO := &ReportJsonDTO{
					Date:           date,
					Channel:        channelMap[channel],
					Language:       languageMap[language],
					CountryCode:    countryCode,
					CountV3:        msgSendData.CountV3,
					CountV22:       msgSendData.CountV22,
					CountV36:       msgSendData.CountV36,
					SuccessCount:   msgSendData.SuccessCount,
					FailedCount:    msgSendData.FailedCount,
					TimeoutCount:   msgSendData.TimeoutCount,
					InterceptCount: msgSendData.InterceptCount,
				}
				reportJsonDTO.GenerationCount[0] = userInfoData.GenerationCount[2] + userInfoData.GenerationCount[3] + userInfoData.GenerationCount[4] + userInfoData.GenerationCount[5] + userInfoData.GenerationCount[6]
				for i := range userInfoData.GenerationCount {
					if i == 0 {
						continue
					}
					reportJsonDTO.GenerationCount[i] = userInfoData.GenerationCount[i]
				}
				for i := range userInfoData.TotalJoinCount {
					if i == 0 {
						continue
					}
					reportJsonDTO.TotalJoinCount[i] = userInfoData.TotalJoinCount[i]
				}
				for i := range userJoinGroupData.DailyJoinCount {
					if i == 0 {
						continue
					}
					reportJsonDTO.DailyJoinCount[i] = userJoinGroupData.DailyJoinCount[i]
				}
				reportJsonDTOList = append(reportJsonDTOList, reportJsonDTO)
			}
		}
	}
	return reportJsonDTOList, nil
}

func (e *EmailReportJob) buildTotalSendList(totalUserInfoData *DataByUserInfo, totalUserJoinGroupData *DataByUserJoinGroup, totalMsgData *MsgData, date string) (*ReportJsonDTO, error) {
	// 累计
	reportJsonDTO := &ReportJsonDTO{
		Date:           date,
		Channel:        "去重合并多渠道",
		Language:       "去重合并多语言",
		CountryCode:    "去重合并多国码",
		CountV3:        totalMsgData.CountV3,
		CountV22:       totalMsgData.CountV22,
		CountV36:       totalMsgData.CountV36,
		SuccessCount:   totalMsgData.SuccessCount,
		FailedCount:    totalMsgData.FailedCount,
		TimeoutCount:   totalMsgData.TimeoutCount,
		InterceptCount: totalMsgData.InterceptCount,
	}
	reportJsonDTO.GenerationCount[0] = totalUserInfoData.GenerationCount[2] + totalUserInfoData.GenerationCount[3] + totalUserInfoData.GenerationCount[4] + totalUserInfoData.GenerationCount[5] + totalUserInfoData.GenerationCount[6]
	for i := range totalUserInfoData.GenerationCount {
		if i == 0 {
			continue
		}
		reportJsonDTO.GenerationCount[i] = totalUserInfoData.GenerationCount[i]
	}
	for i := range totalUserInfoData.TotalJoinCount {
		if i == 0 {
			continue
		}
		reportJsonDTO.TotalJoinCount[i] = totalUserInfoData.TotalJoinCount[i]
	}
	for i := range totalUserJoinGroupData.DailyJoinCount {
		if i == 0 {
			continue
		}
		reportJsonDTO.DailyJoinCount[i] = totalUserJoinGroupData.DailyJoinCount[i]
	}
	return reportJsonDTO, nil
}

func (e *EmailReportJob) mergeCell(file *excelize.File, sheetName, startIndex, endIndex, value string) error {
	methodName := "mergeCell"

	err := file.MergeCell(sheetName, startIndex, endIndex)
	if err != nil {
		e.l.Errorf("method:%s,合并单元格%v 失败，err:%v", methodName, startIndex, err)
		return err
	}

	// 设置合并单元格的值
	err = file.SetCellValue(sheetName, startIndex, value)
	if err != nil {
		e.l.Errorf("method:%s,设置合并单元格%v 的值失败，err:%v", methodName, startIndex, err)
		return err
	}
	return nil
}

func getCloKey(index int) string {
	if index < 26 {
		return string(rune(int('A') + index))
	} else {
		index = index - 26
		return "A" + string(rune(int('A')+index))
	}
}

// GetExcelColumnName 根据给定的序号获取对应的Excel列名
func generateExcelColumnNames(count int) []string {
	columnNames := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		columnName := ""
		num := i
		for num > 0 {
			num--
			remainder := num % 26
			columnName = string(rune(remainder+65)) + columnName
			num = num / 26
		}
		columnNames = append(columnNames, columnName)
	}
	return columnNames
}

// 创建Excel文件并返回文件流
func (e *EmailReportJob) generateExcelFile(reportJsonDtoList []*ReportJsonDTO) ([]byte, error) {
	// 创建一个新的 Excel 文件
	f := excelize.NewFile()
	methodName := "sendReport-generateExcelFile"
	// 在第一个工作表中设置表头（两行）
	sheetName := "数据日报"
	err := f.SetSheetName("Sheet1", sheetName)
	if err != nil {
		e.l.Warnf("method:%s,创建sheet失败，err:%v", methodName, err)
		return nil, err
	}

	// 设置第二行的标题
	headers := []string{"日期", "语言", "渠道", "国码",
		"初代种子引入", "2代人数", "3代人数", "4代人数", "5代人数", "6+代人数", "总裂变人数（2代及之后累计）",
		"拉1人人数", "拉2人人数", "拉3人人数", "拉4人人数", "拉5人人数", "拉6人人数", "拉7人人数", "拉8人人数",
		"拉9人人数", "拉10人人数", "拉11人人数", "拉12人人数", "拉13人人数", "拉14人人数", "拉15人人数",
		"拉1人人数", "拉2人人数", "拉3人人数", "拉4人人数", "拉5人人数", "拉6人人数", "拉7人人数", "拉8人人数",
		"拉9人人数", "拉10人人数", "拉11人人数", "拉12人人数", "拉13人人数", "拉14人人数", "拉15人人数",
	}

	cells := generateExcelColumnNames(len(headers) + 7)

	// 设置宽度
	for _, cell := range cells {
		err = f.SetColWidth(sheetName, cell, cell, 15) // 设置每列宽度为 20
		if err != nil {
			e.l.Warnf("method:%s,设置sheet样式失败，err:%v", methodName, err)
			return nil, err
		}
	}

	err = e.mergeCell(f, sheetName, "A1", "D1", "基础类目")
	if err != nil {
		return nil, err
	}

	err = e.mergeCell(f, sheetName, "E1", "K1", "导入和裂变情况")
	if err != nil {
		return nil, err
	}

	err = e.mergeCell(f, sheetName, "L1", "Z1", "助力滞留情况【A单日】")
	if err != nil {
		return nil, err
	}

	err = e.mergeCell(f, sheetName, "AA1", "AO1", "助力滞留情况【B累加】")
	if err != nil {
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AP1", "AP1", 15)
	err = f.SetCellValue(sheetName, "AP1", "催促成团下发数")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AP1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AQ1", "AQ1", 15)
	err = f.SetCellValue(sheetName, "AQ1", "免费续时下发数")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AQ1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AR1", "AR1", 15)
	err = f.SetCellValue(sheetName, "AR1", "付费续时下发数")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AR1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AS1", "AS1", 15)
	err = f.SetCellValue(sheetName, "AS1", "发送成功")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AS1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AT1", "AT1", 15)
	err = f.SetCellValue(sheetName, "AT1", "发送失败")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AT1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AU1", "AU1", 15)
	err = f.SetCellValue(sheetName, "AU1", "发送超时")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AU1值失败，err:%v", methodName, err)
		return nil, err
	}

	_ = f.SetColWidth(sheetName, "AV1", "AV1", 15)
	err = f.SetCellValue(sheetName, "AV1", "非白拦截")
	if err != nil {
		e.l.Errorf("method:%s,设置单元格AV1值失败，err:%v", methodName, err)
		return nil, err
	}

	cells = generateExcelColumnNames(len(headers))
	for i, header := range headers {
		cell := fmt.Sprintf("%s2", cells[i])
		err = f.SetCellValue(sheetName, cell, header)
		if err != nil {
			e.l.Errorf("method:%s,设置第二行的标题失败，err:%v", methodName, err)
			return nil, err
		}
	}

	// 填充数据
	for i, reportJsonDto := range reportJsonDtoList {
		row := i + 3 // 从第 3 行开始填充数据
		index := 0
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Date)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Language)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Channel)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.CountryCode)
		for j := range reportJsonDto.GenerationCount {
			if j == 0 {
				continue
			}
			index++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.GenerationCount[j])
		}
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.GenerationCount[0])
		for j := range reportJsonDto.DailyJoinCount {
			if j == 0 {
				continue
			}
			index++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.DailyJoinCount[j])
		}
		for j := range reportJsonDto.TotalJoinCount {
			if j == 0 {
				continue
			}
			index++
			_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.TotalJoinCount[j])
		}
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.CountV3)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.CountV22)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.CountV36)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.SuccessCount)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.FailedCount)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.TimeoutCount)
		index++
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.InterceptCount)
	}

	// 将Excel文件保存到内存中的字节切片
	var buf bytes.Buffer
	err = f.Write(&buf)
	if err != nil {
		e.l.Errorf("method:%s,xlsx写字节数组失败，err:%v", methodName, err)
		return nil, err
	}

	// 返回字节切片
	return buf.Bytes(), nil
}

func (e *EmailReportJob) sendEmail(fileData []byte, utc string, date string) error {
	methodName := "sendReport-sendEmail"
	// 创建一个新的邮件对象
	mailer := gomail.NewMessage()

	emailConfig := e.dataConf.EmailConfig

	// 设置发件人和收件人
	mailer.SetHeader("From", emailConfig.FromAddress)                       // 发件人地址
	mailer.SetHeader("To", emailConfig.ToAddressList...)                    // 收件人地址
	mailer.SetHeader("Subject", "MLBB25031裂变活动-活动每日统计数据-"+date+" UTC "+utc) // 邮件主题
	mailer.SetBody("text/plain", "统计数据详见附件")                                // 邮件内容

	// 将字节数组作为附件附加到邮件中
	// 使用 AttachReader 方法将字节数组包装为一个 io.Reader 来作为附件
	mailer.Attach("wa日报 UTC"+utc+" "+date+".xlsx", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(fileData)
		return err
	}))

	// 设置 SMTP 服务器的配置信息
	dialer := gomail.NewDialer(
		emailConfig.ServerHost,      //邮箱的 SMTP 服务器地址
		int(emailConfig.ServerPort), //邮箱的 SMTP 端口
		emailConfig.ApiUser,         // user
		emailConfig.ApiKey,          // 密码（或应用专用密码）
	)
	// 设置 SSL 加密
	//dialer.SSL = true

	// 发送邮件
	if err := dialer.DialAndSend(mailer); err != nil {
		e.l.Warnf("method:%s,send email failed, activity id:%v，err:%v", methodName, e.bizConf.Activity.Id, err)
		return err
	}
	e.l.Infof("method:%s,send email success, activity id:%v", methodName, e.bizConf.Activity.Id)
	return nil
}
