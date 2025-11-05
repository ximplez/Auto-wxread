package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
)

type ReadingStatistic struct {
	*ErrorResp
	ReadTimes map[string]int `json:"readTimes"`
	ReadDays  int            `json:"readDays"`
}

type ReadingStatisticService struct {
	headers map[string]string
	params  map[string]string
}

func NewReadingStatisticService() *ReadingStatisticService {
	headers := map[string]string{
		"Host":            "i.weread.qq.com",
		"accept":          "*/*",
		"channelid":       "AppStore",
		"basever":         "8.3.3.20",
		"v":               "8.3.3.20",
		"accept-language": "zh-Hans-CN;q=1, en-CN;q=0.9",
		"user-agent":      "WeRead/8.3.3 (iPhone; iOS 18.5; Scale/3.00)",
	}
	params := map[string]string{
		"baseTime":          "0",
		"defaultPreferBook": "0",
		"mode":              "weekly",
	}
	return &ReadingStatisticService{headers: headers, params: params}
}

func (s *ReadingStatisticService) GetReadingStatistic(ctx context.Context, vid, skey string) (*ReadingStatistic, error) {
	s.headers["vid"] = vid
	s.headers["skey"] = skey
	_, bs, err := http.Get("https://i.weread.qq.com/readdata/detail", s.params, s.headers)
	if err != nil {
		return nil, err
	}
	// log.Printf("resp: %s", string(bs))
	res := json_tool.PhaseJson[ReadingStatistic](bs)
	if res.ErrorResp != nil {
		s := fmt.Sprintf("获取阅读统计失败: %s", res.ErrorResp.Errmsg)
		return nil, errors.New(s)
	}
	return res, nil
}
