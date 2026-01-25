import time
from typing import Dict, Any, List
from utils.config import get_source_config
from .base import SearchSource
from .types import SearchInput


class BochaSource(SearchSource):
    """
    Bocha搜索源实现
    """
    
    def __init__(self):
        super().__init__()
        source_config = get_source_config("bocha")
        self.api_key = source_config.get("api_key", "")
        self.url = source_config.get("url", "https://api.bochaai.com/v1/web-search")
        self.headers = {
            "Content-Type": "application/json"
        }
        if self.api_key:
            self.headers["Authorization"] = f"Bearer {self.api_key}"
    
    async def get_links(self, query: str, **kwargs) -> Dict[str, Any]:
        """
        获取搜索链接
        """
        start_time = time.time()
        
        # 解析参数
        count = kwargs.get("count", 5)
        freshness = kwargs.get("freshness", "oneWeek")
        summary = kwargs.get("summary", False)
        
        body = {
            "query": query,
            "freshness": freshness,
            "summary": summary,
            "count": count
        }
        
        resp = await self.client.post(self.url, headers=self.headers, json=body)
        
        if resp.status_code != 200:
            print(f"Bocha请求失败: {resp.status_code}, {resp.text}")
            return {"code": -1, "msg": f"Bocha请求失败: {resp.status_code}", "data": {}}
        
        raw_data = resp.json()
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
    
    def parse_result(self, raw_data: Dict[str, Any]) -> List:
        """
        解析Bocha搜索结果
        """
        results = []
        
        if "data" in raw_data and "webPages" in raw_data["data"]:
            for item in raw_data["data"]["webPages"].get("value", []):
                results.append(
                    self.create_link_result(
                        title=item.get("name"),
                        url=item.get("url"),
                        snippet=item.get("snippet"),
                        site_name=item.get("siteName"),
                        published_time=item.get("datePublished")
                    )
                )
        
        return results


# 兼容旧接口
async def get_links(query: str, freshness: str = "oneWeek", summary: bool = False, count: int = 5) -> dict:
    async with BochaSource() as source:
        return await source.get_links(query, freshness=freshness, summary=summary, count=count)





import asyncio

if __name__ == "__main__":
    async def main():
        links = await get_links("AI未来十大风口", "oneDay", True, 10)
        print(links)

    asyncio.run(main())
