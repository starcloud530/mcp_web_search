# file: mcp_service.py

import os
import yaml
import aiohttp
from typing import List, Dict, Any

# MCP
from mcp.server.fastmcp import FastMCP

# ===================== 读取配置 =====================
# 统一配置文件路径
CONFIG_PATH = os.path.join(os.path.dirname(__file__), "config.yaml")

# 读取统一配置文件
with open(CONFIG_PATH, 'r', encoding='utf-8') as f:
    config = yaml.load(f, Loader=yaml.FullLoader)

# 获取各服务配置
context_crawl_config = config['context_crawl']
links_search_config = config['links_search']
mcp_config = config['mcp']

# 构建服务API地址
host_1 = context_crawl_config['host']
port_1 = context_crawl_config['port']
crawl_server_api = f"http://{host_1}:{port_1}/crawl"

host_2 = links_search_config['host']
port_2 = links_search_config['port']
search_server_api = f"http://{host_2}:{port_2}/get_links"

print(f"crawl_server_api: {crawl_server_api}")
print(f"search_server_api: {search_server_api}")

# ===================== 创建 MCP =====================
mcp = FastMCP("web-search", host=mcp_config['host'], port=mcp_config['port'])

# is_calling_search : 后续是否允许使用这些链接进行搜索工具的调用
# ===================== 工具函数 =====================
@mcp.tool()
async def get_links(query: str, count: int = 5) -> str:
    """
    获取与关键词相关的互联网链接，并通过摘要预览链接的信息
    Args:
        query: 搜索关键词，例如 "python 爬虫 教程"
        count: 需要返回的链接数量
    Returns:
        格式化后的链接信息

    """
    async with aiohttp.ClientSession() as session:
        async with session.post(
            search_server_api,
            json={"query": query, "count": count}
        ) as resp:
            if resp.status != 200:
                raise RuntimeError(f"Search API failed: {resp.status}")
            data = await resp.json()
            links = data["data"]["links"]
    #return links

    # 转换为文本
    text_list = []
    for link in links:
        title = link.get("title", "")
        url = link.get("url", "")
        snippet = link.get("snippet", "")
        text = f"链接: {url}\n标题: {title}\n摘要: {snippet}\n"
        text_list.append(text)
    
    result = "\n".join(text_list)
    #result += "\n\n请使用get_page_content工具并传入上述URL来获取完整内容"
    return result



@mcp.tool()
async def get_page_content(urls: List[str]) -> str:
    """
    爬取指定URL的网页完整内容，这在已经从摘要中捕捉到重要信息，想要进一步了解更加全面的内容时非常有用
    Args:
        urls: 你要全文浏览的URL列表，例如 ["https://www.baidu.com", "https://www.google.com"]
    Returns:
        网页的完整内容
    """
    async with aiohttp.ClientSession() as session:
        async with session.post(
            crawl_server_api,
            json={"urls": urls}
        ) as resp:
            if resp.status != 200:
                raise RuntimeError(f"Crawl API failed: {resp.status}")
            data = await resp.json()
            results = data["data"].get("results", [])

    if not results:
        return "查询结果为空，当前链接中无有效信息，请尝试其他关键词或者其他链接。"

    text_list = []
    for page in results:
        url = page.get("url", "")
        text = page.get("text", "")
        text_list.append(f"URL: {url}\n{text}")

    return "\n\n===\n\n".join(text_list)

if __name__ == '__main__':
    print("MCP:web-search is running on port 8006.")
    mcp.run(transport="sse")

