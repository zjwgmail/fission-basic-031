package util

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// countryCodeToTimeZone 存储国码与时区的映射
var countryCodeToTimeZone = map[string]string{
	"60":  "Asia/Kuala_Lumpur",
	"62":  "Asia/Jakarta",
	"63":  "Asia/Manila",
	"65":  "Asia/Singapore",
	"66":  "Asia/Bangkok",
	"84":  "Asia/Ho_Chi_Minh",
	"7":   "Europe/Moscow",
	"90":  "Europe/Istanbul",
	"966": "Asia/Riyadh",
	"380": "Europe/Kiev",
	"375": "Europe/Minsk",
	"998": "Asia/Tashkent",
	"996": "Asia/Bishkek",
	"994": "Asia/Baku",
	"373": "Europe/Chisinau",
	"992": "Asia/Dushanbe",
	"374": "Asia/Yerevan",
	"971": "Asia/Dubai",
	"973": "Asia/Bahrain",
	"974": "Asia/Qatar",
	"965": "Asia/Kuwait",
	"968": "Asia/Muscat",
	"20":  "Africa/Cairo",
	"216": "Africa/Tunis",
	"213": "Africa/Algiers",
	"92":  "Asia/Karachi",
	"880": "Asia/Dhaka",
	"852": "Asia/Hong_Kong",
}

func GetNowTimeByWaId(ctx context.Context, waId string, now time.Time) (time.Time, error) {
	timeZone, err := getTimeZoneByPhoneNumber(waId)
	if err != nil {
		log.Context(ctx).Error(fmt.Sprintf("getTimeZoneByPhoneNumber error,err:%v", err))
		return time.Now(), err
	}
	// 加载指定时区
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		log.Context(ctx).Error(fmt.Sprintf("cann't load timeZone,timeZone:%v,err:%v", timeZone, err))
		return time.Now(), fmt.Errorf("cann't load timeZone %s: %w", timeZone, err)
	}

	// 获取当前时间并转换到指定时区
	now2 := now.In(loc)
	return now2, nil
}

func GetTimeByTimeStr(timestampStr string) (time.Time, error) {
	// 将字符串转换为 int64
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		fmt.Println("GetCustomTimeByTime,Error parsing timestamp:", err)
		return time.Now(), errors.New(fmt.Sprintf("Error parsing timestamp:%v", err))
	}

	// 创建 time.Time 对象
	newTime := time.Unix(timestamp, 0)

	// 获取当前地区的时间
	//localTime := newTime.In(time.Local) // 使用本地时区

	return newTime, nil
}

// IsNotDisturbTime 查询是否在免打扰时间，isDebug 传配置中的 business.activity.isDebug
func IsNotDisturbTime(ctx context.Context, isDebug bool, waId string) (bool, error) {
	if isDebug {
		return false, nil
	}

	// 获取当前时间并转换到指定时区
	now, err := GetNowTimeByWaId(ctx, waId, time.Now())
	if err != nil {
		log.Context(ctx).Error(fmt.Sprintf("get nowTime by waId's timeZone error,waId:%v,err:%v", waId, err))
		return true, errors.New("get nowTime by waId's timeZone error")
	}

	// 提取小时数
	hour := now.Hour()

	// 判断时间是否在晚上 22 点到次日 10 点之间
	return hour >= 22 || hour < 10, nil
}

// GetSendRenewMsgTime 获取续免费消息的发送的时间（用户每次向商户号发送消息，就需要重新计算）
func GetSendRenewMsgTime(ctx context.Context, waId string, afterHour int, nowUnix int64) (int64, error) {

	now := time.Unix(nowUnix, 0)
	// 获取当前时间并转换到指定时区
	now, err := GetNowTimeByWaId(ctx, waId, now)
	if err != nil {
		return -1, err
	}

	afterTime := now.Add(time.Duration(afterHour) * time.Hour)

	// 获取时间的小时
	hour := afterTime.Hour()
	// 判断时间是否在晚上 11 点到次日 9 点之间
	if hour >= 22 && hour <= 24 {
		afterTime = time.Date(afterTime.Year(), afterTime.Month(), afterTime.Day(), 21, 0, 0, 0, afterTime.Location())
	} else if hour >= 0 && hour < 10 {
		afterTime = time.Date(afterTime.Year(), afterTime.Month(), afterTime.Day()-1, 21, 0, 0, 0, afterTime.Location())
	}

	return afterTime.Unix(), nil
}

// GetSendClusteringTime 获取发送催促成团消息时间、免费cdk时间
func GetSendClusteringTime(ctx context.Context, waId string, afterHour int, nowUnix int64) (int64, error) {

	now := time.Unix(nowUnix, 0)
	// 获取当前时间并转换到指定时区
	now, err := GetNowTimeByWaId(ctx, waId, now)
	if err != nil {
		return -1, err
	}
	// 获取当前时间后5小时的时间
	afterTime := now.Add(time.Duration(afterHour) * time.Hour)

	// 获取时间的小时
	hour := afterTime.Hour()
	// 判断时间是否在晚上 11 点到次日 9 点之间
	if hour >= 22 && hour <= 24 {
		afterTime = time.Date(afterTime.Year(), afterTime.Month(), afterTime.Day()+1, 10, 0, 0, 0, afterTime.Location())
	} else if hour >= 0 && hour < 10 {
		afterTime = time.Date(afterTime.Year(), afterTime.Month(), afterTime.Day(), 10, 0, 0, 0, afterTime.Location())
	}

	return afterTime.Unix(), nil
}

// getCountryByPhoneNumber 根据手机号获取所属时区
func getTimeZoneByPhoneNumber(waId string) (string, error) {
	// 去除可能存在的空格
	phoneNumber := strings.ReplaceAll(waId, " ", "")

	// 尝试匹配三位区号
	for code, timeZone := range countryCodeToTimeZone {
		if len(code) == 3 && strings.HasPrefix(phoneNumber, code) {
			return timeZone, nil
		}
	}

	// 尝试匹配两位区号
	for code, country := range countryCodeToTimeZone {
		if len(code) == 2 && strings.HasPrefix(phoneNumber, code) {
			return country, nil
		}
	}

	// 尝试匹配1位区号
	for code, country := range countryCodeToTimeZone {
		if len(code) == 1 && strings.HasPrefix(phoneNumber, code) {
			return country, nil
		}
	}

	return "", errors.New("no timezone found")
}

func GetPtTimeList() []string {
	now := time.Now()
	// 格式化时间为 "2006-01-02" 格式
	todayDate := now.Format("20060102")

	yesterday := now.Add(-24 * time.Hour)
	yesterdayDate := yesterday.Format("20060102")

	return []string{todayDate, yesterdayDate}
}
