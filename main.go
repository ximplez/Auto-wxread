package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/schollz/progressbar/v3"
	tool_chromedp "github.com/ximplez/wxread/chromedp"
	"github.com/ximplez/wxread/device_cfg"
	"github.com/ximplez/wxread/io"
	"github.com/ximplez/wxread/json_tool"
)

var (
	url = "https://weread.qq.com/"
	// 书标题
	bookTitle string
	// 目标阅读时间
	targetReadTime time.Duration
	// 最终阅读时间
	totalReadTime int64
	// 总阅读页数
	totalReadPageCnt int64
	// 飞书机器人通知链接
	feishuBotUrl string
	// cookies
	cookies string

	bar          *progressbar.ProgressBar
	deviceCfg    = device_cfg.IPadPro
	finishedBook bool
)

func main() {
	tt := flag.Int64("t", 5, "目标阅读时间(分钟)")
	b := flag.String("b", "", "目标书名")
	fb := flag.String("fb", "", "飞书机器人通知链接")
	c := flag.String("c", "", "cookies")
	flag.Parse()
	if tt == nil || *tt <= 0 {
		log.Fatalln("targetTime 非法")
	}
	if fb != nil && strings.TrimSpace(*fb) != "" {
		feishuBotUrl = strings.TrimSpace(*fb)
	}
	if c != nil && strings.TrimSpace(*c) != "" {
		cookies = strings.TrimSpace(*c)
	}
	targetReadTime = time.Minute * time.Duration(*tt)
	if b != nil && strings.TrimSpace(*b) != "" {
		bookTitle = strings.TrimSpace(*b)
	}
	log.Printf("目标阅读时间: %s, 目标书名: %s", targetReadTime.String(), bookTitle)

	// 访问网页
	err := accessWeb()
	if err != nil {
		log.Fatalf("err: %v", err)
	}
}

func accessWeb() error {
	ctx, cancel := context.WithTimeout(context.Background(), targetReadTime)
	defer cancel()
	err := tool_chromedp.AccessWebWithCtx(ctx, chromedp.Tasks{
		// 设置设备模拟
		chromedp.Emulate(deviceCfg.Device),
		loadCookies(),
		// 页面导航
		chromedp.Navigate(url),
		deviceCfg.AfterNavigate,
		login(),
		findBook(),
		beforeRead(),
		startRead(),
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			end()
			return nil
		}
		NotifyFeishu(NewFeishuMsg("微信读书", "❌ 阅读失败", err.Error(), ""))
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
		finishedText = RedText("全书阅读完毕") + " 🎉🎉🎉"
	}
	atc := 0
	if totalReadPageCnt == 0 {
		atc = 0
	} else {
		atc = int(totalReadTime / 1000 / totalReadPageCnt)
	}
	summary := fmt.Sprintf(`📕书名: %s %s
	本次阅读时间: %s
	本次阅读页数: %s 页
	本次平均阅读时间: %s 秒`, BoldText(BlueText(bookTitle)), finishedText,
		BoldText(GreenText((time.Millisecond * time.Duration(totalReadTime)).String())), BoldText(strconv.FormatInt(totalReadPageCnt, 10)),
		BoldText(strconv.FormatInt(int64(atc), 10)))
	log.Printf(summary)
	NotifyFeishu(NewFeishuMsg("微信读书", "🎉结束阅读", summary, ""))
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
		}
		return nil
	}
}

func beforeRead() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Printf("📕书名: %s，目标阅读时间: %v", bookTitle, targetReadTime.String())
		return deviceCfg.BeforeRead(ctx)
	}
}

// 保存Cookies
func saveCookies(ctx context.Context) (err error) {
	// cookies的获取对应是在devTools的network面板中
	// 1. 获取cookies
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return
	}

	// 2. 序列化
	cookiesData := json_tool.ToJson(network.GetCookiesReturns{Cookies: cookies}, false)

	// 3. 存储到临时文件
	if err = io.WriteFile("cookies.tmp", cookiesData); err != nil {
		return
	}
	return
}

// 加载Cookies
func loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var cookiesParams *network.SetCookiesParams
		if cookies != "" {
			cookiesParams = json_tool.PhaseJsonFromString[network.SetCookiesParams](cookies)
		} else {
			// 如果cookies临时文件不存在则直接跳过
			if _, _err := os.Stat("cookies.tmp"); os.IsNotExist(_err) {
				return
			}

			// 如果存在则读取cookies的数据
			cookiesParams, err = io.ReadJsonFile[network.SetCookiesParams]("cookies.tmp")
			if err != nil {
				return
			}
		}

		// 设置cookies
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
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
		} else {
			log.Printf("✅ 已登录")
		}
		return
	}
}

func isLogin(ctx context.Context) (bool, error) {
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return false, err
	}
	for _, cookie := range cookies {
		if cookie.Name == "wr_name" && cookie.Value != "" {
			return true, nil
		}
	}
	return false, err
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
		// 二维码监控
		for {
			if err := qrcodeRefresh(ctx); err != nil {
				return err
			}
			log.Printf("🍪登录中")
			if err := chromedp.Sleep(10 * time.Second).Do(ctx); err != nil {
				return err
			}
			if ok, err := isLogin(ctx); err != nil {
				return err
			} else if ok {
				log.Printf("✅登录成功")
				// 保存cookies
				if err := saveCookies(ctx); err != nil {
					return err
				}
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
		NotifyFeishu(NewFeishuMsg("微信读书", "🍪扫码登录", "",
			fmt.Sprintf("https://ximplez.github.io/base64-image-viewer/?target=%s", qrcode)))
		log.Printf("🍪已发送登录二维码通知")
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
		log.Printf("✅ 开始阅读")
		bar = progressbar.Default(-1, "阅读中...")
		NotifyFeishu(NewFeishuMsg("微信读书", "📕开始阅读", fmt.Sprintf("📕书名: %s，目标阅读时间: %v",
			BlueText(BoldText(bookTitle)), GreenText(BoldText(targetReadTime.String()))), ""))
		startTime := time.Now()
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
				return err
			}
			if end, err := deviceCfg.IsEndPage(ctx); err != nil {
				return err
			} else if end {
				finishedBook = true
				break
			}
			if err := deviceCfg.NextPage(ctx); err != nil {
				return err
			}
			totalReadPageCnt++
			if err := bar.Add(1); err != nil {
				log.Printf("progress err. %v", err)
			}
		}
		return nil
	}
}
