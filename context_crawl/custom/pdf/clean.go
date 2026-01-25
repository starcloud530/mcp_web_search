package pdf

import (
	"strings"

	"context_crawl/types"
)

// PDFCleaner 负责清洗PDF文本
type PDFCleaner struct {}

// NewPDFCleaner 创建一个新的PDFCleaner实例
func NewPDFCleaner() *PDFCleaner {
	return &PDFCleaner{}
}

// Clean 清洗PDF文本，实现types.Cleaner接口
func (c *PDFCleaner) Clean(input types.Type) (types.Type, error) {
	text := input.Text

	// 移除多余的空行
	text = c.removeExtraEmptyLines(text)

	// 移除多余的空格
	text = c.removeExtraSpaces(text)

	// 移除PDF特有的标记和噪声
	text = c.removePDFSpecificNoise(text)

	return types.Type{
		Url:  input.Url,
		Text: text,
	}, nil
}

// removeExtraEmptyLines 移除多余的空行
func (c *PDFCleaner) removeExtraEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			cleanedLines = append(cleanedLines, trimmedLine)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

// removeExtraSpaces 移除多余的空格
func (c *PDFCleaner) removeExtraSpaces(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// removePDFSpecificNoise 移除PDF特有的标记和噪声
func (c *PDFCleaner) removePDFSpecificNoise(text string) string {
	// 移除常见的PDF标记
	noisePatterns := []string{
		"[PDF文本提取功能需要PDF处理库支持]",
		"Page ",
		"PDF",
	}

	for _, pattern := range noisePatterns {
		text = strings.ReplaceAll(text, pattern, "")
	}

	return text
}
