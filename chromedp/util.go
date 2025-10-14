package tool_chromedp

import "github.com/chromedp/chromedp"

// elementScreenshot 截取特定元素的屏幕截图。
func elementScreenshot(sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}

// fullScreenshot 会截取整个浏览器视口的屏幕截图。
//
// 注意: chromedp.FullScreenshot 会覆盖设备的emulation 设置。 Use
// 使用 device.Reset 重置 emulation 和 viewport 设置。
func fullScreenshot(quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.FullScreenshot(res, quality),
	}
}
