// ================== URL处理service ===================

package service

import (
	"context"
	"context_crawl/core"
	"context_crawl/types"
	"fmt"
	"sync"
	"time"
)

// HandleURL 处理单个URL
// 根据URL选择合适的pipeline，然后使用该pipeline处理数据
func HandleURL(input types.Type) (types.Type, error) {
	// 选择合适的pipeline
	pipeline, found := core.ChoosePipeline(input.Url)
	if !found {
		return types.Type{}, fmt.Errorf("no pipeline found for url: %s", input.Url)
	}

	// 使用pipeline处理数据
	result, err := pipeline.Process(input)
	if err != nil {
		return types.Type{}, fmt.Errorf("pipeline process failed: %w", err)
	}

	return result, nil
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
				// 处理成功，添加到结果切片
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			case <-errChan:
				// 处理失败，忽略
			case <-ctx.Done():
				// 超时，忽略
			}
		}(input)
	}

	// 等待所有goroutine完成
	wg.Wait()

	return results
}