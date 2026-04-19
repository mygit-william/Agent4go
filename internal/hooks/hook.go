package hooks

const (
	EventPreAction  = "PRE_ACTION"
	EventPostAction = "POST_ACTION"
)

// Interface Hook 接口
type Interface interface {
	Handle(event string, context map[string]interface{}) map[string]interface{}
}

// BaseHook 基础 Hook
type BaseHook struct{}

func (h *BaseHook) Handle(event string, context map[string]interface{}) map[string]interface{} {
	return context
}
