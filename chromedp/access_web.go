package tool_chromedp

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func AccessWebWithCtx(ctx context.Context, tasks chromedp.Tasks) error {
	// 设置浏览器选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(ctx, opts...)

	// 创建一个浏览器实例
	ctx, cancel := chromedp.NewContext(allocCtx,
		// 设置日志方法
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	defer chromedp.Cancel(ctx)

	return chromedp.Run(ctx,
		tasks,
	)
}
