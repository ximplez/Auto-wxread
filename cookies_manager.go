package main

import (
	"context"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/ximplez/wxread/utils"
	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/io"
	"github.com/ximplez/wxread/utils/json_tool"
)

// 加载Cookies
func loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		if cookies != "" {
			if utils.IsUrl(cookies) {
				// 从http中读取cookies
				if cookies, err = readCookiesFromHttp(cookies); err != nil {
					return err
				}
			}
		} else {
			// 从文件中读取cookies
			if cookies, err = readCookiesFromFile(); err != nil {
				return err
			}
		}
		if cookies != "" {
			cookiesParams := json_tool.PhaseJsonFromString[network.SetCookiesParams](cookies)
			// 设置cookies
			return network.SetCookies(cookiesParams.Cookies).Do(ctx)
		}
		return nil
	}
}

// 保存Cookies
func saveCookies() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cks, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}

		// 2. 序列化
		cookiesData := json_tool.ToJson(network.GetCookiesReturns{Cookies: cks}, false)

		// 3. 存储到临时文件
		if err = io.WriteFile("cookies.tmp", cookiesData); err != nil {
			return err
		}
		return nil
	}
}

func readCookiesFromHttp(url string) (string, error) {
	_, bytes, err := http.Get(url, nil, nil)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func readCookiesFromFile() (string, error) {
	// 如果cookies临时文件不存在则直接跳过
	if _, _err := os.Stat("cookies.tmp"); os.IsNotExist(_err) {
		return "", nil
	}
	bytes, err := io.ReadFile("cookies.tmp")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
