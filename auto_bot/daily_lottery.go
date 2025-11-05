package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
)

type DailyLottery struct {
	*ErrorResp
	IsExpired    bool   `json:"isExpired"`
	IsSpecial    bool   `json:"isSpecial"`
	Issue        string `json:"issue"`
	IsPaid       int    `json:"isPaid"`
	GiftReceived `json:"giftReceived"`
}

type GiftReceived struct {
	*ErrorResp
	Name        string `json:"name"`
	Money       int    `json:"money"`
	ReceiveTime int    `json:"receiveTime"`
}

type DailyLotteryService struct {
	headers map[string]string
	params  map[string]string
}

func NewDailyLotteryService() *DailyLotteryService {
	headers := map[string]string{
		"Host":           "weread.qq.com",
		"accept":         "application/json, text/plain, */*",
		"sec-fetch-site": "same-origin",
		// "if-none-match":   `W/"df0-DMtlna7tm6ujPSlQpmdzYNJ5pBI"`,
		"sec-fetch-mode":  "cors",
		"accept-language": "zh-CN,zh-Hans;q=0.9",
		"user-agent":      "Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;WeRead/8.3.3 (iPhone; iOS 18.5; Scale/3.00)",
		"referer":         "https://weread.qq.com/membership-promotions?backgroundColor=%252321253D&isAnimateNavBarBackground=1&isShowNavBarShadow=0&isStatusbarLight=1&navBarBackgroundColor=%252321253D&navBarTintColor=%2523F2D2A1&navBarTitleColor=%2523ffffff",
		"sec-fetch-dest":  "empty",
	}
	params := map[string]string{
		"platform": "ios_html",
	}
	return &DailyLotteryService{headers: headers, params: params}
}

func (s *DailyLotteryService) buildCookie(vid, skey string) string {
	return fmt.Sprintf("wr_skey=%s;wr_vid=%s", skey, vid)
}

func (s *DailyLotteryService) GetDailyLottery(ctx context.Context, vid, skey string) (*DailyLottery, error) {
	s.headers["Cookie"] = s.buildCookie(vid, skey)
	_, bs, err := http.Get("https://weread.qq.com/membership-promotions/api/list", s.params, s.headers)
	if err != nil {
		return nil, err
	}
	res := json_tool.PhaseJson[DailyLottery](bs)
	if res.ErrorResp != nil {
		s := fmt.Sprintf("获取每日抽奖情况错误: %s", res.ErrorResp.Errmsg)
		return nil, errors.New(s)
	}
	return res, nil
}

func (s *DailyLotteryService) ExecuteDailyLottery(ctx context.Context, vid, skey, issue string) (*GiftReceived, error) {
	s.headers["Cookie"] = s.buildCookie(vid, skey)
	s.headers["Content-Type"] = "application/json"
	_, bs, err := http.PostWithParam("https://weread.qq.com/membership-promotions/api/receive", json_tool.ToJson(map[string]any{
		"issue": issue,
	}, false), s.params, s.headers)
	if err != nil {
		return nil, err
	}
	res := json_tool.PhaseJsonFromString[GiftReceived](bs)
	if res.ErrorResp != nil {
		s := fmt.Sprintf("执行每日抽奖错误: %s", res.ErrorResp.Errmsg)
		return nil, errors.New(s)
	}
	return res, nil
}
