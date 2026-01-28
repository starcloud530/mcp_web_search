// ================== URL处理service ===================

package service

import (
	"context"
	"context_crawl/core"
	"context_crawl/types"
	"fmt"
	"log"
	"sync"
	"time"
)

// HandleURL 处理单个URL
// 根据URL选择合适的pipeline，然后使用该pipeline处理数据
// 如果当前pipeline失败或返回空，会尝试下一个pipeline（保底机制）
func HandleURL(input types.Type) (types.Type, error) {
	// 获取第一个匹配的pipeline
	pipeline, found := core.ChoosePipeline(input.Url)
	if !found {
		return types.Type{}, fmt.Errorf("no pipeline found for url: %s", input.Url)
	}

	// 保底机制：尝试所有匹配的pipeline，直到成功
	fallbackPipelines := core.GetPipelinesAfter(input.Url)
	allPipelines := append([]core.PipelineEntry{{Index: 0, Name: "primary", Pipeline: pipeline}}, fallbackPipelines...)

	for i, entry := range allPipelines {
		p := entry.Pipeline

		// 使用pipeline处理数据
		result, err := p.Process(input)
		if err != nil {
			log.Printf("⚠️ 第%d个pipeline(%s)处理失败: %v, URL: %s",
				i+1, entry.Name, err, input.Url)
			continue // 尝试下一个pipeline
		}

		// 检查返回的text是否为空
		if result.Text == "" {
			log.Printf("⚠️ 第%d个pipeline(%s)返回空内容, URL: %s",
				i+1, entry.Name, input.Url)
			continue // 尝试下一个pipeline
		}

		// 成功获取内容
		log.Printf("✅ 第%d个pipeline(%s)成功获取内容, URL: %s",
			i+1, entry.Name, input.Url)
		return result, nil
	}

	// 所有pipeline都失败了
	return types.Type{}, fmt.Errorf("all pipelines failed for url: %s", input.Url)
}

// HandleURLs 并发处理多个URL
// 设置超时时间，只返回成功的内容
func HandleURLs(inputs []types.Type, timeout time.Duration) []types.Type {
	// 创建结果切片
	var results []types.Type

	// 创建互斥锁，保护results切片
	var mu sync.Mutex

	// 创建等待组
	var wg sync.WaitGroup

	// 创建上下文，用于控制超时
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 并发处理每个URL
	for _, input := range inputs {
		wg.Add(1)

		go func(input types.Type) {
			defer wg.Done()

			// 创建通道，用于接收处理结果
			resultChan := make(chan types.Type, 1)
			errChan := make(chan error, 1)

			// 启动goroutine处理单个URL
			go func() {
				result, err := HandleURL(input)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- result
			}()

			// 等待处理结果或超时
			select {
			case result := <-resultChan:
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			case err := <-errChan:
				log.Printf("❌ 处理失败: %v, URL: %s", err, input.Url)
			case <-ctx.Done():
				log.Printf("⏰ 处理超时: %s", input.Url)
			}
		}(input)
	}

	// 等待所有goroutine完成
	wg.Wait()

	return results
}
