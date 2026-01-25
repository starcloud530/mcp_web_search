// ================== 抓取管道选择器，根据URL选择合适的pipeline ===================

package core

import (
	"context_crawl/types"
)

// ChoosePipeline 根据URL选择合适的pipeline
// 使用预排序的列表，返回第一个匹配的pipeline
// 如果没有找到匹配的，返回名称为default的pipeline
func ChoosePipeline(url string) (types.Pipeline, bool) {
	// 遍历预排序的pipeline列表，返回第一个匹配的
	for _, entry := range pipelines {
		if entry.Pipeline.Match(url) {
			return entry.Pipeline, true
		}
	}
	// 不存在时
	return nil, false
}
