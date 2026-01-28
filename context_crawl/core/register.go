// ******** 统一在此处注册 pipeline **************

package core

import (
	"context_crawl/base/colly"
	"context_crawl/custom/github"
	"context_crawl/custom/md"
	"context_crawl/custom/pdf"
	"context_crawl/types"
)

// ================= 注册入口 =============

// PipelineEntry 表示一个pipeline注册条目
type PipelineEntry struct {
	Index    int            // 匹配优先级，数值越大优先级越高
	Name     string         // pipeline标识符
	Pipeline types.Pipeline // 对应的pipeline实例
}

// 维护一个全局的pipelines注册器
var pipelines []PipelineEntry // 按优先级排序的pipeline列表

// init 函数在包初始化时调用，注册所有pipeline
func init() {
	register()
}

func register() {

	// 1️⃣ 注册通用 colly pipeline
	RegisterPipeline(10, "colly", colly.NewCollyPipeline())

	// 2️⃣ 注册GitHub pipeline
	RegisterPipeline(15, "github", github.NewGitHubPipeline(""))

	// 3️⃣ 注册markdown pipeline
	RegisterPipeline(20, "markdown", md.NewMarkdownPipeline())

	// 4️⃣ 注册PDF pipeline
	RegisterPipeline(25, "pdf", pdf.NewPDFPipeline())

}

// RegisterPipeline 注册一个pipeline
// 直接添加到列表并排序
func RegisterPipeline(index int, name string, pipeline types.Pipeline) {
	// 创建新的pipeline条目
	entry := PipelineEntry{
		Index:    index,
		Name:     name,
		Pipeline: pipeline,
	}

	// 添加到列表
	pipelines = append(pipelines, entry)

	// 按Index排序（优先级高的在前）
	sortPipelines()
}

// sortPipelines 对pipeline列表按优先级排序
func sortPipelines() {
	// 按Index排序（优先级高的在前，数值越大优先级越高）
	for i := 0; i < len(pipelines); i++ {
		for j := i + 1; j < len(pipelines); j++ {
			if pipelines[i].Index < pipelines[j].Index {
				pipelines[i], pipelines[j] = pipelines[j], pipelines[i]
			}
		}
	}
}

// GetAllPipelines 返回所有已注册的pipeline（按优先级排序）
func GetAllPipelines() []PipelineEntry {
	return pipelines
}

// GetPipelinesAfter 返回第一个匹配指定URL的pipeline之后，下一个也匹配该URL的pipeline
// 用于保底机制：当前pipeline失败后，尝试下一个匹配的pipeline
func GetPipelinesAfter(url string) []PipelineEntry {
	found := false

	for _, entry := range pipelines {
		if found {
			// 只返回下一个也匹配该URL的pipeline
			if entry.Pipeline.Match(url) {
				return []PipelineEntry{entry}
			}
		}
		if !found && entry.Pipeline.Match(url) {
			found = true
		}
	}

	return []PipelineEntry{}
}
