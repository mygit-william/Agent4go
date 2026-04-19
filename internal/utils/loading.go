package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// LoadingAnimation 加载动画
type LoadingAnimation struct {
	message string
	stop    chan struct{}
	done    chan struct{}
	once    sync.Once
	started bool
}

// NewLoadingAnimation 创建加载动画
func NewLoadingAnimation(message string) *LoadingAnimation {
	return &LoadingAnimation{
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start 启动动画
func (l *LoadingAnimation) Start() {
	if l.started {
		return
	}
	l.started = true

	go func() {
		defer close(l.done)

		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		cyan := color.New(color.FgCyan)

		for {
			select {
			case <-l.stop:
				// 清除当前行
				fmt.Printf("\r\x1b[K")
				return
			default:
				cyan.Printf("\r%s %s...", frames[i], l.message)
				i = (i + 1) % len(frames)
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Stop 停止动画并显示完成标记
func (l *LoadingAnimation) Stop() {
	l.once.Do(func() {
		close(l.stop)
		<-l.done // 等待 goroutine 退出
	})
}

// StopWithResult 停止动画并显示结果
func (l *LoadingAnimation) StopWithResult(success bool, result string) {
	l.Stop()

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	if success {
		// 截断过长的结果
		display := result
		lines := strings.Split(display, "\n")
		if len(lines) > 5 {
			display = strings.Join(lines[:5], "\n") + fmt.Sprintf("\n  ... (共 %d 行，已截断)", len(lines))
		}
		if len(display) > 200 {
			display = display[:200] + "..."
		}
		green.Printf("✅ %s 完成\n", l.message)
		if display != "" {
			fmt.Printf("   %s\n", display)
		}
	} else {
		red.Printf("❌ %s 失败: %s\n", l.message, result)
	}
}
