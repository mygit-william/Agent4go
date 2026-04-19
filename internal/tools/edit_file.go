package tools

import (
	"fmt"
	"os"
	"strings"
)

// EditFile 编辑文件工具
type EditFile struct{}

// NewEditFile 创建工具
func NewEditFile() *EditFile {
	return &EditFile{}
}

// Name 工具名称
func (t *EditFile) Name() string {
	return "edit_file"
}

// Description 工具描述
func (t *EditFile) Description() string {
	return "Edit file by replacing, inserting or appending content. The path must be absolute path."
}

// Parameters 参数定义
func (t *EditFile) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "文件绝对路径",
		},
		"operation": map[string]interface{}{
			"type":        "string",
			"enum":        []string{"replace", "insert", "append", "prepend"},
			"description": "操作类型: replace/insert/append/prepend",
		},
		"new_content": map[string]interface{}{
			"type":        "string",
			"description": "新内容",
		},
		"old_content": map[string]interface{}{
			"type":        "string",
			"description": "旧内容(用于 replace 操作)",
		},
		"line_number": map[string]interface{}{
			"type":        "integer",
			"description": "插入时的目标行号",
		},
	}
}

// Required 必填参数
func (t *EditFile) Required() []string {
	return []string{"path", "operation", "new_content"}
}

// Execute 执行工具
func (t *EditFile) Execute(args map[string]interface{}) string {
	path, ok := args["path"].(string)
	if !ok {
		return "错误: 缺少 path 参数"
	}

	operation, ok := args["operation"].(string)
	if !ok {
		return "错误: 缺少 operation 参数"
	}

	newContent, ok := args["content"].(string)
	if !ok {
		if newContent, ok = args["new_content"].(string); !ok {
			return "错误: 缺少 content/new_content 参数"
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprintf("错误: 文件不存在 -> %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("错误: 无法读取文件 -> %s", path)
	}

	contentStr := string(content)
	var result string

	switch operation {
	case "replace":
		oldContent, ok := args["old_content"].(string)
		if !ok {
			return "错误: replace 操作需要 old_content 参数"
		}
		if !strings.Contains(contentStr, oldContent) {
			return "错误: 未找到要替换的内容"
		}
		result = strings.Replace(contentStr, oldContent, newContent, 1)

	case "append":
		result = contentStr + newContent

	case "prepend":
		result = newContent + contentStr

	case "insert":
		lineNumber, ok := args["line_number"].(float64)
		if !ok {
			return "错误: insert 操作需要 line_number 参数"
		}
		lines := strings.Split(contentStr, "\n")
		lineIdx := int(lineNumber)
		if lineIdx < 1 || lineIdx > len(lines)+1 {
			return "错误: 行号超出范围"
		}
		lines = append(lines[:lineIdx-1], append([]string{newContent}, lines[lineIdx-1:]...)...)
		result = strings.Join(lines, "\n")

	default:
		return fmt.Sprintf("错误: 不支持的操作类型 -> %s", operation)
	}

	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		return "错误: 文件编辑失败"
	}

	return fmt.Sprintf("文件编辑成功(%s操作)", operation)
}
