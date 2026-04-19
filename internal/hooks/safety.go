package hooks

import "log"

// SafetyHook 安全检查 Hook
type SafetyHook struct {
	BaseHook
	dangerousPatterns []string
}

// NewSafetyHook 创建安全 Hook
func NewSafetyHook() *SafetyHook {
	return &SafetyHook{
		dangerousPatterns: []string{
			"rm -rf", "dd if=", "mkfs", "fdisk",
			"sudo rm", "chmod 777", "> /dev/",
			"curl | bash", "wget | bash",
		},
	}
}

// Handle 处理事件
func (h *SafetyHook) Handle(event string, context map[string]interface{}) map[string]interface{} {
	if event != EventPreAction {
		return context
	}

	tool, ok := context["tool"].(map[string]interface{})
	if !ok {
		return context
	}

	args, _ := tool["arguments"].(string)

	// 检查危险模式
	for _, pattern := range h.dangerousPatterns {
		if contains(args, pattern) {
			log.Printf("[SafetyHook] 检测到危险操作: %s", pattern)
			return map[string]interface{}{
				"decision": "deny",
				"reason":   "检测到危险操作模式: " + pattern,
				"tool":     tool,
			}
		}
	}

	return context
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
