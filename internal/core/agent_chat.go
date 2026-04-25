package core

import (
	"fmt"
	"strings"

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
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("任务开始")
	fmt.Printf("└─ 用户输入: %s\n", compactLine(input, 140))

	for step := 0; step < MaxExecutionSteps; step++ {
		fmt.Printf("\n├─ 步骤 %d/%d\n", step+1, MaxExecutionSteps)
		// 调用 LLM
		loading := utils.NewLoadingAnimation("思考中")
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
			green.Println("✔ 任务完成，回复已生成")
			fmt.Println(strings.Repeat("=", 60))
			a.saveToLongTermMemory(sessionID, input, resp.Reply)

			// 飞书通知：任务完成
			a.notifyComplete(input, resp.Reply)

			return resp.Reply
		}

		// 有工具调用
		yellow := color.New(color.FgYellow)
		yellow.Printf("│  ├─ 执行工具: %d 个\n", len(resp.Tool))

		// 构建 Assistant 消息
		*messages = append(*messages, llm.Message{
			Role:      "assistant",
			Content:   resp.Reply,
			ToolCalls: resp.Tool,
		})

		// 执行所有工具调用
		for i, toolCall := range resp.Tool {
			isLast := i == len(resp.Tool)-1
			branch := "│  ├─"
			if isLast {
				branch = "│  └─"
			}
			fmt.Printf("%s [%d/%d] %s\n", branch, i+1, len(resp.Tool), toolCall.Function.Name)
			context := map[string]interface{}{
				"tool": map[string]interface{}{
					"name":      toolCall.Function.Name,
					"arguments": toolCall.Function.Arguments,
				},
			}

			// 触发 PRE_ACTION Hook
			context = a.triggerHooks("PRE_ACTION", context)
			if a.shouldStop(context) {
				fmt.Printf("│     ✖ 工具 %s 被策略拒绝\n", toolCall.Function.Name)
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

	// 飞书通知：达到步骤上限
	a.notifyComplete(input, "执行达到最大步骤限制")

	return "执行达到最大步骤限制"
}

// notifyComplete 任务完成后发送通知
func (a *Agent) notifyComplete(input, output string) {
	if a.notifier != nil && a.notifier.IsEnabled() {
		summary := fmt.Sprintf("输入: %s\n输出: %s", compactLine(input, 200), compactLine(output, 500))
		if err := a.notifier.NotifyTaskComplete(summary); err != nil {
			red := color.New(color.FgRed)
			red.Printf("⚠ 飞书通知发送失败: %v\n", err)
		}
	}
}

func compactLine(s string, maxLen int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
