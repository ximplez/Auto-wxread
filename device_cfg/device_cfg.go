package device_cfg

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/chromedp/chromedp"
)

type DeviceCfg struct {
	chromedp.Device
	AfterNavigate        chromedp.QueryAction
	BeforeClickLogin     chromedp.QueryAction
	ClickLogin           chromedp.QueryAction
	FetchLoginQrCode     func(ctx context.Context) (string, error)
	IsInvalidLoginQrCode func(ctx context.Context) (bool, error)
	RefreshLoginQrCode   func(ctx context.Context) error
	FindBookAndClick     func(ctx context.Context, bookName string) (string, error)
	BeforeRead           func(ctx context.Context) error
	StartRead            func(ctx context.Context) error
	NextPage             func(ctx context.Context) error
}

type emptyQueryAction struct{}

func (e *emptyQueryAction) Do(ctx context.Context) error {
	return nil
}

var (
	emptyQuery = &emptyQueryAction{}
)

func random(min, max int64) int64 {
	if min >= max {
		return 0
	}
	return min + rand.N[int64](max-min)
}

func randomReadTime(min, max int64) time.Duration {
	return time.Duration(random(min, max)) * time.Second
}
