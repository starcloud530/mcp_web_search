# file: search_api.py
from fastapi import FastAPI, Query
from typing import List, Dict, Any
import asyncio

from utils.bm25 import bm25_rank_links



# 导入搜索源
from source.bocha_source import get_links as bocha_get_links
from source.mita_source import get_links as mita_get_links
from source.duckgo_source import get_links as duckgo_get_links
from utils.config import is_source_enabled



app = FastAPI(title="AI 联网搜索 API")

from pydantic import BaseModel

class LINKREQUEST(BaseModel):
    query: str
    count: int = 5


async def fetch_links(query: str, count: int = 5) -> List:
    """
    异步调用启用的搜索源，并去重
    单个源失败不影响其他结果返回
    """
    tasks = []
    
    # 检查并添加启用的搜索源
    if is_source_enabled("bocha"):
        tasks.append(bocha_get_links(query, count=count))
    
    if is_source_enabled("mita"):
        # 添加异常处理，防止Mita源失败导致整个请求失败
        async def mita_with_exception_handling():
            try:
                return await mita_get_links(query, size=count)
            except Exception as e:
                print(f"Mita搜索失败: {e}")
                return {"code": -1, "msg": "Mita搜索失败", "data": {}}
        tasks.append(mita_with_exception_handling())
    
    if is_source_enabled("duckgo"):
        # DuckDuckGo是同步函数，需要使用asyncio.to_thread来异步运行
        # 添加异常处理，防止单个源失败导致整个请求失败
        async def duckgo_with_exception_handling():
            try:
                return await asyncio.to_thread(duckgo_get_links, query, max_results=count)
            except Exception as e:
                print(f"DuckDuckGo搜索失败: {e}")
                return {"code": -1, "msg": "DuckDuckGo搜索失败", "data": {}}
        tasks.append(duckgo_with_exception_handling())
    
    # 使用return_exceptions=True，确保单个任务失败不影响其他任务
    results = await asyncio.gather(*tasks, return_exceptions=True)
    
    links = []
    seen = set()
    for result in results:
        # 处理异常结果
        if isinstance(result, Exception):
            print(f"搜索源异常: {result}")
            continue
        
        # 处理正常结果
        if result["code"] == 0:
            for link in result["data"]["links"]:
                if link.url not in seen:
                    seen.add(link.url)
                    links.append(link)
    
    # 如果所有源都失败，返回空列表
    if not links:
        print("警告: 所有搜索源都失败，返回空结果")
    
    return links


@app.post("/get_links")
async def search(req : LINKREQUEST) -> Dict[str, Any]:
    """
    搜索接口：
    1. 调用多个搜索源获取链接
    2. 对结果进行 BM25 排序
    3. 返回排序后的链接列表
    """
    start_time = asyncio.get_event_loop().time()
    
    links = await fetch_links(req.query, count=req.count)
    
    # BM25 排序
    sorted_links = bm25_rank_links(links, req.query)
    
    duration = asyncio.get_event_loop().time() - start_time
    
    # 返回结果，保留 BM25 分数
    data = []
    for link in sorted_links:
        data.append({
            "title": link.title,
            "url": link.url,
            "snippet": link.snippet,
            "site_name": link.site_name,
            "published_time": link.published_time,
            "bm25_score": link.extra_fields.get("bm25_score"),
            "bm25_rank": link.extra_fields.get("bm25_rank")
        })
    
    return {"code": 0, "msg": "success", "data": {"duration": round(duration, 2), "links": data}}

"""
curl -X POST http://localhost:8000/get_links \
-H "Content-Type: application/json" \
-d '{"query":"golang web framework","count":5}'

{
  "code": 0,
  "msg": "success",
  "data": {
    "duration": 0.2345,
    "links": [
      {
        "title": "Gin Web Framework",
        "url": "https://gin-gonic.com/",
        "snippet": "Gin is a HTTP web framework written in Go...",
        "site_name": "gin-gonic.com",
        "published_time": "2022-01-01",
        "bm25_score": 3.45,
        "bm25_rank": 1
      },
      ...
    ]
  }
}


"""
