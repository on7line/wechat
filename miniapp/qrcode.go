package miniapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/silenceper/wechat/util"
)

const codeURL = "https://api.weixin.qq.com/wxa/getwxacode?access_token=%s"
const codeUnlimitURL = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s"

type WxaCode struct {
	Path      string `json:"path,omitempty"`
	Page      string `json:"page,omitempty"`
	Scene     string `json:"scene,omitempty"` // 永久二维码，path不能带参数
	Width     int    `json:"width,omitempty"`
	AutoColor bool   `json:"auto_color"`
	LineColor struct {
		R string `json:"r"`
		G string `json:"g"`
		B string `json:"b"`
	} `json:"line_color,omitempty"`
	IsHyaline bool `json:"is_hyaline,omitempty"` // 是否需要透明底色， is_hyaline 为true时，生成透明底色的小程序码
}

// https://developers.weixin.qq.com/miniprogram/dev/api/open-api/qr-code/createWXAQRCode.html
// WxaCode 生成小程序二维码
func (app *MiniAPP) WxaCode(param WxaCode, tmp bool) (wxacode []byte, err error) {
	var u string
	var accessToken string
	accessToken, err = app.GetAccessToken()
	if err != nil {
		return
	}
	if tmp {
		u = fmt.Sprintf(codeURL, accessToken)
	} else {
		param.Page = param.Path
		param.Path = ""
		u = fmt.Sprintf(codeUnlimitURL, accessToken)
	}

	response, err := util.PostJSON(u, param)
	if err != nil {
		return
	}

	// 判断是否是json
	if bytes.Index(response, []byte("{")) == 0 {
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
	}
	return response, err
}

// SaveQrcodeImage 保存qrcode图片
func (app *MiniAPP) SaveQrcodeImage(path, filename string, src io.Reader) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		fmt.Println(err.Error())
		return err
	}

	out, err := os.Create(fmt.Sprintf("%s/%s", path, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
