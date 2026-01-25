package colly

import (
	"fmt"
	"regexp"
	"strings"

	"context_crawl/types"
)

// ScoredChunker 实现了基于质量评分的文本分块器
type ScoredChunker struct {
	ScoreThreshold float64 // 分块质量分数阈值
}

// NewScoredChunker 创建一个新的ScoredChunker实例
func NewScoredChunker(scoreThreshold float64) *ScoredChunker {
	return &ScoredChunker{
		ScoreThreshold: scoreThreshold,
	}
}

// NewDefaultScoredChunker 创建一个默认的ScoredChunker实例，scoreThreshold为0.0
func NewDefaultScoredChunker() *ScoredChunker {
	return &ScoredChunker{
		ScoreThreshold: 0.0,
	}
}

// Chunk 对文本进行分块处理，实现types.Chunker接口
func (sc *ScoredChunker) Chunk(input types.Type) (types.Type, error) {
	text := input.Text
	chunkSize := 500 // 默认分块大小

	// 定义Chunk结构体
	type Chunk struct {
		Text   string
		Score  float64
		IsCode bool
	}

	var chunks []Chunk

	// 先把占位符单独分离，防止被正则切句拆开
	reCodePlaceholder := regexp.MustCompile(`@CODE_\d+@`)
	segments := reCodePlaceholder.Split(text, -1)
	placeholders := reCodePlaceholder.FindAllString(text, -1)

	// 使用输入Type中的代码映射
	codeMap := input.CodeMap
	if codeMap == nil {
		codeMap = make(map[string]string)
	}

	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			// 按句子切分普通文本
			reSentence := regexp.MustCompile(`[^。！？.!?]+[。！？.!?]?`)
			sentences := reSentence.FindAllString(segment, -1)

			var current string
			for _, s := range sentences {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}

				if len(current)+len(s) <= chunkSize {
					current += s
				} else {
					if current != "" {
						score := sc.chunkScore(current)
						if score >= sc.ScoreThreshold {
							chunks = append(chunks, Chunk{
								Text:   current,
								Score:  score,
								IsCode: false,
							})
						}
					}
					current = s
				}
			}

			// 收尾
			if current != "" {
				score := sc.chunkScore(current)
				if score >= sc.ScoreThreshold {
					chunks = append(chunks, Chunk{
						Text:   current,
						Score:  score,
						IsCode: false,
					})
				}
			}
		}

		// 插入对应的占位符代码块（保证每个 segment 结束后才插一次）
		if i < len(placeholders) {
			ph := placeholders[i]
			codeText := codeMap[ph]
			chunks = append(chunks, Chunk{
				Text:   codeText,
				Score:  1.0,
				IsCode: true,
			})
		}
	}

	// 如果没有分块，返回提示信息
	if len(chunks) == 0 {
		chunks = append(chunks, Chunk{
			Text:   "查询结果为空，当前链接中无有效信息，请尝试其他关键词或者其他链接。",
			Score:  0.0,
			IsCode: false,
		})
	}

	// 格式化分块结果
	var organizedText strings.Builder
	for i, chunk := range chunks {
		organizedText.WriteString(fmt.Sprintf("### chunk %d (recall_score:%.3f is_code:%t):\n", i+1, chunk.Score, chunk.IsCode))
		organizedText.WriteString(chunk.Text + "\n\n")
	}

	return types.Type{
		Url:  input.Url,
		Text: organizedText.String(),
	}, nil
}

// chunkScore chunk 质量评分函数（私有方法）
func (sc *ScoredChunker) chunkScore(text string) float64 {
	textLen := len(strings.TrimSpace(text))
	if textLen == 0 {
		return 0
	}

	reAlpha := regexp.MustCompile(`[A-Za-z0-9]`)
	alphaCount := len(reAlpha.FindAllString(text, -1))
	reCJK := regexp.MustCompile(`[\p{Han}]`)
	cjkCount := len(reCJK.FindAllString(text, -1))
	score := float64(cjkCount+alphaCount) / float64(textLen)
	if score > 1.0 {
		score = 1.0
	}
	return score
}
