#!/bin/bash

# MCP Web Search Server 启动脚本
# 根据 README.md 启动两个子服务和一个 MCP 服务

echo "正在启动 MCP Web Search Server..."

# 检查 Python 是否安装
if ! command -v python &> /dev/null; then
    echo "错误: Python 未安装"
    exit 1
fi

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装"
    exit 1
fi

# 安装 Python 依赖
echo "正在安装 Python 依赖..."
pip install -r requirements.txt

# 安装 Go 依赖
echo "正在安装 Go 依赖..."
cd context_crawl
go mod tidy

# 编译 Go 程序
echo "正在编译 Go 程序..."
go build -o context_crawl main.go
cd ..

# 创建 PID 目录
mkdir -p pids

# 启动链接搜索服务 (端口: 8004)
echo "正在启动链接搜索服务..."
cd links_search
nohup python main.py > ../logs/links_search.log 2>&1 &
echo $! > ../pids/links_search.pid
cd ..

# 启动网页爬取服务 (端口: 8003)
echo "正在启动网页爬取服务..."
cd context_crawl
nohup ./context_crawl > ../logs/context_crawl.log 2>&1 &
echo $! > ../pids/context_crawl.pid
cd ..

# 启动 MCP 服务器 (端口: 8006)
echo "正在启动 MCP 服务器..."
nohup python server.py > logs/mcp_server.log 2>&1 &
echo $! > pids/mcp_server.pid

echo "所有服务已启动成功!"
echo "- 链接搜索服务: http://localhost:8004"
echo "- 网页爬取服务: http://localhost:8003"
echo "- MCP 服务器: http://localhost:8006"
echo "日志文件位于 logs/ 目录"
echo "PID 文件位于 pids/ 目录"
