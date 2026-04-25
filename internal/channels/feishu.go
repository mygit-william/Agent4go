package channels

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// FeishuNotifier 飞书 Webhook 通知器
type FeishuNotifier struct {
	webhookURL string
	enabled    bool
	client     *resty.Client
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhook_url"`
}

// NewFeishuNotifier 创建飞书通知器
func NewFeishuNotifier(cfg FeishuConfig) *FeishuNotifier {
	return &FeishuNotifier{
		webhookURL: cfg.WebhookURL,
		enabled:    cfg.Enabled && cfg.WebhookURL != "",
		client:     resty.New(),
	}
}

// feishuTextRequest 飞书文本消息请求体
type feishuTextRequest struct {
	MsgType string             `json:"msg_type"`
	Content feishuTextContent  `json:"content"`
}

type feishuTextContent struct {
	Text string `json:"text"`
}

// Notify 发送飞书通知
func (f *FeishuNotifier) Notify(text string) error {
	if !f.enabled {
		return nil
	}

	resp, err := f.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(feishuTextRequest{
			MsgType: "text",
			Content: feishuTextContent{
				Text: text,
			},
		}).
		Post(f.webhookURL)

	if err != nil {
		return fmt.Errorf("飞书通知发送失败: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("飞书通知发送失败, HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// NotifyTaskComplete 任务完成通知
func (f *FeishuNotifier) NotifyTaskComplete(taskSummary string) error {
	text := fmt.Sprintf("🤖 Nanobot 任务完成\n📝 %s", taskSummary)
	return f.Notify(text)
}

// IsEnabled 是否启用
func (f *FeishuNotifier) IsEnabled() bool {
	return f.enabled
}
