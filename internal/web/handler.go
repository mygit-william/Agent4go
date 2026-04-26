package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mygit-william/nanobot-go/internal/core"
	"github.com/mygit-william/nanobot-go/internal/llm"
)

// ========== 会话管理 API ==========

// handleSessions GET/POST /api/sessions
func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listSessions(w, r)
	case "POST":
		s.createSession(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

// listSessions 列出所有会话
func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	type sessionSummary struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		UpdatedAt time.Time `json:"updated_at"`
		MessageCount int    `json:"message_count"`
	}

	var result []sessionSummary
	for _, session := range s.sessions {
		msgCount := 0
		for _, msg := range session.Messages {
			if msg.Role != "system" {
				msgCount++
			}
		}
		result = append(result, sessionSummary{
			ID:           session.ID,
			Title:        session.Title,
			UpdatedAt:    session.UpdatedAt,
			MessageCount: msgCount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": result,
	})
}

// createSession 创建新会话
func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()

	session := &Session{
		ID:        sessionID,
		Title:     "新会话",
		Messages:  []llm.Message{{Role: "system", Content: s.systemPrompt}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()

	// 持久化
	s.saveSession(session)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"session": session,
	})
}

// handleSessionDetail GET/DELETE /api/sessions/{id}
func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "缺少会话 ID")
		return
	}

	sessionID := parts[0]

	switch r.Method {
	case "GET":
		s.getSession(w, r, sessionID)
	case "DELETE":
		s.deleteSession(w, r, sessionID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

// getSession 获取会话详情
func (s *Server) getSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := s.getOrCreateSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "会话不存在")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"session": session,
	})
}

// deleteSession 删除会话
func (s *Server) deleteSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	s.sessionsMu.Lock()
	delete(s.sessions, sessionID)
	s.sessionsMu.Unlock()

	// 删除持久化文件
	s.memManager.DeleteConversation(sessionID)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ========== 聊天 API ==========

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Reply     string        `json:"reply"`
	SessionID string        `json:"session_id"`
	Messages  []llm.Message `json:"messages"`
}

// handleChat POST /api/chat (非流式)
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求解析失败")
		return
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	session, err := s.getOrCreateSession(req.SessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取会话失败")
		return
	}

	// 更新标题（如果是第一条用户消息）
	if session.Title == "新会话" {
		session.Title = s.generateSessionTitle([]llm.Message{{Role: "user", Content: req.Message}})
	}

	// 使用 NoOpStreamHandler（非流式，直接返回结果）
	handler := &core.NoOpStreamHandler{}
	reply := s.agent.ChatStream(req.SessionID, req.Message, &session.Messages, handler)

	// 保存会话
	s.saveSession(session)

	writeJSON(w, http.StatusOK, ChatResponse{
		Reply:     reply,
		SessionID: req.SessionID,
		Messages:  session.Messages,
	})
}

// SSEEvent SSE 事件
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// handleChatStream GET /api/chat/stream (SSE 流式)
func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// 从 query 参数获取数据
	sessionID := r.URL.Query().Get("session_id")
	message := r.URL.Query().Get("message")

	if message == "" {
		writeError(w, http.StatusBadRequest, "消息不能为空")
		return
	}

	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	session, err := s.getOrCreateSession(sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取会话失败")
		return
	}

	// 更新标题
	if session.Title == "新会话" {
		session.Title = s.generateSessionTitle([]llm.Message{{Role: "user", Content: message}})
	}

	// 设置 SSE 头 (强制 UTF-8)
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// 确保能立即刷新
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// 创建 SSE Handler
	sseHandler := &SSEStreamHandler{
		writer:  w,
		flusher: w.(http.Flusher),
		done:    make(chan struct{}),
	}

	// 在 goroutine 中执行 Agent
	var finalReply string
	go func() {
		finalReply = s.agent.ChatStream(sessionID, message, &session.Messages, sseHandler)
		close(sseHandler.done)
	}()

	// 等待完成并保存
	<-sseHandler.done

	// 保存会话
	s.saveSession(session)

	// 发送完成事件（包含最终回复和 session_id）
	sseHandler.sendEvent("done", map[string]string{
		"reply":      finalReply,
		"session_id": sessionID,
	})
}

// SSEStreamHandler SSE 流式处理器
type SSEStreamHandler struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

// OnEvent 发送事件到 SSE
func (h *SSEStreamHandler) OnEvent(event core.StreamEvent) {
	// 将 StreamEvent 转为 map 以便前端统一解析
	data := map[string]interface{}{
		"type":    string(event.Type),
		"content": event.Content,
		"step":    event.Step,
		"max_steps": event.MaxSteps,
		"tool_name": event.ToolName,
		"tool_args": event.ToolArgs,
		"tool_index": event.ToolIndex,
		"tool_total": event.ToolTotal,
		"result":  event.Result,
		"success": event.Success,
		"error":   event.Error,
	}
	h.sendEvent(string(event.Type), data)
}

// handleChatStreamPost POST /api/chat/stream (SSE 流式，支持长消息)
func (s *Server) handleChatStreamPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求解析失败")
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "消息不能为空")
		return
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	session, err := s.getOrCreateSession(req.SessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取会话失败")
		return
	}

	// 更新标题
	if session.Title == "新会话" {
		session.Title = s.generateSessionTitle([]llm.Message{{Role: "user", Content: req.Message}})
	}

	// 设置 SSE 头 (强制 UTF-8)
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// 确保能立即刷新
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// 创建 SSE Handler
	sseHandler := &SSEStreamHandler{
		writer:  w,
		flusher: w.(http.Flusher),
		done:    make(chan struct{}),
	}

	// 在 goroutine 中执行 Agent
	var finalReply string
	go func() {
		finalReply = s.agent.ChatStream(req.SessionID, req.Message, &session.Messages, sseHandler)
		close(sseHandler.done)
	}()

	// 等待完成并保存
	<-sseHandler.done

	// 保存会话
	s.saveSession(session)

	// 发送完成事件（包含最终回复和 session_id）
	sseHandler.sendEvent("done", map[string]string{
		"reply":      finalReply,
		"session_id": req.SessionID,
	})
}

// OnComplete 任务完成
func (h *SSEStreamHandler) OnComplete(finalReply string) {
	// 在 done channel 中处理，无需额外操作
}

// OnError 发生错误
func (h *SSEStreamHandler) OnError(err error) {
	// 错误已通过 OnEvent 发送，无需额外操作
}

// sendEvent 发送 SSE 事件
func (h *SSEStreamHandler) sendEvent(eventType string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(h.writer, "event: %s\n", eventType)
	fmt.Fprintf(h.writer, "data: %s\n\n", string(jsonData))

	if h.flusher != nil {
		h.flusher.Flush()
	}
}
