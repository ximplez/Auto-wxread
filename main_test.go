package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ximplez/wxread/device_cfg"
	"github.com/ximplez/wxread/notify"
)

type recordNotifier struct {
	upserts  []notify.CardMessage
	progress []notify.CardMessage
}

func (r *recordNotifier) Upsert(message notify.CardMessage) {
	r.upserts = append(r.upserts, message)
}

func (r *recordNotifier) NotifyProgress(message notify.CardMessage) {
	r.progress = append(r.progress, message)
}

func resetWxReadTestState(t *testing.T) *recordNotifier {
	t.Helper()

	oldBookTitle := bookTitle
	oldTargetReadTime := targetReadTime
	oldCookies := cookies
	oldDebug := debug
	oldDeviceCfg := deviceCfg
	oldFinishedBook := finishedBook
	oldTotalReadTime := totalReadTime
	oldTotalReadPageCnt := totalReadPageCnt
	oldCurCatalog := curCatalog
	oldNotifier := notifier
	oldAccessWebWithCtx := accessWebWithCtx
	oldCheckLogin := checkLogin
	oldDoLoginAction := doLoginAction

	recorder := &recordNotifier{}
	bookTitle = "book"
	targetReadTime = time.Minute
	cookies = ""
	debug = false
	deviceCfg = device_cfg.IPadPro
	finishedBook = false
	totalReadTime = 0
	totalReadPageCnt = 0
	curCatalog = nil
	notifier = recorder
	accessWebWithCtx = func(context.Context, chromedp.Tasks, bool) error {
		return nil
	}
	checkLogin = func(context.Context) (bool, error) {
		return true, nil
	}
	doLoginAction = func() chromedp.ActionFunc {
		return func(context.Context) error {
			return nil
		}
	}

	t.Cleanup(func() {
		bookTitle = oldBookTitle
		targetReadTime = oldTargetReadTime
		cookies = oldCookies
		debug = oldDebug
		deviceCfg = oldDeviceCfg
		finishedBook = oldFinishedBook
		totalReadTime = oldTotalReadTime
		totalReadPageCnt = oldTotalReadPageCnt
		curCatalog = oldCurCatalog
		notifier = oldNotifier
		accessWebWithCtx = oldAccessWebWithCtx
		checkLogin = oldCheckLogin
		doLoginAction = oldDoLoginAction
	})

	return recorder
}

func TestNotifyWxReadWrappersBuildCards(t *testing.T) {
	recorder := resetWxReadTestState(t)

	notifyWxRead(notify.WxReadStatusStarting, wxReadState(0, "", "starting"))
	notifyWxReadProgress(wxReadState(30*time.Second, "", "reading"))

	assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusStarting)})
	assertStatuses(t, recorder.progress, []string{string(notify.WxReadStatusReading)})
}

func TestAccessWebChromeStartFailureUpdatesLoadingAndFailedOnly(t *testing.T) {
	recorder := resetWxReadTestState(t)
	accessWebWithCtx = func(context.Context, chromedp.Tasks, bool) error {
		return errors.New("chrome failed to start:\nexit status 1")
	}

	err := accessWeb()

	if err == nil || !strings.Contains(err.Error(), "chrome failed to start") {
		t.Fatalf("accessWeb() err = %v, want chrome failed to start", err)
	}
	assertStatuses(t, recorder.upserts, []string{
		string(notify.WxReadStatusLoading),
		string(notify.WxReadStatusFailed),
	})
	assertStatuses(t, recorder.progress, nil)
}

func TestAccessWebInProgressFailureUpdatesWarningThenFailed(t *testing.T) {
	recorder := resetWxReadTestState(t)
	totalReadPageCnt = 3
	accessWebWithCtx = func(context.Context, chromedp.Tasks, bool) error {
		return errors.New("page action failed")
	}

	err := accessWeb()

	if err == nil {
		t.Fatal("accessWeb() err = nil, want error")
	}
	assertStatuses(t, recorder.upserts, []string{
		string(notify.WxReadStatusLoading),
		string(notify.WxReadStatusProgressWarning),
		string(notify.WxReadStatusFailed),
	})
}

func TestAccessWebDeadlineExceededUpdatesFinished(t *testing.T) {
	recorder := resetWxReadTestState(t)
	accessWebWithCtx = func(context.Context, chromedp.Tasks, bool) error {
		return context.DeadlineExceeded
	}

	if err := accessWeb(); err != nil {
		t.Fatalf("accessWeb() err = %v, want nil", err)
	}
	assertStatuses(t, recorder.upserts, []string{
		string(notify.WxReadStatusLoading),
		string(notify.WxReadStatusFinished),
	})
}

func TestFindBookUpdatesBookFound(t *testing.T) {
	recorder := resetWxReadTestState(t)
	deviceCfg.FindBookAndClick = func(context.Context, string) (string, error) {
		return "found book", nil
	}

	if err := findBook().Do(context.Background()); err != nil {
		t.Fatalf("findBook() err = %v", err)
	}
	assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusBookFound)})
	if bookTitle != "found book" {
		t.Fatalf("bookTitle = %s, want found book", bookTitle)
	}
}

func TestBeforeReadUpdatesReady(t *testing.T) {
	recorder := resetWxReadTestState(t)
	deviceCfg.BeforeRead = func(context.Context) error {
		return nil
	}

	if err := beforeRead().Do(context.Background()); err != nil {
		t.Fatalf("beforeRead() err = %v", err)
	}
	assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusReady)})
}

func TestLoginUpdatesSuccessForCookieAndQRCodeFlows(t *testing.T) {
	t.Run("cookie login", func(t *testing.T) {
		recorder := resetWxReadTestState(t)
		checkLogin = func(context.Context) (bool, error) {
			return true, nil
		}

		if err := login().Do(context.Background()); err != nil {
			t.Fatalf("login() err = %v", err)
		}
		assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusLoginSuccess)})
	})

	t.Run("qrcode login", func(t *testing.T) {
		recorder := resetWxReadTestState(t)
		checkLogin = func(context.Context) (bool, error) {
			return false, nil
		}
		doLoginAction = func() chromedp.ActionFunc {
			return func(context.Context) error {
				return nil
			}
		}

		if err := login().Do(context.Background()); err != nil {
			t.Fatalf("login() err = %v", err)
		}
		assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusLoginSuccess)})
	})
}

func TestRenderLoginUpdatesLoginRequiredWithQRCode(t *testing.T) {
	recorder := resetWxReadTestState(t)
	deviceCfg.FetchLoginQrCode = func(context.Context) (string, error) {
		return "encoded-qrcode", nil
	}

	if err := renderLogin(context.Background()); err != nil {
		t.Fatalf("renderLogin() err = %v", err)
	}
	assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusLoginRequired)})
	if got := recorder.upserts[0].SubButtonURL; !strings.Contains(got, "encoded-qrcode") {
		t.Fatalf("SubButtonURL = %s, want qrcode URL", got)
	}
}

func TestStartReadUpdatesReadingAndProgress(t *testing.T) {
	recorder := resetWxReadTestState(t)
	endChecks := 0
	deviceCfg.GetCatalogInfo = func(context.Context) (*device_cfg.CatalogInfo, error) {
		return &device_cfg.CatalogInfo{}, nil
	}
	deviceCfg.StartRead = func(context.Context) error {
		return nil
	}
	deviceCfg.IsEndPage = func(context.Context) (bool, error) {
		endChecks++
		return endChecks >= 2, nil
	}
	deviceCfg.NextPage = func(context.Context) error {
		return nil
	}

	if err := startRead().Do(context.Background()); err != nil {
		t.Fatalf("startRead() err = %v", err)
	}
	assertStatuses(t, recorder.upserts, []string{string(notify.WxReadStatusReading)})
	assertStatuses(t, recorder.progress, []string{string(notify.WxReadStatusReading)})
	if !finishedBook {
		t.Fatal("finishedBook = false, want true")
	}
	if totalReadPageCnt != 1 {
		t.Fatalf("totalReadPageCnt = %d, want 1", totalReadPageCnt)
	}
}

func TestStartReadFailureUpdatesProgressWarning(t *testing.T) {
	recorder := resetWxReadTestState(t)
	deviceCfg.GetCatalogInfo = func(context.Context) (*device_cfg.CatalogInfo, error) {
		return &device_cfg.CatalogInfo{}, nil
	}
	deviceCfg.StartRead = func(context.Context) error {
		return errors.New("read failed")
	}

	err := startRead().Do(context.Background())

	if err == nil {
		t.Fatal("startRead() err = nil, want error")
	}
	assertStatuses(t, recorder.upserts, []string{
		string(notify.WxReadStatusReading),
		string(notify.WxReadStatusProgressWarning),
	})
}

func assertStatuses(t *testing.T, messages []notify.CardMessage, want []string) {
	t.Helper()

	var got []string
	for _, message := range messages {
		got = append(got, message.Status)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("statuses = %#v, want %#v", got, want)
	}
}
