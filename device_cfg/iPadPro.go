package device_cfg

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

var IPadPro = DeviceCfg{
	Device:           device.IPadPro,
	AfterNavigate:    clean,
	BeforeClickLogin: chromedp.WaitReady("#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_header_wrapper > div.wr_index_page_top_section_header_action > a:nth-child(3)"),
	ClickLogin:       chromedp.Click("#__nuxt > div > div > div > div.wr_index_page_content_wrapper > div.wr_index_page_top_section_wrapper > div.wr_index_page_top_section_header_wrapper > div.wr_index_page_top_section_header_action > a:nth-child(3)"),
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
		if err := clean.Do(ctx); err != nil {
			return err
		}
		if err := chromedp.WaitReady(`#routerView > div > div.app_content > div.wr_various_font_provider_wrapper`).Do(ctx); err != nil {
			return err
		}
		return nil
	},
	StartRead: func(ctx context.Context) error {
		if err := chromedp.QueryAfter(`#routerView > div > div.app_content > div.wr_various_font_provider_wrapper`, func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			return clean.Do(ctx)
		}).Do(ctx); err != nil {
			return err
		}
		cury, prey := float64(0), float64(0)
		width, height := float64(0), float64(0)
		if err := chromedp.Evaluate("window.innerWidth", &width).Do(ctx); err != nil {
			return err
		}
		if err := chromedp.Evaluate("window.innerHeight", &height).Do(ctx); err != nil {
			return err
		}
		for {
			if err := chromedp.Evaluate("window.scrollY", &prey).Do(ctx); err != nil {
				return err
			}
			// if err := chromedp.MouseEvent(input.MouseWheel, width/2, height/2, func(params *input.DispatchMouseEventParams) *input.DispatchMouseEventParams {
			// 	params.WithDeltaY(float64(random(50, 200)))
			// 	return params
			// }).Do(ctx); err != nil {
			// 	return err
			// }
			if err := chromedp.Evaluate(fmt.Sprintf("window.scrollTo({\n  top: %f,\n  behavior: \"smooth\",\n})", prey+float64(random(50, 200))), nil).
				Do(ctx); err != nil {
				return err
			}
			if err := chromedp.Sleep(1 * time.Second).Do(ctx); err != nil {
				return err
			}
			if err := chromedp.Evaluate("window.scrollY", &cury).Do(ctx); err != nil {
				return err
			}
			if cury == prey {
				break
			}
		}
		return nil
	},
	NextPage: func(ctx context.Context) error {
		return chromedp.Click("#routerView > div > div.app_content > div.readerFooter > div > button:nth-child(1)").Do(ctx)
	},
	IsEndPage: func(ctx context.Context) (bool, error) {
		var end bool
		if err := chromedp.QueryAfter("#routerView > div > div.app_content > div.readerFooter", func(ctx context.Context, id runtime.ExecutionContextID, node ...*cdp.Node) error {
			n := node[0]
			if clz, ok := n.Attribute("class"); ok {
				if strings.Contains(clz, "readerFooter_last_page") {
					end = true
				}
			}
			return nil
		}).Do(ctx); err != nil {
			return false, err
		}
		return end, nil
	},
}
