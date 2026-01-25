// ============== chunk 统一 输入、输出、接口规范 ==============
package types

// Chunker 接口定义分块组件的统一行为
type Chunker interface {
	Chunk(Type) (Type, error)
}
