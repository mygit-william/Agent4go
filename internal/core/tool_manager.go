package core

import (
	"encoding/json"
	"fmt"
	"strings"
	// "time"
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
	actionLine := fmt.Sprintf("工具 %s", toolName)
	if detail != "" {
		actionLine = fmt.Sprintf("%s (%s)", actionLine, detail)
	}
	fmt.Printf("│     ├─ 调用: %s\n", actionLine)
	loading := utils.NewLoadingAnimation("处理中")
	loading.Start()
	//sleep 3 seconds
	// time.Sleep(10 * time.Second)
	// 执行工具
	result := tool.Execute(args)

	// 判断是否成功
	success := !strings.HasPrefix(result, "错误")

	// 停止动画并显示结果摘要
	loading.Stop()
	if !success {
		fmt.Println("│     └─ 状态: 失败")
	} else {
		fmt.Println("│     └─ 状态: 完成")
	}
	printToolResultSummary(toolName, args, result)

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

func printToolResultSummary(toolName string, args map[string]interface{}, result string) {
	trimmed := strings.TrimSpace(result)
	if trimmed == "" {
		fmt.Println("│        结果: 空输出")
		return
	}

	switch toolName {
	case "read_file":
		printReadFileSummary(args, trimmed)
	case "write_file", "edit_file":
		printWriteLikeSummary(args, trimmed)
	case "bash":
		printBashSummary(args, trimmed)
	default:
		printGenericSummary(trimmed)
	}
}

func printReadFileSummary(args map[string]interface{}, result string) {
	path, _ := args["path"].(string)
	lines := strings.Split(result, "\n")
	fmt.Printf("│        已读取: %s\n", path)
	fmt.Printf("│        行数: %d\n", len(lines))

	previewCount := minInt(5, len(lines))
	if previewCount == 0 {
		fmt.Println("│        预览: (空文件)")
		return
	}
	fmt.Println("│        预览:")
	for i := 0; i < previewCount; i++ {
		fmt.Printf("│          %s\n", lines[i])
	}
	if len(lines) > previewCount {
		fmt.Printf("│          ... (还有 %d 行)\n", len(lines)-previewCount)
	}
}

func printWriteLikeSummary(args map[string]interface{}, result string) {
	path, _ := args["path"].(string)
	fmt.Printf("│        目标文件: %s\n", path)

	content := ""
	if c, ok := args["content"].(string); ok {
		content = c
	} else if c, ok := args["new_content"].(string); ok {
		content = c
	}

	content = strings.TrimSpace(content)
	if content == "" {
		fmt.Printf("│        执行结果: %s\n", firstLine(result))
		return
	}

	lines := strings.Split(content, "\n")
	previewCount := minInt(4, len(lines))
	fmt.Printf("│        写入内容预览 (%d 行):\n", len(lines))
	for i := 0; i < previewCount; i++ {
		fmt.Printf("│          %s\n", lines[i])
	}
	if len(lines) > previewCount {
		fmt.Printf("│          ... (还有 %d 行)\n", len(lines)-previewCount)
	}
	fmt.Printf("│        执行结果: %s\n", firstLine(result))
}

func printBashSummary(args map[string]interface{}, result string) {
	command, _ := args["command"].(string)
	fmt.Printf("│        命令: %s\n", cleanCommand(command))

	lines := strings.Split(result, "\n")
	previewCount := minInt(8, len(lines))
	if strings.TrimSpace(result) == "" || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		fmt.Println("│        输出: (无输出)")
		return
	}
	fmt.Println("│        输出预览:")
	for i := 0; i < previewCount; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		fmt.Printf("│          %s\n", lines[i])
	}
	if len(lines) > previewCount {
		fmt.Printf("│          ... (还有 %d 行)\n", len(lines)-previewCount)
	}
}

func printGenericSummary(result string) {
	lines := strings.Split(result, "\n")
	display := firstLine(result)
	if len(display) > 120 {
		display = display[:120] + "..."
	}
	if len(lines) > 1 {
		fmt.Printf("│        结果: %s (另有 %d 行)\n", display, len(lines)-1)
		return
	}
	fmt.Printf("│        结果: %s\n", display)
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	return strings.TrimSpace(lines[0])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
