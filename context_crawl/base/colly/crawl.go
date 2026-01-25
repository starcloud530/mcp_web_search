// file: colly_crawler.go
/*

简易 Web 爬虫，使用 Colly 库，并限制最大并发数。

整理设计倾向于代码搜索，针对代码的提取进行了一系列优化

*/
package colly

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"

	"context_crawl/types"
)

// CollyCrawler 实现了基于Colly的爬虫

type CollyCrawler struct {
	MaxConcurrent  int           // 最大并发数
	RequestTimeout time.Duration // 请求超时时间
}

// NewCollyCrawler 创建一个新的CollyCrawler实例
func NewCollyCrawler() *CollyCrawler {
	return &CollyCrawler{
		MaxConcurrent:  60,               // 默认最大并发数
		RequestTimeout: 15 * time.Second, // 默认请求超时15秒
	}
}

// 异步Chromedp封装
type FetchResult struct {
	URL  string
	HTML string
	Err  error
}

// semaphore 控制最大并发
var chromedpSem = make(chan struct{}, 3) // 最多 3 个动态页面同时抓取
var wg sync.WaitGroup

// 默认重试次数（总共尝试4次）
const MaxRetries = 3

// 指数退避时间（较长的时间）
func getBackoffDelay(retryCount int) time.Duration {
	// 使用较长的指数退避：1s, 2s, 4s
	return time.Duration(1<<uint(retryCount)) * 1000 * time.Millisecond
}

// 判断错误是否应该重试
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误类型
	errStr := err.Error()

	// 连接错误
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "timeout") {
		return true
	}

	// HTTP状态码错误
	if strings.Contains(errStr, "408") || // 请求超时
		strings.Contains(errStr, "409") || // 冲突
		strings.Contains(errStr, "429") || // 速率限制
		strings.Contains(errStr, "500") || // 内部错误
		strings.Contains(errStr, "502") || // 网关错误
		strings.Contains(errStr, "503") || // 服务不可用
		strings.Contains(errStr, "504") { // 网关超时
		return true
	}

	return false
}

func FetchPageAsync(url string, resultChan chan<- FetchResult) {
	wg.Add(1)
	chromedpSem <- struct{}{} // 获取 token
	go func() {
		defer wg.Done()
		defer func() { <-chromedpSem }() // 释放 token

		var html string
		var err error

		// 重试机制：默认重试3次（总共4次尝试）
		for i := 0; i <= MaxRetries; i++ {
			// 设置代理选项
			opts := []chromedp.ExecAllocatorOption{
				chromedp.NoFirstRun,
				chromedp.NoDefaultBrowserCheck,
				chromedp.Headless,
			}

			// 添加代理设置
			if proxy := os.Getenv("http_proxy"); proxy != "" {
				opts = append(opts, chromedp.ProxyServer(proxy))
			} else if proxy := os.Getenv("https_proxy"); proxy != "" {
				opts = append(opts, chromedp.ProxyServer(proxy))
			}

			allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
			defer cancel()
			ctx, cancel := chromedp.NewContext(allocCtx)
			defer cancel()
			ctx, cancel = context.WithTimeout(ctx, 10*time.Second) // 单次请求10秒超时
			defer cancel()

			// 等待时间根据重试次数递增：2s, 3s, 4s, 5s
			waitTime := 2 + time.Duration(i)*time.Second
			err = chromedp.Run(ctx,
				chromedp.Navigate(url),
				chromedp.Sleep(waitTime), // 根据重试次数递增等待时间
				chromedp.OuterHTML("html", &html),
			)

			if err == nil {
				// 检查是否包含错误信息
				if !containsErrorMessages(html) {
					break // 成功且无错误信息则退出重试循环
				}
				// 如果包含错误信息，继续重试
				log.Printf("⚠️ 检测到页面错误信息，第%d次重试", i+1)
			}

			// 检查是否应该重试
			if i < MaxRetries && shouldRetry(err) {
				backoff := getBackoffDelay(i)
				log.Printf("⚠️ 第%d次抓取失败，%v秒后重试: %v", i+1, backoff.Seconds(), err)
				time.Sleep(backoff)
			} else {
				// 不应该重试或已达到最大重试次数，直接退出
				if i < MaxRetries {
					log.Printf("❌ 错误不可重试，放弃抓取: %v", err)
				}
				break
			}
		}

		resultChan <- FetchResult{
			URL:  url,
			HTML: html,
			Err:  err,
		}
	}()
}
func (cc *CollyCrawler) Crawl(input types.Type) (types.Type, error) {
	fmt.Println("🚀 开始爬取网页...")
	start := time.Now()
	var result types.Type
	resultChan := make(chan types.Type, 1)
	// dynamicResultChan := make(chan FetchResult, 1) // 禁用动态抓取后不再需要

	// 全局超时：60 秒内必须结束（考虑到重试机制）
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 多个用户代理轮换
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	}

	// Colly 初始化 - 使用同步模式
	c := colly.NewCollector(
		colly.Async(false), // 使用同步模式
		colly.MaxDepth(1),
	)
	// 随机选择用户代理
	c.UserAgent = userAgents[time.Now().UnixNano()%int64(len(userAgents))]
	c.AllowURLRevisit = false

	// 不使用代理，直接连接
	log.Printf("不使用代理，直接连接")
	c.WithTransport(&http.Transport{
		Proxy: nil, // 禁用代理
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	})

	// 简化限制设置
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,                      // 同步模式下设置为1
		Delay:       500 * time.Millisecond, // 适当延迟
	})

	// 静态页面
	c.OnHTML("body", func(e *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		e.DOM.Find("script, style, noscript").Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
		html, err := e.DOM.Html()
		if err != nil {
			log.Println("❌ 网页解析失败:", e.Request.URL, err)
			return
		}
		// 暂时禁用动态抓取，优先返回静态结果
		// if needsJS(html) {
		//  FetchPageAsync(e.Request.URL.String(), dynamicResultChan)
		//  return
		// }
		select {
		case resultChan <- types.Type{Url: e.Request.URL.String(), Text: html}:
		case <-ctx.Done():
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("❌ Error: %v, URL: %s, StatusCode: %d", err, r.Request.URL, r.StatusCode)
		// 对于超时错误，记录更详细的信息
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("⏰ 请求超时，可能是网络问题或页面响应慢")
		}
	})

	// 处理单个输入
	url := input.Url
	// 使用Colly爬取网络页面
	c.Visit(url)
	// 同步模式下不需要Wait()
	// c.Wait()

	// 关闭resultChan，因为我们已经禁用了动态抓取
	close(resultChan)

	// 处理网络爬取的结果
	done := make(chan struct{})
	go func() {
		for r := range resultChan {
			result = r
			break // 只取第一个结果
		}
		close(done)
	}()

	select {
	case <-ctx.Done(): // 超时直接返回
		fmt.Println("⏰ 爬取超时，返回空结果")
		return types.Type{}, fmt.Errorf("爬取超时")
	case <-done:
		fmt.Println("✅ 网页爬取完成")
	}

	fmt.Println("总耗时:", time.Since(start))
	return result, nil
}

// 清洗文本，保留代码块占位符
// 清洗文本，保留代码块占位符
// 辅助函数
// 判断网页是否需要 JS 渲染
func needsJS(html string) bool {
	// 如果 body 内容太短，就认为是动态页面
	if len(html) < 100 {
		return true
	}

	// 判断是否有常见的 JS 框架标识
	jsFrameworks := []string{
		"id=\"app\"",
		"id=\"root\"",
		"id=\"__next\"",
		"react-root",
		"vue-app",
		"ng-app",
		"data-reactroot",
	}

	for _, framework := range jsFrameworks {
		if strings.Contains(strings.ToLower(html), strings.ToLower(framework)) {
			log.Printf("🔍 检测到JS框架标识: %s", framework)
			return true
		}
	}

	// 检查是否包含大量JavaScript代码
	if strings.Count(html, "<script") > 10 {
		log.Printf("🔍 检测到大量script标签，可能是动态页面")
		return true
	}

	return false
}

// 检测页面是否包含错误信息
func containsErrorMessages(html string) bool {
	errorPatterns := []string{
		"Uh oh! There was an error while loading",
		"Please reload this page",
		"Something went wrong",
		"An error occurred",
		"Error loading",
		"Failed to load",
	}

	lowerHTML := strings.ToLower(html)
	for _, pattern := range errorPatterns {
		if strings.Contains(lowerHTML, strings.ToLower(pattern)) {
			log.Printf("⚠️ 检测到错误信息: %s", pattern)
			return true
		}
	}

	return false
}
