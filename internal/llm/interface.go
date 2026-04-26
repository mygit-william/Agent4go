package llm

// Message 消息结构
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string    `json:"tool_call_id,omitempty"`
	Name      string     `json:"name,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
	Result  string `json:"result,omitempty"`  // 执行结果
	Success bool   `json:"success,omitempty"` // 是否成功
}

// Response LLM 响应
type Response struct {
	Reply string     `json:"reply"`
	Tool  []ToolCall `json:"tool"`
}

// Interface LLM 接口
type Interface interface {
	Chat(messages *[]Message, tools []ToolDefinition) (Response, error)
	ChatStream(messages *[]Message, tools []ToolDefinition, onChunk func(string)) (string, error)
}

// ToolDefinition 工具定义
type ToolDefinition struct {
	Type     string `json:"type"`
	Function struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"function"`
}

// Config LLM 配置
type Config struct {
	DefaultProvider string                    `json:"default_provider"`
	Providers       map[string]ProviderConfig `json:"providers"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Driver  string `json:"driver"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}
