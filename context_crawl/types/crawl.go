// ================= 定义所有pipeline中 crawl.go 的输入、输出 接口规范 =====================
package types

// Crawler 接口定义爬虫组件的统一行为
type Crawler interface {
	Crawl(Type) (Type, error)
}
