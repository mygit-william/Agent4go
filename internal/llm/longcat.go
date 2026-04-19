package llm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// LongcatAdapter Longcat 适配器
type LongcatAdapter struct {
	baseURL string
	model   string
	apiKey  string
	client  *resty.Client
}

// NewLongcatAdapter 创建 Longcat 适配器
func NewLongcatAdapter(baseURL, model, apiKey string) (*LongcatAdapter, error) {
	client := resty.New().
		SetTimeout(120 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(3 * time.Second)

	return &LongcatAdapter{
		baseURL: baseURL,
		model:   model,
		apiKey:  apiKey,
		client:  client,
	}, nil
}

// Chat 发送对话请求
func (a *LongcatAdapter) Chat(messages *[]Message, tools []ToolDefinition) (Response, error) {
	if len(*messages) == 0 {
		return Response{}, fmt.Errorf("messages 不能为空")
	}

	payload := map[string]interface{}{
		"model":      a.model,
		"messages":   *messages,
		"stream":     false,
		"temperature": 0.7,
	}

	if len(tools) > 0 {
		payload["tools"] = tools
	}

	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+a.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(a.baseURL + "/v1/chat/completions")

	if err != nil {
		return Response{}, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return Response{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content   interface{} `json:"content"`
				ToolCalls []ToolCall  `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return Response{}, fmt.Errorf("JSON 解析失败: %v", err)
	}

	if len(result.Choices) == 0 {
		return Response{Reply: "LLM 调用失败，未获取到响应"}, nil
	}

	reply := ""
	switch v := result.Choices[0].Message.Content.(type) {
	case string:
		reply = v
	default:
		data, _ := json.Marshal(v)
		reply = string(data)
	}

	return Response{
		Reply: reply,
		Tool:  result.Choices[0].Message.ToolCalls,
	}, nil
}

// ChatStream 流式对话
func (a *LongcatAdapter) ChatStream(messages *[]Message, tools []ToolDefinition, onChunk func(string)) (string, error) {
	return "", fmt.Errorf("stream mode not implemented")
}
