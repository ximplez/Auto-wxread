package notify

import (
	"log"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
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

func RedText(text string) string {
	return "<font color=\"red\">" + text + "</font>"
}

func GreenText(text string) string {
	return "<font color=\"green\">" + text + "</font>"
}

func BlueText(text string) string {
	return "<font color=\"blue\">" + text + "</font>"
}

func YellowText(text string) string {
	return "<font color=\"yellow\">" + text + "</font>"
}

func GreyText(text string) string {
	return "<font color=\"grey\">" + text + "</font>"
}
func BoldText(text string) string {
	return "**" + text + "**"
}

func NotifyFeishu(feishuBotUrl string, msg *FeishuMsg) {
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
