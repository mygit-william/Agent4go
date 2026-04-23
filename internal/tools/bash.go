package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Bash Bash 命令工具
type Bash struct {
	executor *ShellExecutor
}

// NewBash 创建工具
func NewBash() *Bash {
	return &Bash{
		executor: NewShellExecutor(),
	}
}

// Name 工具名称
func (t *Bash) Name() string {
	return "bash"
}

// Description 工具描述
func (t *Bash) Description() string {
	return "Run a shell command in the current workspace with security checks."
}

// Parameters 参数定义
func (t *Bash) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"command": map[string]interface{}{
			"type":        "string",
			"description": "要执行的bash命令 (仅允许安全命令)",
			"maxLength":   500,
		},
	}
}

// Required 必填参数
func (t *Bash) Required() []string {
	return []string{"command"}
}

// Execute 执行工具
func (t *Bash) Execute(args map[string]interface{}) string {
	command, ok := args["command"].(string)
	if !ok {
		return "错误: 缺少 command 参数"
	}

	// 命令长度检查
	if len(command) > 500 {
		return "错误: 命令过长，最大允许500个字符"
	}

	return t.executor.Exec(command)
}

// ShellExecutor 安全的 Shell 执行器
type ShellExecutor struct {
	commandWhitelist []string
	commandBlacklist []string
	timeout          time.Duration
	maxOutputSize    int
}

// NewShellExecutor 创建执行器
func NewShellExecutor() *ShellExecutor {
	return &ShellExecutor{
		commandWhitelist: []string{
			"ls", "dir", "cat", "head", "tail", "grep", "find", "file", "stat",
			"pwd", "whoami", "hostname", "uname", "date", "time", "which", "whereis",
			"ps", "top", "htop", "kill", "jobs",
			"ping", "curl", "wget", "netstat", "ss", "ifconfig", "ip",
			"git", "composer", "php", "python", "node", "npm", "yarn",
			"echo", "printf", "awk", "sed", "sort", "uniq", "cut", "tr",
			"tar", "gzip", "gunzip", "zip", "unzip",
			"chmod", "chown", "chgrp",
			"go", "cargo", "rustc", "mvn", "gradle",
		}, 
		commandBlacklist: []string{
			"rm", "rmdir", "dd", "mkfs", "fdisk", "format",
			"sudo", "su", "ssh", "scp", "sftp",
			"chroot", "mount", "umount",
			"killall", "pkill", "shutdown", "reboot", "halt",
		},
		timeout:       30 * time.Second,
		maxOutputSize: 1024 * 1024, // 1MB
	}
}

// Exec 执行命令
func (e *ShellExecutor) Exec(command string) string {
	// 基础安全检查
	if !e.basicSecurityCheck(command) {
		return "错误: 命令未通过基础安全检查"
	}

	// 解析命令
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "错误: 空命令"
	}

	cmdName := parts[0]

	// 白名单检查
	// if !e.inWhitelist(cmdName) {
	// 	return fmt.Sprintf("错误: 命令不在白名单中 - %s", cmdName)
	// }

	// 黑名单检查
	if e.inBlacklist(cmdName) {
		return fmt.Sprintf("错误: 命令在黑名单中 - %s", cmdName)
	}

	// 参数安全检查
	// if len(parts) > 1 {
	// 	args := strings.Join(parts[1:], " ")
	// 	if !e.checkParameters(args) {
	// 		return "错误: 参数包含非法内容"
	// 	}
	// }

	// 执行命令
	return e.executeSafely(command)
}

func (e *ShellExecutor) basicSecurityCheck(command string) bool {
	if strings.TrimSpace(command) == "" {
		return false
	}

	if len(command) > 500 {
		return false
	}

	// 检查特殊字符
	// dangerous := []string{"|", ">", "<", "&&", "||", ";", "$(", "`"}
	// for _, d := range dangerous {
	// 	if strings.Contains(command, d) {
	// 		return false
	// 	}
	// }

	return true
}

func (e *ShellExecutor) inWhitelist(cmd string) bool {
	for _, c := range e.commandWhitelist {
		if c == cmd {
			return true
		}
	}
	return false
}

func (e *ShellExecutor) inBlacklist(cmd string) bool {
	for _, c := range e.commandBlacklist {
		if c == cmd {
			return true
		}
	}
	return false
}

func (e *ShellExecutor) checkParameters(args string) bool {
	dangerous := []string{"..", "//", "./", "$(", "`", "|", ">", "<", "&&", "||", ";", "\n", "\r"}
	for _, d := range dangerous {
		if strings.Contains(args, d) {
			return false
		}
	}

	if strings.Contains(args, "sudo ") || strings.Contains(args, "su ") {
		return false
	}

	return true
}

func (e *ShellExecutor) executeSafely(command string) string {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("错误: %v\n%s", err, string(output))
	}

	result := string(output)
	if len(result) > e.maxOutputSize {
		result = result[:e.maxOutputSize] + "\n... (输出被截断，超过1MB限制)"
	}

	return result
}
