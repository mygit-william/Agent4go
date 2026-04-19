package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile 写入文件工具
type WriteFile struct{}

// NewWriteFile 创建工具
func NewWriteFile() *WriteFile {
	return &WriteFile{}
}

// Name 工具名称
func (t *WriteFile) Name() string {
	return "write_file"
}

// Description 工具描述
func (t *WriteFile) Description() string {
	return "Write content to file. For partial edits, prefer edit_file instead. The path must be absolute path."
}

// Parameters 参数定义
func (t *WriteFile) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "要写入的文件的绝对路径",
		},
		"content": map[string]interface{}{
			"type":        "string",
			"description": "要写入的内容",
		},
	}
}

// Required 必填参数
func (t *WriteFile) Required() []string {
	return []string{"path", "content"}
}

// Execute 执行工具
func (t *WriteFile) Execute(args map[string]interface{}) string {
	path, ok := args["path"].(string)
	if !ok {
		return "错误: 缺少 path 参数"
	}

	content, ok := args["content"].(string)
	if !ok {
		return "错误: 缺少 content 参数"
	}

	// 创建目录
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Sprintf("错误: 无法创建目录 -> %s", dir)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf("错误: 文件写入失败 -> %s", path)
	}

	return "文件写入成功"
}
