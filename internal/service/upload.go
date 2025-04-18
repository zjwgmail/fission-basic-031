package service

import (
	"context"
	"errors"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gocarina/gocsv"
	"mime/multipart"
	"strings"
	"time"
)

type UploadService struct {
	uploadUserInfoRepo biz.UploadUserInfoRepo
	drainageJob        *biz.DrainageJob
	emailReportJob     *biz.EmailReportJob
	redisService       *redis.RedisService
	bizConf            *conf.Business
	l                  *log.Helper
}
type CSVRecord struct {
	AppName        string `csv:"应用名称"`
	Type           string `csv:"类型"`
	Direction      string `csv:"方向"`
	SendNumber     string `csv:"发送方号码"`
	ReceivedNumber string `csv:"接收方号码"`
	Time           string `csv:"时间"`
}

func NewUploadService(
	uploadUserInfoRepo biz.UploadUserInfoRepo,
	drainageJob *biz.DrainageJob,
	emailReportJob *biz.EmailReportJob,
	redisService *redis.RedisService,
	c *conf.Business,
	l log.Logger,
) *UploadService {
	return &UploadService{
		uploadUserInfoRepo: uploadUserInfoRepo,
		drainageJob:        drainageJob,
		emailReportJob:     emailReportJob,
		redisService:       redisService,
		bizConf:            c,
		l:                  log.NewHelper(l),
	}
}
func (u *UploadService) UploadFile(ctx context.Context, req *v1.UploadRequest) (*v1.UploadResponse, error) {
	timestamp := req.GetTimestamp()
	utc := int(req.GetUtc())
	if timestamp == 0 {
		return &v1.UploadResponse{
			Message: "错误的时间戳",
		}, nil
	}
	u.emailReportJob.ManualSend(ctx, utc, timestamp, req.GetSendEmail())
	return &v1.UploadResponse{
		Message: "开始执行",
	}, nil
}

func (u *UploadService) UploadFileV1(ctx context.Context, file *multipart.FileHeader) (*v1.UploadResponse, error) {
	csvFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	var records []CSVRecord
	if err := gocsv.Unmarshal(csvFile, &records); err != nil {
		return nil, errors.New("error parsing CSV")
	}
	uploadUserInfoList := make([]*biz.UploadUserInfoDTO, 0)
	i := 0
	loc, _ := time.LoadLocation("Asia/Shanghai")
	validWaIdPrefixList := u.bizConf.PushEvent1.CountryCodes
	for _, record := range records {
		if record.Type != "营销会话" {
			continue
		}
		if record.Direction != "下行" {
			continue
		}
		// 过滤国码
		if !validCountryCode(validWaIdPrefixList, record.ReceivedNumber) {
			continue
		}
		sendTime, err := time.ParseInLocation("2006-01-02 15:04:05", record.Time, loc)
		if err != nil {
			fmt.Printf("解析时间出错: %v\n", err)
			return nil, err
		}
		uploadUserInfo := &biz.UploadUserInfoDTO{
			PhoneNumber:  strings.TrimSpace(record.ReceivedNumber),
			LastSendTime: sendTime,
		}
		uploadUserInfoList = append(uploadUserInfoList, uploadUserInfo)
		i++
		if i == 1000 && len(uploadUserInfoList) > 0 {
			err = u.uploadUserInfoRepo.InsertBatch(ctx, uploadUserInfoList)
			if err != nil {
				u.l.Error("保存上传用户失败", err)
				return nil, err
			}
			uploadUserInfoList = make([]*biz.UploadUserInfoDTO, 0)
			i = 0
		}
	}
	if len(uploadUserInfoList) > 0 {
		err = u.uploadUserInfoRepo.InsertBatch(ctx, uploadUserInfoList)
		if err != nil {
			u.l.Error("保存上传用户失败", err)
			return nil, err
		}
	}
	defer func(csvFile multipart.File) {
		_ = csvFile.Close()
	}(csvFile)
	return &v1.UploadResponse{
		Message: "上传文件成功",
	}, nil
}

func validCountryCode(validWaIdPrefixList []string, waId string) bool {
	valid := false
	for _, prefix := range validWaIdPrefixList {
		if strings.HasPrefix(waId, prefix) {
			valid = true
			break
		}
	}
	return valid
}
