package service

import (
	"context"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/nxcloud"
	"fission-basic/internal/pkg/queue"
	"fission-basic/internal/pkg/redis"
	iutil "fission-basic/internal/util"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-kratos/kratos/v2/log"
)

type NxCloudService struct {
	nxUsecase    *biz.NXCloudUsecase
	l            *log.Helper
	confData     *conf.Data
	confBusiness *conf.Business
	redisService *redis.RedisService
	gw           *queue.GW
	gwRecall     *queue.GWRecall
	gwUnknown    *queue.GWUnknown

	aiUc *biz.ActivityInfoUsecase

	// parall
	parallReqs chan *v1.UserAttendInfoRequest
	ctx        context.Context
	parallSize int
	closeReq   chan struct{}
}

func NewNxCloudService(
	nxUsecase *biz.NXCloudUsecase,
	l log.Logger,
	confData *conf.Data,
	confBusiness *conf.Business,
	redisService *redis.RedisService,
	gw *queue.GW,
	gwRecall *queue.GWRecall,
	gwUnknown *queue.GWUnknown,
	aiUc *biz.ActivityInfoUsecase,
) (*NxCloudService, func()) {
	parallSize := 15
	ctx, cancel := context.WithCancel(context.Background())
	n := &NxCloudService{
		nxUsecase:    nxUsecase,
		l:            log.NewHelper(l),
		confData:     confData,
		redisService: redisService,
		confBusiness: confBusiness,
		gw:           gw,
		gwRecall:     gwRecall,
		gwUnknown:    gwUnknown,
		aiUc:         aiUc,
		parallReqs:   make(chan *v1.UserAttendInfoRequest, parallSize*200),
		ctx:          ctx,
		parallSize:   parallSize,
		closeReq:     make(chan struct{}),
	}

	f := func() {
		close(n.parallReqs)
		cancel()
		<-n.closeReq
	}

	go n.run(ctx)
	go n.chanLen()

	return n, f
}

var (
	ErrorNxCloudRepeated = errors.New("消息重复")
)

func (n *NxCloudService) Sign(ctx context.Context, req *v1.UserAttendInfoRequest) error {
	//tr, ok := transport.FromServerContext(ctx)
	//if !ok {
	//	n.l.WithContext(ctx).Errorf("获取header失败")
	//	return errors.New("获取header失败")
	//}
	//md := tr.RequestHeader()
	//
	//signNiuxin := md.Get("Sign")
	//
	//if !n.confData.Nx.IsVerifySign && signNiuxin != "" {
	//	exists, err := n.redisService.CheckMsgSignIsExists(ctx, signNiuxin)
	//	if err != nil {
	//		return err
	//	}
	//	if exists {
	//		return ErrorNxCloudRepeated
	//	}
	//}
	//
	//if !n.confData.Nx.IsVerifySign {
	//	return nil
	//}
	//
	//commonHeaders := map[string]string{
	//	"accessKey": md.Get("AccessKey"),
	//	"ts":        md.Get("Ts"),
	//	"bizType":   md.Get("BizType"),
	//	"action":    md.Get("Action"),
	//}
	//
	//encode, err := json.NewEncoder().Encode(req)
	//if err != nil {
	//	n.l.WithContext(ctx).Errorf("req转json失败")
	//	return errors.New("req转json失败")
	//}
	//messageStr := string(encode)
	//
	//sign := util.CallSign(commonHeaders, messageStr, n.confData.Nx.Sk)
	//if sign != signNiuxin {
	//	n.l.WithContext(ctx).Errorf("方法[UserAttendInfo]，验签失败,commonHeaders:%v，messageStr:%v，sign:%v,传过来的sign:%v", commonHeaders, messageStr, sign, signNiuxin)
	//	return errors.New("验签失败")
	//}

	return nil
}

func (n *NxCloudService) userAttendInfo(
	ctx context.Context,
	req *v1.UserAttendInfoRequest,
) (*v1.UserAttendInfoResponse, error) {
	n.l.WithContext(ctx).Infof("consumer.UserAttendInfo req=%+v", req)

	err := n.Sign(ctx, req)
	if err != nil {
		if errors.Is(err, ErrorNxCloudRepeated) {
			return &v1.UserAttendInfoResponse{}, nil
		}
		return nil, err
	}

	info, err := nxcloud.ParseNXCloud(ctx, n.l, n.confData, n.confBusiness, req)
	if err != nil {
		n.l.WithContext(ctx).Errorf("parse nxCloud failed, err=%v", err)
		return &v1.UserAttendInfoResponse{}, nil
	}
	info.SendTime = time.Now().Unix()

	n.l.WithContext(ctx).Infof("info=%+v", info)

	msgType := info.MsgType

	// 助力消息
	if msgType == nxcloud.MsgTypeRallyCode ||
		msgType == nxcloud.MsgTypeAttend {
		err = n.nxUsecase.CreateMsg(ctx, info)
		if err != nil {
			n.l.WithContext(ctx).Errorf("createMsg failed, err=%v", err)
			return &v1.UserAttendInfoResponse{}, nil
		}

		return &v1.UserAttendInfoResponse{}, nil
	}

	// 免费续时消息
	if msgType == nxcloud.MsgTypeRenewMsgReply {
		err = n.nxUsecase.RenewMsg(ctx, info)
		if err != nil {
			n.l.WithContext(ctx).Errorf("RenewMsg failed, err=%v", err)
			return &v1.UserAttendInfoResponse{}, nil
		}
		return &v1.UserAttendInfoResponse{}, nil
	}

	// 回执消息
	if msgType == nxcloud.MsgTypeCallback {
		if constants.NxStatusSent == info.Status ||
			constants.NxStatusFailed == info.Status ||
			constants.NxStatusTimeout == info.Status {
			err = n.nxUsecase.Recall(ctx, info)
			if err != nil {
				n.l.WithContext(ctx).Errorf("callbackMsg failed, err=%v", err)
				return &v1.UserAttendInfoResponse{}, nil
			}
			n.l.WithContext(ctx).Infof("other callbackMsg status=%v", info.Status)
			return &v1.UserAttendInfoResponse{}, nil
		} else {
			n.l.WithContext(ctx).Infof("not sent fail msg not doing it, err=%v", err)
			return &v1.UserAttendInfoResponse{}, nil
		}
	}

	// 非白消息
	if msgType == nxcloud.MsgTypeNotWhite {
		err = n.nxUsecase.NotWhiteMsg(ctx, info)
		if err != nil {
			n.l.WithContext(ctx).Errorf("NotWhiteMsg failed, err=%v", err)
			return &v1.UserAttendInfoResponse{}, nil
		}
		n.l.WithContext(ctx).Infof("NotWhiteMsg gw end status=%v", info.Status)
		return &v1.UserAttendInfoResponse{}, nil
	}

	// todo zsj 新增收到消息表,看着逻辑已经有了
	// todo zsj 查询用户信息表看用户是否已经存在，若存在 则更新用户提醒表的最后收到消息的时间，

	// 未知消息：只保留用户提醒表、接收消息
	err = n.nxUsecase.OnlySaveMsg(ctx, info)
	if err != nil {
		n.l.WithContext(ctx).Errorf("unknow failed, err=%v", err)
		return &v1.UserAttendInfoResponse{}, nil
	}

	return &v1.UserAttendInfoResponse{}, nil
}

func (n *NxCloudService) UserAttendInfo2(
	ctx context.Context,
	req *v1.UserAttendInfoRequest,
) (*v1.UserAttendInfoResponse, error) {
	cost := iutil.MethodCost(ctx, n.l, "NxCloudService.UserAttendInfo")
	defer cost()

	n.l.WithContext(ctx).Infof("UserAttendInfo-gw req=%+v", req)

	data, err := jsoniter.Marshal(req)
	if err != nil {
		n.l.WithContext(ctx).Errorf("UserAttendInfo-gw-marshal-failed, err=%v", err)
		return &v1.UserAttendInfoResponse{}, nil
	}

	err = n.gw.Queue.SendFront([]string{string(data)}, false)
	if err != nil {
		n.l.WithContext(ctx).Errorf("UserAttendInfo send back failed, data=%s", data)
		return nil, err
	}

	n.l.WithContext(ctx).Infof("UserAttendInfo send back success, data=%s", data)

	return &v1.UserAttendInfoResponse{}, nil
}

// UserAttendInfo implements v1.NXCloudHTTPServer.
// webhook
func (n *NxCloudService) UserAttendInfo(
	ctx context.Context,
	req *v1.UserAttendInfoRequest,
) (*v1.UserAttendInfoResponse, error) {
	cost := iutil.MethodCost(ctx, n.l, "NxCloudService.UserAttendInfo")
	defer cost()

	n.l.WithContext(ctx).Infof("UserAttendInfo-gw req=%+v", req)
	_, err := nxcloud.ParseNXCloud(ctx, n.l, n.confData, n.confBusiness, req)
	if err != nil {
		n.l.WithContext(ctx).Errorf("parse nxCloud failed, err=%v", err)
		return &v1.UserAttendInfoResponse{}, nil
	}

	n.parallReqs <- req
	return &v1.UserAttendInfoResponse{}, nil

	// n.l.WithContext(ctx).Infof("UserAttendInfo send back start marshal, cost=%s", time.Since(start).String())
	// data, err := jsoniter.Marshal(req)
	// if err != nil {
	// 	n.l.WithContext(ctx).Errorf("UserAttendInfo-gw-marshal-failed, err=%v", err)
	// 	return &v1.UserAttendInfoResponse{}, nil
	// }

	// n.l.WithContext(ctx).Infof("UserAttendInfo send back start, cost=%s", time.Since(start).String())
	// err = tempQueue.SendBack([]string{string(data)}, false)
	// if err != nil {
	// 	n.l.WithContext(ctx).Errorf("UserAttendInfo send back failed, data=%s, cost=%s", data, time.Since(start).String())
	// 	return nil, err
	// }

	// n.l.WithContext(ctx).Infof("UserAttendInfo send back success, cost=%s", data, time.Since(start).String())

	// return &v1.UserAttendInfoResponse{}, nil
}

func (n *NxCloudService) chanLen() error {
	for {
		select {
		case <-n.ctx.Done():
			return n.ctx.Err()
		case <-time.After(time.Minute * 1):
			n.l.WithContext(n.ctx).Infof("gw chan len=%d", len(n.parallReqs))
		}
	}
}

func (n *NxCloudService) run(ctx context.Context) error {
	defer func() {
		n.closeReq <- struct{}{}
	}()

	parallSize := n.parallSize
	reqs := make([]*v1.UserAttendInfoRequest, 0, parallSize)
	results := make([]string, 0, parallSize)

	sendQ := func() {
		if len(reqs) == 0 {
			return
		}
		for i := range reqs {
			data, err := jsoniter.Marshal(reqs[i])
			if err != nil {
				n.l.WithContext(ctx).Errorf("UserAttendInfo-gw-marshal-failed, err=%v", err)
				continue
			}
			results = append(results, string(data))
		}
		start := time.Now()
		err := n.gw.Queue.SendBack(results, false)
		if err != nil {
			n.l.WithContext(ctx).Error("UserAttendInfo send back failed, data=%v, err=%v", results, err)
		}
		n.l.WithContext(ctx).Infof("gw send back success, size=%v , cost=%s", len(results), time.Since(start).String())

		reqs = reqs[:0]
		results = results[:0]
	}

	for {
		select {
		case <-ctx.Done():
			sendQ()
			return ctx.Err()
		default:
		}

		select {
		case r, ok := <-n.parallReqs:
			if !ok {
				sendQ()
				return nil
			}
			reqs = append(reqs, r)
		case <-time.After(time.Second * 1):
			sendQ()
		}

		if len(reqs) >= parallSize {
			sendQ()
		}
	}
}
