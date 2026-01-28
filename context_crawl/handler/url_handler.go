// ================== URL处理handler ===================
package handler

import (
	"context_crawl/handler/models"
	"context_crawl/service"
	"context_crawl/types"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// ProcessURLs 处理多个URL的请求
func ProcessURLs(request models.Request) models.Response {
	// 将URL列表转换为[]types.Type
	var inputs []types.Type
	for _, url := range request.Urls {
		inputs = append(inputs, types.Type{Url: url})
	}

	// 调用service.HandleURLs处理多个URL，设置10秒超时
	results := service.HandleURLs(inputs, 10*time.Second)

	// 构建响应数据
	data := make(map[string]interface{})
	// 初始化为空数组而不是nil
	processedResults := make([]map[string]interface{}, 0)

	for _, result := range results {
		processedResults = append(processedResults, map[string]interface{}{
			"url":  result.Url,
			"text": result.Text,
		})
	}

	data["results"] = processedResults

	if len(processedResults) == 0 {
		log.Printf("⚠️ 没有爬取到任何内容，URLs: %v", request.Urls)
	}

	// 构建响应
	return models.Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}

// HandleProcessURLs 处理多个URL的HTTP请求
func HandleProcessURLs(c *gin.Context) {
	var request models.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, models.Response{Code: -1, Msg: "Invalid request body", Data: nil})
		return
	}

	response := ProcessURLs(request)
	c.JSON(200, response)
}
