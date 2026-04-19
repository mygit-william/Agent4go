package hooks

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PermissionCheckHook 权限检查 Hook
type PermissionCheckHook struct {
	BaseHook
	mode          string
	writeTools    []string
	readOnlyTools []string
}

// NewPermissionCheckHook 创建权限检查 Hook
func NewPermissionCheckHook(mode string) *PermissionCheckHook {
	return &PermissionCheckHook{
		mode: mode,
		writeTools: []string{
			"write_file", "edit_file", "delete_file", "bash",
		},
		readOnlyTools: []string{
			"read_file", "web_search", "web_fetch",
		},
	}
}

// Handle 处理事件
func (h *PermissionCheckHook) Handle(event string, context map[string]interface{}) map[string]interface{} {
	if event != EventPreAction {
		return context
	}

	tool, ok := context["tool"].(map[string]interface{})
	if !ok {
		return context
	}

	toolName, _ := tool["name"].(string)

	// 根据模式检查权限
	switch h.mode {
	case "plan":
		// plan 模式：只允许只读工具
		if h.isReadOnly(toolName) {
			return context
		}
		return map[string]interface{}{
			"decision": "deny",
			"reason":   "plan 模式禁止写操作",
			"tool":     tool,
		}

	case "default":
		// default 模式：读自动通过，写操作需要确认
		if h.isReadOnly(toolName) {
			return context
		}
		if h.confirmOperation(toolName, tool) {
			return context
		}
		return map[string]interface{}{
			"decision": "deny",
			"reason":   "用户取消操作",
			"tool":     tool,
		}

	case "auto":
		// auto 模式：所有操作自动通过
		return context

	default:
		return context
	}
}

// confirmOperation 交互式确认操作
func (h *PermissionCheckHook) confirmOperation(toolName string, tool map[string]interface{}) bool {
	yellow := color.New(color.FgYellow, color.Bold)
	cyan := color.New(color.FgCyan)
	white := color.New(color.FgWhite)

	yellow.Println("\n⚠️  需要确认")
	white.Printf("工具: ")
	cyan.Println(toolName)

	// 显示参数
	if args, ok := tool["arguments"].(string); ok && args != "" {
		white.Printf("参数: ")
		// 格式化 JSON 参数
		args = strings.ReplaceAll(args, "\n", " ")
		if len(args) > 100 {
			args = args[:100] + "..."
		}
		fmt.Println(args)
	}

	// 显示提示
	white.Print("是否执行? [y/N]: ")

	// 读取用户输入
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes"
}

func (h *PermissionCheckHook) isReadOnly(toolName string) bool {
	for _, t := range h.readOnlyTools {
		if t == toolName {
			return true
		}
	}
	return false
}

func (h *PermissionCheckHook) isWrite(toolName string) bool {
	for _, t := range h.writeTools {
		if t == toolName {
			return true
		}
	}
	return false
}
