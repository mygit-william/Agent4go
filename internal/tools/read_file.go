package tools

import (
	"fmt"
	"os"
	"strings"
)

// ReadFile 读取文件工具
type ReadFile struct{}

// NewReadFile 创建工具
func NewReadFile() *ReadFile {
	return &ReadFile{}
}

// Name 工具名称
func (t *ReadFile) Name() string {
	return "read_file"
}

// Description 工具描述
func (t *ReadFile) Description() string {
	return "Read a file by the path. 结果将以带行号的格式返回,行号从1开始. The path must be absolute path."
}

// Parameters 参数定义
func (t *ReadFile) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "要读取的文件的绝对路径",
		},
	}
}

// Required 必填参数
func (t *ReadFile) Required() []string {
	return []string{"path"}
}

// Execute 执行工具
func (t *ReadFile) Execute(args map[string]interface{}) string {
	path, ok := args["path"].(string)
	if !ok {
		return "错误: 缺少 path 参数"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Sprintf("错误: 文件不存在 -> %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("错误: 无法读取文件 -> %s", path)
	}

	lines := strings.Split(string(data), "\n")
	var output strings.Builder
	for i, line := range lines {
		output.WriteString(fmt.Sprintf("%d\t%s\n", i+1, line))
	}

	return strings.TrimRight(output.String(), "\n")
}
