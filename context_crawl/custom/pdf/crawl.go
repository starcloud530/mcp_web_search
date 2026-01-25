package pdf

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"context_crawl/types"
)

// PDFCrawler 负责爬取PDF文件
type PDFCrawler struct {
	TempDir string
}

// NewPDFCrawler 创建一个新的PDFCrawler实例
func NewPDFCrawler() *PDFCrawler {
	return &PDFCrawler{
		TempDir: os.TempDir(),
	}
}

// Crawl 爬取单个PDF文件，实现types.Crawler接口
func (pc *PDFCrawler) Crawl(input types.Type) (types.Type, error) {
	if IsPDFFile(input.Url) {
		return pc.CrawlPDFFile(input.Url)
	}
	return types.Type{}, fmt.Errorf("not a PDF file: %s", input.Url)
}

// CrawlPDFFile 爬取单个PDF文件
func (pc *PDFCrawler) CrawlPDFFile(url string) (types.Type, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return pc.downloadPDFFile(url)
	} else {
		return pc.readLocalPDFFile(url)
	}
}

// downloadPDFFile 下载远程PDF文件
func (pc *PDFCrawler) downloadPDFFile(url string) (types.Type, error) {
	// 创建临时文件
	tempFile, err := os.CreateTemp(pc.TempDir, "pdf_*.pdf")
	if err != nil {
		return types.Type{}, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	// 设置请求上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.Type{}, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 执行请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Type{}, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return types.Type{}, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 检查Content-Type是否为PDF
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/pdf" && !strings.Contains(contentType, "pdf") {
		return types.Type{}, fmt.Errorf("不是PDF文件，Content-Type: %s", contentType)
	}

	// 下载文件内容
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return types.Type{}, fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 提取PDF文本
	text, err := ExtractTextFromPDF(tempFile.Name())
	if err != nil {
		return types.Type{}, fmt.Errorf("提取PDF文本失败: %v", err)
	}

	// 删除临时文件
	os.Remove(tempFile.Name())

	return types.Type{
		Url:  url,
		Text: text,
	}, nil
}

// readLocalPDFFile 读取本地PDF文件
func (pc *PDFCrawler) readLocalPDFFile(filePath string) (types.Type, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return types.Type{}, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 提取PDF文本
	text, err := ExtractTextFromPDF(filePath)
	if err != nil {
		return types.Type{}, fmt.Errorf("提取PDF文本失败: %v", err)
	}

	return types.Type{
		Url:  filePath,
		Text: text,
	}, nil
}
