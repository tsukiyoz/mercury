package ioc

import (
	"context"
	"os"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/lazywoo/mercury/internal/payment/events"
	"github.com/lazywoo/mercury/internal/payment/repository"
	"github.com/lazywoo/mercury/internal/payment/service/wechat"
	"github.com/lazywoo/mercury/pkg/logger"
)

type WechatConfig struct {
	AppID        string
	MchID        string
	MchKey       string
	MchSerialNum string
	CertPath     string
	KeyPath      string
}

func InitWechatConfig() WechatConfig {
	return WechatConfig{
		AppID:        os.Getenv("WECHAT_APP_ID"),
		MchID:        os.Getenv("WECHAT_MCH_ID"),
		MchKey:       os.Getenv("WECHAT_MCH_KEY"),
		MchSerialNum: os.Getenv("WECHAT_MCH_SERIAL_NUM"),
		CertPath:     "./config/cert/apiclient_cert.pem",
		KeyPath:      "./config/cert/apiclient_key.pem",
	}
}

func InitWechatClient(cfg WechatConfig) *core.Client {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(cfg.KeyPath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client, err := core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(
		cfg.MchID, cfg.MchSerialNum, mchPrivateKey, cfg.MchKey,
	))
	if err != nil {
		panic(err)
	}
	return client
}

func InitWechatNativeService(
	cli *core.Client,
	repo repository.PaymentRepository,
	l logger.Logger,
	producer events.Producer,
	cfg WechatConfig,
) *wechat.NativePaymentService {
	return wechat.NewNativePaymentService(
		&native.NativeApiService{
			Client: cli,
		},
		cfg.AppID, cfg.MchID, repo, l, producer,
	)
}

func InitWechatNotifyHandler(cfg WechatConfig) *notify.Handler {
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
	hdl, err := notify.NewRSANotifyHandler(cfg.MchKey, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	if err != nil {
		panic(err)
	}
	return hdl
}
