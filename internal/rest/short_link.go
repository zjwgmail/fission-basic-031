package rest

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fission-basic/internal/conf"
	"fission-basic/internal/pojo/dto"
	"fission-basic/internal/pojo/response"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/util"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

type ShortLink struct {
}

func ShortUrlSign(value string, signKey string) string {
	hashMac := hmac.New(sha1.New, []byte(signKey))
	hashMac.Write([]byte(value))
	out := hashMac.Sum(nil)
	return base64.StdEncoding.EncodeToString(out)
}

func getSignParamValue(longUrl string, activityId string, project string, expireAt int64) string {
	return fmt.Sprintf("url=%s&expire_at=%d&activity_id=%s&project_id=%s", longUrl, expireAt, activityId, project)
}

func (u ShortLink) GetShortUrlByUrl(ctx context.Context, longUrl string, confBiz *conf.Business) (string, error) {
	log.Infof("start create short link,longUrl:%v", longUrl)
	shortDto := dto.ShortDto{
		LongUrl:          longUrl,
		ActivityId:       confBiz.Activity.Id,
		ProjectId:        confBiz.Activity.Wa.ShortProject,
		ShortLinkApi:     confBiz.Activity.Wa.ShortLinkApi,
		ShortLinkBaseUrl: confBiz.Activity.Wa.ShortLinkBaseUrl,
		SignKey:          confBiz.Activity.Wa.ShortLinkSignKey,
	}
	shortUrl, err := u.getShortUrl(ctx, shortDto)

	log.Infof("create short link success,longUrl:%v", longUrl)
	return shortUrl, err
}

func (u ShortLink) getShortUrl(ctx context.Context, shortDto dto.ShortDto) (string, error) {
	methodName := util.GetCurrentFuncName()
	value := getSignParamValue(shortDto.LongUrl, shortDto.ActivityId, shortDto.ProjectId, 1745078400)
	sign := ShortUrlSign(value, shortDto.SignKey)

	params := map[string]any{
		"long_url":    shortDto.LongUrl,
		"expire_at":   1745078400,
		"activity_id": shortDto.ActivityId,
		"project_id":  shortDto.ProjectId,
		"sign":        sign,
	}

	log.Infof("方法[%s]，开始调用生成短链接接口,请求：params:%v", methodName, params)

	var res string
	var nxErr error
	maxRetries := 3               // 设置最大重试次数
	retryDelay := 2 * time.Second // 设置每次重试的延迟

	for i := 0; i < maxRetries; i++ {
		res, nxErr = DoPostSSL(shortDto.ShortLinkApi, params, nil, 10*1000*time.Second, 10*1000*time.Second)
		if nxErr == nil {
			break // 如果请求成功，退出重试循环
		}

		log.Warnf("方法[%s]，调用生成短链接接口http失败,params:%v,err:%v,正在重试[%d/%d]", methodName, params, nxErr, i+1, maxRetries)
		time.Sleep(retryDelay) // 等待一段时间后再重试
	}

	if nxErr != nil {
		log.Warnf("方法[%s]，调用生成短链接接口http失败,params:%v,err:%v", methodName, params, nxErr)
		return "", nxErr
	}
	log.Infof("方法[%s]，结束调用生成短链接接口,请求：params:%v,返回: %v", methodName, params, res)

	resNx := &response.ShortLinkResponse{}
	nxErr = json.NewEncoder().Decode([]byte(res), resNx)
	if nxErr != nil {
		log.Warnf("方法[%s]，生成短链接接口返回转实体报错,res:%v,err：%v", methodName, res, nxErr)
		return "", nxErr
	}
	if 0 != resNx.Code {
		log.Warnf("方法[%s]，调用生成短链接接口http失败,params:%v,res:%v", methodName, params, resNx)
		return "", errors.New("生成短链接失败")
	}
	return shortDto.ShortLinkBaseUrl + resNx.Data.ShortUrl, nil
}
