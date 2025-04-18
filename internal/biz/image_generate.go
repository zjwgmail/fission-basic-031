package biz

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fission-basic/api/constants"
	v1 "fission-basic/api/fission/v1"
	"fission-basic/internal/conf"
	"fission-basic/internal/pkg/redis"
	"fission-basic/internal/util/encoder/json"
	"fission-basic/util"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"time"
	"unicode/utf8"

	"github.com/go-kratos/kratos/v2/log"
)

type ImageGenerate struct {
	b            *conf.Business
	l            *log.Helper
	redisService *redis.RedisService
}

func NewImageGenerate(
	b *conf.Business,
	l log.Logger,
	redisService *redis.RedisService,
) *ImageGenerate {
	InitImageService()
	return &ImageGenerate{
		l:            log.NewHelper(l),
		b:            b,
		redisService: redisService,
	}
}

type SynthesisParam struct {
	BizType         int      `json:"bizType"`
	LangNum         string   `json:"langNum"`
	NicknameList    []string `json:"nicknameList"`
	CurrentProgress int64    `json:"currentProgress"`
	FilePath        string   `json:"filePath"`
	FilePaths       []string `json:"filePaths"`
	FileDir         string   `json:"fileDir"`
}

type PreSignResponse struct {

	//状态码
	Code int `json:"code"`

	//返回消息
	Message string `json:"message"`

	//返回消息
	TraceId string `json:"traceId"`

	// 响应信息
	Data interface{} `json:"data"`
}

var langMap = map[string]string{
	"01": "zh_CN",
	"02": "en",
	"03": "in",
	"04": "my",
	"05": "hi",
}

var langFont = map[string]string{
	"zh_CN": "./configs/image/zh_CN/AlibabaPuHuiTi-3-55-Regular.ttf",
	"en":    "./configs/image/en/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"in":    "./configs/image/in/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"my":    "./configs/image/my/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"hi":    "./configs/image/hi/NotoSans-VariableFont_wdth,wght.ttf",
}

var langBoldFont = map[string]string{
	"zh_CN": "./configs/image/zh_CN/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"en":    "./configs/image/en/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"bm":    "./configs/image/hi/AlibabaPuHuiTi-3-75-SemiBold.ttf",
	"hi":    "./configs/image/hi/AlibabaPuHuiTi-3-75-SemiBold.ttf",
}

var coverCopywriting = map[string]map[string]string{
	"zh_CN": {
		"left":  "你的好友【",
		"right": "】为你助力成功！",
	},
	"en": {
		"left":  "Your friend ",
		"right": " has successfully helped you!",
	},
	"my": {
		"left":  "Rakan anda ",
		"right": " telah berjaya membantu anda!",
	},
	"in": {
		"left":  "Temanmu ",
		"right": " berhasil membantumu!",
	},
}

var langFixedCover = map[string]map[int64]string{
	"en": {
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740140228280-DYTuACPlkg.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740127981539-VLyJ57UCVn.png",
	},
	"my": {
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740140204294-BgPi8yVdOp.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740140229984-1wPr096qF9.png",
	},
	"in": {
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740140245714-PJpUVhQkgn.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740140247285-Ih2U29rtMu.png",
	},
}

const BASE_PATH = "./configs/image/"

var imageMap = map[string]map[int64]string{
	"my": {
		1:  "https://mlbbmy.outweb.mobilelegends.com/1740127778327-V1IRgLGhxy.png",
		2:  "https://mlbbmy.outweb.mobilelegends.com/1740127778547-4GFe28UBen.png",
		3:  "https://mlbbmy.outweb.mobilelegends.com/1740127778763-6fP9mQDN1c.png",
		4:  "https://mlbbmy.outweb.mobilelegends.com/1740127778941-iKmXrimSAq.png",
		5:  "https://mlbbmy.outweb.mobilelegends.com/1740127779144-Ec2sxUepi0.png",
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740501243702-BMnrKADB3y.png",
		7:  "https://mlbbmy.outweb.mobilelegends.com/1740127779630-CpJ6eb9Uvx.png",
		8:  "https://mlbbmy.outweb.mobilelegends.com/1740127779838-JcLUFASxHi.png",
		9:  "https://mlbbmy.outweb.mobilelegends.com/1740127780032-3ezQgoycmt.png",
		10: "https://mlbbmy.outweb.mobilelegends.com/1740127780217-SjhpkseDuh.png",
		11: "https://mlbbmy.outweb.mobilelegends.com/1740127780464-mSIaJF5iu6.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740501245088-QC10jFXRtr.png",
		13: "https://mlbbmy.outweb.mobilelegends.com/1740127781014-BXzTBc3obi.png",
		14: "https://mlbbmy.outweb.mobilelegends.com/1740127781253-3yClIQuiWg.png",
		15: "https://mlbbmy.outweb.mobilelegends.com/1740127781463-sZT0uVQ6LY.png",
	},
	"in": {
		1:  "https://mlbbmy.outweb.mobilelegends.com/1740127949327-C8TXvQrzr9.png",
		2:  "https://mlbbmy.outweb.mobilelegends.com/1740127949574-6981W3YBxN.png",
		3:  "https://mlbbmy.outweb.mobilelegends.com/1740127949781-MqpTPhm7v3.png",
		4:  "https://mlbbmy.outweb.mobilelegends.com/1740127949979-0Vrcx5geZx.png",
		5:  "https://mlbbmy.outweb.mobilelegends.com/1740127950199-r0PaW1d6ax.png",
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740501249277-EcsSDGQw9x.png",
		7:  "https://mlbbmy.outweb.mobilelegends.com/1740127950659-aVD30mCTJY.png",
		8:  "https://mlbbmy.outweb.mobilelegends.com/1740127950839-VQvjXOjpIt.png",
		9:  "https://mlbbmy.outweb.mobilelegends.com/1740127951014-TimrllTvzs.png",
		10: "https://mlbbmy.outweb.mobilelegends.com/1740127951200-Uz0b6OnSfh.png",
		11: "https://mlbbmy.outweb.mobilelegends.com/1740127951411-quwDDSp2sL.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740501250267-WKo8qvl5zS.png",
		13: "https://mlbbmy.outweb.mobilelegends.com/1740127951964-jQDuU0BJkt.png",
		14: "https://mlbbmy.outweb.mobilelegends.com/1740127952142-az0FAQQ0xu.png",
		15: "https://mlbbmy.outweb.mobilelegends.com/1740127952330-9L99vuql4D.png",
	},
	"en": {
		1:  "https://mlbbmy.outweb.mobilelegends.com/1740127978346-kd2gfKGjBp.png",
		2:  "https://mlbbmy.outweb.mobilelegends.com/1740127978622-qMZt43jdCX.png",
		3:  "https://mlbbmy.outweb.mobilelegends.com/1740127978915-n77rWFam58.png",
		4:  "https://mlbbmy.outweb.mobilelegends.com/1740127979152-4kPYtFhXQ4.png",
		5:  "https://mlbbmy.outweb.mobilelegends.com/1740127979414-nJrrNN0BWn.png",
		6:  "https://mlbbmy.outweb.mobilelegends.com/1740501247046-CRdjbxiOmU.png",
		7:  "https://mlbbmy.outweb.mobilelegends.com/1740127980028-pTBbQQTv3k.png",
		8:  "https://mlbbmy.outweb.mobilelegends.com/1740127980262-8dKnsyzkzT.png",
		9:  "https://mlbbmy.outweb.mobilelegends.com/1740127980586-QkoR3rd488.png",
		10: "https://mlbbmy.outweb.mobilelegends.com/1740127980838-daAKB4uIvq.png",
		11: "https://mlbbmy.outweb.mobilelegends.com/1740127981264-tMCVWNnpQS.png",
		12: "https://mlbbmy.outweb.mobilelegends.com/1740501248140-HeyenUQKz9.png",
		13: "https://mlbbmy.outweb.mobilelegends.com/1740127981906-mQ631G0P1X.png",
		14: "https://mlbbmy.outweb.mobilelegends.com/1740127982186-CFdLtT747a.png",
		15: "https://mlbbmy.outweb.mobilelegends.com/1740127982461-RzDSAPLSjK.png",
	},
}

func (u *ImageGenerate) ImageDowngrade(ctx context.Context, req *v1.SynthesisParamRequest, waId string) (string, error) {
	if req == nil {
		return "", nil
	}
	if req.ImagedDowngrade == 1 {
		u.l.Info(fmt.Sprintf("method[%s]，ImageDowngrade success req：%v", util.GetCurrentFuncName(), req))
		u.redisService.Set(constants.ImageDowngrade, "1")
	} else {
		u.l.Info(fmt.Sprintf("method[%s]，ImageDowngrade del req：%v", util.GetCurrentFuncName(), req))
		u.redisService.Del(constants.ImageDowngrade)
	}
	return "OK", nil
}

// 得到互动消息的图片url
func (u *ImageGenerate) GetInteractiveImageUrl(ctx context.Context, req *v1.SynthesisParamRequest, waId string) (string, error) {
	methodName := util.GetCurrentFuncName()
	u.l.Info(fmt.Sprintf("method[%s] 获取图片url,waId:%v,req:%v", methodName, waId, req))

	if req.NicknameList == nil || len(req.NicknameList) == 0 {
		u.l.Errorf("method[%s]，nickname is empty,req：%v", methodName, req)
		return "", errors.New("nickname is empty")
	}
	if req.CurrentProgress < 0 || req.CurrentProgress > 15 {
		u.l.Errorf("method[%s]，currentProgress is error,req：%v", methodName, req)
		return "", errors.New("currentProgress is error")
	}
	progress := req.CurrentProgress
	lang := langMap[req.LangNum]
	exits := u.redisService.Exits(constants.ImageDowngrade)
	if exits {
		s := imageMap[lang][progress]
		if s == "" {
			u.l.Errorf("method[%s]，ImageDowngrade error req：%v", methodName, req)
			return "", errors.New("helpCodeKey is not exists")
		}
		return s, nil
	}
	imageData, err := u.CreateProgressCoverReturnBytes(ctx, req, waId)
	if err != nil {
		return "", errors.New("create progress cover error")
	}
	compressImage, err := compressImage(imageData, 70)
	if err != nil {
		return "", errors.New("compressImage cover error")
	}

	imageUrl, err := u.GenerateImageAndUpload2S3WithBytes(compressImage, ctx, req, waId)
	if err != nil {
		u.l.Errorf("method[%s]，DealImagemethod，result：%v err:%v", methodName, "generate image error", err)
		return "", errors.New("generate image error")
	}
	u.l.Info(fmt.Sprintf("method[%s]，获取图片url,waId:%v,req:%v,imageUrl:%v", methodName, waId, req, imageUrl))
	return imageUrl, nil
}

func (u *ImageGenerate) GenerateImageAndUpload2S3WithBytes(imageData []byte, ctx context.Context, req *v1.SynthesisParamRequest, waId string) (string, error) {
	// 生成10个字符的随机字符串
	methodName := util.GetCurrentFuncName()
	randomString := generateRandomString(10)
	u.l.Info(fmt.Sprintf("method[%s]，获取图片url,waId:%v,req:%v", methodName, waId, req))
	// 组合时间戳、随机字符串和文件后缀
	fileName := fmt.Sprintf("%d-%s.png", time.Now().UnixNano()/int64(time.Millisecond), randomString)

	//获取预签名URL

	preSignParam := map[string]string{
		//"bucket": config.ApplicationConfig.S3Config.Bucket,
		"key": fileName,
	}

	preSignUrl, err := u.GetPreSignUrl(preSignParam)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "generate image error", err)
		return "", errors.New("get preSignUrl error")
	}
	u.l.Info(fmt.Sprintf("method[%s] 调用生成图片method-获取预签名-结束,waId:%v,req:%v time:%v", methodName, waId, req, time.Now().UnixNano()/int64(time.Millisecond)))
	u.l.Info(fmt.Sprintf("调用生成图片method-上传文件-开始,waId:%v,req:%v", waId, req))

	err = putObject2S3(preSignUrl, imageData)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "upload to s3 error", err)
		return "", errors.New("pre sign url upload to s3 error")
	}

	u.l.Info(fmt.Sprintf("调用生成图片method-上传文x件-结束,waId:%v,url:%v,req:%v time:%v", methodName, waId, req, time.Now().UnixNano()/int64(time.Millisecond)))
	return u.b.S3Config.DonAmin + fileName, nil
}

func (u *ImageGenerate) GenerateImageAndUpload2S3(coverFilePath string, ctx context.Context, req *v1.SynthesisParamRequest) (string, error) {
	methodName := util.GetCurrentFuncName()

	// 读取图片文件
	imageData, err := ioutil.ReadFile(coverFilePath)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "read file error", err)
		return "", errors.New("read file error")
	}

	// 生成10个字符的随机字符串
	randomString := generateRandomString(10)

	// 组合时间戳、随机字符串和文件后缀
	fileName := fmt.Sprintf("%d-%s.png", time.Now().UnixNano()/int64(time.Millisecond), randomString)

	//获取预签名URL

	preSignParam := map[string]string{
		//"bucket": config.ApplicationConfig.S3Config.Bucket,
		"key": fileName,
	}

	preSignUrl, err := u.GetPreSignUrl(preSignParam)
	if err != nil {
		u.l.Errorf("method[%s]，getPreSignUrl err:%v", methodName, err)
		return "", errors.New("get preSignUrl error")
	}

	err = putObject2S3(preSignUrl, imageData)
	if err != nil {
		u.l.Errorf("method[%s]， putObject2S3，err :%v", methodName, err)
		return "", errors.New("pre sign url upload to s3 error")
	}

	// 删除临时文件
	err = os.Remove(coverFilePath)
	if err != nil {
		u.l.Errorf("method[%s]， delete file error,err:%v", methodName, err)
		return "", errors.New("delete file error")
	}
	return u.b.S3Config.DonAmin + fileName, nil
}

var fontBytesMap = make(map[string][]byte)
var boldFontBytesMap = make(map[string][]byte)

func getFont(lang string) ([]byte, error) {
	if fontBytesMap[lang] != nil {
		return fontBytesMap[lang], nil
	}
	fontBytes, err := os.ReadFile(langFont[lang])
	if err != nil {
		return nil, err
	}
	fontBytesMap[lang] = fontBytes
	return fontBytes, nil
}

var httpTransportClient = http.Client{}

var preSignClient = http.Client{}

func InitImageService() {
	for lang, fontPath := range langFont {
		fontBytes, err := os.ReadFile(fontPath)
		if err != nil {
			// 处理错误，例如记录日志或者panic
			panic(err) // 这里选择panic，因为初始化失败是严重错误
		}
		fontBytesMap[lang] = fontBytes
	}

	for lang, fontPath := range langBoldFont {
		fontBytes, err := os.ReadFile(fontPath)
		if err != nil {
			// 处理错误，例如记录日志或者panic
			panic(err) // 这里选择panic，因为初始化失败是严重错误
		}
		boldFontBytesMap[lang] = fontBytes
	}

	httpTransportClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,              // 最大空闲连接数
			IdleConnTimeout:    10 * time.Second, // 空闲连接超时时间
			MaxConnsPerHost:    200,
			DisableCompression: true, // 禁用压缩，因为压缩和解压缩会消耗CPU资源
		},
	}

	preSignClient = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,              // 最大空闲连接数
			IdleConnTimeout:    20 * time.Second, // 空闲连接超时时间
			MaxConnsPerHost:    200,
			DisableCompression: true, // 禁用压缩，因为压缩和解压缩会消耗CPU资源
		},
	}
}

func getBoldFont(lang string) ([]byte, error) {
	if fontBytesMap[lang] != nil {
		return fontBytesMap[lang], nil
	}
	fontBytes, err := os.ReadFile(langBoldFont[lang])
	if err != nil {
		return nil, err
	}
	fontBytesMap[lang] = fontBytes
	return fontBytes, nil
}

func truncateString(s string, maxRunes int) string {
	// 将字符串转换为rune切片，以正确处理UTF-8字符
	runes := []rune(s)

	// 如果rune切片的长度小于或等于maxRunes，则直接返回原始字符串
	if len(runes) <= maxRunes {
		return s
	}

	// 截取前maxRunes个rune，并转换回字符串
	truncated := string(runes[:maxRunes])

	// 如果截取后的字符串长度小于原始字符串长度，追加"..."
	if utf8.RuneCountInString(truncated) < utf8.RuneCountInString(s) {
		truncated += "..."
	}

	return truncated
}

func (u *ImageGenerate) CreateProgressCoverReturnBytes(ctx context.Context, req *v1.SynthesisParamRequest, waId string) ([]byte, error) {
	methodName := util.GetCurrentFuncName()
	lang := langMap[req.LangNum]
	if lang == "en" || lang == "in" || lang == "my" {
		return u.CreateProgressCoverWithOTFReturnBytes(ctx, req)
	}

	u.l.Info(fmt.Sprintf("调用生成图片method-读取本地图片-开始,waId:%v,req:%v", waId, req))
	startTime := time.Now()

	file, err := os.Open(BASE_PATH + lang + "/banner" + fmt.Sprintf("%d", req.CurrentProgress) + ".jpg")
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "open file error", err)
		return nil, err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "decode jpeg error", err)
		return nil, err
	}
	u.l.Info(fmt.Sprintf("调用生成图片method-读取本地图片-结束,waId:%v,req:%v time:%v", waId, req, time.Since(startTime)))
	startTime = time.Now()
	u.l.Info(fmt.Sprintf("调用生成图片method-合成图片-开始,waId:%v,req:%v", waId, req))
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	leftText := coverCopywriting[lang]["left"]
	rightText := coverCopywriting[lang]["right"]

	regularFontBytes, err := getFont(lang)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "font file error", err)
		return nil, err
	}
	boldFontBytes, err := getBoldFont(lang)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "bold font file error", err)
		return nil, err
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return nil, err
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return nil, err
	}

	fontSize := 36.0
	//(139, 9, 50)。
	fixedColor := image.NewUniform(color.RGBA{R: 139, G: 9, B: 50, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	centerX, centerY := 570, 355

	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += leftWidth
	c.SetFont(boldFont)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += nicknameWidth
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, err
	}
	u.l.Info(fmt.Sprintf("调用生成图片method-合成图片-结束,waId:%v,req:%v time:%v", waId, req, time.Since(startTime)))
	return buf.Bytes(), nil
}

var langFileMap = map[string]string{
	"en": "/en-",
	"in": "/in-",
	"my": "/MY-",
}

func (u *ImageGenerate) CreateProgressCoverWithOTFReturnBytes(ctx context.Context, req *v1.SynthesisParamRequest) ([]byte, error) {
	methodName := util.GetCurrentFuncName()
	lang := langMap[req.LangNum]
	file, err := os.Open(BASE_PATH + lang + langFileMap[lang] + fmt.Sprintf("%02d", req.CurrentProgress) + ".jpg")
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "open file error", err)
		return nil, err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "decode jpeg error", err)
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	leftText := coverCopywriting[lang]["left"]
	rightText := coverCopywriting[lang]["right"]

	regularFontBytes, err := getFont(lang)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "font file error", err)
		return nil, err
	}
	boldFontBytes, err := getBoldFont(lang)
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "bold font file error", err)
		return nil, err
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return nil, err
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return nil, err
	}

	fontSize := 30.0
	fixedColor := image.NewUniform(color.RGBA{R: 139, G: 9, B: 50, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	centerX, centerY := 570, 355

	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += leftWidth
	c.SetFont(boldFont)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	startX += nicknameWidth
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 压缩图片并返回压缩后的图片数据
func compressImage(imageData []byte, quality int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	// 创建一个新的bytes.Buffer用于存储压缩后的图片数据
	var buf bytes.Buffer
	// 使用JPEG格式压缩图片，并设置压缩质量
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (u *ImageGenerate) CreateProgressCoverWithOTF(ctx context.Context, req *v1.SynthesisParamRequest) (string, error) {
	methodName := util.GetCurrentFuncName()
	// Open the base image file
	lang := langMap[req.LangNum]
	file, err := os.Open(BASE_PATH + lang + "/banner" + fmt.Sprintf("%d", req.CurrentProgress) + ".jpg")
	if err != nil {
		u.l.Errorf("method[%s]， generateImageAndUpload2S3，result ：%v err:%v", methodName, "open file error", err)
		return "", errors.New("open file error")
	}
	defer file.Close()
	var img image.Image
	img, err = jpeg.Decode(file)

	// Create a new RGBA image for drawing
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	// Get the nickname
	nicknames := req.NicknameList
	nickname := nicknames[req.CurrentProgress-1]
	nickname = truncateString(nickname, 8)

	// Left and right fixed copywriting
	leftText := coverCopywriting[lang]["left"]   // e.g., "YOUR FRIEND ["
	rightText := coverCopywriting[lang]["right"] // e.g., "] HAVE SUCCESSFULLY SUPPORTED YOU!"

	// Load fonts
	regularFontBytes, err := getFont(lang) // Font for left and right fixed copywriting
	if err != nil {
		u.l.Errorf("method[%s]， getFont，result ：%v err:%v", methodName, "font file error", err)
		return "", errors.New("font file error")
	}
	boldFontBytes, err := getBoldFont(lang) // Bold font for the nickname
	if err != nil {
		u.l.Errorf("method[%s]， getBoldFont，result ：%v err:%v", methodName, "bold font file error", err)
		return "", errors.New("bold font file error")
	}

	regularFont, err := freetype.ParseFont(regularFontBytes)
	if err != nil {
		return "", errors.New("parse font error for regular font")
	}
	boldFont, err := freetype.ParseFont(boldFontBytes)
	if err != nil {
		return "", errors.New("parse font error for bold font")
	}

	// Define font size and colors
	fontSize := 23.0
	fixedColor := image.NewUniform(color.RGBA{R: 139, G: 9, B: 50, A: 255})
	nicknameColor := image.NewUniform(color.RGBA{R: 255, G: 30, B: 30, A: 255})

	// Create a freetype context
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	// Define text positions
	centerX, centerY := 570, 355

	// Measure text width
	leftFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})
	boldFace := truetype.NewFace(boldFont, &truetype.Options{Size: fontSize})
	rightFace := truetype.NewFace(regularFont, &truetype.Options{Size: fontSize})

	drawer := font.Drawer{
		Face: leftFace,
	}
	leftWidth := drawer.MeasureString(leftText).Ceil()

	drawer.Face = boldFace
	nicknameWidth := drawer.MeasureString(nickname).Ceil()

	drawer.Face = rightFace
	rightWidth := drawer.MeasureString(rightText).Ceil()

	totalWidth := leftWidth + nicknameWidth + rightWidth

	// Calculate the starting coordinates
	startX := centerX - totalWidth/2
	startY := centerY + int(c.PointToFixed(fontSize)>>6)/2

	// Draw left copywriting
	c.SetFont(regularFont)
	c.SetFontSize(fontSize)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(leftText, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw left text error")
	}

	// Update startX for nickname
	startX += leftWidth

	// Draw nickname
	c.SetFont(boldFont)
	c.SetFontSize(fontSize)
	c.SetSrc(nicknameColor)
	_, err = c.DrawString(nickname, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw nickname error")
	}

	// Update startX for right text
	startX += nicknameWidth

	// Draw right copywriting
	c.SetFont(regularFont)
	c.SetSrc(fixedColor)
	_, err = c.DrawString(rightText, freetype.Pt(startX, startY))
	if err != nil {
		return "", errors.New("draw right text error")
	}

	// Save the output image
	tmpPath := BASE_PATH + generateRandomString(10) + ".png"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return "", errors.New("error creating output file")
	}
	defer outFile.Close()

	err = png.Encode(outFile, rgba)
	if err != nil {
		return "", errors.New("error encoding image")
	}
	return tmpPath, nil
}

const RandomStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// 生成随机的字母和数字序列
func generateRandomString(n int) string {
	var letters = []rune(RandomStr)
	s := make([]rune, n)
	for i := range s {
		b, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		s[i] = letters[b.Int64()]
	}
	return string(s)
}

func (u *ImageGenerate) GetPreSignUrl(bodyData map[string]string) (string, error) {
	bodyBytes, err := json.NewEncoder().Encode(bodyData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error encoding body data: %v", err))
	}
	// 创建请求
	req, err := http.NewRequest("POST", u.b.S3Config.PreSignUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error creating request: %v", err))
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := preSignClient
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error sending request: %v", err))
	}
	defer resp.Body.Close()
	// 读取响应体
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resultResponse := &PreSignResponse{}
	err = json.NewEncoder().Decode(responseBody, resultResponse)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error decoding response: %v", err))
	}
	if resultResponse.Code != 200 {
		return "", errors.New(fmt.Sprintf("调用getPreSignUrl,返回结果失败，报错信息: %v", resultResponse.Message))
	}
	return resultResponse.Data.(string), nil
}

// PutObject2S3 使用预签名 URL 上传文件到 S3
func putObject2S3(preSignUrl string, fileData []byte) error {
	// 使用预签名 URL 上传文件
	req, err := http.NewRequest("PUT", preSignUrl, bytes.NewReader(fileData))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create request: %v", err))
	}

	req.Header.Set("Content-Type", "image/png") // 设置文件类型

	client := httpTransportClient
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to upload file: %v", err))
	}
	defer resp.Body.Close()

	// 检查上传响应
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Failed to upload file, status: %s, body: %s", resp.Status, body))
	}
	return nil
}
