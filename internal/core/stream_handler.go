package core

// StreamEventType 流式事件类型
type StreamEventType string

const (
	EventTypeStart       StreamEventType = "start"        // 任务开始
	EventTypeThinking    StreamEventType = "thinking"     // LLM 思考中（步骤信息）
	EventTypeToolCall    StreamEventType = "tool_call"    // 工具调用
	EventTypeToolResult  StreamEventType = "tool_result"  // 工具执行结果
	EventTypeAssistant   StreamEventType = "assistant"    // Assistant 回复内容
	EventTypeComplete    StreamEventType = "complete"     // 任务完成
	EventTypeError       StreamEventType = "error"        // 错误
	EventTypeStepLimit   StreamEventType = "step_limit"   // 达到步骤上限
)

// StreamEvent 流式事件
type StreamEvent struct {
	Type      StreamEventType `json:"type"`
	Step      int             `json:"step,omitempty"`
	MaxSteps  int             `json:"max_steps,omitempty"`
	Content   string          `json:"content,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolArgs  string          `json:"tool_args,omitempty"`
	ToolIndex int             `json:"tool_index,omitempty"`
	ToolTotal int             `json:"tool_total,omitempty"`
	Result    string          `json:"result,omitempty"`
	Success   bool            `json:"success,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// StreamHandler 流式处理接口
type StreamHandler interface {
	// OnEvent 收到事件时调用
	OnEvent(event StreamEvent)
	// OnComplete 任务完成时调用（返回最终结果）
	OnComplete(finalReply string)
	// OnError 发生错误时调用
	OnError(err error)
}

// NoOpStreamHandler 空实现（用于兼容旧代码）
type NoOpStreamHandler struct{}

func (n *NoOpStreamHandler) OnEvent(event StreamEvent) {}
func (n *NoOpStreamHandler) OnComplete(finalReply string) {}
func (n *NoOpStreamHandler) OnError(err error) {}
