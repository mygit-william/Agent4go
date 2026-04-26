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
let currentTheme = 'dark';
let currentAssistantMsgEl = null;

// ========== DOM 元素 ==========
const els = {
    sidebar: document.getElementById('sidebar'),
    sessionList: document.getElementById('session-list'),
    btnNewChat: document.getElementById('btn-new-chat'),
    btnToggleSidebar: document.getElementById('btn-toggle-sidebar'),
    btnThemeToggle: document.getElementById('theme-toggle'),
    chatTitle: document.getElementById('chat-title'),
    messages: document.getElementById('messages'),
    messageInput: document.getElementById('message-input'),
    btnSend: document.getElementById('btn-send'),
};

// ========== 初始化 ==========
document.addEventListener('DOMContentLoaded', () => {
    loadTheme();
    loadSessions();
    setupEventListeners();
    autoResizeTextarea();
});

function setupEventListeners() {
    els.btnNewChat.addEventListener('click', createNewSession);
    els.btnToggleSidebar.addEventListener('click', toggleSidebar);
    els.btnSend.addEventListener('click', sendMessage);
    els.btnThemeToggle.addEventListener('click', toggleTheme);

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
        els.messageInput.style.height = Math.min(els.messageInput.scrollHeight, 200) + 'px';
    });
}

function toggleSidebar() {
    els.sidebar.classList.toggle('collapsed');
}

// ========== THEME ==========

function loadTheme() {
    const saved = localStorage.getItem('nanobot-theme');
    if (saved) {
        currentTheme = saved;
        applyTheme(currentTheme);
    }
}

function toggleTheme() {
    currentTheme = currentTheme === 'dark' ? 'light' : 'dark';
    applyTheme(currentTheme);
    localStorage.setItem('nanobot-theme', currentTheme);
}

function applyTheme(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    if (els.btnThemeToggle) {
        els.btnThemeToggle.textContent = theme === 'dark' ? '🌙' : '☀️';
    }
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

    // Group sessions by date
    const groups = groupSessionsByDate(sessions);

    groups.forEach(group => {
        const groupLabel = document.createElement('div');
        groupLabel.className = 'session-group-label';
        groupLabel.textContent = group.label;
        els.sessionList.appendChild(groupLabel);

        group.sessions.forEach(session => {
            const item = document.createElement('div');
            item.className = `session-item ${session.id === currentSessionId ? 'active' : ''}`;
            item.innerHTML = `
                <span class="session-icon">💬</span>
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
    });
}

function groupSessionsByDate(sessions) {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const groups = [];
    const todaySessions = [];
    const yesterdaySessions = [];
    const olderSessions = [];

    sessions.forEach(session => {
        const sessionDate = new Date(session.updated_at);
        const sessionDay = new Date(sessionDate.getFullYear(), sessionDate.getMonth(), sessionDate.getDate());

        if (sessionDay.getTime() === today.getTime()) {
            todaySessions.push(session);
        } else if (sessionDay.getTime() === yesterday.getTime()) {
            yesterdaySessions.push(session);
        } else {
            olderSessions.push(session);
        }
    });

    if (todaySessions.length > 0) {
        groups.push({ label: '今天', sessions: todaySessions });
    }
    if (yesterdaySessions.length > 0) {
        groups.push({ label: '昨天', sessions: yesterdaySessions });
    }
    if (olderSessions.length > 0) {
        groups.push({ label: '更早', sessions: olderSessions });
    }

    return groups;
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
        <div class="messages-inner">
            <div class="welcome">
                <div class="welcome-logo">🤖</div>
                <h2>你好！我是 Nanobot</h2>
                <p>输入任务开始执行，我可以帮你读写文件、执行命令、编辑代码等。</p>
                <div class="welcome-suggestions">
                    <button class="suggestion-btn">📝 帮我写一个 Go 的 HTTP 服务器</button>
                    <button class="suggestion-btn">🔧 分析项目目录结构并优化</button>
                    <button class="suggestion-btn">📊 读取并总结最近的日志文件</button>
                </div>
            </div>
        </div>
    `;

    // Bind suggestion buttons
    els.messages.querySelectorAll('.suggestion-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            els.messageInput.value = btn.textContent.replace(/^[^\s]+\s/, '');
            sendMessage();
        });
    });
}

function renderMessages(messages) {
    els.messages.innerHTML = '<div class="messages-inner"></div>';
    const inner = els.messages.querySelector('.messages-inner');

    // 先收集所有 tool 消息的结果，按 tool_call_id 索引
    const toolResults = {};
    messages.forEach(msg => {
        if (msg.role === 'tool' && msg.tool_call_id) {
            toolResults[msg.tool_call_id] = {
                content: msg.content,
                name: msg.name
            };
        }
    });

    messages.forEach(msg => {
        if (msg.role === 'system' || msg.role === 'tool') return;
        
        // 如果有 tool_calls，尝试匹配 tool 结果
        let toolCalls = msg.tool_calls || [];
        if (toolCalls.length > 0) {
            toolCalls = toolCalls.map(tool => {
                const result = toolResults[tool.id];
                if (result) {
                    return {
                        ...tool,
                        result: result.content,
                        success: !result.content.startsWith('错误')
                    };
                }
                return tool;
            });
        }
        
        const msgEl = createMessageElement(msg.role, msg.content || '', toolCalls);
        inner.appendChild(msgEl);
    });

    scrollToBottom();
}

function createMessageElement(role, content, toolCalls) {
    const msgDiv = document.createElement('div');
    msgDiv.className = `message ${role}`;

    const avatar = role === 'user' ? '👤' : role === 'assistant' ? '🤖' : '⚙️';

    // 处理 content 为 null 的情况
    const safeContent = content || '';
    let contentHtml = formatContent(safeContent);

    // 工具调用展示
    if (toolCalls && toolCalls.length > 0) {
        contentHtml += renderToolChain(toolCalls);
    }

    msgDiv.innerHTML = `
        <div class="message-avatar">${avatar}</div>
        <div class="message-content-wrapper">
            <div class="message-content">${contentHtml}</div>
            <div class="message-meta">
                <span>${formatTime(new Date().toISOString())}</span>
                <div class="message-actions">
                    <button class="msg-action-btn" title="复制">📋</button>
                    ${role === 'assistant' ? '<button class="msg-action-btn" title="重新生成">🔄</button>' : ''}
                </div>
            </div>
        </div>
    `;

    // Bind copy action
    msgDiv.querySelectorAll('.msg-action-btn').forEach(btn => {
        if (btn.title === '复制') {
            btn.addEventListener('click', () => {
                navigator.clipboard.writeText(safeContent).then(() => {
                    btn.textContent = '✓';
                    setTimeout(() => btn.textContent = '📋', 2000);
                });
            });
        }
    });

    // Bind tool call toggles - click header to show/hide result and args
    msgDiv.querySelectorAll('.tool-call').forEach(toolEl => {
        const header = toolEl.querySelector('.tool-call-header');
        if (!header) return;
        header.addEventListener('click', () => {
            header.classList.toggle('expanded');
            const result = toolEl.querySelector('.tool-result');
            if (result) result.classList.toggle('visible');
            const args = toolEl.querySelector('.tool-args');
            if (args) args.classList.toggle('visible');
        });
    });

    // Bind code copy buttons
    msgDiv.querySelectorAll('.code-copy-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const code = btn.closest('.code-block').querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                btn.textContent = '✓ 已复制';
                setTimeout(() => btn.textContent = '📋 复制', 2000);
            });
        });
    });

    return msgDiv;
}

function appendMessage(role, content, toolCalls) {
    // 移除欢迎语
    const welcome = els.messages.querySelector('.welcome');
    if (welcome) {
        welcome.closest('.messages-inner').innerHTML = '';
    }

    let inner = els.messages.querySelector('.messages-inner');
    if (!inner) {
        els.messages.innerHTML = '<div class="messages-inner"></div>';
        inner = els.messages.querySelector('.messages-inner');
    }

    const msgEl = createMessageElement(role, content, toolCalls);
    inner.appendChild(msgEl);
    scrollToBottom();

    return msgEl;
}

// ========== TOOL CALL RENDERING ==========

function renderToolChain(toolCalls) {
    if (!toolCalls || toolCalls.length === 0) return '';

    let html = '';
    toolCalls.forEach((tool, idx) => {
        html += renderToolCall(tool, idx);
    });
    return html;
}

function renderToolCall(tool, idx) {
    const args = typeof tool.function.arguments === 'string'
        ? tool.function.arguments
        : JSON.stringify(tool.function.arguments, null, 2);

    // 解析参数，提取文件路径或命令
    let displayLine = '';
    try {
        const parsedArgs = JSON.parse(args);
        displayLine = formatToolDisplay(tool.function.name, parsedArgs);
    } catch (e) {
        displayLine = tool.function.name;
    }

    const hasResult = tool.result !== undefined && tool.result !== null;
    const isExecuting = !hasResult && tool.result === undefined;
    const resultHtml = hasResult ? renderToolResult(tool.result, tool.success) : '';

    return `
        <div class="tool-call ${isExecuting ? 'executing' : ''}" data-tool-index="${idx}" data-tool-name="${escapeHtml(tool.function.name)}">
            <div class="tool-call-header ${hasResult ? '' : 'expanded'}">
                <span class="toggle-icon">▶</span>
                <span class="tool-display-line">${escapeHtml(displayLine)}</span>
                ${isExecuting ? '<span class="tool-status">⏳ 执行中...</span>' : ''}
            </div>
            <div class="tool-args">${escapeHtml(args)}</div>
            ${resultHtml}
        </div>
    `;
}

// 格式化工具显示行
function formatToolDisplay(toolName, args) {
    switch (toolName) {
        case 'read_file':
            return `read_file ${args.path || ''}`;
        case 'write_file':
            return `write_file ${args.path || ''}`;
        case 'edit_file':
            return `edit_file ${args.path || ''}`;
        case 'bash':
            return `bash ${args.command || ''}`;
        default:
            return `${toolName} ${JSON.stringify(args)}`;
    }
}

// 格式化工具显示行
function formatToolDisplay(toolName, args) {
    switch (toolName) {
        case 'read_file':
            return `read_file ${args.path || ''}`;
        case 'write_file':
            return `write_file ${args.path || ''}`;
        case 'edit_file':
            return `edit_file ${args.path || ''}`;
        case 'bash':
            return `bash ${args.command || ''}`;
        default:
            return `${toolName} ${JSON.stringify(args)}`;
    }
}

function renderToolResult(result, success) {
    const resultText = typeof result === 'string' ? result : JSON.stringify(result, null, 2);
    return `
        <div class="tool-result ${success ? 'success' : 'error'} visible">
            <div class="tool-result-header ${success ? 'success' : 'error'}">
                ${success ? '✓' : '✗'} 执行结果
            </div>
            <div>${escapeHtml(resultText.substring(0, 500))}${resultText.length > 500 ? '...' : ''}</div>
        </div>
    `;
}

function renderToolResultLegacy(toolName, result, success) {
    // 找到对应的工具块（通过 tool-name 匹配）
    const toolCalls = els.messages.querySelectorAll('.tool-call');
    let targetTool = null;
    
    // 优先找没有结果且名字匹配的工具块
    for (let i = toolCalls.length - 1; i >= 0; i--) {
        const nameAttr = toolCalls[i].getAttribute('data-tool-name');
        if (nameAttr === toolName && !toolCalls[i].querySelector('.tool-result')) {
            targetTool = toolCalls[i];
            break;
        }
    }
    
    // 如果没找到匹配的，找最后一个没有结果的工具块
    if (!targetTool) {
        for (let i = toolCalls.length - 1; i >= 0; i--) {
            if (!toolCalls[i].querySelector('.tool-result')) {
                targetTool = toolCalls[i];
                break;
            }
        }
    }
    
    if (!targetTool) return;

    // 移除执行中状态
    targetTool.classList.remove('executing');
    const statusEl = targetTool.querySelector('.tool-status');
    if (statusEl) statusEl.remove();

    const resultText = typeof result === 'string' ? result : JSON.stringify(result, null, 2);
    const resultDiv = document.createElement('div');
    resultDiv.className = `tool-result ${success ? 'success' : 'error'}`;
    resultDiv.innerHTML = `
        <div class="tool-result-header ${success ? 'success' : 'error'}">
            ${success ? '✓' : '✗'} 执行结果
        </div>
        <div>${escapeHtml(resultText).substring(0, 500)}${resultText.length > 500 ? '...' : ''}</div>
    `;

    targetTool.appendChild(resultDiv);

    // 绑定 toggle：点击 header 展开/折叠 result
    const header = targetTool.querySelector('.tool-call-header');
    if (header) {
        // 移除之前的 expanded 类（流式时默认展开的）
        header.classList.remove('expanded');
        // 移除旧的点击事件（通过克隆替换）
        const newHeader = header.cloneNode(true);
        header.parentNode.replaceChild(newHeader, header);
        newHeader.addEventListener('click', () => {
            newHeader.classList.toggle('expanded');
            resultDiv.classList.toggle('visible');
            const args = targetTool.querySelector('.tool-args');
            if (args) args.classList.toggle('visible');
        });
    }

    scrollToBottom();
}

// ========== THINKING / LOADING ==========

function showThinking(step, maxSteps, content) {
    const existing = document.getElementById('thinking-indicator');
    if (existing) existing.remove();

    const inner = els.messages.querySelector('.messages-inner');
    if (!inner) return;

    const div = document.createElement('div');
    div.id = 'thinking-indicator';
    div.className = 'message assistant';
    div.innerHTML = `
        <div class="message-avatar">🤖</div>
        <div class="message-content-wrapper">
            <div class="thinking">
                <div class="thinking-dots">
                    <span></span><span></span><span></span>
                </div>
                <span>${content || '思考中...'}</span>
            </div>
        </div>
    `;

    inner.appendChild(div);
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
            buffer = lines.pop();

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();
                if (!line) continue;

                if (line.startsWith('event:')) {
                    const eventType = line.substring(6).trim();
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
    contentDiv.innerHTML = formatContent(content);

    // Re-bind code copy buttons
    currentMsg.querySelectorAll('.code-copy-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const code = btn.closest('.code-block').querySelector('code').textContent;
            navigator.clipboard.writeText(code).then(() => {
                btn.textContent = '✓ 已复制';
                setTimeout(() => btn.textContent = '📋 复制', 2000);
            });
        });
    });

    scrollToBottom();
    return currentMsg;
}



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
            const toolEl = toolDiv.firstElementChild;
            currentAssistantMsgEl.querySelector('.message-content').appendChild(toolEl);
            
            // 绑定点击展开/折叠
            const header = toolEl.querySelector('.tool-call-header');
            if (header) {
                header.addEventListener('click', () => {
                    header.classList.toggle('expanded');
                    const result = toolEl.querySelector('.tool-result');
                    if (result) result.classList.toggle('visible');
                    const args = toolEl.querySelector('.tool-args');
                    if (args) args.classList.toggle('visible');
                });
            }
            
            scrollToBottom();
            break;
        case 'tool_result':
            renderToolResultLegacy(data.tool_name, data.result, data.success);
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

// ========== 工具函数 ==========

function formatContent(text) {
    if (!text) return '';

    let html = escapeHtml(text);

    // Code blocks with language
    html = html.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
        const language = lang || 'text';
        return `
            <div class="code-block">
                <div class="code-header">
                    <span class="code-lang">${language}</span>
                    <button class="code-copy-btn">📋 复制</button>
                </div>
                <pre><code>${code.trim()}</code></pre>
            </div>
        `;
    });

    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

    // Bold
    html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');

    // Italic
    html = html.replace(/\*(.*?)\*/g, '<em>$1</em>');

    // Headers
    html = html.replace(/^### (.*$)/gim, '<h3>$1</h3>');
    html = html.replace(/^## (.*$)/gim, '<h2>$1</h2>');
    html = html.replace(/^# (.*$)/gim, '<h1>$1</h1>');

    // Lists
    html = html.replace(/^\s*[-*+]\s+(.*$)/gim, '<li>$1</li>');
    html = html.replace(/(<li>.*<\/li>\n?)+/g, '<ul>$&</ul>');

    // Paragraphs
    html = html.split('\n\n').map(p => {
        p = p.trim();
        if (!p || p.startsWith('<')) return p;
        return `<p>${p.replace(/\n/g, '<br>')}</p>`;
    }).join('');

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
    let inner = els.messages.querySelector('.messages-inner');
    if (!inner) {
        els.messages.innerHTML = '<div class="messages-inner"></div>';
        inner = els.messages.querySelector('.messages-inner');
    }

    const div = document.createElement('div');
    div.className = 'message system';
    div.innerHTML = `
        <div class="message-avatar">⚠️</div>
        <div class="message-content-wrapper">
            <div class="message-content" style="color:var(--error)">${escapeHtml(message)}</div>
        </div>
    `;
    inner.appendChild(div);
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
