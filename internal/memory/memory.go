package memory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mygit-william/nanobot-go/internal/llm"
)

// Config 记忆系统配置
type Config struct {
	StorageDir     string // 存储根目录 (storage/)
	MaxContextSize int    // 触发摘要的消息数阈值
	SummaryEnabled bool   // 是否启用自动摘要
}

// Manager 记忆管理器，统一管理对话持久化和长期记忆
type Manager struct {
	config Config
	mu     sync.RWMutex
}

// Conversation 对话持久化结构
type Conversation struct {
	SessionID string        `json:"session_id"`
	Messages  []llm.Message `json:"messages"`
	Summary   string        `json:"summary,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewManager 创建记忆管理器
func NewManager(cfg Config) *Manager {
	// 确保目录存在
	os.MkdirAll(filepath.Join(cfg.StorageDir, "context"), 0755)
	os.MkdirAll(filepath.Join(cfg.StorageDir, "memory"), 0755)

	if cfg.MaxContextSize <= 0 {
		cfg.MaxContextSize = 20
	}

	return &Manager{config: cfg}
}

// ============================================================
// 对话持久化
// ============================================================

// conversationPath 对话文件路径
func (m *Manager) conversationPath(sessionID string) string {
	safeID := sanitizeFilename(sessionID)
	return filepath.Join(m.config.StorageDir, "context", safeID+".json")
}

// LoadConversation 加载对话历史
func (m *Manager) LoadConversation(sessionID string) (*Conversation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.conversationPath(sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Conversation{
				SessionID: sessionID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("读取对话文件失败: %w", err)
	}

	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, fmt.Errorf("解析对话文件失败: %w", err)
	}
	return &conv, nil
}

// SaveConversation 保存对话
func (m *Manager) SaveConversation(conv *Conversation) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conv.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化对话失败: %w", err)
	}

	path := m.conversationPath(conv.SessionID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入对话文件失败: %w", err)
	}
	return nil
}

// SaveMessages 保存消息列表（过滤掉 system 消息，因为 system prompt 每次重新生成）
func (m *Manager) SaveMessages(sessionID string, messages []llm.Message) error {
	// 只保存非 system 消息
	var saveable []llm.Message
	for _, msg := range messages {
		if msg.Role != "system" {
			saveable = append(saveable, msg)
		}
	}

	// 读取已有对话以保留摘要和创建时间
	existing, err := m.LoadConversation(sessionID)
	if err != nil {
		existing = &Conversation{
			SessionID: sessionID,
			CreatedAt: time.Now(),
		}
	}

	existing.Messages = saveable
	return m.SaveConversation(existing)
}

// LoadMessages 加载消息列表（注入 system prompt + 摘要 + 历史消息）
func (m *Manager) LoadMessages(sessionID string, systemPrompt string) ([]llm.Message, error) {
	conv, err := m.LoadConversation(sessionID)
	if err != nil {
		return nil, err
	}

	// 1. 生成 fresh system prompt（含长期记忆）
	longTermMem, _ := m.LoadLongTermMemory()
	fullSystemPrompt := systemPrompt
	if longTermMem != "" {
		fullSystemPrompt += "\n\n### 长期记忆\n" + longTermMem
	}

	var messages []llm.Message
	messages = append(messages, llm.Message{Role: "system", Content: fullSystemPrompt})

	// 2. 附加对话摘要（如果有）
	if conv.Summary != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "以下是之前对话的摘要:\n" + conv.Summary,
		})
	}

	// 3. 附加近期对话消息
	messages = append(messages, conv.Messages...)
	return messages, nil
}

// ============================================================
// 上下文窗口管理 / 自动摘要
// ============================================================

// ShouldSummarize 检查是否需要摘要
func (m *Manager) ShouldSummarize(sessionID string) bool {
	if !m.config.SummaryEnabled {
		return false
	}

	conv, err := m.LoadConversation(sessionID)
	if err != nil {
		return false
	}
	return len(conv.Messages) > m.config.MaxContextSize
}

// SummarizeConversation 使用 LLM 摘要旧消息
// 如果 LLM 摘要失败，退化为简单截断
func (m *Manager) SummarizeConversation(sessionID string, llmAdapter llm.Interface) error {
	conv, err := m.LoadConversation(sessionID)
	if err != nil {
		return err
	}

	if len(conv.Messages) <= m.config.MaxContextSize/2 {
		return nil // 消息太少，不摘要
	}

	// 保留最近一半消息，摘要前一半
	halfPoint := len(conv.Messages) / 2
	oldMessages := conv.Messages[:halfPoint]
	recentMessages := conv.Messages[halfPoint:]

	// 构建摘要请求
	var conversationText strings.Builder
	for _, msg := range oldMessages {
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		roleLabel := map[string]string{
			"user":     "用户",
			"assistant": "AI",
			"tool":     "工具",
		}[msg.Role]
		if roleLabel == "" {
			roleLabel = msg.Role
		}
		conversationText.WriteString(fmt.Sprintf("[%s]: %s\n", roleLabel, content))
	}

	summaryPrompt := fmt.Sprintf(
		"请用简洁的中文总结以下对话的关键信息（包括用户需求、重要决策、关键事实），不要遗漏重要细节：\n\n%s",
		conversationText.String(),
	)

	summaryMessages := []llm.Message{
		{Role: "system", Content: "你是一个对话摘要助手，负责提取对话中的关键信息。"},
		{Role: "user", Content: summaryPrompt},
	}

	resp, err := llmAdapter.Chat(&summaryMessages, nil)
	if err != nil {
		// LLM 摘要失败 → 退化为简单截断
		log.Printf("[Memory] 摘要生成失败 (%v)，退化为截断", err)
		conv.Messages = recentMessages
		conv.Summary += "\n[摘要生成失败，旧消息已截断]"
		return m.SaveConversation(conv)
	}

	// 拼接摘要
	if conv.Summary != "" {
		conv.Summary += "\n\n"
	}
	conv.Summary += fmt.Sprintf("[%s 摘要]\n%s", time.Now().Format("2006-01-02 15:04"), resp.Reply)

	// 只保留近期消息
	conv.Messages = recentMessages

	log.Printf("[Memory] 摘要完成，保留 %d 条近期消息", len(recentMessages))
	return m.SaveConversation(conv)
}

// TrimConversation 简单截断（保留最近 N 条消息，不用 LLM）
func (m *Manager) TrimConversation(sessionID string, keepCount int) error {
	conv, err := m.LoadConversation(sessionID)
	if err != nil {
		return err
	}

	if len(conv.Messages) <= keepCount {
		return nil
	}

	trimmed := conv.Messages[len(conv.Messages)-keepCount:]
	conv.Messages = trimmed

	return m.SaveConversation(conv)
}

// ============================================================
// 长期记忆 (MEMORY.md)
// ============================================================

// memoryPath 长期记忆文件路径
func (m *Manager) memoryPath() string {
	return filepath.Join(m.config.StorageDir, "memory", "MEMORY.md")
}

// LoadLongTermMemory 加载长期记忆
func (m *Manager) LoadLongTermMemory() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.memoryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("读取长期记忆失败: %w", err)
	}
	return string(data), nil
}

// AppendLongTermMemory 追加长期记忆条目
func (m *Manager) AppendLongTermMemory(content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.memoryPath()

	existing := ""
	data, err := os.ReadFile(path)
	if err == nil {
		existing = string(data)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newContent := existing
	if !strings.HasSuffix(newContent, "\n") && newContent != "" {
		newContent += "\n"
	}
	newContent += fmt.Sprintf("- **%s**: %s\n", timestamp, content)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入长期记忆失败: %w", err)
	}
	return nil
}

// DeleteConversation 删除对话
func (m *Manager) DeleteConversation(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.conversationPath(sessionID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除对话文件失败: %w", err)
	}
	return nil
}

// SearchLongTermMemory 搜索长期记忆
func (m *Manager) SearchLongTermMemory(query string) ([]string, error) {
	content, err := m.LoadLongTermMemory()
	if err != nil {
		return nil, err
	}
	if content == "" {
		return nil, nil
	}

	var results []string
	lines := strings.Split(content, "\n")
	queryLower := strings.ToLower(query)

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), queryLower) && strings.TrimSpace(line) != "" {
			results = append(results, line)
		}
	}
	return results, nil
}

// ============================================================
// 工具函数
// ============================================================

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}
