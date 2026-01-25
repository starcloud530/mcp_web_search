package md

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

// MarkdownCrawler 负责爬取Markdown文件
type MarkdownCrawler struct {
	TempDir string
}

// NewMarkdownCrawler 创建一个新的MarkdownCrawler实例
func NewMarkdownCrawler() *MarkdownCrawler {
	return &MarkdownCrawler{
		TempDir: os.TempDir(),
	}
}

// Crawl 爬取单个页面，实现types.Crawler接口
func (mc *MarkdownCrawler) Crawl(input types.Type) (types.Type, error) {
	if IsMarkdownFile(input.Url) {
		return mc.CrawlMarkdownFile(input.Url)
	}
	return types.Type{}, fmt.Errorf("not a markdown file: %s", input.Url)
}

// CrawlMarkdownFiles 爬取多个Markdown文件
func (mc *MarkdownCrawler) CrawlMarkdownFiles(urls []string) ([]types.Type, error) {
	results := make([]types.Type, 0, len(urls))

	for _, url := range urls {
		result, err := mc.CrawlMarkdownFile(url)
		if err != nil {
			fmt.Printf("爬取Markdown文件失败: %v, URL: %s\n", err, url)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// CrawlMarkdownFile 爬取单个Markdown文件
func (mc *MarkdownCrawler) CrawlMarkdownFile(url string) (types.Type, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return mc.downloadMarkdownFile(url)
	} else {
		return mc.readLocalMarkdownFile(url)
	}
}

// downloadMarkdownFile 下载远程Markdown文件
func (mc *MarkdownCrawler) downloadMarkdownFile(url string) (types.Type, error) {
	tempFile, err := os.CreateTemp(mc.TempDir, "markdown_*.md")
	if err != nil {
		return types.Type{}, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.Type{}, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.Type{}, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Type{}, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return types.Type{}, fmt.Errorf("写入临时文件失败: %v", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return types.Type{}, fmt.Errorf("读取临时文件失败: %v", err)
	}

	os.Remove(tempFile.Name())

	text := fmt.Sprintf("<div>%s</div>", string(content))

	return types.Type{
		Url:  url,
		Text: text,
	}, nil
}

// readLocalMarkdownFile 读取本地Markdown文件
func (mc *MarkdownCrawler) readLocalMarkdownFile(filePath string) (types.Type, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return types.Type{}, fmt.Errorf("读取本地文件失败: %v", err)
	}

	text := fmt.Sprintf("<div>%s</div>", string(content))

	return types.Type{
		Url:  filePath,
		Text: text,
	}, nil
}
