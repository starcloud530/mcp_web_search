// ================ pipeline:组合代码执行逻辑的管道 =====================
package types

// Pipeline 接口定义了Pipeline的统一行为
type Pipeline interface {
	Process(Type) (Type, error)
	Match(url string) bool // 匹配方法，用于判断是否处理该URL
}

// PipelineRegister 接口定义了Pipeline包的注册行为
// 每个pipeline包应该导出一个Register函数，该函数会将pipeline注册到系统中
type PipelineRegister interface {
	Register() // 注册函数，用于将pipeline注册到系统中
}
