package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/util/encoder/rsa"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CDKService struct {
	userInfoUsecase     *biz.UserInfoUsecase
	systemConfigUsecase *biz.SystemConfigUsecase
	bizConfig           *conf.Business
	redisService        *redis.RedisService
	l                   *log.Helper
	privateKey          string
}

func NewCDKService(
	d *conf.Data,
	userUsecase *biz.UserInfoUsecase,
	systemConfigUsecase *biz.SystemConfigUsecase,
	bizConfig *conf.Business,
	redisService *redis.RedisService,
	l log.Logger,
) *CDKService {
	return &CDKService{
		userInfoUsecase:     userUsecase,
		systemConfigUsecase: systemConfigUsecase,
		bizConfig:           bizConfig,
		redisService:        redisService,
		l:                   log.NewHelper(l),
		privateKey:          d.Rsa.PrivateKey,
	}
}

// GetCDK implements v1.CDKHTTPServer.
func (c *CDKService) GetCDK(ctx context.Context, req *v1.GetCDKRequest) (*v1.GetCDKResponse, error) {
	if req.Param == "" {
		return &v1.GetCDKResponse{Code: 400, Message: "empty params"}, fmt.Errorf("empty params")
	}

	params, err := rsa.Decrypt(req.Param, c.privateKey)
	if err != nil {
		c.l.WithContext(ctx).Errorf("decrypt params failed, err=%v, params=%s", err, req.Param)
		return &v1.GetCDKResponse{Code: 400, Message: "decrypt params failed"}, err
	}

	ret := &CDKQuery{}
	err = json.Unmarshal([]byte(params), &ret)
	if err != nil {
		c.l.WithContext(ctx).Errorf("unmarshal params failed, err=%v, params=%s", err, params)
		return &v1.GetCDKResponse{Code: 400, Message: "unmarshal params failed"}, err
	}

	userInfo, err := c.userInfoUsecase.GetUserInfoByHelpCode(ctx, ret.Code)
	if err != nil {
		return &v1.GetCDKResponse{Code: 400, Message: "user not exit"}, err
	}

	var cdk string
	switch ret.Mode {
	case 1:
		cdk = userInfo.CDKv0
	case 3:
		cdk = userInfo.CDKv3
	case 6:
		cdk = userInfo.CDKv6
	case 9:
		cdk = userInfo.CDKv9
	case 12:
		cdk = userInfo.CDKv12
	case 15:
		cdk = userInfo.CDKv15
	}
	data := v1.GetCDKResponseData{
		RallyCode:  userInfo.Channel + userInfo.Language + fmt.Sprintf("%02d", userInfo.Generation) + ret.Code,
		Language:   userInfo.Language,
		Channel:    userInfo.Channel,
		Generation: int32(userInfo.Generation),
		WaName:     userInfo.Nickname,
		Cdk:        cdk,
	}
	return &v1.GetCDKResponse{
		Code:    200,
		Data:    &data,
		Message: "success",
	}, nil
}

func (c *CDKService) ImportCDK(ctx context.Context, req *v1.ImportCDKRequest) (*v1.ImportCDKResponse, error) {
	ctx = context.WithoutCancel(ctx)
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V0
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V3
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V6
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V9
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V12
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()
	//
	//go func() {
	//	cdk := c.bizConfig.Cdk.V15
	//	c.l.Infof("import cdk by dir, cdk=%v", cdk)
	//	start := time.Now().Unix()
	//	err := c.importCDKByDir(cdk)
	//	if err != nil {
	//		c.l.Errorf("import cdk by dir failed, err=%v", err)
	//	}
	//	c.l.Infof("import cdk=%v time=%v", cdk, time.Now().Unix()-start)
	//}()

	return &v1.ImportCDKResponse{}, nil
}

func (c *CDKService) importCDKByDir(cdk *conf.Business_CDK_CDKType) error {
	countKey := cdk.QueueName + constants.CdkTotalCountKeySuffix
	//初始化cdk库存计数
	value := c.redisService.Get(countKey)
	count, _ := strconv.Atoi(value)
	//替换~目录 测试用
	dir := cdk.Dir
	if strings.HasPrefix(dir, "~") {
		home, _ := os.UserHomeDir()
		dir = strings.Replace(dir, "~", home, 1)
	}
	if !isDir(dir) {
		return fmt.Errorf("dir not exist: %s", dir)
	}
	//扫描目录下所有文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		//处理单个文件
		countTmp, err := c.importCDKByFilePath(path, cdk.QueueName)
		count += countTmp
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = c.redisService.Set(countKey, strconv.Itoa(count))
	return nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (c *CDKService) importCDKByFilePath(filePath string, queueName string) (int, error) {
	count := 0
	batchSize := 30000
	var batch []string

	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		return count, err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			c.l.Errorf("close file failed, err=%v", err)
		}
	}(file)

	if !strings.HasSuffix(filePath, ".txt") {
		return count, nil
	}

	scanner := bufio.NewScanner(file)
	background := context.Background()
	for scanner.Scan() {
		cdk := strings.TrimSpace(string(scanner.Bytes())) // 直接操作字节
		if cdk == "" {
			continue
		}
		batch = append(batch, cdk)
		count++

		// 当批量数据达到10000条时，进行批量处理
		if len(batch) == batchSize {
			err = c.redisService.CreateQueueAndBatchPut(background, queueName, batch...)

			if err != nil {
				return count, err
			}
			total, err := c.redisService.GetQueueSize(background, queueName)
			if err != nil {
				return count, err
			}
			c.l.Infof("queueName =%v  filePath %v batch size: %d total%d", queueName, filePath, len(batch), total)

			// 清空当前批次
			batch = batch[:0]
		}
	}

	// 文件读取完成后，处理剩余的批量数据
	if len(batch) > 0 {
		err = c.redisService.CreateQueueAndBatchPut(context.Background(), queueName, batch...)
		if err != nil {
			return count, err
		}
	}

	if err = scanner.Err(); err != nil {
		return count, err
	}
	return count, nil
}

func (c *CDKService) cleanCDK() {
	c.redisService.Del("cdk_v0")
	c.redisService.Del("cdk_v3")
	c.redisService.Del("cdk_v6")
	c.redisService.Del("cdk_v9")
	c.redisService.Del("cdk_v12")
	c.redisService.Del("cdk_v15")
}

func (c *CDKService) CDKTest(ctx context.Context, req *v1.CDKTestRequest) (*v1.CDKTestResponse, error) {
	return &v1.CDKTestResponse{}, nil
}

type CDKQuery struct {
	Code string `json:"code,omitempty"`
	Mode int    `json:"mode,omitempty"`
}
