package colly

import (
	"fmt"
	"regexp"
	"strings"

	"context_crawl/types"
)

// BasicCleaner 实现了基于正则表达式的文本清洗器
type BasicCleaner struct{}

// NewBasicCleaner 创建一个新的BasicCleaner实例
func NewBasicCleaner() *BasicCleaner {
	return &BasicCleaner{}
}

// Clean 清洗HTML内容，实现types.Cleaner接口
func (bc *BasicCleaner) Clean(input types.Type) (types.Type, error) {
	html := input.Text
	// 匹配 <pre ...>...</pre> 和 <code ...>...</code>，支持带属性和多行
	reCode := regexp.MustCompile(`(?s)<pre[^>]*>(.*?)</pre>|<code[^>]*>(.*?)</code>`)
	matches := reCode.FindAllStringSubmatch(html, -1)
	codeMap := make(map[string]string)
	for i, m := range matches {
		var code string
		if m[1] != "" {
			code = m[1]
		} else if m[2] != "" {
			code = m[2]
		} else {
			continue
		}
		// 删除 HTML 标签
		reHTML := regexp.MustCompile(`(?s)<[^>]*>`)
		code = reHTML.ReplaceAllString(code, "")
		code = strings.TrimSpace(code)
		placeholder := fmt.Sprintf("@CODE_%d@", i)
		codeMap[placeholder] = code
		// 替换原始代码为占位符，保留换行和缩进
		html = strings.Replace(html, code, placeholder, 1)
	}

	// 去掉 HTML 标签
	reHTML := regexp.MustCompile(`(?s)<[^>]*>`)
	html = reHTML.ReplaceAllString(html, "")
	reURL := regexp.MustCompile(`(https?:\\*\/\*\/*[^\s\"']+)`)
	html = reURL.ReplaceAllString(html, "")
	// 压缩多余空白，但保持占位符两侧至少一个空格
	for placeholder := range codeMap {
		html = strings.ReplaceAll(html, placeholder, " "+placeholder+" ")
	}
	reSpace := regexp.MustCompile(`\s+`)
	html = reSpace.ReplaceAllString(html, " ")
	html = strings.TrimSpace(html)

	return types.Type{
		Url:     input.Url,
		Text:    html,
		CodeMap: codeMap,
	}, nil
}
