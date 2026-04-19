package core

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/utils"
)

// Chat 核心对话循环
func (a *Agent) Chat(sessionID, input string, messages *[]llm.Message) string {
	*messages = append(*messages, llm.Message{
		Role:    "user",
		Content: input,
	})

	toolsDef := a.toolManager.GetFunctionDefinitions()

	for step := 0; step < MaxExecutionSteps; step++ {
		// 调用 LLM
		loading := utils.NewLoadingAnimation("AI 思考中")
		loading.Start()
		resp, err := a.llm.Chat(messages, toolsDef)
		loading.Stop()

		if err != nil {
			red := color.New(color.FgRed)
			red.Printf("❌ LLM 调用失败: %v\n", err)
			return fmt.Sprintf("LLM 调用出错: %v", err)
		}

		// 没有工具调用，直接返回
		if len(resp.Tool) == 0 {
			green := color.New(color.FgGreen)
			green.Println("✅ AI 回复完成")
			a.saveToLongTermMemory(sessionID, input, resp.Reply)
			return resp.Reply
		}

		// 有工具调用
		yellow := color.New(color.FgYellow)
		yellow.Printf("🔧 AI 请求执行 %d 个工具\n", len(resp.Tool))

		// 构建 Assistant 消息
		*messages = append(*messages, llm.Message{
			Role:      "assistant",
			Content:   resp.Reply,
			ToolCalls: resp.Tool,
		})

		// 执行所有工具调用
		for _, toolCall := range resp.Tool {
			context := map[string]interface{}{
				"tool": map[string]interface{}{
					"name":      toolCall.Function.Name,
					"arguments": toolCall.Function.Arguments,
				},
			}

			// 触发 PRE_ACTION Hook
			context = a.triggerHooks("PRE_ACTION", context)
			if a.shouldStop(context) {
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

			// 收集结果
			*messages = append(*messages, llm.Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
				Content:     output,
			})
		}
	}

	return "执行达到最大步骤限制"
}

// triggerHooks 触发所有 Hook
func (a *Agent) triggerHooks(event string, context map[string]interface{}) map[string]interface{} {
	for _, hook := range a.hooks {
		context = hook.Handle(event, context)
		if a.shouldStop(context) {
			return context
		}
	}
	return context
}

// shouldStop 检查是否应停止
func (a *Agent) shouldStop(context map[string]interface{}) bool {
	decision, ok := context["decision"].(string)
	return ok && decision == "deny"
}

// saveToLongTermMemory 保存长期记忆
func (a *Agent) saveToLongTermMemory(sessionID, input, output string) {
	// TODO: 实现文件存储
}
