// ================== 主程序入口 ===================

package main

import (
	"context_crawl/app"
	"context_crawl/utils"
	"fmt"
	"os"
)

func main() {
	// 加载配置文件
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "../config.yaml"
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v, using default port 7008\n", err)
		// 设置路由
		router := app.RouterAPI()
		// 启动服务器
		port := ":7008"
		fmt.Printf("Server is running on port %s\n", port)
		if err := router.Run(port); err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
		return
	}

	// 设置路由
	router := app.RouterAPI()

	// 启动服务器
	port := fmt.Sprintf(":%d", config.Server.Port)
	fmt.Printf("Server is running on port %s\n", port)
	if err := router.Run(port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
