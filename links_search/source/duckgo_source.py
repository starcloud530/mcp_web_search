import asyncio
from typing import List, Dict, Any
from duckduckgo_search import DDGS
from database.models.link_result import LinkResult
import time
from .base import SearchSource


class DuckDuckGoSource(SearchSource):
    """
    DuckDuckGo搜索源实现
    """
    
    def __init__(self):
        super().__init__()
        self.use_async_client = False  # DuckDuckGo使用同步API
    
    async def get_links(self, query: str, **kwargs) -> Dict[str, Any]:
        """
        获取搜索链接
        """
        start_time = time.time()
        
        # 解析参数
        max_results = kwargs.get("max_results", 5)
        
        # 使用同步函数搜索
        raw_data = await self.run_sync_function(
            self._sync_search,
            query,
            max_results
        )
        
        links = self.parse_result(raw_data)
        
        duration = time.time() - start_time
        return {
            "code": 0,
            "msg": "success",
            "data": {
                "links": links,
                "duration": round(duration, 2)
            }
        }
    
    def _sync_search(self, query: str, max_results: int) -> List[Dict[str, Any]]:
        """
        同步搜索函数
        """
        search_term = query
        results = []
        
        with DDGS() as ddgs:
            for item in ddgs.text(
                keywords=search_term,
                max_results=max_results,
                safesearch=False,
            ):
                results.append(item)
        
        return results
    
    def parse_result(self, raw_data: List[Dict[str, Any]]) -> List:
        """
        解析DuckDuckGo搜索结果
        """
        results = []
        
        for item in raw_data:
            results.append(
                self.create_link_result(
                    title=item.get("title"),
                    url=item.get("href"),
                    snippet=item.get("body"),
                    site_name=None,  # DuckDuckGo API 没有返回来源站点，可以不填
                    published_time=None
                )
            )
        
        return results


# 兼容旧接口
def get_links(query: str, max_results: int = 5) -> dict:
    """
    使用 DuckDuckGo 搜索，并返回 LinkResult 列表
    Args:
        query: 搜索关键字或者提问
        max_results: 返回的结果数量
    Returns:
        LinkResult 列表
    """
    start_time = time.time()
    search_term = query
    results: List[LinkResult] = []
    with DDGS() as ddgs:
        for item in ddgs.text(
            keywords=search_term,
            max_results=max_results,
            safesearch=False,
        ):
            results.append(
                LinkResult(
                    title=item.get("title"),
                    url=item.get("href"),
                    snippet=item.get("body").replace('\n', '') if item.get("body") else None,
                    site_name=None,  # DuckDuckGo API 没有返回来源站点，可以不填
                    published_time=None
                )
            )
    duration = time.time() - start_time

    return {
        "code":0,
        "msg":"success",
        "data":{
            "links":results,
            "duration":round(duration, 2)
        }
    }

if __name__ == '__main__':
    async def test():
        async with DuckDuckGoSource() as source:
            results = await source.get_links("湖南大学", max_results=3)
            print(results)
    
    asyncio.run(test())
