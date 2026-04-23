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

		base := color.New(color.FgHiBlack)
		soft := color.New(color.FgCyan)
		mid := color.New(color.FgHiCyan, color.Bold)
		shine := color.New(color.FgHiWhite, color.Bold)
		runes := []rune(l.message)
		if len(runes) == 0 {
			runes = []rune("处理中")
		}

		// 给光效前后留一点空白，看起来更像“扫过”
		padding := 3
		frame := 0
		direction := 1

		for {
			select {
			case <-l.stop:
				// 清除当前行
				fmt.Printf("\r\x1b[K")
				return
			default:
				shinePos := frame - padding

				var b strings.Builder
				b.WriteString("\r")
				for idx, r := range runes {
					dist := absInt(idx - shinePos)
					switch dist {
					case 0:
						b.WriteString(shine.Sprint(string(r)))
					case 1:
						b.WriteString(mid.Sprint(string(r)))
					case 2:
						b.WriteString(soft.Sprint(string(r)))
					default:
						b.WriteString(base.Sprint(string(r)))
					}
				}
				fmt.Print(b.String())

				frame += direction
				maxFrame := len(runes) + padding*2 - 1
				if frame >= maxFrame {
					frame = maxFrame
					direction = -1
				} else if frame <= 0 {
					frame = 0
					direction = 1
				}
				time.Sleep(95 * time.Millisecond)
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
		green.Printf("✔ %s 完成\n", l.message)
	} else {
		red.Printf("✖ %s 失败\n", l.message)
		if result != "" {
			fmt.Printf("  %s\n", result)
		}
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
