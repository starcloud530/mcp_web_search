// ======== 使用colly的通用抓取规则 ============ //
package colly

import (
	"context_crawl/types"
)

// CollyPipeline 定义了一个完整的处理流程，包括爬取、清洗和分块
type CollyPipeline struct {
	Crawler types.Crawler // 爬虫组件
	Cleaner types.Cleaner // 清洗组件
	Chunker types.Chunker // 分块组件
}

// NewCollyPipeline 创建一个新的Pipeline实例，使用colly作为爬虫组件
func NewCollyPipeline() *CollyPipeline {
	crawler := NewCollyCrawler()
	cleaner := NewBasicCleaner()
	chunker := NewScoredChunker(0.0)
	return &CollyPipeline{
		Crawler: crawler,
		Cleaner: cleaner,
		Chunker: chunker,
	}
}

// Process 执行collypipeline处理流程，实现types.Pipeline接口
func (p *CollyPipeline) Process(input types.Type) (types.Type, error) {
	// 1. 爬取页面
	pageResult, err := p.Crawler.Crawl(input)
	if err != nil {
		return types.Type{}, err
	}

	// 2. 清洗和分块处理
	cleanResult, err := p.Cleaner.Clean(pageResult)
	if err != nil {
		return types.Type{}, err
	}

	chunkResult, err := p.Chunker.Chunk(cleanResult)
	if err != nil {
		return types.Type{}, err
	}

	return chunkResult, nil
}

// Match 匹配方法，默认匹配所有URL
func (p *CollyPipeline) Match(url string) bool {
	return true
}
