# Nanobot-Go

> 一个基于 Go 语言的智能 Agent 系统，支持多种 LLM 后端。

## 特性

- 🚀 多模型支持: Zhipu、OpenAI、Ollama、Longcat
- 🔒 安全执行: 命令白名单/黑名单过滤
- 🧠 长期记忆: 对话历史持久化
- 📱 多通道: 支持 CLI 和钉钉
- 🛠️ 工具丰富: 文件操作、命令执行

## 项目结构

```
nanobot-go/
├── cmd/nanobot/main.go    # 入口
├── internal/
│   ├── core/             # 核心逻辑
│   ├── llm/              # LLM 适配器
│   ├── channels/         # 通信通道
│   ├── tools/            # 工具执行
│   ├── hooks/            # 事件钩子
│   └── utils/            # 工具函数
├── storage/              # 数据存储
└── config/               # 配置文件
```

## 快速开始

```bash
# 安装依赖
go mod tidy

# 运行 CLI 模式
go run cmd/nanobot/main.go -mode cli

# 运行服务模式
go run cmd/nanobot/main.go -mode serve
```

## 编译可执行文件

```bash
# 编译 CLI 版本
go build -o nanobot.exe cmd/nanobot/main.go

# 编译服务版本
go build -o nanobot-server.exe cmd/nanobot/main.go

# 跨平台编译
# Windows
GOOS=windows GOARCH=amd64 go build -o nanobot.exe cmd/nanobot/main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o nanobot cmd/nanobot/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o nanobot cmd/nanobot/main.go


## 配置

编辑 `config/config.json` 设置 LLM API 密钥。

## 从 PHP 版本迁移

本项目从 `nanobot-php` 迁移而来，功能完全一致。