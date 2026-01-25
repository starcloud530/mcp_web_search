// ================ clean.go 输入、输出、接口规范 =====================
package types

// Cleaner 接口定义清洗组件的统一行为
type Cleaner interface {
	Clean(Type) (Type, error)
}
