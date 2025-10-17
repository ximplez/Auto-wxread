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
	IsEndPage            func(ctx context.Context) (bool, error)
}

type doQueryAction struct {
	do func(ctx context.Context) error
}

func (e *doQueryAction) Do(ctx context.Context) error {
	if e.do != nil {
		return e.do(ctx)
	}
	return nil
}

var (
	emptyQuery = &doQueryAction{}
	clean      = &doQueryAction{do: cleanAction}
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

func cleanAction(ctx context.Context) error {
	if err := chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {
    get: () => false,
	configurable: true,
  });`, nil).Do(ctx); err != nil {
		return err
	}
	if err := chromedp.Evaluate(`window.navigator.chrome = {
    app: {
        isInstalled: false,
        InstallState: {
            DISABLED: "disabled",
            INSTALLED: "installed",
            NOT_INSTALLED: "not_installed"
        },
        RunningState: {
            CANNOT_RUN: "cannot_run",
            READY_TO_RUN: "ready_to_run",
            RUNNING: "running"
        }
    }
  };`, nil).Do(ctx); err != nil {
		return err
	}
	// 	if err := chromedp.Evaluate(`await page.evaluateOnNewDocument(() => {
	//   const originalQuery = window.navigator.permissions.query;
	//   return window.navigator.permissions.query = (parameters) => (
	//     parameters.name === 'notifications' ?
	//       Promise.resolve({ state: Notification.permission }) :
	//       originalQuery(parameters)
	//   );
	// });`, nil).Do(ctx); err != nil {
	// 		return err
	// 	}
	if err := chromedp.Evaluate(`Object.defineProperty(navigator, 'languages', {
    get: () => ['zh-CN', 'zh'],
	configurable: true,
  });`, nil).Do(ctx); err != nil {
		return err
	}
	if err := chromedp.Evaluate(`Object.defineProperty(navigator, 'plugins', {
    get: () => [1, 2, 3, 4, 5,6],
	configurable: true,
  });`, nil).Do(ctx); err != nil {
		return err
	}
	return nil
}
