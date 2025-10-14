package main

import (
	"log"

	"github.com/ximplez/wxread/http"
	"github.com/ximplez/wxread/json_tool"
)

type FeishuMsg struct {
	App       string `json:"app"`
	Title     string `json:"title"`
	Msg       string `json:"msg"`
	TargetUrl string `json:"targetUrl"`
}

func NewFeishuMsg(app, title, msg, targetUrl string) *FeishuMsg {
	return &FeishuMsg{
		App:       app,
		Title:     title,
		Msg:       msg,
		TargetUrl: targetUrl,
	}
}

func NotifyFeishu(msg *FeishuMsg) {
	if msg == nil || feishuBotUrl == "" {
		return
	}
	// 发送飞书消息
	_, _, err := http.Post(feishuBotUrl, json_tool.ToJson(msg, false), nil)
	if err != nil {
		log.Printf("NotifyFeishu err:%v, msg:%s", err, json_tool.ToJson(msg, false))
		return
	}
}
