package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/schollz/progressbar/v3"
	tool_chromedp "github.com/ximplez/wxread/chromedp"
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

	bar *progressbar.ProgressBar
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
		chromedp.Emulate(device.KindleFireHDX),
		loadCookies(),
		// 页面导航
		chromedp.Navigate(url),
		chromedp.WaitReady(`#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_header_wrapper`),
		login(),
		findBook(),
		chromedp.WaitReady(`#routerView > div > div.wr_horizontalReader_app_content > div.wr_various_font_provider_wrapper > div > div > div.renderTargetContainer > div.renderTarget_pager > div.renderTarget_pager_content.renderTarget_pager_content_right > button`),
		beforeRead(),
		startRead(),
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			summary := fmt.Sprintf("📕书名: %s，总阅读时间: %s, 总阅读页数: %d 页, 平均阅读时间: %d 秒", bookTitle,
				(time.Millisecond * time.Duration(totalReadTime)).String(), totalReadPageCnt, totalReadTime/1000/totalReadPageCnt)
			log.Printf(summary)
			NotifyFeishu(NewFeishuMsg("微信读书", "🎉结束阅读", summary, ""))
			return nil
		}
		NotifyFeishu(NewFeishuMsg("微信读书", "❌阅读失败", err.Error(), ""))
		return err
	}
	return nil
}
func findBook() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var book string
		ind := "#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_content_wrapper > div > div.wr_index_mini_shelf_wrapper"
		if err := chromedp.WaitVisible(ind, chromedp.After(func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			n := node[0]
			for i := int64(0); i < n.ChildNodeCount; i++ {
				// 获取书标题
				if err := chromedp.Text(fmt.Sprintf("%s > div:nth-child(%d) > a > div > div.wr_index_mini_shelf_card_content_info > div.wr_index_mini_shelf_card_content_title",
					ind, i+1), &book).Do(ctx); err != nil {
					log.Printf("err: %v", err)
					continue
				}
				if bookTitle == "" || book == bookTitle {
					if err := chromedp.Click(fmt.Sprintf("%s > div:nth-child(%d) > a", ind, i+1)).Do(ctx); err != nil {
						return err
					}
					bookTitle = book
					break
				}
			}
			return nil
		})).Do(ctx); err != nil {
			return err
		}
		if book == "" {
			return fmt.Errorf("❌未找到书: %s", bookTitle)
		}
		log.Printf("✅ 找到书: %s", book)
		return nil
	}
}

func beforeRead() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 获取书标题
		if err := chromedp.Text("#routerView > div > div.wr_horizontalReader_app_content > div.readerTopBar > div > div.readerTopBar_left > span > span", &bookTitle).Do(ctx); err != nil {
			return err
		}
		log.Printf("📕书名: %s，目标阅读时间: %v", bookTitle, targetReadTime.String())
		return nil
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
			log.Printf("❌未登录")
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
		// 点击登录
		if err := chromedp.Click("#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper " +
			"> div.wr_index_page_top_section_header_wrapper > div.wr_index_page_top_section_header_action > a:nth-child(3)").Do(ctx); err != nil {
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
	var qrcode string
	if err := chromedp.QueryAfter("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container "+
		"> div.wr_login_modal_qr_wrapper_old > div.wr_login_modal_qr_wrapper > div > img",
		func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			for _, v := range node {
				qrcode = v.AttributeValue("src")
			}
			NotifyFeishu(NewFeishuMsg("微信读书", "🍪扫码登录", "", fmt.Sprintf("https://ximplez.github.io/base64-image-viewer/?target=%s", qrcode)))
			log.Printf("🍪已发送登录二维码通知")
			return nil
		}).Do(ctx); err != nil {
		return err
	}
	return nil
}

func qrcodeRefresh(ctx context.Context) error {
	return chromedp.QueryAfter("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container > div.wr_login_modal_qr_wrapper_old",
		func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			for _, child := range node[0].Children {
				if child.AttributeValue("class") == "wr_login_modal_lang" && child.Children[0].NodeValue == "二维码已失效" {
					log.Printf("🍪二维码失效，刷新中...")
					if err := chromedp.Click("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container " +
						"> div.login_dialog_retry_delegate").Do(ctx); err != nil {
						return err
					}
					if err := renderLogin(ctx); err != nil {
						return err
					}
					log.Printf("✅二维码已刷新")
				}
			}
			return nil
		}).Do(ctx)
}

func randomReadTime(min, max int64) time.Duration {
	if min >= max {
		return 0
	}
	t := min + rand.N[int64](max-min)
	return time.Duration(t) * time.Second
}

func startRead() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Printf("✅ 开始阅读")
		bar = progressbar.Default(-1, "阅读中...")
		NotifyFeishu(NewFeishuMsg("微信读书", "📕开始阅读", fmt.Sprintf("📕书名: %s，目标阅读时间: %v", bookTitle, targetReadTime.String()), ""))
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
			if err := chromedp.Sleep(randomReadTime(10, 60)).Do(ctx); err != nil {
				return err
			}
			if err := nextPage(ctx); err != nil {
				return err
			}
			totalReadPageCnt++
			if err := bar.Add(1); err != nil {
				log.Printf("progress err. %v", err)
			}
		}
	}
}
func nextPage(ctx context.Context) error {
	return chromedp.Click("#routerView > div > div.wr_horizontalReader_app_content > " +
		"div.wr_various_font_provider_wrapper > div > div > div.renderTargetContainer > div.renderTarget_pager > " +
		"div.renderTarget_pager_content.renderTarget_pager_content_right > button").Do(ctx)
}
