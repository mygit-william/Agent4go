package tools

// Tool 工具接口
type Tool interface {
	// Name 工具名称
	Name() string

	// Description 工具描述
	Description() string

	// Parameters 参数定义
	Parameters() map[string]interface{}

	// Required 必填参数
	Required() []string

	// Execute 执行工具
	Execute(args map[string]interface{}) string
}
