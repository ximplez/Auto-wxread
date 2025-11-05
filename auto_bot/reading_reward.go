package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
)

type awardStatus int

const (
	awardStatus_Unreceivable awardStatus = iota
	awardStatus_Unreceived
	awardStatus_Received
)

type Award struct {
	AwardLevelId     int         `json:"awardLevelId"`
	AwardChooseType  int         `json:"awardChooseType"`
	AwardStatus      awardStatus `json:"awardStatus"`
	AwardStatusDesc  string      `json:"awardStatusDesc"`
	AwardLevelDesc   string      `json:"awardLevelDesc"`
	AwardChoicesDesc string      `json:"awardChoicesDesc"`
	AwardChoices     []struct {
		ChoiceType int `json:"choiceType"`
		AwardNum   int `json:"awardNum"`
		CanChoice  int `json:"canChoice"`
	} `json:"awardChoices"`
}

type ReadingReward struct {
	*ErrorResp
	ReadtimeAwards []*Award `json:"readtimeAwards"`
	ReaddayAwards  []*Award `json:"readdayAwards"`
}

type readingRewardBody struct {
	AwardLevelId    int    `json:"awardLevelId"`
	Unread          int    `json:"unread"`
	IsExchangeAward int    `json:"isExchangeAward"`
	Pf              string `json:"pf"`
	IsVisitReadGoal int    `json:"isVisitReadGoal"`
	AwardChoiceType int    `json:"awardChoiceType"`
}

type ReadingRewardService struct {
	headers map[string]string
	params  map[string]string
}

func NewReadingRewardService() *ReadingRewardService {
	headers := map[string]string{
		"Host":            "i.weread.qq.com",
		"channelid":       "AppStore",
		"accept":          "*/*",
		"accept-language": "zh-Hans-CN;q=1, en-CN;q=0.9",
		"basever":         "8.3.3.20",
		"user-agent":      "WeRead/8.3.3 (iPhone; iOS 18.5; Scale/3.00)",
		"v":               "8.3.3.20",
		"Cookie":          "wr_logined=1",
		"content-type":    "application/json",
	}
	return &ReadingRewardService{headers: headers}
}

func (s *ReadingRewardService) GetReadingReward(ctx context.Context, vid, skey string) (*ReadingReward, error) {
	s.headers["vid"] = vid
	s.headers["skey"] = skey
	_, bs, err := http.Post("https://i.weread.qq.com/weekly/exchange", json_tool.ToJson(readingRewardBody{
		AwardLevelId:    0,
		Unread:          1,
		IsExchangeAward: 0,
		Pf:              "weread_wx-2001-iap-2001-iphone",
		IsVisitReadGoal: 1,
		AwardChoiceType: 0,
	}, false), s.headers)
	if err != nil {
		return nil, err
	}
	res := json_tool.PhaseJsonFromString[ReadingReward](bs)
	if res.ErrorResp != nil {
		s := fmt.Sprintf("获取阅读奖励失败: %s", res.ErrorResp.Errmsg)
		return nil, errors.New(s)
	}
	return res, nil
}

func (s *ReadingRewardService) ExchangeReadingReward(ctx context.Context, vid, skey string, levelId, choiceType int) error {
	s.headers["vid"] = vid
	s.headers["skey"] = skey
	_, bs, err := http.Post("https://i.weread.qq.com/weekly/exchange", json_tool.ToJson(readingRewardBody{
		AwardLevelId:    levelId,
		Unread:          1,
		IsExchangeAward: 1,
		Pf:              "weread_wx-2001-iap-2001-iphone",
		AwardChoiceType: choiceType,
	}, false), s.headers)
	if err != nil {
		return err
	}
	res := json_tool.PhaseJsonFromString[ReadingReward](bs)
	if res.ErrorResp != nil {
		s := fmt.Sprintf("兑换阅读奖励失败: %s", res.ErrorResp.Errmsg)
		return errors.New(s)
	}
	return nil
}
