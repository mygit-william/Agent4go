package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mygit-william/nanobot-go/internal/core"
	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/memory"
)

// Server Web 服务器
type Server struct {
	agent        *core.Agent
	memManager   *memory.Manager
	projectRoot  string
	port         string
	sessions     map[string]*Session
	sessionsMu   sync.RWMutex
	systemPrompt string
}

// Session Web 会话
type Session struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Messages  []llm.Message `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// NewServer 创建 Web 服务器
func NewServer(agent *core.Agent, memManager *memory.Manager, projectRoot, port string) *Server {
	if port == "" {
		port = "8080"
	}

	s := &Server{
		agent:       agent,
		memManager:  memManager,
		projectRoot: projectRoot,
		port:        port,
		sessions:    make(map[string]*Session),
	}

	s.loadSystemPrompt()
	s.loadExistingSessions()

	return s
}

// loadSystemPrompt 加载系统提示词
func (s *Server) loadSystemPrompt() {
	agentsFile := filepath.Join(s.projectRoot, "storage", "AGENTS.md")
	data, err := os.ReadFile(agentsFile)
	if err == nil {
		s.systemPrompt = string(data)
	}

	machineInfo := fmt.Sprintf(`
### 工作目录
工作目录在 %s 下, MEMORY.md 在 %s/memory 目录下.
当前系统: %s
`, s.projectRoot, s.projectRoot, runtime.GOOS)

	s.systemPrompt += machineInfo
}

// loadExistingSessions 加载已有会话
func (s *Server) loadExistingSessions() {
	ctxDir := filepath.Join(s.projectRoot, "storage", "context")
	entries, err := os.ReadDir(ctxDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".json")
		conv, err := s.memManager.LoadConversation(sessionID)
		if err != nil {
			continue
		}

		title := s.generateSessionTitle(conv.Messages)
		s.sessions[sessionID] = &Session{
			ID:        sessionID,
			Title:     title,
			Messages:  conv.Messages,
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
		}
	}
}

// generateSessionTitle 从消息生成会话标题
func (s *Server) generateSessionTitle(messages []llm.Message) string {
	for _, msg := range messages {
		if msg.Role == "user" && msg.Content != "" {
			title := msg.Content
			if len(title) > 30 {
				title = title[:30] + "..."
			}
			return title
		}
	}
	return "新会话"
}

// getOrCreateSession 获取或创建会话
func (s *Server) getOrCreateSession(sessionID string) (*Session, error) {
	s.sessionsMu.RLock()
	session, exists := s.sessions[sessionID]
	s.sessionsMu.RUnlock()

	if exists {
		return session, nil
	}

	// 从持久化加载
	conv, err := s.memManager.LoadConversation(sessionID)
	if err != nil {
		return nil, err
	}

	// 注入 system prompt
	messages, err := s.memManager.LoadMessages(sessionID, s.systemPrompt)
	if err != nil {
		// 回退：自己组装
		messages = []llm.Message{
			{Role: "system", Content: s.systemPrompt},
		}
		messages = append(messages, conv.Messages...)
	}

	session = &Session{
		ID:        sessionID,
		Title:     "新会话",
		Messages:  messages,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}

	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	return session, nil
}

// saveSession 保存会话
func (s *Server) saveSession(session *Session) error {
	// 过滤 system 消息
	var saveable []llm.Message
	for _, msg := range session.Messages {
		if msg.Role != "system" {
			saveable = append(saveable, msg)
		}
	}
	
	// DEBUG: 记录保存的消息
	fmt.Printf("[DEBUG] saveSession: sessionID=%s, total messages=%d, saveable=%d\n", 
		session.ID, len(session.Messages), len(saveable))
	for i, msg := range saveable {
		fmt.Printf("[DEBUG]   msg[%d]: role=%s, content=%q, has_tool_calls=%v\n", 
			i, msg.Role, msg.Content, len(msg.ToolCalls) > 0)
	}

	session.UpdatedAt = time.Now()

	conv := &memory.Conversation{
		SessionID: session.ID,
		Messages:  saveable,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
	}

	return s.memManager.SaveConversation(conv)
}

// Run 启动服务器
func (s *Server) Run() error {
	mux := http.NewServeMux()

	// 静态文件
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/app.js", s.handleJS)
	mux.HandleFunc("/style.css", s.handleCSS)

	// API
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/sessions/", s.handleSessionDetail)
	mux.HandleFunc("/api/chat", s.handleChat)
	mux.HandleFunc("/api/chat/stream", s.handleChatStream)
	mux.HandleFunc("/api/chat/stream/post", s.handleChatStreamPost)

	addr := ":" + s.port
	log.Printf("🚀 Web 服务器启动: http://localhost%s", addr)
	return http.ListenAndServe(addr, s.corsMiddleware(mux))
}

// corsMiddleware CORS 中间件
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeJSON 写入 JSON 响应 (强制 UTF-8)
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
