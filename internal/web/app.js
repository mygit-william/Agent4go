/**
 * Nanobot Web 前端
 * SSE 流式聊天 + 会话管理
 */

const API_BASE = '';

// ========== 状态 ==========
let currentSessionId = null;
let sessions = [];
let isStreaming = false;
let eventSource = null;

// ========== DOM 元素 ==========
const els = {
    sidebar: document.getElementById('sidebar'),
    sessionList: document.getElementById('session-list'),
    btnNewChat: document.getElementById('btn-new-chat'),
    btnToggleSidebar: document.getElementById('btn-toggle-sidebar'),
    chatTitle: document.getElementById('chat-title'),
    messages: document.getElementById('messages'),
    messageInput: document.getElementById('message-input'),
    btnSend: document.getElementById('btn-send'),
};

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    loadSessions();
    setupEventListeners();
    autoResizeTextarea();
});

function setupEventListeners() {
    els.btnNewChat.addEventListener('click', createNewSession);
    els.btnToggleSidebar.addEventListener('click', toggleSidebar);
    els.btnSend.addEventListener('click', sendMessage);

    els.messageInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
}

function autoResizeTextarea() {
    els.messageInput.addEventListener('input', () => {
        els.messageInput.style.height = 'auto';
        els.messageInput.style.height = els.messageInput.scrollHeight + 'px';
    });
}

function toggleSidebar() {
    els.sidebar.classList.toggle('collapsed');
}

// ========== 会话管理 ==========

async function loadSessions() {
    try {
        const res = await fetch(`${API_BASE}/api/sessions`);
        const data = await res.json();
        sessions = data.sessions || [];
        renderSessionList();
    } catch (err) {
        console.error('加载会话失败:', err);
    }
}

function renderSessionList() {
    els.sessionList.innerHTML = '';

    sessions.forEach(session => {
        const item = document.createElement('div');
        item.className = `session-item ${session.id === currentSessionId ? 'active' : ''}`;
        item.innerHTML = `
            <div class="session-info">
                <div class="session-title">${escapeHtml(session.title)}</div>
                <div class="session-meta">${formatTime(session.updated_at)} · ${session.message_count} 条消息</div>
            </div>
            <div class="session-actions">
                <button class="btn-delete" title="删除">×</button>
            </div>
        `;

        item.addEventListener('click', (e) => {
            if (e.target.classList.contains('btn-delete')) {
                e.stopPropagation();
                deleteSession(session.id);
            } else {
                switchSession(session.id);
            }
        });

        els.sessionList.appendChild(item);
    });
}

async function createNewSession() {
    try {
        const res = await fetch(`${API_BASE}/api/sessions`, { method: 'POST' });
        const data = await res.json();
        const session = data.session;

        sessions.unshift({
            id: session.id,
            title: session.title,
            updated_at: session.updated_at,
            message_count: 0,
        });

        currentSessionId = session.id;
        renderSessionList();
        clearMessages();
        els.chatTitle.textContent = '新会话';
    } catch (err) {
        console.error('创建会话失败:', err);
        showError('创建会话失败');
    }
}

async function switchSession(sessionId) {
    if (isStreaming) return;

    currentSessionId = sessionId;
    renderSessionList();

    try {
        const res = await fetch(`${API_BASE}/api/sessions/${sessionId}`);
        const data = await res.json();
        const session = data.session;

        els.chatTitle.textContent = session.title;
        renderMessages(session.messages);
    } catch (err) {
        console.error('切换会话失败:', err);
    }
}

async function deleteSession(sessionId) {
    if (!confirm('确定要删除这个会话吗？')) return;

    try {
        await fetch(`${API_BASE}/api/sessions/${sessionId}`, { method: 'DELETE' });
        sessions = sessions.filter(s => s.id !== sessionId);

        if (currentSessionId === sessionId) {
            currentSessionId = null;
            clearMessages();
            els.chatTitle.textContent = '新会话';
        }

        renderSessionList();
    } catch (err) {
        console.error('删除会话失败:', err);
    }
}

// ========== 消息渲染 ==========

function clearMessages() {
    els.messages.innerHTML = `
        <div class="welcome">
            <h2>👋 你好！我是 Nanobot</h2>
            <p>输入任务开始执行，我可以帮你读写文件、执行命令、编辑代码等。</p>
        </div>
    `;
}

function renderMessages(messages) {
    els.messages.innerHTML = '';

    messages.forEach(msg => {
        if (msg.role === 'system') return;
        appendMessage(msg.role, msg.content, msg.tool_calls);
    });

    scrollToBottom();
}

function appendMessage(role, content, toolCalls) {
    // 移除欢迎语
    const welcome = els.messages.querySelector('.welcome');
    if (welcome) welcome.remove();

    const msgDiv = document.createElement('div');
    msgDiv.className = `message ${role}`;

    const avatar = role === 'user' ? '👤' : role === 'assistant' ? '🤖' : '⚙️';

    let contentHtml = formatContent(content);

    // 工具调用展示
    if (toolCalls && toolCalls.length > 0) {
        toolCalls.forEach((tool, idx) => {
            contentHtml += renderToolCall(tool, idx);
        });
    }

    msgDiv.innerHTML = `
        <div class="message-avatar">${avatar}</div>
        <div class="message-content">${contentHtml}</div>
    `;

    els.messages.appendChild(msgDiv);
    scrollToBottom();

    // 绑定工具调用折叠事件
    msgDiv.querySelectorAll('.tool-call-header').forEach(header => {
        header.addEventListener('click', () => {
            header.classList.toggle('expanded');
            const args = header.nextElementSibling;
            if (args) args.classList.toggle('visible');
        });
    });

    return msgDiv;
}

function renderToolCall(tool, idx) {
    const args = typeof tool.function.arguments === 'string'
        ? tool.function.arguments
        : JSON.stringify(tool.function.arguments, null, 2);

    return `
        <div class="tool-call">
            <div class="tool-call-header">
                <span class="toggle-icon">▶</span>
                <span class="tool-name">${escapeHtml(tool.function.name)}</span>
            </div>
            <div class="tool-args">${escapeHtml(args)}</div>
        </div>
    `;
}

function renderToolResult(toolName, result, success) {
    const toolCalls = els.messages.querySelectorAll('.tool-call');
    const lastTool = toolCalls[toolCalls.length - 1];
    if (!lastTool) return;

    const resultDiv = document.createElement('div');
    resultDiv.className = `tool-result ${success ? 'success' : 'error'}`;
    resultDiv.innerHTML = `
        <div style="font-weight:600;margin-bottom:4px;">${success ? '✓' : '✗'} ${escapeHtml(toolName)}</div>
        <pre style="margin:0;white-space:pre-wrap;word-break:break-all;">${escapeHtml(result.substring(0, 500))}${result.length > 500 ? '...' : ''}</pre>
    `;

    lastTool.appendChild(resultDiv);
    scrollToBottom();
}

function showThinking(step, maxSteps, content) {
    const existing = document.getElementById('thinking-indicator');
    if (existing) existing.remove();

    const div = document.createElement('div');
    div.id = 'thinking-indicator';
    div.className = 'message system';
    div.innerHTML = `
        <div class="message-avatar">⏳</div>
        <div class="message-content">
            <div class="thinking">
                <div class="thinking-dots">
                    <span></span><span></span><span></span>
                </div>
                <span>步骤 ${step}/${maxSteps} · ${escapeHtml(content)}</span>
            </div>
        </div>
    `;

    els.messages.appendChild(div);
    scrollToBottom();
}

function hideThinking() {
    const el = document.getElementById('thinking-indicator');
    if (el) el.remove();
}

// ========== 发送消息 ==========

async function sendMessage() {
    const text = els.messageInput.value.trim();
    if (!text || isStreaming) return;

    // 如果没有当前会话，先创建一个
    if (!currentSessionId) {
        await createNewSession();
    }

    // 添加用户消息
    appendMessage('user', text);
    els.messageInput.value = '';
    els.messageInput.style.height = 'auto';

    // 开始 SSE 流式接收
    startStream(text);
}

function startStream(message) {
    isStreaming = true;
    els.btnSend.disabled = true;

    // 使用 POST 方式启动 SSE（支持长消息）
    startStreamPost(message);
}

// 使用 POST 方式启动 SSE（用于长消息）
async function startStreamPost(message) {
    isStreaming = true;
    els.btnSend.disabled = true;

    try {
        const response = await fetch(`${API_BASE}/api/chat/stream/post`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                session_id: currentSessionId,
                message: message,
            }),
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        let currentAssistantMsg = null;
        let assistantContent = '';

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop(); // 保留未完成的行

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                if (!line) continue;

                // 解析 SSE 格式
                if (line.startsWith('event:')) {
                    const eventType = line.substring(6).trim();
                    // 读取下一行的 data
                    if (i + 1 < lines.length && lines[i + 1].startsWith('data:')) {
                        const dataStr = lines[i + 1].substring(5).trim();
                        i++;
                        try {
                            const data = JSON.parse(dataStr);
                            if (eventType === 'assistant') {
                                assistantContent += (data.content || '');
                                currentAssistantMsg = updateOrCreateAssistantMsg(assistantContent, currentAssistantMsg);
                            } else {
                                handleSSEEvent(eventType, data);
                            }
                        } catch (e) {
                            console.error('解析事件数据失败:', e);
                        }
                    }
                }
            }
        }

        // 流结束后，如果没有 assistant 消息但有内容，确保显示
        if (assistantContent && !currentAssistantMsg) {
            appendMessage('assistant', assistantContent);
        }

        hideThinking();
        isStreaming = false;
        els.btnSend.disabled = false;
        loadSessions();
    } catch (err) {
        console.error('POST SSE 错误:', err);
        hideThinking();
        isStreaming = false;
        els.btnSend.disabled = false;
        showError('连接中断: ' + err.message);
    }
}

function updateOrCreateAssistantMsg(content, currentMsg) {
    hideThinking();

    if (!currentMsg) {
        return appendMessage('assistant', content);
    }

    const contentDiv = currentMsg.querySelector('.message-content');
    const textNodes = Array.from(contentDiv.childNodes).filter(n =>
        n.nodeType === Node.TEXT_NODE || (n.nodeType === Node.ELEMENT_NODE && !n.classList.contains('tool-call'))
    );

    if (textNodes.length === 0) {
        const p = document.createElement('p');
        p.textContent = content;
        contentDiv.insertBefore(p, contentDiv.firstChild);
    } else {
        const first = textNodes[0];
        if (first.tagName === 'P') {
            first.textContent = content;
        } else {
            const p = document.createElement('p');
            p.textContent = content;
            contentDiv.insertBefore(p, contentDiv.firstChild);
        }
    }

    scrollToBottom();
    return currentMsg;
}

// 全局变量，跟踪当前 assistant 消息元素
let currentAssistantMsgEl = null;

function handleSSEEvent(eventType, data) {
    switch (eventType) {
        case 'start':
            showThinking(0, 1000, '任务开始');
            currentAssistantMsgEl = null;
            break;
        case 'thinking':
            showThinking(data.step, data.max_steps, data.content);
            break;
        case 'tool_call':
            hideThinking();
            if (!currentAssistantMsgEl) {
                currentAssistantMsgEl = appendMessage('assistant', '');
            }
            const toolDiv = document.createElement('div');
            toolDiv.innerHTML = renderToolCall({
                function: { name: data.tool_name, arguments: data.tool_args }
            }, data.tool_index);
            currentAssistantMsgEl.querySelector('.message-content').appendChild(toolDiv);
            toolDiv.querySelector('.tool-call-header').addEventListener('click', function() {
                this.classList.toggle('expanded');
                this.nextElementSibling.classList.toggle('visible');
            });
            scrollToBottom();
            break;
        case 'tool_result':
            renderToolResult(data.tool_name, data.result, data.success);
            break;
        case 'complete':
            hideThinking();
            break;
        case 'error':
            hideThinking();
            showError(data.error || '发生错误');
            break;
        case 'done':
            hideThinking();
            isStreaming = false;
            els.btnSend.disabled = false;
            currentAssistantMsgEl = null;
            loadSessions();
            break;
        case 'step_limit':
            hideThinking();
            showError('执行达到最大步骤限制');
            isStreaming = false;
            els.btnSend.disabled = false;
            currentAssistantMsgEl = null;
            break;
    }
}

function handleStreamEvent(data, currentAssistantMsg) {
    // 通用事件处理（备用）
}

// ========== 工具函数 ==========

function formatContent(text) {
    if (!text) return '';

    // 简单的 Markdown 支持
    let html = escapeHtml(text);

    // 代码块
    html = html.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>');

    // 行内代码
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

    // 段落
    html = html.split('\n\n').map(p => `<p>${p.replace(/\n/g, '<br>')}</p>`).join('');

    return html;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function scrollToBottom() {
    els.messages.scrollTop = els.messages.scrollHeight;
}

function showError(message) {
    const div = document.createElement('div');
    div.className = 'message system';
    div.innerHTML = `
        <div class="message-avatar">⚠️</div>
        <div class="message-content" style="color:var(--error)">${escapeHtml(message)}</div>
    `;
    els.messages.appendChild(div);
    scrollToBottom();
}

function formatTime(isoString) {
    if (!isoString) return '';
    const date = new Date(isoString);
    const now = new Date();
    const diff = now - date;

    if (diff < 60000) return '刚刚';
    if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`;

    return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}
