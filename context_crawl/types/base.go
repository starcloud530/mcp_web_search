package types

// 定义输入输出通用类，关于网页的处理无外乎网址 和 文本
type Type struct {
	Url     string            // URL
	Text    string            // 任意类型的文本
	CodeMap map[string]string // 代码映射，用于存储代码占位符和实际代码内容的映射
}
