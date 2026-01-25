// ================== API路由注册 ===================

package v1

import (
	"context_crawl/handler"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册v1版本的路由
func RegisterRoutes(router *gin.Engine) {
	// 注册处理多个URL的接口
	router.POST("/crawl", handler.HandleProcessURLs)
}
