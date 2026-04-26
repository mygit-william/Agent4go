package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mygit-william/nanobot-go/internal/channels"
	"github.com/mygit-william/nanobot-go/internal/core"
	"github.com/mygit-william/nanobot-go/internal/llm"
	"github.com/mygit-william/nanobot-go/internal/memory"
	"github.com/mygit-william/nanobot-go/internal/web"
)

func main() {
	mode := flag.String("mode", "cli", "运行模式: cli, web 或 serve")
	configPath := flag.String("config", "config/config.json", "配置文件路径")
	port := flag.String("port", "8080", "Web 模式端口")
	flag.Parse()

	// 获取项目根目录：优先使用当前工作目录
	projectRoot, _ := os.Getwd()
	
	// 检查配置文件是否存在，如果不存在尝试使用可执行文件目录
	configFile := filepath.Join(projectRoot, *configPath)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		execPath, _ := os.Executable()
		projectRoot = filepath.Dir(filepath.Dir(execPath))
		configFile = filepath.Join(projectRoot, *configPath)
	}

	// 加载配置文件
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("❌ 配置文件不存在: %s\n", configFile)
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("❌ 配置文件解析失败: %v\n", err)
		os.Exit(1)
	}

	// 创建 LLM 工厂
	llmFactory := llm.NewFactory(config.LLM)
	llmAdapter, err := llmFactory.Make()
	if err != nil {
		fmt.Printf("❌ 创建 LLM 适配器失败: %v\n", err)
		os.Exit(1)
	}

	// 创建 Agent
	permissionMode := config.Permissions.Mode
	if permissionMode == "" {
		permissionMode = "default"
	}
	agent := core.NewAgent(llmAdapter, filepath.Join(projectRoot, "storage"), permissionMode)

	// 创建记忆管理器
	memManager := memory.NewManager(memory.Config{
		StorageDir:     filepath.Join(projectRoot, "storage"),
		MaxContextSize: config.Agent.MaxContextSize,
		SummaryEnabled: config.Agent.SummaryEnabled,
	})

	// 设置飞书通知器
	feishuNotifier := channels.NewFeishuNotifier(config.Channels.Feishu)
	agent.SetNotifier(feishuNotifier)
	if feishuNotifier.IsEnabled() {
		fmt.Println("📢 飞书通知已启用")
	}

	// Web 模式下权限设为 auto（不弹窗确认）
	if *mode == "web" {
		agent.SetPermissionMode("auto")
		fmt.Println("🔓 Web 模式: 权限设为 auto（自动执行）")
	}

	switch *mode {
	case "cli":
		cliChannel := channels.NewCLIChannel(agent, projectRoot)
		cliChannel.Receive()
	case "web":
		fmt.Println("🚀 启动 Web 服务...")
		server := web.NewServer(agent, memManager, projectRoot, *port)
		if err := server.Run(); err != nil {
			fmt.Printf("❌ Web 服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		fmt.Println("🚀 启动钉钉服务...")
		// TODO: 实现钉钉服务模式
	default:
		fmt.Printf("❌ 未知模式: %s\n", *mode)
		os.Exit(1)
	}
}

// Config 配置结构
type Config struct {
	App struct {
		Name  string `json:"name"`
		Debug bool   `json:"debug"`
	} `json:"app"`
	LLM         llm.Config `json:"llm"`
	Channels    struct {
		DingTalk struct {
			Enabled    bool   `json:"enabled"`
			AppKey     string `json:"app_key"`
			AppSecret  string `json:"app_secret"`
			WebhookURL string `json:"webhook_url"`
		} `json:"dingtalk"`
		Feishu channels.FeishuConfig `json:"feishu"`
	} `json:"channels"`
	Permissions struct {
		Mode        string `json:"mode"`
		Description string `json:"description"`
	} `json:"permissions"`
	Agent struct {
		MaxContextSize int  `json:"max_context_size"`
		MaxTokens      int  `json:"max_tokens"`
		SummaryEnabled bool `json:"summary_enabled"`
	} `json:"agent"`
}
