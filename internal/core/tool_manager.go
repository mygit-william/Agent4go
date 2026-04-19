package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/tools"
	"github.com/mygit-william/nanobot-go/internal/utils"
)

// ToolManager 工具管理器
type ToolManager struct {
	tools map[string]tools.Tool
}

// NewToolManager 创建工具管理器
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]tools.Tool),
	}
}

// Register 注册工具
func (tm *ToolManager) Register(tool tools.Tool) {
	tm.tools[tool.Name()] = tool
}

// GetFunctionDefinitions 获取工具定义
func (tm *ToolManager) GetFunctionDefinitions() []llm.ToolDefinition {
	definitions := make([]llm.ToolDefinition, 0, len(tm.tools))

	for _, tool := range tm.tools {
		params := tool.Parameters()
		def := llm.ToolDefinition{
			Type: "function",
		}
		def.Function.Name = tool.Name()
		def.Function.Description = tool.Description()
		def.Function.Parameters = map[string]interface{}{
			"type":       "object",
			"properties": params,
			"required":   tool.Required(),
		}
		definitions = append(definitions, def)
	}

	return definitions
}

// Run 执行工具（带动画）
func (tm *ToolManager) Run(toolName string, arguments string) string {
	tool, ok := tm.tools[toolName]
	if !ok {
		red := color.New(color.FgRed)
		red.Printf("❌ 工具不存在: %s\n", toolName)
		return fmt.Sprintf("错误: 工具不存在 - %s", toolName)
	}

	// 解析参数
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			red := color.New(color.FgRed)
			red.Printf("❌ 参数解析失败: %v\n", err)
			return fmt.Sprintf("错误: 参数解析失败 - %v", err)
		}
	}

	// 显示开始动画 - 显示工具名和关键参数
	detail := extractToolDetail(toolName, args)
	loading := utils.NewLoadingAnimation(fmt.Sprintf("执行 %s %s", toolName, detail))
	loading.Start()

	// 执行工具
	result := tool.Execute(args)

	// 判断是否成功
	success := !strings.HasPrefix(result, "错误")

	// 停止动画并显示结果摘要
	loading.StopWithResult(success, result)

	return result
}

// HasTool 检查工具是否存在
func (tm *ToolManager) HasTool(name string) bool {
	_, ok := tm.tools[name]
	return ok
}

// extractToolDetail 提取工具的关键参数用于显示
func extractToolDetail(toolName string, args map[string]interface{}) string {
	switch toolName {
	case "bash":
		if cmd, ok := args["command"].(string); ok {
			// 清理命令中的换行符并截断
			cmd = cleanCommand(cmd)
			if len(cmd) > 50 {
				return cmd[:50] + "..."
			}
			return cmd
		}
	case "read_file":
		if path, ok := args["path"].(string); ok {
			return "读取 " + path
		}
	case "write_file":
		if path, ok := args["path"].(string); ok {
			return "写入 " + path
		}
	case "edit_file":
		if path, ok := args["path"].(string); ok {
			operation := "编辑"
			if op, ok := args["operation"].(string); ok {
				operation = mapOperation(op)
			}
			return fmt.Sprintf("%s %s", operation, path)
		}
	}
	return ""
}

// cleanCommand 清理命令显示
func cleanCommand(cmd string) string {
	// 替换换行为空格，多个空格为一个
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	cmd = strings.ReplaceAll(cmd, "\r", "")
	for strings.Contains(cmd, "  ") {
		cmd = strings.ReplaceAll(cmd, "  ", " ")
	}
	return strings.TrimSpace(cmd)
}

// mapOperation 映射操作名称
func mapOperation(op string) string {
	switch op {
	case "replace":
		return "替换"
	case "insert":
		return "插入"
	case "append":
		return "追加"
	case "prepend":
		return "前置"
	default:
		return op
	}
}
