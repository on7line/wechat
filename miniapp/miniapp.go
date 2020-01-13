package miniapp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/silenceper/wechat/context"
	"github.com/silenceper/wechat/util"
)

const (
	templateSendURL               = "https://api.weixin.qq.com/cgi-bin/message/wxopen/template/send"
	sessionkeyURL                 = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"
	getAnalysisDailyVisitTrendURL = "https://api.weixin.qq.com/datacube/getweanalysisappiddailyvisittrend"
)

//MiniAPP 模板消息
type MiniAPP struct {
	*context.Context
}

//NewMiniApp 实例化
func NewMiniApp(context *context.Context) *MiniAPP {
	app := new(MiniAPP)
	app.Context = context
	return app
}

// WXAppSession 微信小程序会话
type WXAppSession struct {
	ErrCode    int
	ErrMsg     string
	OpenID     string
	SessionKey string `json:"session_key"`
}

// WXAnalysisItem 数据
type WXAnalysisItem struct {
	RefDate         string  `json:"ref_date"`          // ref_date	string	日期，格式为 yyyymmdd
	SessionCNT      float64 `json:"session_cnt"`       //	number	打开次数
	VisitPV         float64 `json:"visit_pv"`          //	number	访问次数
	VisitUV         float64 `json:"visit_uv"`          //	number	访问人数
	VisitUVNew      float64 `json:"visit_uv_new"`      // number	新用户数
	StayTimeUV      float64 `json:"stay_time_uv"`      // number	人均停留时长 (浮点型，单位：秒)
	StayTimeSession float64 `json:"stay_time_session"` // number	次均停留时长 (浮点型，单位：秒)
	VisitDepth      float64 `json:"visit_depth"`       // number	平均访问深度 (浮点型)
}

// WXAnalysis 微信小程序会话
type WXAnalysis struct {
	ErrCode int
	ErrMsg  string
	List    []WXAnalysisItem `json:"list"`
}

//DataItem 模版内某个 .DATA 的值
type DataItem struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// MiniMessage 微信小程序模板消息
type MiniMessage struct {
	ToUser     string               `json:"touser"`
	TemplateID string               `json:"template_id"`
	Page       string               `json:"page"`
	FormID     string               `json:"form_id"`          // Prepayid or form id
	Keyword    string               `json:"emphasis_keyword"` // 模板需要放大的关键词，不填则默认无放大
	Data       map[string]*DataItem `json:"data"`
	Color      string               `json:"color"` // 模板内容字体的颜色，不填默认黑色
}

// NewMiniMessage 创建小程序模板消息
func NewMiniMessage(openid, templateid, formid string) *MiniMessage {
	return &MiniMessage{
		ToUser:     openid,
		TemplateID: templateid,
		FormID:     formid,
		Data:       make(map[string]*DataItem),
	}
}

// NewMiniMessageEx 创建小程序模板消息加强版
func NewMiniMessageEx(openid, templateid, formid string, msgs []string) *MiniMessage {
	msg := &MiniMessage{
		ToUser:     openid,
		TemplateID: templateid,
		FormID:     formid,
		Data:       make(map[string]*DataItem),
	}
	for i, v := range msgs {
		key := fmt.Sprintf("keyword%d", i+1)
		msg.Data[key] = &DataItem{
			Value: v,
		}
	}
	return msg
}

func (app *MiniAPP) GetAnalysisDailyVisitTrend(sDate string) (result WXAnalysis, err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken()
	if err != nil {
		return
	}

	args := struct {
		BeginDate string `json:"begin_date"`
		EndDate   string `json:"end_date"`
	}{}
	args.BeginDate = sDate
	args.EndDate = sDate
	uri := fmt.Sprintf("%s?access_token=%s", getAnalysisDailyVisitTrendURL, accessToken)
	fmt.Println("GetAnalysisDailyVisitTrend:", uri, args)
	response, err := util.PostJSON(uri, args)
	if err != nil {
		return
	}

	err = json.Unmarshal(response, &result)
	if err != nil {
		return result, err
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}

	return
}

// SendTemplate 发送小程序模板消息
func (app *MiniAPP) SendTemplate(msg *MiniMessage) (err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken()
	if err != nil {
		return
	}
	uri := fmt.Sprintf("%s?access_token=%s", templateSendURL, accessToken)
	fmt.Println("send template:", uri, msg)
	response, err := util.PostJSON(uri, msg)
	if err != nil {
		return
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

// GetSessionKey 获取小程序session key
func (app *MiniAPP) GetSessionKey(code string) (result WXAppSession, err error) {
	uri := fmt.Sprintf(sessionkeyURL, app.AppID, app.AppSecret, code)
	response, err := util.PostJSON(uri, nil)
	if err != nil {
		return
	}

	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("template msg send error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

// WXAppSign 小程序签名验证
func (app *MiniAPP) WXAppSign(rawdata, sessionkey string) string {
	var cipherTxt bytes.Buffer
	cipherTxt.WriteString(rawdata)
	cipherTxt.WriteString(sessionkey)
	return util.SHA1(cipherTxt.Bytes())
}

// WXAppDecript 小程序解密
func (app *MiniAPP) WXAppDecript(crypted, sessionkey, iv string) ([]byte, error) {
	cryptedByte, _ := base64.StdEncoding.DecodeString(crypted)
	key, _ := base64.StdEncoding.DecodeString(sessionkey)
	ivbyte, _ := base64.StdEncoding.DecodeString(iv)
	return util.AESDecrypt(cryptedByte, key, ivbyte)
}

type wxSexType byte

func (t wxSexType) String() string {
	switch t {
	case 1:
		return "M"
	case 2:
		return "F"
	default:
		return "U"
	}
}

// WXUserInfo 微信小程序用户信息
type WXUserInfo struct {
	OpenID     string    `json:"openid"`
	NickName   string    `json:"nickname"`
	Gender     wxSexType `json:"gender"`
	Language   string    `json:"language"`
	City       string    `json:"city"`
	Province   string    `json:"province"`
	Country    string    `json:"country"`
	HeadImgURL string    `json:"avatarUrl"`
	UnionID    string    `json:"unionId"`
	WaterMark  struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// GetUserInfo 获取微信用户信息
func (app *MiniAPP) GetUserInfo(iv, cipherTxt, sessionKey string) (WXUserInfo, error) {
	data, err := app.WXAppDecript(cipherTxt, sessionKey, iv)
	if err != nil {
		return WXUserInfo{}, err
	}
	var user WXUserInfo
	err = json.Unmarshal(data, &user)
	return user, err
}

// WXPhoneInfo 微信账号绑定电话信息
type WXPhoneInfo struct {
	Phone     string `json:"phoneNumber"`
	PurePhone string `json:"purePhoneNumber"`
	Country   string `json:"countryCode"`
	WaterMark struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// GetPhoneNumber 获取微信绑定电话号码
func (app *MiniAPP) GetPhoneNumber(iv, cipherTxt, sessionKey string) (WXPhoneInfo, error) {
	data, err := app.WXAppDecript(cipherTxt, sessionKey, iv)
	if err != nil {
		return WXPhoneInfo{}, err
	}
	var phone WXPhoneInfo
	err = json.Unmarshal(data, &phone)
	return phone, err
}
