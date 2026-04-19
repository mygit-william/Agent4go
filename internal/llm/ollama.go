package llm

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// OllamaAdapter Ollama 本地模型适配器
type OllamaAdapter struct {
	baseURL string
	model   string
	client  *resty.Client
}

// NewOllamaAdapter 创建 Ollama 适配器
func NewOllamaAdapter(baseURL, model string) (*OllamaAdapter, error) {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := resty.New().
		SetTimeout(120 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(3 * time.Second)

	return &OllamaAdapter{
		baseURL: baseURL,
		model:   model,
		client:  client,
	}, nil
}

// Chat 发送对话请求
func (a *OllamaAdapter) Chat(messages *[]Message, tools []ToolDefinition) (Response, error) {
	if len(*messages) == 0 {
		return Response{}, fmt.Errorf("messages 不能为空")
	}

	payload := map[string]interface{}{
		"model":    a.model,
		"messages": *messages,
		"stream":   false,
	}

	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(a.baseURL + "/api/chat")

	if err != nil {
		return Response{}, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return Response{}, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return Response{}, fmt.Errorf("JSON 解析失败: %v", err)
	}

	return Response{
		Reply: result.Message.Content,
		Tool:  nil,
	}, nil
}

// ChatStream 流式对话
func (a *OllamaAdapter) ChatStream(messages *[]Message, tools []ToolDefinition, onChunk func(string)) (string, error) {
	return "", fmt.Errorf("stream mode not implemented")
}
