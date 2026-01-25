package pdf

import (
	"context_crawl/base/colly"
	"context_crawl/types"
)

// PDFPipeline 定义了PDF文件的完整处理流程
type PDFPipeline struct {
	Crawler types.Crawler
	Cleaner types.Cleaner
	Chunker types.Chunker
}

// NewPDFPipeline 创建一个新的PDFPipeline实例
func NewPDFPipeline() *PDFPipeline {
	crawler := NewPDFCrawler()
	cleaner := NewPDFCleaner()
	// 复用colly的ScoredChunker
	chunker := colly.NewScoredChunker(0.0)
	return &PDFPipeline{
		Crawler: crawler,
		Cleaner: cleaner,
		Chunker: chunker,
	}
}

// Process 执行PDF pipeline处理流程，实现types.Pipeline接口
func (p *PDFPipeline) Process(input types.Type) (types.Type, error) {
	// 1. 爬取PDF文件
	pdfResult, err := p.Crawler.Crawl(input)
	if err != nil {
		return types.Type{
			Url:  input.Url,
			Text: "PDF文件爬取失败: " + err.Error(),
		}, nil
	}

	// 2. 清洗PDF文本
	cleanResult, err := p.Cleaner.Clean(pdfResult)
	if err != nil {
		return types.Type{
			Url:  input.Url,
			Text: "PDF文本清洗失败: " + err.Error(),
		}, nil
	}

	// 3. 分块处理（直接使用colly的ScoredChunker）
	chunkResult, err := p.Chunker.Chunk(cleanResult)
	if err != nil {
		return types.Type{
			Url:  input.Url,
			Text: "PDF文本分块失败: " + err.Error(),
		}, nil
	}

	return chunkResult, nil
}

// Match 匹配方法，实现types.Pipeline接口
func (p *PDFPipeline) Match(url string) bool {
	return IsPDFFile(url)
}
