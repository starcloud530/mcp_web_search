#!/bin/bash

# MCP Web Search Server 停止脚本
# 停止所有服务并清理 PID 文件

echo "正在停止 MCP Web Search Server..."

# 检查 PID 目录是否存在
if [ ! -d "pids" ]; then
    echo "警告: pids 目录不存在，可能没有服务在运行"
    exit 0
fi

# 停止链接搜索服务
if [ -f "pids/links_search.pid" ]; then
    PID=$(cat pids/links_search.pid)
    echo "正在停止链接搜索服务 (PID: $PID)..."
    kill $PID 2>/dev/null || echo "警告: 链接搜索服务可能已停止"
    rm -f pids/links_search.pid
fi

# 停止网页爬取服务
if [ -f "pids/context_crawl.pid" ]; then
    PID=$(cat pids/context_crawl.pid)
    echo "正在停止网页爬取服务 (PID: $PID)..."
    kill $PID 2>/dev/null || echo "警告: 网页爬取服务可能已停止"
    rm -f pids/context_crawl.pid
fi

# 停止 MCP 服务器
if [ -f "pids/mcp_server.pid" ]; then
    PID=$(cat pids/mcp_server.pid)
    echo "正在停止 MCP 服务器 (PID: $PID)..."
    kill $PID 2>/dev/null || echo "警告: MCP 服务器可能已停止"
    rm -f pids/mcp_server.pid
fi

# 清理空的 pids 目录
if [ -z "$(ls -A pids)" ]; then
    rm -rf pids
fi

echo "所有服务已停止成功!"
