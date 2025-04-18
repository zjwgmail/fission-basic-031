package service

import (
	"context"
	"fission-basic/internal/conf"
	"fission-basic/internal/util/encoder/rsa"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fission-basic/api/constants"
	v1 "fission-basic/api/helloworld/v1"
	"fission-basic/internal/biz"
	"fission-basic/internal/pojo/dto"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
)

type StudentService struct {
	stuUC         *biz.StudentUsecase
	waMsgService  *biz.WaMsgService
	waMsgSendRepo biz.WaMsgSendRepo
	activity      *biz.ActivityInfoUsecase
	userUsecase   *biz.UserInfoUsecase
	l             *log.Helper
	bizConfig     *conf.Business
	privateKey    string
}

func NewStudentService(d *conf.Data, stu *biz.StudentUsecase, waMsgService *biz.WaMsgService, waMsgSendRepo biz.WaMsgSendRepo, usecase *biz.ActivityInfoUsecase, userinfo *biz.UserInfoUsecase, l log.Logger, conf *conf.Business) *StudentService {
	return &StudentService{
		stuUC:         stu,
		waMsgService:  waMsgService,
		l:             log.NewHelper(l),
		waMsgSendRepo: waMsgSendRepo,
		activity:      usecase,
		userUsecase:   userinfo,
		privateKey:    d.Rsa.PrivateKey,
		bizConfig:     conf,
	}
}

func (s *StudentService) AddStudent(ctx context.Context, req *v1.AddStudentRequest) (*v1.AddStudentResponse, error) {
	name := req.Name
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	err := s.stuUC.AddStudent(ctx, name)
	if err != nil {
		return nil, err
	}

	return &v1.AddStudentResponse{
		Name: name,
	}, nil
}

func (s *StudentService) GetStudent(ctx context.Context, req *v1.GetStudentRequest) (*v1.GetStudentRespose, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	stu, err := s.stuUC.GetStudent(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if stu == nil {
		return nil, fmt.Errorf("student not found")
	}

	return &v1.GetStudentRespose{
		Name: stu.Name,
	}, nil
}

func (s *StudentService) ListStudents(ctx context.Context, req *v1.ListStudentsRequest) (*v1.ListStudentsResponse, error) {
	offset := req.Offset
	length := req.Lenth
	if length == 0 {
		length = 10
	}

	stus, total, err := s.stuUC.ListStudents(ctx, uint(offset), uint(length))
	if err != nil {
		return nil, err
	}

	return &v1.ListStudentsResponse{
		Stus: lo.Map(stus, func(stu *biz.Student, _ int) *v1.Stu {
			return &v1.Stu{
				Name:      stu.Name,
				CreatedAt: stu.CreatedAt.UnixMilli(),
			}
		}),
		Total: total,
	}, nil
}

func (s *StudentService) MessageSend(ctx context.Context, req *v1.InvitationRequest) (*v1.InvitationResponse, error) {
	buildMsgInfo := &dto.BuildMsgInfo{
		WaId:       "85257481920",
		MsgType:    constants.ActivityTaskMsg,
		Channel:    "01",
		Language:   "01",
		Generation: "01",
		RallyCode:  "a010100000",
	}

	//help := &dto.HelpParam{
	//	WaId:      "85257481920",
	//	RallyCode: "a010100000",
	//	IsHelp:    false,
	//}

	//helpNickName := []*dto.HelpNickNameInfo{
	//	{Id: 111, UserNickname: "ceshi", RallyCode: "a010100001"},
	//}
	//nx, err := s.waMsgService.HelpOverMsg2NX(ctx, buildMsgInfo, "xxxx", helpNickName, constants.BizTypeInteractive, "http:xxxx")

	//nx, err := s.waMsgService.HelpTaskSingleSuccessMsg2NX(ctx, buildMsgInfo, "xxxx", helpNickName)

	//nx, err := s.waMsgService.ActivityTask2NX(ctx, buildMsgInfo, help)
	nx, err := s.waMsgService.CannotAttendActivity2NX(ctx, buildMsgInfo)
	if err != nil {
		return nil, err
	}

	// 新增消息表
	for _, paramsDto := range nx {
		id, err := s.waMsgSendRepo.AddWaMsgSend(ctx, paramsDto.WaMsgSend)
		if err != nil {
			return nil, err
		}
		paramsDto.WaMsgSend.ID = id
	}

	// 后续会更新消息表，需要消息id
	_, err = s.waMsgService.SendMsgList2NX(ctx, nx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *StudentService) ActivityGet(ctx context.Context, req *v1.InvitationRequest) (*v1.InvitationResponse, error) {

	update := &biz.UpdateActivityInfoDto{
		Id:             "mlbb25031",
		ActivityStatus: constants.ATStatusBuffer,
	}
	s.l.Infof(fmt.Sprintf("method[%s],ActivityGet activity to ATStatusBuffer activity's id:%v", "ActivityJobHandle", update.Id))
	err := s.activity.UpdateActivityInfo(ctx, update)
	if err != nil {
		return nil, err

	}
	return nil, nil
}

func (s *StudentService) TimeGet(ctx context.Context, req *v1.InvitationRequest) (*v1.InvitationResponse, error) {

	now := time.Now()
	s.l.Infof("当前时间：%v", now)
	return &v1.InvitationResponse{
		HtmlText: now.Format(time.DateTime),
	}, nil
}

func (s *StudentService) Invitation(ctx context.Context, req *v1.InvitationRequest) (*v1.InvitationResponse, error) {
	title := s.bizConfig.Activity.Title
	desc := s.bizConfig.Activity.Desc
	imageLink := s.bizConfig.Activity.ImageLink
	domain := s.bizConfig.Activity.ShowDomain
	codeEntry := req.Code
	code, err := rsa.Decrypt(codeEntry, s.privateKey)
	if err != nil {
		s.l.Warnf("decrypt code failed, err=%v, code=%s", err, codeEntry)
		return nil, err
	}
	info, err := s.userUsecase.GetUserInfoByHelpCode(ctx, code)
	if err != nil {
		s.l.Warnf("get user info failed, err=%v, code=%s", err, code)
		return nil, err
	}
	metaOgTitle := ""
	metaOgDescription := ""
	metaOgUrl := ""
	metaOgImage := ""

	js := "https://sg-play.mobilelegends.com/events/mlbb25031/invitate.js?t=" + strconv.FormatInt(time.Now().Unix(), 10)
	language := info.Language
	langFlag := "L02"
	if "03" == language {
		langFlag = "L03"
	} else if "04" == language {
		langFlag = "L04"
	} else {
		langFlag = "L02"
	}
	metaOgTitle = title[langFlag]
	metaOgDescription = desc[langFlag]
	metaOgUrl = domain[langFlag]
	metaOgImage = imageLink[langFlag]

	htmlFile, err := os.ReadFile("./configs/html/mlbb25031AsyncTemplate.html")
	if err != nil {
		s.l.Warnf("read html file failed, err=%v", err)
		return nil, err
	}
	html := strings.ReplaceAll(string(htmlFile), "{{meta_og_title}}", metaOgTitle)
	html = strings.ReplaceAll(html, "{{meta_og_description}}", metaOgDescription)
	html = strings.ReplaceAll(html, "{{meta_og_url}}", metaOgUrl)
	html = strings.ReplaceAll(html, "{{meta_og_image}}", metaOgImage)
	html = strings.ReplaceAll(html, "{{js}}", js)
	htmlResponse := v1.InvitationResponse{
		HtmlText: html,
	}
	return &htmlResponse, nil
}

func ReplacePlaceholders(template string, values ...string) string {
	re := regexp.MustCompile(`\{\{(\d+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// 获取占位符的数字
		index := match[2] - '1' // match[2] 是数字字符
		if int(index) < len(values) {
			return values[int(index)]
		}
		return match // 如果没有匹配，保持原样
	})
	return result
}
