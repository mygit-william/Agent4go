package channels

import (
	"github.com/mygit-william/nanobot-go/internal/core"
)

// Interface 通道接口
type Interface interface {
	GetName() string
	Receive()
	Send(sessionID, message string)
}

// CLIChannel 命令行通道
type CLIChannel struct {
	agent       *core.Agent
	projectRoot string
}

// NewCLIChannel 创建 CLI 通道
func NewCLIChannel(agent *core.Agent, projectRoot string) *CLIChannel {
	return &CLIChannel{
		agent:       agent,
		projectRoot: projectRoot,
	}
}

// GetName 获取通道名称
func (c *CLIChannel) GetName() string {
	return "cli"
}

// Send 发送消息
func (c *CLIChannel) Send(sessionID, message string) {
	// 在 CLI 模式下直接打印
	println("📢 系统:", message)
}
