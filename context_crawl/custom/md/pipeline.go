package md

import (
	"context_crawl/base/colly"
	"context_crawl/types"
)

type MarkdownPipeline struct {
	Crawler types.Crawler // 爬虫组件
	Chunker types.Chunker // 分块组件
	Cleaner types.Cleaner // 清洗组件
}

// NewMarkdownPipeline 创建一个新的 MarkdownPipeline 实例
func NewMarkdownPipeline() *MarkdownPipeline {
	// 自定义markdown抓取逻辑
	crawler := NewMarkdownCrawler()
	// 复用 colly 其余组件
	cleaner := colly.NewBasicCleaner()
	chunker := colly.NewScoredChunker(0.0) // 设置scoreThreshold为0.0

	return &MarkdownPipeline{
		Crawler: crawler,
		Chunker: chunker,
		Cleaner: cleaner,
	}
}
func (p *MarkdownPipeline) Process(input types.Type) (types.Type, error) {
	// 直接使用Crawler爬取页面
	result, err := p.Crawler.Crawl(input)
	if err != nil {
		return types.Type{}, err
	}

	// 清洗和分块处理
	cleanResult, err := p.Cleaner.Clean(result)
	if err != nil {
		return types.Type{}, err
	}

	chunkResult, err := p.Chunker.Chunk(cleanResult)
	if err != nil {
		return types.Type{}, err
	}

	return chunkResult, nil
}

// Match 匹配方法，检查URL是否是Markdown文件
func (p *MarkdownPipeline) Match(url string) bool {
	return IsMarkdownFile(url)
}
