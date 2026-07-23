package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
}

func TestBuildWxReadLoginCardUsesQRCodeButton(t *testing.T) {
	card := BuildWxReadCard(WxReadStatusLoginRequired, WxReadCardState{
		BookTitle:      "测试书籍",
		TargetReadTime: DefaultProgressNotifyEvery,
		QRCodeURL:      "https://example.com/qrcode",
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
