package pdf

import (
	"fmt"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// IsPDFFile 判断是否是PDF文件
func IsPDFFile(url string) bool {
	// 检查URL是否以.pdf结尾
	if strings.HasSuffix(strings.ToLower(url), ".pdf") {
		return true
	}
	// 检查URL是否包含arxiv.org/pdf
	if strings.Contains(url, "arxiv.org/pdf") {
		return true
	}
	return false
}

// GetPDFFileName 获取PDF文件的名称
func GetPDFFileName(url string) string {
	if strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]
		if IsPDFFile(filename) {
			return filename
		}
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("pdf_%d.pdf", timestamp)
}

// ExtractTextFromPDF 从PDF文件中提取文本
// 使用github.com/ledongthuc/pdf库实现PDF文本提取
func ExtractTextFromPDF(filePath string) (string, error) {
	// 打开PDF文件
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Sprintf("[PDF文件打开失败]\n错误: %v\n文件路径: %s", err, filePath), nil
	}
	defer f.Close()

	var textBuilder strings.Builder

	// 遍历所有页面提取文本
	for pageNum := 1; pageNum <= r.NumPage(); pageNum++ {
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}

		// 提取页面文本
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}

		// 添加页面文本到结果
		textBuilder.WriteString(text)
		textBuilder.WriteString("\n")
	}

	// 如果没有提取到文本，返回提示信息
	if textBuilder.Len() == 0 {
		return fmt.Sprintf("[PDF文本提取为空]\n文件路径: %s", filePath), nil
	}

	return textBuilder.String(), nil
}
