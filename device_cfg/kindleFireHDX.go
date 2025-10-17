package device_cfg

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

var KindleFireHDX = DeviceCfg{
	Device:        device.KindleFireHDX,
	AfterNavigate: chromedp.WaitReady(`#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_header_wrapper`),
	BeforeClickLogin: chromedp.WaitReady("#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper " +
		"> div.wr_index_page_top_section_header_wrapper > div.wr_index_page_top_section_header_action > a:nth-child(3)"),
	ClickLogin: chromedp.Click("#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper " +
		"> div.wr_index_page_top_section_header_wrapper > div.wr_index_page_top_section_header_action > a:nth-child(3)"),
	FetchLoginQrCode: func(ctx context.Context) (string, error) {
		var qrcode string
		if err := chromedp.QueryAfter("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container "+
			"> div.wr_login_modal_qr_wrapper_old > div.wr_login_modal_qr_wrapper > div > img",
			func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
				for _, v := range node {
					qrcode = v.AttributeValue("src")
				}
				return nil
			}).Do(ctx); err != nil {
			return "", err
		}
		return qrcode, nil
	},
	IsInvalidLoginQrCode: func(ctx context.Context) (bool, error) {
		var invalid bool
		if err := chromedp.QueryAfter("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container > div.wr_login_modal_qr_wrapper_old",
			func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
				for _, child := range node[0].Children {
					if child.AttributeValue("class") == "wr_login_modal_lang" && child.Children[0].NodeValue == "二维码已失效" {
						invalid = true
					}
				}
				return nil
			}).Do(ctx); err != nil {
			return false, err
		}
		return invalid, nil
	},
	RefreshLoginQrCode: func(ctx context.Context) error {
		return chromedp.Click("body > div.wr_mask > div > div > div.wr_login_modal_qr_wrapper_container " +
			"> div.login_dialog_retry_delegate").Do(ctx)
	},
	FindBookAndClick: func(ctx context.Context, bookName string) (string, error) {
		var book string
		ind := "#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_content_wrapper > div > div.wr_index_mini_shelf_wrapper"
		if err := chromedp.WaitReady(ind, chromedp.After(func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			n := node[0]
			for i := int64(0); i < n.ChildNodeCount; i++ {
				// 获取书标题
				if err := chromedp.Text(fmt.Sprintf("%s > div:nth-child(%d) > a > div > div.wr_index_mini_shelf_card_content_info > div.wr_index_mini_shelf_card_content_title",
					ind, i+1), &book).Do(ctx); err != nil {
					log.Printf("err: %v", err)
					continue
				}
				if bookName == "" || book == bookName {
					if err := chromedp.Click(fmt.Sprintf("%s > div:nth-child(%d) > a", ind, i+1)).Do(ctx); err != nil {
						return err
					}
					break
				}
			}
			return nil
		})).Do(ctx); err != nil {
			return "", err
		}
		return book, nil
	},
	BeforeRead: func(ctx context.Context) error {
		return chromedp.WaitReady(`#routerView > div > div.wr_horizontalReader_app_content > div.wr_various_font_provider_wrapper > div > div > div.renderTargetContainer > div.renderTarget_pager > div.renderTarget_pager_content.renderTarget_pager_content_right > button`).Do(ctx)
	},
	StartRead: func(ctx context.Context) error {
		return chromedp.Sleep(randomReadTime(10, 30)).Do(ctx)
	},
	NextPage: func(ctx context.Context) error {
		return chromedp.Click("#routerView > div > div.wr_horizontalReader_app_content > " +
			"div.wr_various_font_provider_wrapper > div > div > div.renderTargetContainer > div.renderTarget_pager > " +
			"div.renderTarget_pager_content.renderTarget_pager_content_right > button").Do(ctx)
	},
	IsEndPage: func(ctx context.Context) (bool, error) {
		return false, nil
	},
}
