package channels

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/mygit-william/nanobot-go/internal/llm"
)

// Receive 接收输入
func (c *CLIChannel) Receive() {
	// 设置编码
	c.setupEncoding()

	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	cyan.Println("Nanobot CLI")
	fmt.Println("输入任务开始执行，或使用: help | clear | exit")
	fmt.Println()

	// 构建系统提示词
	messages := c.buildSystemPrompt()

	reader := bufio.NewReader(os.Stdin)

	for {
		green.Print("› ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			yellow.Println("会话已结束。")
			break
		}

		if input == "" {
			fmt.Println("提示: 输入任务内容，或使用 help/exit。")
			continue
		}

		if input == "help" {
			fmt.Println("可用命令:")
			fmt.Println("  • help  - 显示帮助")
			fmt.Println("  • clear - 清屏")
			fmt.Println("  • exit  - 退出")
			fmt.Println("  • 其他输入会作为任务交给 Agent")
			fmt.Println()
			continue
		}

		if input == "clear" {
			cmd := exec.Command("clear")
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", "cls")
			}
			cmd.Stdout = os.Stdout
			cmd.Run()
			continue
		}

		// 调用 Agent
		reply := c.agent.Chat("cli_user", input, &messages)
		cyan.Printf("\n助手回复:\n%s\n\n", strings.TrimSpace(reply))
	}
}

func (c *CLIChannel) setupEncoding() {
	// Windows 下设置 UTF-8 编码
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "chcp", "65001").Run()
	}
}

func (c *CLIChannel) buildSystemPrompt() []llm.Message {
	agentsFile := c.projectRoot + "/storage/AGENTS.md"
	data, err := os.ReadFile(agentsFile)
	systemContent := ""
	if err == nil {
		systemContent = string(data)
	}

	// 添加机器信息
	machineInfo := fmt.Sprintf(`
### 工作目录
工作目录在 %s 下, MEMORY.md 在 %s/memory 目录下.
当前系统: %s
`, c.projectRoot, c.projectRoot, runtime.GOOS)

	return []llm.Message{
		{Role: "system", Content: systemContent + machineInfo},
	}
}
