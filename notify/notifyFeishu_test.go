package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFeishuCardNotifierUpsertSendsThenUpdates(t *testing.T) {
	var requests []sendCardRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send_card" {
			t.Fatalf("path = %s, want /send_card", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("Authorization = %s", got)
		}
		var req sendCardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, req)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"messageId": "om_test",
			},
		})
	}))
	defer server.Close()

	notifier := NewFeishuCardNotifier(CardConfig{
		Enabled:          true,
		GatewayBaseURL:   server.URL,
		GatewayAuthToken: "token",
		AppID:            "cli_test",
		TemplateID:       "ctp_test",
	})
	notifier.Upsert(BuildWxReadCard(WxReadStatusStarting, WxReadCardState{
		BookTitle:      "测试书籍",
		TargetReadTime: DefaultProgressNotifyEvery,
	}))
	notifier.Upsert(BuildWxReadCard(WxReadStatusReading, WxReadCardState{
		BookTitle:      "测试书籍",
		TargetReadTime: DefaultProgressNotifyEvery,
		TotalReadTime:  DefaultProgressNotifyEvery / 2,
	}))

	if len(requests) != 2 {
		t.Fatalf("request count = %d, want 2", len(requests))
	}
	if requests[0].MessageID != "" {
		t.Fatalf("first request messageId = %s, want empty", requests[0].MessageID)
	}
	if requests[1].MessageID != "om_test" {
		t.Fatalf("second request messageId = %s, want om_test", requests[1].MessageID)
	}
	if got := requests[1].TemplateVariable["title"]; got != "阅读进行中" {
		t.Fatalf("updated title = %v, want 阅读进行中", got)
	}
	if len(requests[0].Images) != 0 {
		t.Fatalf("first images length = %d, want 0", len(requests[0].Images))
	}
	if len(requests[1].Images) != 0 {
		t.Fatalf("updated images length = %d, want 0", len(requests[1].Images))
	}
}

func TestBuildWxReadLoginCardUsesQRCodeButton(t *testing.T) {
	card := BuildWxReadCard(WxReadStatusLoginRequired, WxReadCardState{
		BookTitle:         "测试书籍",
		TargetReadTime:    DefaultProgressNotifyEvery,
		QRCodeURL:         "https://example.com/qrcode",
		QRCodeImageBase64: "AQID",
	})
	vars := card.toTemplateVariable()

	if got := vars["sub_button_text"]; got != "打开登录二维码" {
		t.Fatalf("sub_button_text = %v, want 打开登录二维码", got)
	}
	if got := vars["sub_button_url"]; got != "https://example.com/qrcode" {
		t.Fatalf("sub_button_url = %v, want qrcode URL", got)
	}
	if got := vars["title_style"]; got != "orange" {
		t.Fatalf("title_style = %v, want orange", got)
	}
	images := card.toGatewayImages(DefaultQRCodeImageVariable)
	if len(images) != 1 {
		t.Fatalf("images length = %d, want 1", len(images))
	}
	if images[0].Variable != DefaultQRCodeImageVariable {
		t.Fatalf("images[0].Variable = %s, want %s", images[0].Variable, DefaultQRCodeImageVariable)
	}
	if images[0].Base64 != "AQID" {
		t.Fatalf("images[0].Base64 = %s, want qrcode base64", images[0].Base64)
	}
	if images[0].FileName != "wxread-login-qrcode.png" {
		t.Fatalf("images[0].FileName = %s, want wxread-login-qrcode.png", images[0].FileName)
	}
}

func TestFeishuCardNotifierSendsQRCodeImageThroughGateway(t *testing.T) {
	var request sendCardRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"messageId": "om_test",
			},
		})
	}))
	defer server.Close()

	notifier := NewFeishuCardNotifier(CardConfig{
		Enabled:             true,
		GatewayBaseURL:      server.URL,
		GatewayAuthToken:    "token",
		AppID:               "cli_test",
		TemplateID:          "ctp_test",
		QRCodeImageVariable: "login_qrcode",
	})
	notifier.Upsert(BuildWxReadCard(WxReadStatusLoginRequired, WxReadCardState{
		BookTitle:         "测试书籍",
		TargetReadTime:    DefaultProgressNotifyEvery,
		QRCodeURL:         "https://example.com/qrcode",
		QRCodeImageBase64: "AQID",
	}))

	if len(request.Images) != 1 {
		t.Fatalf("request.Images length = %d, want 1", len(request.Images))
	}
	if request.Images[0].Variable != "login_qrcode" {
		t.Fatalf("request.Images[0].Variable = %s, want login_qrcode", request.Images[0].Variable)
	}
	if request.Images[0].Base64 != "AQID" {
		t.Fatalf("request.Images[0].Base64 = %s, want qrcode base64", request.Images[0].Base64)
	}
	if request.Images[0].ContentType != "image/png" {
		t.Fatalf("request.Images[0].ContentType = %s, want image/png", request.Images[0].ContentType)
	}
}

func TestBuildWxReadReadingCardIncludesProgressSummary(t *testing.T) {
	card := BuildWxReadCard(WxReadStatusReading, WxReadCardState{
		BookTitle:        "测试书籍",
		TargetReadTime:   time.Hour,
		TotalReadTime:    15 * time.Minute,
		TotalReadPageCnt: 12,
		CatalogName:      "第一章",
		CatalogProgress:  "25.00% (1/4)",
	})

	for _, want := range []string{
		"🟡 **状态**：<font color='grey'>阅读中</font>",
		"📄 **已读**：12 页",
		"⏱️ **时长**：15分0秒",
		"📌 **章节**：第一章",
		"📊 **总进度**：25.00% (1/4)",
		"🎯 **目标**：",
	} {
		if !strings.Contains(card.Content, want) {
			t.Fatalf("card.Content missing %q:\n%s", want, card.Content)
		}
	}
}

func TestBuildWxReadFailedCardKeepsErrorSummaryOnly(t *testing.T) {
	card := BuildWxReadCard(WxReadStatusFailed, WxReadCardState{
		BookTitle:      "测试书籍",
		TargetReadTime: time.Hour,
		Error:          "chrome failed to start",
		Detail:         "任务已停止。",
	})

	for _, want := range []string{
		"🔴 **状态**：<font color='red'>阅读失败</font>",
		"⚠️ **异常**：chrome failed to start",
		"💬 **提示**：任务已停止。",
	} {
		if !strings.Contains(card.Content, want) {
			t.Fatalf("card.Content missing %q:\n%s", want, card.Content)
		}
	}
}

func TestNewCardConfigFromEnvParsesNotificationConfigJSON(t *testing.T) {
	env := map[string]string{
		NotificationConfigEnv: `{
			"gatewayBaseUrl": "https://gateway.example.com",
			"gatewayAuthToken": "token",
			"appId": "cli_test",
			"templateId": "ctp_test",
			"templateVersionName": "1.0.0",
			"receiveIdType": "email",
			"receiveId": "name@example.com",
			"appName": "微信读书",
			"openId": "ou_test",
			"defaultUrl": "https://weread.qq.com/",
			"progressNotifySeconds": 30
		}`,
	}

	cfg := NewCardConfigFromEnv(func(key string) string {
		return env[key]
	})

	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if cfg.GatewayBaseURL != "https://gateway.example.com" {
		t.Fatalf("GatewayBaseURL = %s", cfg.GatewayBaseURL)
	}
	if cfg.ProgressNotifyEvery != 30*time.Second {
		t.Fatalf("ProgressNotifyEvery = %s, want 30s", cfg.ProgressNotifyEvery)
	}
}

func TestNewCardConfigFromEnvMissingFieldsStaysEnabledForErrorLog(t *testing.T) {
	cfg := NewCardConfigFromEnv(func(key string) string {
		if key == NotificationConfigEnv {
			return `{}`
		}
		return ""
	})

	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true for configured but invalid notification JSON")
	}
	if err := cfg.validate(); err == nil {
		t.Fatal("validate() err = nil, want missing field error")
	}
}

func TestTemplateVariableNormalizesEmptyMainButtonEvent(t *testing.T) {
	card := BuildWxReadCard(WxReadStatusStarting, WxReadCardState{
		BookTitle:      "测试书籍",
		TargetReadTime: DefaultProgressNotifyEvery,
	})
	vars := card.toTemplateVariable()

	event, ok := vars["main_button_event"].(map[string]any)
	if !ok {
		t.Fatalf("main_button_event = %T, want object", vars["main_button_event"])
	}
	if event["action"] != "noop" {
		t.Fatalf("main_button_event.action = %v, want noop", event["action"])
	}
}
