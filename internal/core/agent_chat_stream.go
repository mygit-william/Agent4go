package core

import (
	"fmt"
	"strings"

	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/utils"
)

// ChatStream 流式对话循环（供 Web/SSE 使用）
// 与 Chat() 逻辑一致，但通过 StreamHandler 输出事件而非直接打印到终端
func (a *Agent) ChatStream(sessionID, input string, messages *[]llm.Message, handler StreamHandler) string {
	*messages = append(*messages, llm.Message{
		Role:    "user",
		Content: input,
	})

	toolsDef := a.toolManager.GetFunctionDefinitions()

	// 发送任务开始事件
	handler.OnEvent(StreamEvent{
		Type:    EventTypeStart,
		Content: input,
	})

	for step := 0; step < MaxExecutionSteps; step++ {
		// 发送步骤开始事件
		handler.OnEvent(StreamEvent{
			Type:     EventTypeThinking,
			Step:     step + 1,
			MaxSteps: MaxExecutionSteps,
			Content:  "思考中...",
		})

		// 调用 LLM
		loading := utils.NewLoadingAnimation("思考中")
		loading.Start()
		resp, err := a.llm.Chat(messages, toolsDef)
		loading.Stop()

		if err != nil {
			handler.OnEvent(StreamEvent{
				Type:  EventTypeError,
				Error: fmt.Sprintf("LLM 调用失败: %v", err),
			})
			handler.OnError(err)
			return fmt.Sprintf("LLM 调用出错: %v", err)
		}

		// DEBUG: 记录 LLM 响应
		fmt.Printf("[DEBUG] ChatStream step %d: reply=%q, tool_calls=%d\n", step+1, resp.Reply, len(resp.Tool))

		// 没有工具调用，直接返回
		if len(resp.Tool) == 0 {
			// 追加 assistant 回复到消息历史
			*messages = append(*messages, llm.Message{
				Role:    "assistant",
				Content: resp.Reply,
			})

			handler.OnEvent(StreamEvent{
				Type:    EventTypeAssistant,
				Content: resp.Reply,
			})
			handler.OnEvent(StreamEvent{
				Type: EventTypeComplete,
			})

			a.saveToLongTermMemory(sessionID, input, resp.Reply)
			a.notifyComplete(input, resp.Reply)

			return resp.Reply
		}

		// 有工具调用
		handler.OnEvent(StreamEvent{
			Type:      EventTypeThinking,
			Step:      step + 1,
			MaxSteps:  MaxExecutionSteps,
			Content:   fmt.Sprintf("执行 %d 个工具", len(resp.Tool)),
			ToolTotal: len(resp.Tool),
		})

		// 构建 Assistant 消息
		*messages = append(*messages, llm.Message{
			Role:      "assistant",
			Content:   resp.Reply,
			ToolCalls: resp.Tool,
		})

		// 执行所有工具调用
		for i, toolCall := range resp.Tool {
			// 发送工具调用事件
			handler.OnEvent(StreamEvent{
				Type:      EventTypeToolCall,
				Step:      step + 1,
				ToolName:  toolCall.Function.Name,
				ToolArgs:  toolCall.Function.Arguments,
				ToolIndex: i + 1,
				ToolTotal: len(resp.Tool),
			})

			context := map[string]interface{}{
				"tool": map[string]interface{}{
					"name":      toolCall.Function.Name,
					"arguments": toolCall.Function.Arguments,
				},
			}

			// 触发 PRE_ACTION Hook
			context = a.triggerHooks("PRE_ACTION", context)
			if a.shouldStop(context) {
				handler.OnEvent(StreamEvent{
					Type:     EventTypeToolResult,
					ToolName: toolCall.Function.Name,
					Result:   "系统拒绝执行该操作",
					Success:  false,
				})
				*messages = append(*messages, llm.Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Name:       toolCall.Function.Name,
					Content:    "系统拒绝执行该操作",
				})
				continue
			}

			// 执行工具
			output := a.toolManager.Run(toolCall.Function.Name, toolCall.Function.Arguments)

			// 触发 POST_ACTION Hook
			context["tool_execution"] = map[string]interface{}{
				"tool_name": toolCall.Function.Name,
				"output":    output,
			}
			a.triggerHooks("POST_ACTION", context)

			// 发送工具结果事件
			success := !strings.HasPrefix(output, "错误")
			handler.OnEvent(StreamEvent{
				Type:     EventTypeToolResult,
				ToolName: toolCall.Function.Name,
				Result:   output,
				Success:  success,
			})

			// 收集结果
			*messages = append(*messages, llm.Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
				Content:    output,
			})
		}
	}

	// 达到步骤上限
	handler.OnEvent(StreamEvent{
		Type:  EventTypeStepLimit,
		Error: "执行达到最大步骤限制",
	})
	a.notifyComplete(input, "执行达到最大步骤限制")

	return "执行达到最大步骤限制"
}
