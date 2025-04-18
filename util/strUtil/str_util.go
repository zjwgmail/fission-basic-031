package strUtil

import (
	"regexp"
	"strings"
)

// ReplacePlaceholders 根据占位符替换字符串中的变量
func ReplacePlaceholders(template string, values ...string) string {
	re := regexp.MustCompile(`\{\{(\d+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// 获取占位符的数字
		index := match[2] - '1' // match[2] 是数字字符
		if int(index) < len(values) {
			return values[int(index)]
		}
		return match // 如果没有匹配，保持原样
	})
	return result
}

func RemoveDirectionalFormatting(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\u202A' || r == '\u202C' {
			return -1
		}
		return r
	}, s)
}
