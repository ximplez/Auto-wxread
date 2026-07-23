package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/schollz/progressbar/v3"
	tool_chromedp "github.com/ximplez/wxread/chromedp"
	"github.com/ximplez/wxread/device_cfg"
	"github.com/ximplez/wxread/notify"
	"github.com/ximplez/wxread/utils/env_utils"
)

var (
	url = "https://weread.qq.com/"
	// 书标题
	bookTitle string
	// 目标阅读时间
	targetReadTime time.Duration
	// cookies or cookies url
	cookies string
	// debug模式
	debug bool

	bar          *progressbar.ProgressBar
	deviceCfg    = device_cfg.IPadPro
	finishedBook bool
	// 最终阅读时间
	totalReadTime int64
	// 总阅读页数
	totalReadPageCnt int64
	// 当前阅读章节
	curCatalog *device_cfg.CatalogInfo
	notifier   *notify.FeishuCardNotifier
)

const (
	envCookiesKey = "COOKIES"
	envBookTitle  = "BOOK_NAME"
)

func main() {
	tt := flag.Int64("t", 5, "目标阅读时间(分钟)")
	flag.BoolFunc("debug", "开启debug模式", func(s string) error {
		debug = true
		return nil
	})
	flag.Parse()
	if tt == nil || *tt <= 0 {
		log.Fatalln("targetTime 非法")
	}
	bookTitle = env_utils.GetEnv(envBookTitle)
	cookies = env_utils.GetEnv(envCookiesKey)
	notifier = notify.NewFeishuCardNotifier(notify.NewCardConfigFromEnv(env_utils.GetEnv))
	targetReadTime = time.Minute * time.Duration(*tt)
	log.Printf("目标阅读时间: %s, 目标书名: %s", targetReadTime.String(), bookTitle)
	notifyWxRead(notify.WxReadStatusStarting, wxReadState(0, "", "任务已创建，正在准备浏览器环境。"))

	// 访问网页
	err := accessWeb()
	if err != nil {
		log.Fatalf("err: %v", err)
	}
}

func accessWeb() error {
	ctx, cancel := context.WithTimeout(context.Background(), targetReadTime)
	defer cancel()
	notifyWxRead(notify.WxReadStatusLoading, wxReadState(0, "", "正在打开微信读书并恢复 cookies。"))
	err := tool_chromedp.AccessWebWithCtx(ctx, chromedp.Tasks{
		// 设置设备模拟
		chromedp.Emulate(deviceCfg.Device),
		loadCookies(),
		// 页面导航
		chromedp.Navigate(url),
		deviceCfg.AfterNavigate,
		login(),
		saveCookies(),
		findBook(),
		beforeRead(),
		startRead(),
	}, debug)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			end()
			return nil
		}
		if totalReadPageCnt > 0 || curCatalog != nil {
			notifyWxRead(notify.WxReadStatusProgressWarning, wxReadState(0, err.Error(), "阅读过程中出现异常，已记录当前进度。"))
		}
		notifyWxRead(notify.WxReadStatusFailed, wxReadState(0, err.Error(), "任务已停止。"))
		return err
	}
	if finishedBook {
		end()
	}
	return nil
}

func end() {
	finishedText := ""
	if finishedBook {
		finishedText = "全书阅读完毕"
	}
	atc := 0
	if totalReadPageCnt == 0 {
		atc = 0
	} else {
		atc = int(totalReadTime / 1000 / totalReadPageCnt)
	}
	catalogStr := ""
	if curCatalog != nil {
		catalogStr = fmt.Sprintf(`
	当前章节: %s
	当前进度: %s
`, curCatalog.CurCatalog(), curCatalog.CurProgress())
	}
	summary := fmt.Sprintf(`书名: %s %s
本次阅读时间: %s
本次阅读页数: %d 页
本次平均阅读时间: %d 秒
阅读进度: %s`, bookTitle, finishedText,
		(time.Millisecond * time.Duration(totalReadTime)).String(), totalReadPageCnt, atc, catalogStr)
	log.Print(summary)
	notifyWxRead(notify.WxReadStatusFinished, wxReadState(0, "", "阅读总结已生成。"))
}
func findBook() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if book, err := deviceCfg.FindBookAndClick(ctx, bookTitle); err != nil {
			return err
		} else {
			if book == "" {
				return fmt.Errorf("❌ 未找到书: %s", bookTitle)
			}
			log.Printf("✅ 找到书: %s", book)
			bookTitle = book
			notifyWxRead(notify.WxReadStatusBookFound, wxReadState(0, "", "目标书籍已定位。"))
		}
		return nil
	}
}

func beforeRead() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Printf("📕书名: %s，目标阅读时间: %v", bookTitle, targetReadTime.String())
		if err := deviceCfg.BeforeRead(ctx); err != nil {
			return err
		}
		notifyWxRead(notify.WxReadStatusReady, wxReadState(0, "", "阅读页加载完成。"))
		return nil
	}
}

// 检查是否登陆
func login() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if ok, err := isLogin(ctx); err != nil {
			return err
		} else if !ok {
			log.Printf("❌ 未登录")
			if err := doLogin().Do(ctx); err != nil {
				return err
			}
			notifyWxRead(notify.WxReadStatusLoginSuccess, wxReadState(0, "", "扫码登录已完成。"))
		} else {
			log.Printf("✅ 已登录")
			notifyWxRead(notify.WxReadStatusLoginSuccess, wxReadState(0, "", "cookies 登录态有效。"))
		}
		return
	}
}

func isLogin(ctx context.Context) (bool, error) {
	var hasVid, hasSkey, vid, skey bool
	retryCount := 0
	for (!hasVid || !hasSkey || !vid || !skey) && retryCount < 10 {
		cs, err := network.GetCookies().Do(ctx)
		if err != nil {
			return false, err
		}
		for _, cookie := range cs {
			if vid && skey {
				return true, nil
			}
			if cookie.Name == "wr_skey" {
				hasSkey = true
				if cookie.Value != "" {
					skey = true
				}
			}
			if cookie.Name == "wr_vid" {
				hasVid = true
				if cookie.Value != "" {
					vid = true
				}
			}
		}
		if vid && skey {
			return true, nil
		}
		retryCount++
		if err = chromedp.Sleep(2 * time.Second).Do(ctx); err != nil {
			return false, err
		}
		log.Printf("%v... Cookies 尝试加载", retryCount)
	}

	if check, err := deviceCfg.DoubleCheckLogin(ctx); err != nil {
		return false, err
	} else {
		return check, nil
	}
}

func doLogin() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if err := deviceCfg.BeforeClickLogin.Do(ctx); err != nil {
			return err
		}
		// 点击登录
		if err := deviceCfg.ClickLogin.Do(ctx); err != nil {
			return err
		}
		// 渲染登录二维码
		if err := renderLogin(ctx); err != nil {
			return err
		}
		hasLogin := atomic.Bool{}
		// 异步监控二维码过期
		go func() {
			for {
				if hasLogin.Load() {
					return
				}
				if err := qrcodeRefresh(ctx); err != nil {
					log.Printf("err: %v", err)
					return
				}
				if err := chromedp.Sleep(5 * time.Second).Do(ctx); err != nil {
					log.Printf("err: %v", err)
					return
				}
			}
		}()
		for {
			log.Printf("🍪登录中")
			if err := chromedp.Sleep(2 * time.Second).Do(ctx); err != nil {
				return err
			}
			if ok, err := isLogin(ctx); err != nil {
				return err
			} else if ok {
				hasLogin.Store(true)
				log.Printf("✅登录成功")
				break
			}
		}
		return nil
	}
}

// 渲染登录二维码
func renderLogin(ctx context.Context) error {
	if qrcode, err := deviceCfg.FetchLoginQrCode(ctx); err != nil {
		return err
	} else {
		qc := fmt.Sprintf("https://ximplez.github.io/base64-image-viewer/?target=%s", qrcode)
		notifyWxRead(notify.WxReadStatusLoginRequired, wxReadStateWithQRCode(qc, "请打开二维码完成微信读书登录。"))
		log.Printf("🍪已发送登录二维码【%s】", qc)
	}
	return nil
}

func qrcodeRefresh(ctx context.Context) error {
	if invalid, err := deviceCfg.IsInvalidLoginQrCode(ctx); err != nil {
		return err
	} else if invalid {
		log.Printf("🍪二维码失效，刷新中...")
		if err := deviceCfg.RefreshLoginQrCode(ctx); err != nil {
			return err
		}
		if err := renderLogin(ctx); err != nil {
			return err
		}
		log.Printf("✅二维码已刷新")
	}
	return nil
}

func startRead() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var catalogInfoStr string
		if catalogInfo, err := deviceCfg.GetCatalogInfo(ctx); err != nil {
			return err
		} else {
			curCatalog = catalogInfo
			catalogInfoStr = fmt.Sprintf(`
	当前章节: %s
	当前进度: %s
`, catalogInfo.CurCatalog(), catalogInfo.CurProgress())
		}
		log.Printf("✅ 开始阅读 %s", catalogInfoStr)
		bar = progressbar.Default(-1, "阅读中...")
		startTime := time.Now()
		notifyWxRead(notify.WxReadStatusReading, wxReadState(time.Since(startTime), "", "阅读已开始。"))
		defer func() {
			endTime := time.Now()
			totalReadTime = endTime.UnixMilli() - startTime.UnixMilli()
			if err := bar.Finish(); err != nil {
				log.Printf("progress err. %v", err)
			}
			if err := bar.Exit(); err != nil {
				log.Printf("progress err. %v", err)
			}
		}()
		for {
			if err := deviceCfg.StartRead(ctx); err != nil {
				notifyWxRead(notify.WxReadStatusProgressWarning, wxReadState(time.Since(startTime), err.Error(), "章节阅读动作异常。"))
				return err
			}
			if end, err := deviceCfg.IsEndPage(ctx); err != nil {
				notifyWxRead(notify.WxReadStatusProgressWarning, wxReadState(time.Since(startTime), err.Error(), "检查是否到达末页时异常。"))
				return err
			} else if end {
				finishedBook = true
				break
			}
			if err := deviceCfg.NextPage(ctx); err != nil {
				notifyWxRead(notify.WxReadStatusProgressWarning, wxReadState(time.Since(startTime), err.Error(), "翻到下一页时异常。"))
				return err
			}
			if catalogInfo, err := deviceCfg.GetCatalogInfo(ctx); err != nil {
				notifyWxRead(notify.WxReadStatusProgressWarning, wxReadState(time.Since(startTime), err.Error(), "读取章节进度时异常。"))
				return err
			} else {
				curCatalog = catalogInfo
			}
			totalReadPageCnt++
			notifyWxReadProgress(wxReadState(time.Since(startTime), "", "阅读仍在进行中。"))
			if err := bar.Add(1); err != nil {
				log.Printf("progress err. %v", err)
			}
		}
		return nil
	}
}

func notifyWxRead(status notify.WxReadStatus, state notify.WxReadCardState) {
	if notifier == nil {
		return
	}
	notifier.Upsert(notify.BuildWxReadCard(status, state))
}

func notifyWxReadProgress(state notify.WxReadCardState) {
	if notifier == nil {
		return
	}
	notifier.NotifyProgress(notify.BuildWxReadCard(notify.WxReadStatusReading, state))
}

func wxReadState(readTime time.Duration, errText, detail string) notify.WxReadCardState {
	catalogName, catalogProgress := "", ""
	if curCatalog != nil {
		catalogName = curCatalog.CurCatalog()
		catalogProgress = curCatalog.CurProgress()
	}
	if readTime <= 0 && totalReadTime > 0 {
		readTime = time.Millisecond * time.Duration(totalReadTime)
	}
	return notify.WxReadCardState{
		BookTitle:        bookTitle,
		TargetReadTime:   targetReadTime,
		TotalReadTime:    readTime,
		TotalReadPageCnt: totalReadPageCnt,
		CatalogName:      catalogName,
		CatalogProgress:  catalogProgress,
		FinishedBook:     finishedBook,
		Error:            errText,
		Detail:           detail,
	}
}

func wxReadStateWithQRCode(qrcodeURL, detail string) notify.WxReadCardState {
	state := wxReadState(0, "", detail)
	state.QRCodeURL = qrcodeURL
	return state
}
