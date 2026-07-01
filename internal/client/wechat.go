package client

import (
	"context"
	"errors"
	"net/http"

	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3/log"
	"github.com/natholdallas/natools4go/jsons"
	"github.com/natholdallas/natools4go/spew"
	"github.com/shopspring/decimal"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

var (
	wechatInstance *core.Client
	wechatHandler  *notify.Handler
)

func InitWechat() {
	// Load merchant private key from local file using utils, used to sign requests
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(conf.App.WxAPIClientKeyPem)
	if err != nil {
		log.Fatal("load merchant private key error: ", err)
		return
	}
	wechatpayPublicKey, err := utils.LoadPublicKeyWithPath(conf.App.WxPubKeyPem)
	if err != nil {
		log.Fatal("load merchant public key error: ", err)
		return
	}
	client, err := core.NewClient(context.Background(), option.WithWechatPayPublicKeyAuthCipher(
		conf.App.WxMch,
		conf.App.WxCert,
		mchPrivateKey,
		conf.App.WxPubKey,
		wechatpayPublicKey,
	))
	if err != nil {
		log.Fatal("new wechat pay client err: %s", err)
		return
	}
	wechatInstance = client
	wechatHandler = notify.NewNotifyHandler(
		conf.App.WxV3Sercret,
		verifiers.NewSHA256WithRSAPubkeyVerifier(conf.App.WxPubKey, *wechatpayPublicKey),
	)
}

type WxLoginRes struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func WxLogin(code string) (v WxLoginRes, err error) {
	d, _ := client.R().
		SetQueryParams(map[string]string{
			"appid":      conf.App.WxAppID,
			"secret":     conf.App.WxSecret,
			"js_code":    code,
			"grant_type": "authorization_code",
		}).
		Get(conf.App.WxSite + "/sns/jscode2session")
	v = jsons.IUnmarshal[WxLoginRes](d.Bytes())
	if v.ErrCode != 0 {
		return v, errors.New(v.ErrMsg)
	}
	return v, nil
}

type WxUser struct {
	Subscribe     int    `json:"subscribe"`
	Openid        string `json:"openid"`
	Nickname      string `json:"nickname"`
	Sex           int    `json:"sex"`
	Language      string `json:"language"`
	City          string `json:"city"`
	Province      string `json:"province"`
	Country       string `json:"country"`
	Headimgurl    string `json:"headimgurl"`
	SubscribeTime int    `json:"subscribe_time"`
	Unionid       string `json:"unionid"`
	ErrCode       int    `json:"errcode"`
	ErrMsg        string `json:"errmsg"`
}

func WxGetUserInfo(token, openid string) (v WxUser, err error) {
	// https://api.weixin.qq.com/cgi-bin/user/info?access_token=ACCESS_TOKEN&openid=OPENID&lang=zh_CN
	d, _ := client.R().
		SetQueryParams(map[string]string{
			"access_token": token,
			"openid":       openid,
			"lang":         "zh_CN",
		}).
		Get(conf.App.WxSite + "/cgi-bin/user/info")
	v = jsons.IUnmarshal[WxUser](d.Bytes())
	if v.ErrCode != 0 {
		return v, errors.New(v.ErrMsg)
	}
	return v, nil
}

var WxClientCredential = "client_credential"

type WxToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

func WxGetAccessToken(gtype string) (v WxToken, err error) {
	// https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET
	d, _ := client.R().
		SetQueryParams(map[string]string{
			"grant_type": gtype,
			"appid":      conf.App.WxAppID,
			"secret":     conf.App.WxSecret,
		}).
		Get(conf.App.WxSite + "/cgi-bin/token")
	v = jsons.IUnmarshal[WxToken](d.Bytes())
	if v.ErrCode != 0 {
		return v, errors.New(v.ErrMsg)
	}
	return v, nil
}

func WxPay(openid string, amount int64, tradeNo string) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	if conf.App.Debug {
		amount = 1
	}
	svc := jsapi.JsapiApiService{Client: wechatInstance}
	req := jsapi.PrepayRequest{
		Appid:       core.String(conf.App.WxAppID),
		Mchid:       core.String(conf.App.WxMch),
		Description: core.String("none"),
		OutTradeNo:  core.String(tradeNo),
		NotifyUrl:   core.String(conf.App.WxWebhook),
		Amount:      &jsapi.Amount{Total: core.Int64(amount)},
		Payer:       &jsapi.Payer{Openid: core.String(openid)},
	}
	resp, _, err := svc.PrepayWithRequestPayment(context.Background(), req)
	spew.Dump(resp, req)
	return resp, err
}

func WxVerify(request *http.Request) (*payments.Transaction, error) {
	v := new(payments.Transaction)
	_, err := wechatHandler.ParseNotifyRequest(context.Background(), request, &v)
	return v, err
}

// WxAmount The input is always in USD, will be converted to CNY
func WxAmount(amount decimal.Decimal, rate float64) int64 {
	return amount.
		Mul(decimal.NewFromFloat(rate)).
		Mul(decimal.NewFromInt(100)).
		Ceil().
		IntPart()
}
