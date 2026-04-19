package utils

import (
	"strings"
)

// CleanJSON 清理 JSON 字符串
func CleanJSON(s string) string {
	s = strings.TrimSpace(s)
	
	// 移除 markdown 代码块标记
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	
	return s
}

// ExtractJSON 从文本中提取 JSON
func ExtractJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			depth++
		} else if s[i] == '}' {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	
	return ""
}
