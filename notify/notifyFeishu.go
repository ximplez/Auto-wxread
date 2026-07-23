package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultWxReadAppName       = "微信读书"
	DefaultWxReadURL           = "https://weread.qq.com/"
	DefaultProgressNotifyEvery = time.Minute
	NotificationConfigEnv      = "NOTIFICATION_CONFIG_JSON"
)

type CardConfig struct {
	Enabled             bool
	ConfigError         string
	GatewayBaseURL      string
	GatewayAuthToken    string
	AppID               string
	TemplateID          string
	TemplateVersionName string
	ReceiveIDType       string
	ReceiveID           string
	AppName             string
	OpenID              string
	DefaultURL          string
	ProgressNotifyEvery time.Duration
}

type FeishuCardNotifier struct {
	config         CardConfig
	client         *http.Client
	messageID      string
	lastProgressAt time.Time
	disabledLogged bool
	mu             sync.Mutex
}

type CardMessage struct {
	AppName            string
	Title              string
	SubTitle           string
	TitleStyle         string
	Content            string
	Foot               string
	MainButtonText     string
	MainButtonDisabled bool
	MainButtonEvent    any
	SubButtonText      string
	SubButtonDisabled  bool
	SubButtonURL       string
	OpenID             string
	Status             string
	Action             string
	Timestamp          time.Time
}

type WxReadCardState struct {
	BookTitle        string
	TargetReadTime   time.Duration
	TotalReadTime    time.Duration
	TotalReadPageCnt int64
	CatalogName      string
	CatalogProgress  string
	FinishedBook     bool
	QRCodeURL        string
	Error            string
	Detail           string
}

type WxReadStatus string

const (
	WxReadStatusStarting        WxReadStatus = "starting"
	WxReadStatusLoading         WxReadStatus = "loading"
	WxReadStatusLoginRequired   WxReadStatus = "login_required"
	WxReadStatusLoginSuccess    WxReadStatus = "login_success"
	WxReadStatusBookFound       WxReadStatus = "book_found"
	WxReadStatusReady           WxReadStatus = "ready"
	WxReadStatusReading         WxReadStatus = "reading"
	WxReadStatusProgressWarning WxReadStatus = "progress_warning"
	WxReadStatusFailed          WxReadStatus = "failed"
	WxReadStatusFinished        WxReadStatus = "finished"
)

type sendCardRequest struct {
	AppID               string         `json:"appId"`
	ReceiveIDType       string         `json:"receiveIdType,omitempty"`
	ReceiveID           string         `json:"receiveId,omitempty"`
	MessageID           string         `json:"messageId,omitempty"`
	TemplateID          string         `json:"templateId"`
	TemplateVersionName string         `json:"templateVersionName,omitempty"`
	TemplateVariable    map[string]any `json:"templateVariable"`
}

type gatewayCardResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Data    struct {
		MessageID string `json:"messageId"`
	} `json:"data"`
}

type notificationConfigJSON struct {
	Enabled             *bool                `json:"enabled"`
	GatewayBaseURL      string               `json:"gatewayBaseUrl"`
	GatewayAuthToken    string               `json:"gatewayAuthToken"`
	AppID               string               `json:"appId"`
	TemplateID          string               `json:"templateId"`
	TemplateVersionName string               `json:"templateVersionName"`
	ReceiveIDType       string               `json:"receiveIdType"`
	ReceiveID           string               `json:"receiveId"`
	AppName             string               `json:"appName"`
	OpenID              string               `json:"openId"`
	DefaultURL          string               `json:"defaultUrl"`
	ProgressSeconds     any                  `json:"progressNotifySeconds"`
	Card                notificationCardJSON `json:"card"`
}

type notificationCardJSON struct {
	OpenID       string `json:"openId"`
	SubButtonURL string `json:"subButtonUrl"`
}

func NewCardConfigFromEnv(getEnv func(string) string) CardConfig {
	configJSON := strings.TrimSpace(getEnv(NotificationConfigEnv))
	if configJSON == "" {
		return CardConfig{
			Enabled: false,
		}.normalize()
	}

	var raw notificationConfigJSON
	decoder := json.NewDecoder(strings.NewReader(configJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&raw); err != nil {
		return CardConfig{
			Enabled:     true,
			ConfigError: fmt.Sprintf("%s invalid JSON: %v", NotificationConfigEnv, err),
		}.normalize()
	}

	enabled := true
	if raw.Enabled != nil {
		enabled = *raw.Enabled
	}
	cfg := CardConfig{
		Enabled:             enabled,
		GatewayBaseURL:      raw.GatewayBaseURL,
		GatewayAuthToken:    raw.GatewayAuthToken,
		AppID:               raw.AppID,
		TemplateID:          raw.TemplateID,
		TemplateVersionName: raw.TemplateVersionName,
		ReceiveIDType:       raw.ReceiveIDType,
		ReceiveID:           raw.ReceiveID,
		AppName:             firstNonEmpty(raw.AppName, DefaultWxReadAppName),
		OpenID:              firstNonEmpty(raw.OpenID, raw.Card.OpenID),
		DefaultURL:          firstNonEmpty(raw.DefaultURL, raw.Card.SubButtonURL, DefaultWxReadURL),
		ProgressNotifyEvery: parseProgressNotifyEvery(raw.ProgressSeconds),
	}
	return cfg.normalize()
}

func NewFeishuCardNotifier(config CardConfig) *FeishuCardNotifier {
	config = config.normalize()
	return &FeishuCardNotifier{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *FeishuCardNotifier) Upsert(message CardMessage) {
	if n == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.config.Enabled {
		return
	}
	if err := n.config.validate(); err != nil {
		n.logDisabledOnce(err)
		return
	}

	message = n.config.applyMessageDefaults(message)
	if n.messageID != "" {
		if _, err := n.callGateway(n.messageID, message); err == nil {
			return
		} else {
			log.Printf("FeishuCardNotify update err:%v, message_id:%s", err, n.messageID)
			n.messageID = ""
		}
	}

	messageID, err := n.callGateway("", message)
	if err != nil {
		log.Printf("FeishuCardNotify send err:%v, title:%s", err, message.Title)
		return
	}
	n.messageID = messageID
}

func (n *FeishuCardNotifier) NotifyProgress(message CardMessage) {
	if n == nil {
		return
	}
	now := time.Now()
	n.mu.Lock()
	notifyEvery := n.config.ProgressNotifyEvery
	if notifyEvery <= 0 {
		notifyEvery = DefaultProgressNotifyEvery
	}
	if !n.lastProgressAt.IsZero() && now.Sub(n.lastProgressAt) < notifyEvery {
		n.mu.Unlock()
		return
	}
	n.lastProgressAt = now
	n.mu.Unlock()

	n.Upsert(message)
}

func (n *FeishuCardNotifier) callGateway(messageID string, message CardMessage) (string, error) {
	payload := sendCardRequest{
		AppID:               n.config.AppID,
		ReceiveIDType:       n.config.ReceiveIDType,
		ReceiveID:           n.config.ReceiveID,
		MessageID:           messageID,
		TemplateID:          n.config.TemplateID,
		TemplateVersionName: n.config.TemplateVersionName,
		TemplateVariable:    message.toTemplateVariable(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, n.config.gatewayEndpoint(), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+n.config.GatewayAuthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("gateway status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var gatewayResp gatewayCardResponse
	if err := json.Unmarshal(respBody, &gatewayResp); err != nil {
		return "", fmt.Errorf("decode gateway response: %w, body:%s", err, strings.TrimSpace(string(respBody)))
	}
	if !gatewayResp.Success {
		return "", fmt.Errorf("gateway rejected card: %s", gatewayResp.Error)
	}
	if messageID != "" {
		return messageID, nil
	}
	if gatewayResp.Data.MessageID == "" {
		return "", fmt.Errorf("gateway response missing messageId")
	}
	return gatewayResp.Data.MessageID, nil
}

func (n *FeishuCardNotifier) logDisabledOnce(err error) {
	if n.disabledLogged {
		return
	}
	n.disabledLogged = true
	if n.config.Enabled || n.config.hasAnyConfig() {
		log.Printf("FeishuCardNotify disabled: %v", err)
	}
}

func BuildWxReadCard(status WxReadStatus, state WxReadCardState) CardMessage {
	state = state.normalize()
	content := buildWxReadContent(state)
	foot := "后续登录、选书、阅读进度和最终结果都会持续更新在这张卡片。"
	card := CardMessage{
		AppName:            DefaultWxReadAppName,
		Title:              "阅读任务准备开始",
		SubTitle:           fmt.Sprintf("目标阅读 %s", formatDuration(state.TargetReadTime)),
		TitleStyle:         "blue",
		Content:            content,
		Foot:               foot,
		MainButtonText:     "自动执行中",
		MainButtonDisabled: true,
		SubButtonText:      "打开微信读书",
		SubButtonURL:       DefaultWxReadURL,
		Status:             string(status),
		Action:             "wxread",
		Timestamp:          time.Now(),
	}

	switch status {
	case WxReadStatusLoading:
		card.Title = "正在加载微信读书"
		card.SubTitle = "浏览器已启动，正在恢复登录态"
		card.TitleStyle = "blue"
		card.Foot = "正在打开页面并校验 cookies，若登录态失效会自动推送扫码入口。"
	case WxReadStatusLoginRequired:
		card.Title = "需要扫码登录"
		card.SubTitle = "登录态已失效，请打开二维码完成确认"
		card.TitleStyle = "orange"
		card.Foot = "二维码失效后会自动刷新并更新这张卡片。登录完成后阅读任务会继续执行。"
		card.MainButtonText = "等待扫码"
		card.SubButtonText = "打开登录二维码"
		card.SubButtonURL = state.QRCodeURL
	case WxReadStatusLoginSuccess:
		card.Title = "登录成功"
		card.SubTitle = "正在查找目标书籍"
		card.TitleStyle = "turquoise"
		card.Foot = "登录态已确认，接下来会进入书架并定位目标书籍。"
	case WxReadStatusBookFound:
		card.Title = "已找到书籍"
		card.SubTitle = state.BookTitle
		card.TitleStyle = "blue"
		card.Foot = "已进入目标书籍，正在准备阅读页。"
	case WxReadStatusReady:
		card.Title = "阅读页已就绪"
		card.SubTitle = state.BookTitle
		card.TitleStyle = "blue"
		card.Foot = "即将开始模拟阅读，后续会按节奏更新当前章节和累计阅读数据。"
	case WxReadStatusReading:
		card.Title = "阅读进行中"
		card.SubTitle = readingSubtitle(state)
		card.TitleStyle = "wathet"
		card.Foot = "阅读任务仍在执行。卡片会按时间间隔和关键状态变化更新，避免消息刷屏。"
	case WxReadStatusProgressWarning:
		card.Title = "阅读进度异常"
		card.SubTitle = "已记录当前进度，正在收敛异常"
		card.TitleStyle = "orange"
		card.Foot = "程序会继续尝试或进入失败收敛；当前章节、页数和错误信息已保留在卡片中。"
		card.MainButtonText = "检查中"
	case WxReadStatusFailed:
		card.Title = "阅读失败"
		card.SubTitle = "任务已停止"
		card.TitleStyle = "red"
		card.Foot = "请检查登录态、页面结构、网络和 GitHub Secrets 配置。"
		card.MainButtonText = "已失败"
	case WxReadStatusFinished:
		card.Title = "阅读完成"
		if state.FinishedBook {
			card.SubTitle = "全书阅读完毕"
		} else {
			card.SubTitle = "已达到本次目标阅读时长"
		}
		card.TitleStyle = "green"
		card.Foot = "本次任务已结束，卡片不会继续更新。"
		card.MainButtonText = "已完成"
	}
	return card
}

func (c CardConfig) normalize() CardConfig {
	c.GatewayBaseURL = strings.TrimSpace(c.GatewayBaseURL)
	c.GatewayAuthToken = strings.TrimSpace(c.GatewayAuthToken)
	c.AppID = strings.TrimSpace(c.AppID)
	c.TemplateID = strings.TrimSpace(c.TemplateID)
	c.TemplateVersionName = strings.TrimSpace(c.TemplateVersionName)
	c.ReceiveIDType = strings.TrimSpace(c.ReceiveIDType)
	c.ReceiveID = strings.TrimSpace(c.ReceiveID)
	c.AppName = firstNonEmpty(c.AppName, DefaultWxReadAppName)
	c.OpenID = strings.TrimSpace(c.OpenID)
	c.DefaultURL = firstNonEmpty(c.DefaultURL, DefaultWxReadURL)
	if c.ProgressNotifyEvery <= 0 {
		c.ProgressNotifyEvery = DefaultProgressNotifyEvery
	}
	return c
}

func (c CardConfig) validate() error {
	if c.ConfigError != "" {
		return fmt.Errorf("%s", c.ConfigError)
	}
	var missing []string
	if c.GatewayBaseURL == "" {
		missing = append(missing, "gatewayBaseUrl")
	}
	if c.GatewayAuthToken == "" {
		missing = append(missing, "gatewayAuthToken")
	}
	if c.AppID == "" {
		missing = append(missing, "appId")
	}
	if c.TemplateID == "" {
		missing = append(missing, "templateId")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s missing %s", NotificationConfigEnv, strings.Join(missing, ", "))
	}
	return nil
}

func (c CardConfig) hasAnyConfig() bool {
	return c.ConfigError != "" || c.GatewayBaseURL != "" || c.GatewayAuthToken != "" || c.AppID != "" || c.TemplateID != ""
}

func (c CardConfig) gatewayEndpoint() string {
	base := strings.TrimRight(c.GatewayBaseURL, "/")
	if strings.HasSuffix(base, "/send_card") {
		return base
	}
	return base + "/send_card"
}

func (c CardConfig) applyMessageDefaults(message CardMessage) CardMessage {
	if message.AppName == "" {
		message.AppName = c.AppName
	}
	if message.OpenID == "" {
		message.OpenID = c.OpenID
	}
	if message.SubButtonURL == "" {
		message.SubButtonURL = c.DefaultURL
	}
	if message.SubButtonText == "" {
		message.SubButtonText = "打开微信读书"
	}
	if message.TitleStyle == "" {
		message.TitleStyle = "blue"
	}
	if message.MainButtonText == "" {
		message.MainButtonText = "自动执行中"
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	return message
}

func (m CardMessage) toTemplateVariable() map[string]any {
	return map[string]any{
		"app_name":          m.AppName,
		"appName":           m.AppName,
		"title":             m.Title,
		"sub_title":         m.SubTitle,
		"subTitle":          m.SubTitle,
		"title_style":       m.TitleStyle,
		"titleStyle":        m.TitleStyle,
		"content":           withMention(m.OpenID, m.Content),
		"foot":              m.Foot,
		"main_button_text":  m.MainButtonText,
		"mainButtonText":    m.MainButtonText,
		"main_button":       m.MainButtonDisabled,
		"mainButton":        m.MainButtonDisabled,
		"main_button_event": normalizeMainButtonEvent(m.MainButtonEvent),
		"mainButtonEvent":   normalizeMainButtonEvent(m.MainButtonEvent),
		"sub_button_text":   m.SubButtonText,
		"subButtonText":     m.SubButtonText,
		"sub_button":        m.SubButtonDisabled || m.SubButtonURL == "",
		"subButton":         m.SubButtonDisabled || m.SubButtonURL == "",
		"sub_button_url":    m.SubButtonURL,
		"subButtonUrl":      m.SubButtonURL,
		"open_id":           m.OpenID,
		"openId":            m.OpenID,
		"status":            m.Status,
		"action":            m.Action,
		"timestamp":         m.Timestamp.Format(time.RFC3339),
	}
}

func (s WxReadCardState) normalize() WxReadCardState {
	s.BookTitle = firstNonEmpty(s.BookTitle, "默认书籍")
	return s
}

func buildWxReadContent(state WxReadCardState) string {
	lines := []string{
		fmt.Sprintf("**书名**：%s", state.BookTitle),
		fmt.Sprintf("**目标阅读**：%s", formatDuration(state.TargetReadTime)),
	}
	if state.TotalReadTime > 0 {
		lines = append(lines, fmt.Sprintf("**已读时长**：%s", formatDuration(state.TotalReadTime)))
	}
	if state.TotalReadPageCnt > 0 {
		lines = append(lines, fmt.Sprintf("**累计翻页/章节**：%d", state.TotalReadPageCnt))
	}
	if state.CatalogName != "" {
		lines = append(lines, fmt.Sprintf("**当前章节**：%s", state.CatalogName))
	}
	if state.CatalogProgress != "" {
		lines = append(lines, fmt.Sprintf("**目录进度**：%s", state.CatalogProgress))
	}
	if state.TargetReadTime > 0 && state.TotalReadTime > 0 {
		lines = append(lines, fmt.Sprintf("**目标完成度**：%s", progressText(state.TotalReadTime, state.TargetReadTime)))
	}
	if state.TotalReadPageCnt > 0 && state.TotalReadTime > 0 {
		avg := state.TotalReadTime / time.Duration(state.TotalReadPageCnt)
		lines = append(lines, fmt.Sprintf("**平均用时**：%s/页", formatDuration(avg)))
	}
	if state.FinishedBook {
		lines = append(lines, "**阅读结果**：全书阅读完毕")
	}
	if state.Error != "" {
		lines = append(lines, fmt.Sprintf("**异常信息**：%s", state.Error))
	}
	if state.Detail != "" {
		lines = append(lines, state.Detail)
	}
	return strings.Join(lines, "\n")
}

func readingSubtitle(state WxReadCardState) string {
	if state.CatalogProgress != "" {
		return fmt.Sprintf("%s · %s", state.BookTitle, state.CatalogProgress)
	}
	return state.BookTitle
}

func progressText(current, target time.Duration) string {
	if target <= 0 {
		return "-"
	}
	percent := float64(current) / float64(target)
	if percent > 1 {
		percent = 1
	}
	if percent < 0 {
		percent = 0
	}
	const width = 10
	filled := int(percent*width + 0.5)
	if filled > width {
		filled = width
	}
	return fmt.Sprintf("`%s%s` %.0f%%", strings.Repeat("█", filled), strings.Repeat("░", width-filled), percent*100)
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	d = d.Round(time.Second)
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	s := int(d / time.Second)
	if h > 0 {
		return fmt.Sprintf("%d小时%d分%d秒", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d分%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", s)
}

func withMention(openID, content string) string {
	openID = strings.TrimSpace(openID)
	if openID == "" {
		return content
	}
	return fmt.Sprintf("<at id=\"%s\"></at>\n\n%s", openID, content)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeMainButtonEvent(value any) any {
	if isEmptyMainButtonEvent(value) {
		return map[string]any{
			"action": "noop",
			"source": "wxread",
		}
	}
	switch v := value.(type) {
	case json.Number, float64, bool, map[string]any, []any:
		return v
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return map[string]any{
				"action": "noop",
				"source": "wxread",
			}
		}
		var parsed any
		decoder := json.NewDecoder(strings.NewReader(trimmed))
		decoder.UseNumber()
		if err := decoder.Decode(&parsed); err == nil {
			return parsed
		}
		return map[string]any{
			"action": trimmed,
			"source": "wxread",
		}
	default:
		return v
	}
}

func isEmptyMainButtonEvent(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func parseProgressNotifyEvery(value any) time.Duration {
	if value == nil {
		return DefaultProgressNotifyEvery
	}
	var seconds int64
	switch v := value.(type) {
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return DefaultProgressNotifyEvery
		}
		seconds = parsed
	case float64:
		seconds = int64(v)
	case int:
		seconds = int64(v)
	case int64:
		seconds = v
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return DefaultProgressNotifyEvery
		}
		seconds = parsed
	default:
		return DefaultProgressNotifyEvery
	}
	if seconds <= 0 {
		return DefaultProgressNotifyEvery
	}
	return time.Duration(seconds) * time.Second
}
