package service

import (
	"context"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/pkg/feishu"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/util/encoder/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/structpb"
)

type InitService struct {
	init            *biz.Init
	redisService    *redis.RedisService
	l               *log.Helper
	helpCodeUsecase *biz.HelpCodeUsecase
	feishuRest      *feishu.Feishu
}

func NewInitService(init *biz.Init, redisService *redis.RedisService, hcBizService *biz.HelpCodeUsecase, feishuRest *feishu.Feishu, l log.Logger) *InitService {
	return &InitService{
		init:            init,
		redisService:    redisService,
		l:               log.NewHelper(l),
		helpCodeUsecase: hcBizService,
		feishuRest:      feishuRest,
	}
}

func (i *InitService) InitDB(ctx context.Context, req *v1.InitDBRequest) (*v1.InitDBRequestResponse, error) {
	if req == nil {
		return nil, nil
	}
	return nil, nil
	//if pwd != "2020WAINDB.123" {
	//	return nil, nil
	//}
	//keys, redisErr := i.redisService.Keys("*")
	//if redisErr != nil {
	//	log.Errorf("redis keys error:%v", redisErr)
	//	return nil, redisErr
	//}
	//
	//protectedKeys := map[string]struct{}{
	//	constants.GetHelpTextWeightKey("mlbb25031"): {},
	//	//constants.HelpCodeKey:                       {},
	//	constants.CdkV0:        {},
	//	constants.CdkV0_COUNT:  {},
	//	constants.CdkV3:        {},
	//	constants.CdkV3_COUNT:  {},
	//	constants.CdkV6:        {},
	//	constants.CdkV6_COUNT:  {},
	//	constants.CdkV9:        {},
	//	constants.CdkV9_COUNT:  {},
	//	constants.CdkV12:       {},
	//	constants.CdkV12_COUNT: {},
	//	constants.CdkV15:       {},
	//	constants.CdkV15_COUNT: {},
	//}
	//
	//for _, key := range keys {
	//	// 检查key是否在保护的map中
	//	if _, protected := protectedKeys[key]; !protected {
	//		del := i.redisService.Del(key)
	//		if !del {
	//			i.l.Errorf(fmt.Sprintf("删除key %v 失败 ", key))
	//		} else {
	//			i.l.Infof(fmt.Sprintf("删除key %v 成功 ", key))
	//		}
	//	}
	//}
	//
	//_ = i.feishuRest.SendTextMsg(ctx, fmt.Sprintf("初始化缓存完成"))
	//
	//err := i.init.InitDB(context.Background(), req.Pwd)
	//if err != nil {
	//	return nil, err
	//}
	//_ = i.feishuRest.SendTextMsg(ctx, fmt.Sprintf("初始化数据库完成"))
	//return nil, nil
}

type QuerySqlResponse struct {
}

func (i *InitService) QuerySql1(ctx context.Context, req *v1.QuerySqlRequest) ([]byte, error) {
	pwd := req.Pwd
	if pwd != "s32sdAAf@sdf..3" {
		return nil, fmt.Errorf("database user or password error")
	}

	r, err := i.init.QuerySql(ctx, req.Sql)
	if err != nil {
		return nil, err
	}

	data, err := json.NewEncoder().Encode(r)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (i *InitService) QuerySql(ctx context.Context, req *v1.QuerySqlRequest) (*v1.QuerySqlResponse, error) {
	panic("unimplemented")
}

func ConvertMapsToQuerySqlResponse(data []map[string]interface{}) *v1.QuerySqlResponse {
	response := &v1.QuerySqlResponse{}
	for _, rowMap := range data {
		sqlRow := &v1.SqlRow{
			Columns: make(map[string]*structpb.Value),
		}

		for key, value := range rowMap {
			var pbValue *structpb.Value

			switch v := value.(type) {
			case string:
				pbValue = structpb.NewStringValue(v)
			case int, int8, int16, int32, int64:
				pbValue = structpb.NewNumberValue(float64(v.(int64)))
			case float32, float64:
				pbValue = structpb.NewNumberValue(v.(float64))
			case bool:
				pbValue = structpb.NewBoolValue(v)
			default:
				fmt.Println("default>>>>>", key, "===", value)
				var err error
				pbValue, err = structpb.NewValue(value)
				if err != nil {
					panic(err)
				}

				// pbValue = structpb.NewStringValue(fmt.Sprint(value))
			}
			sqlRow.Columns[key] = pbValue
		}

		response.Msg = append(response.Msg, sqlRow)
	}

	return response
}

func (i *InitService) ExeSql(ctx context.Context, req *v1.ExeSqlRequest) (*v1.ExeSqlResponse, error) {
	pwd := req.Pwd
	if pwd != "123" {
		return &v1.ExeSqlResponse{
			Msg: "database user or password error",
		}, nil
	}

	result, err := i.init.ExeSql(context.Background(), req.Sql)
	if err != nil {
		return nil, err
	}

	encode, err := json.NewEncoder().Encode(result)
	if err != nil {
		i.l.Errorf("result转json失败")
		return nil, err
	}

	i.l.Infof("result:%v", encode)

	return &v1.ExeSqlResponse{Msg: "success"}, nil
}
