package llm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// OpenAIAdapter OpenAI 适配器
type OpenAIAdapter struct {
	baseURL string
	model   string
	apiKey  string
	client  *resty.Client
}

// NewOpenAIAdapter 创建 OpenAI 适配器
func NewOpenAIAdapter(baseURL, model, apiKey string) (*OpenAIAdapter, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	client := resty.New().
		SetTimeout(120 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(3 * time.Second)

	return &OpenAIAdapter{
		baseURL: baseURL,
		model:   model,
		apiKey:  apiKey,
		client:  client,
	}, nil
}

// Chat 发送对话请求
func (a *OpenAIAdapter) Chat(messages *[]Message, tools []ToolDefinition) (Response, error) {
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
		Post(a.baseURL + "/chat/completions")

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
func (a *OpenAIAdapter) ChatStream(messages *[]Message, tools []ToolDefinition, onChunk func(string)) (string, error) {
	return "", fmt.Errorf("stream mode not implemented")
}
