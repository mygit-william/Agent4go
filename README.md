# Nanobot

**A lightweight, extensible Go-based LLM Agent framework with built-in tools, security hooks, multi-channel notifications, and Web UI.**

[Features](#-features) · [Quick Start](#-quick-start) · [Configuration](#-configuration) · [Web UI](#-web-ui) · [Development](#-development) · [Deployment](#-deployment)

---

## ✨ Features

| | |
|---|---|
| 🤖 **Multi-LLM Support** | Zhipu AI, OpenAI (any OpenAI-compatible API), Ollama (local), LongCat |
| 🔒 **Security Hooks** | Permission check + safety validation before every tool execution |
| 📁 **Built-in Tools** | `read_file`, `write_file`, `edit_file`, `bash` (with whitelist/blacklist) |
| 🧠 **Memory** | Long-term memory storage + conversation context management |
| 📢 **Notifications** | Feishu (飞书) Webhook — get notified when a task finishes |
| 🌐 **Web UI** | Built-in responsive web interface with SSE streaming, session management, and tool call visualization |
| 🔌 **Extensible** | Add new LLM adapters, tools, or hooks in minutes |
| ⚡ **Go-native** | Single binary, no runtime needed, cross-platform compile |

---

## 📦 Quick Start

### Prerequisites

- Go 1.21 or higher
- An LLM API key (Zhipu AI, OpenAI, Ollama, LongCat…)

### 1 — Clone & Install

```bash
git clone https://github.com/yourname/nanobot-go.git
cd nanobot-go
go mod tidy
```

### 2 — Configure

```bash
cp config/config.json.example config/config.json
# Edit config/config.json and fill in your API key + webhook URL
```

### 3 — Run

**CLI Mode:**
```bash
go run ./cmd/nanobot/ -mode cli
```

**Web Mode:**
```bash
go run ./cmd/nanobot/ -mode web -port 8080
```

---

## 🌐 Web UI

Nanobot includes a built-in web interface with:

- 💬 **Real-time streaming** — SSE (Server-Sent Events) for token-by-token responses
- 📂 **Session management** — Create, switch, rename, and delete conversations
- 🔧 **Tool call visualization** — Expandable tool calls with arguments and results
- 🌙 **Dark theme** — Easy on the eyes
- 📱 **Responsive** — Works on desktop and mobile
- 📦 **Single binary** — Frontend is embedded into the Go binary via `//go:embed`

### Screenshot

```
┌─────────────────────────────────────────────┐
│  🤖 Nanobot                    [New Chat]  │
├──────────┬──────────────────────────────────┤
│ Sessions  │  User: 帮我查一下天气            │
│ ○ Chat 1 │                                  │
│ ○ Chat 2 │  🔧 bash                        │
│          │     {"command": "curl wttr.in"}  │
│          │     ✅ 200 OK                     │
│          │                                  │
│          │  Assistant: 今天北京晴，20-28°C   │
└──────────┴──────────────────────────────────┘
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Web UI |
| GET | `/api/sessions` | List all sessions |
| POST | `/api/sessions` | Create new session |
| GET | `/api/sessions/{id}` | Get session details |
| DELETE | `/api/sessions/{id}` | Delete session |
| POST | `/api/chat/stream/post` | SSE streaming chat (recommended) |
| GET | `/api/chat/stream` | SSE streaming chat (GET, short messages) |

---

## 📁 Project Structure

```
nanobot-go/
├── cmd/nanobot/main.go           # Entry point
├── internal/
│   ├── core/
│   │   ├── agent.go              # Agent core + hook system
│   │   ├── agent_chat.go         # Chat loop (CLI mode)
│   │   ├── agent_chat_stream.go  # Chat loop (streaming mode)
│   │   ├── stream_handler.go     # StreamHandler interface + events
│   │   └── tool_manager.go       # Tool registry & dispatcher
│   ├── llm/                     # LLM adapters (OpenAI-compatible)
│   │   ├── factory.go           # Adapter factory
│   │   ├── interface.go         # Interface definition
│   │   ├── openai.go / ollama.go / zhipu.go / longcat.go
│   ├── tools/                   # Built-in tools
│   │   ├── bash.go              # Shell command (whitelist/blacklist)
│   │   ├── read_file.go
│   │   ├── write_file.go
│   │   └── edit_file.go
│   ├── hooks/
│   │   ├── permission_check.go  # Permission mode enforcement
│   │   └── safety.go            # Content safety validation
│   ├── channels/
│   │   ├── cli.go               # CLI input/output
│   │   └── feishu.go            # Feishu Webhook notifier
│   ├── memory/memory.go         # Memory management
│   └── web/                    # 🆕 Web UI backend
│       ├── server.go            # HTTP server + session management
│       ├── handler.go          # API handlers + SSE
│       ├── static.go           # Embedded frontend files
│       ├── index.html          # Web UI HTML
│       ├── app.js              # Web UI JavaScript
│       └── style.css           # Web UI CSS
├── config/
│   └── config.json.example      # Configuration template
├── storage/                     # Working directory for AI operations
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

## ⚙️ Configuration

### `config/config.json` (minimal)

```json
{
  "llm": {
    "default_provider": "zhipu",
    "providers": {
      "zhipu": {
        "driver": "zhipu",
        "base_url": "https://open.bigmodel.cn/",
        "model": "glm-4.7-flash",
        "api_key": "YOUR_API_KEY"
      }
    }
  },
  "channels": {
    "feishu": {
      "enabled": true,
      "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_WEBHOOK"
    }
  }
}
```

### Full options

| Field | Type | Default | Description |
|---|---|---|---|
| `llm.default_provider` | string | `"zhipu"` | Which provider to use |
| `llm.providers.<name>.driver` | string | — | One of: `openai`, `ollama`, `zhipu`, `longcat` |
| `llm.providers.<name>.base_url` | string | — | API base URL |
| `llm.providers.<name>.model` | string | — | Model name |
| `llm.providers.<name>.api_key` | string | — | API key (not needed for Ollama) |
| `permissions.mode` | string | `"default"` | `default` = confirm writes, `auto` = auto-approve, `plan` = read-only |
| `channels.feishu.enabled` | bool | `false` | Enable Feishu webhook |
| `channels.feishu.webhook_url` | string | — | Your Feishu bot webhook URL |
| `agent.max_context_size` | int | `20` | Max conversation rounds |
| `agent.max_tokens` | int | `8000` | Max response tokens |

### Config file search order

1. `config/config.json` relative to current working directory
2. `config/config.json` relative to the executable
3. Path specified via `-config` flag

### CLI flags

```
-mode cli|web       # cli = interactive terminal, web = web UI (default: cli)
-port <number>      # Port for web mode (default: 8080)
-config <path>      # Path to config file (default: config/config.json)
```

---

## 🛠️ Development

### Add a new LLM adapter

1. Create `internal/llm/myai.go`
2. Implement the `Interface` interface
3. Register it in `internal/llm/factory.go`

```go
type MyAIAdapter struct{ url, key, model string }

func (a *MyAIAdapter) Chat(msgs []llm.Message, tools []map[string]interface{}) (*llm.Response, error) {
    // your implementation
}
```

### Add a new tool

1. Create `internal/tools/mytool.go`
2. Implement the `Tool` interface
3. Register it in `core.NewAgent()`

```go
func (t *MyTool) Name() string       { return "my_tool" }
func (t *MyTool) Description() string { return "..." }
func (t *MyTool) Execute(args map[string]interface{}) string { /* ... */ }
```

### Add a new hook

1. Create `internal/hooks/myhook.go`
2. Implement `core.Hook` interface
3. Add it in `core.NewAgent()`

```go
func (h *MyHook) Handle(event string, ctx map[string]interface{}) map[string]interface{} {
    if event == hooks.EventPreAction {
        // validate/modify context
    }
    return ctx
}
```

### Web UI Development

The frontend files are embedded into the binary. To modify the UI:

1. Edit files in `internal/web/` (`index.html`, `app.js`, `style.css`)
2. Run `go generate ./internal/web/` (if you add a `//go:generate` directive)
3. Rebuild: `go build ./cmd/nanobot/`

The `static.go` file uses `//go:embed` to include the frontend files — no separate build step needed.

---

## 🚀 Deployment

### Build a single binary

```bash
# Windows
go build -ldflags="-s -w" -o nanobot.exe ./cmd/nanobot/

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nanobot ./cmd/nanobot/

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o nanobot ./cmd/nanobot/
```

The binary is fully self-contained — the web UI is embedded inside. Just copy the binary + `config/` + `storage/` to the target machine.

### Run as a service (Linux systemd)

```ini
# /etc/systemd/system/nanobot.service
[Unit]
Description=Nanobot Web Service
After=network.target

[Service]
Type=simple
User=nanobot
WorkingDirectory=/opt/nanobot
ExecStart=/opt/nanobot/nanobot -mode web -port 8080
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now nanobot
```

### Docker

```bash
docker compose up -d
```

Or build manually:

```bash
docker build -t nanobot .
docker run -v $(pwd)/config:/app/config -v $(pwd)/storage:/app/storage nanobot -mode web -port 8080
```

---

## 🔒 Security

- **`bash` tool**: supports command whitelist/blacklist — dangerous commands like `rm -rf /`, `dd`, `shutdown` are blocked by default
- **Permission modes**: `default` asks for confirmation before writes; `plan` is read-only; `auto` approves everything
- **Hook system**: every tool execution goes through `PRE_ACTION` → `POST_ACTION` hooks
- **Web mode**: automatically sets permission mode to `auto` (no confirmation prompts)

> ⚠️ The `bash` tool's whitelist checks are currently **commented out**. Review `internal/tools/bash.go` before running untrusted prompts in production.

---

## 📄 License

MIT — see [LICENSE](LICENSE).
