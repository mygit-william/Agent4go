package channels

// DingTalkChannel 钉钉通道 (占位实现)
type DingTalkChannel struct{}

// NewDingTalkChannel 创建钉钉通道
func NewDingTalkChannel() *DingTalkChannel {
	return &DingTalkChannel{}
}

// GetName 获取通道名称
func (c *DingTalkChannel) GetName() string {
	return "dingtalk"
}

// Receive 接收消息
func (c *DingTalkChannel) Receive() {
	// TODO: 实现钉钉 WebSocket 长连接
}

// Send 发送消息
func (c *DingTalkChannel) Send(sessionID, message string) {
	// TODO: 实现钉钉消息发送
}
