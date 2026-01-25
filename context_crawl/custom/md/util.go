package md

import (
	"fmt"
	"strings"
	"time"
)

// IsMarkdownFile 判断是否是Markdown文件
func IsMarkdownFile(url string) bool {
	extensions := []string{".md", ".markdown", ".mdown", ".mdwn", ".mdtxt", ".mdtext"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return true
		}
	}
	return false
}

// GetMarkdownFileName 获取Markdown文件的名称
func GetMarkdownFileName(url string) string {
	if strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]
		if IsMarkdownFile(filename) {
			return filename
		}
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("markdown_%d.md", timestamp)
}
