// ================== 路由注册 ===================

package app

import (
	"context_crawl/app/v1"
	"github.com/gin-gonic/gin"
)

// RouterAPI 设置路由统一入口
func RouterAPI() *gin.Engine {
	r := gin.Default()

	// 注册v1版本的路由
	v1.RegisterRoutes(r)

	return r
}
