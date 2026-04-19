package core

import (
	"github.com/mygit-william/nanobot-go/internal/hooks"
	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/tools"
)

const (
	MaxExecutionSteps     = 20
	MaxConversationRounds = 20
	MaxLongTermMemories   = 50
)

// Agent 智能代理
type Agent struct {
	llm         llm.Interface
	toolManager *ToolManager
	storageDir  string
	hooks       []Hook
	permissionMode string
}

// NewAgent 创建 Agent
func NewAgent(llmAdapter llm.Interface, storageDir string, permissionMode string) *Agent {
	agent := &Agent{
		llm:            llmAdapter,
		storageDir:     storageDir,
		toolManager:    NewToolManager(),
		permissionMode: permissionMode,
	}

	// 注册内置工具
	agent.toolManager.Register(tools.NewReadFile())
	agent.toolManager.Register(tools.NewWriteFile())
	agent.toolManager.Register(tools.NewBash())
	agent.toolManager.Register(tools.NewEditFile())

	// 添加权限 Hook
	agent.addPermissionHooks()

	return agent
}

// AddHook 添加 Hook
func (a *Agent) AddHook(hook Hook) {
	a.hooks = append(a.hooks, hook)
}

// Hook 接口
type Hook interface {
	Handle(event string, context map[string]interface{}) map[string]interface{}
}

func (a *Agent) addPermissionHooks() {
	mode := a.permissionMode
	if mode == "" {
		mode = "default"
	}
	// 添加权限检查 Hook
	a.hooks = append(a.hooks, hooks.NewPermissionCheckHook(mode))
	// 添加安全检查 Hook
	a.hooks = append(a.hooks, hooks.NewSafetyHook())
}
