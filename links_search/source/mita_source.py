import httpx
import asyncio
from typing import Dict, Any, List
from database.models.link_result import LinkResult
import time
from utils.config import get_source_config
from .base import SearchSource


class MitaSource(SearchSource):
    """
    Mita搜索源实现
    """
    
    def __init__(self):
        super().__init__()
        source_config = get_source_config("mita")
        self.api_key = source_config.get("api_key", "")
        self.url = source_config.get("url", "https://metaso.cn/api/v1/search")
        self.headers = {
            "Accept": "application/json",
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
        scope = kwargs.get("scope", "webpage")
        include_summary = kwargs.get("include_summary", False)
        size = kwargs.get("size", 10)
        include_raw_content = kwargs.get("include_raw_content", False)
        concise_snippet = kwargs.get("concise_snippet", False)
        
        body = {
            "q": query,
            "scope": scope,
            "includeSummary": include_summary,
            "size": str(size),  # Mita要求size是字符串类型
            "includeRawContent": include_raw_content,
            "conciseSnippet": concise_snippet,
            "format": "chat_completions"
        }
        
        try:
            resp = await self.client.post(self.url, headers=self.headers, json=body)
        except httpx.RequestError as e:
            print(f"Mita请求异常: {e}")
            return {"code": -1, "msg": f"Mita请求异常: {e}", "data": {}}
        
        if resp.status_code != 200:
            print(f"Mita请求失败: {resp.status_code}, {resp.text}")
            return {"code": -1, "msg": f"Mita请求失败: {resp.status_code}", "data": {}}
        
        raw_data = resp.json()
        links = self.parse_result(raw_data)
        
        if not links:
            return {"code": -1, "msg": "Mita请求失败", "data": {}}
        
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
        解析Mita搜索结果
        """
        results = []
        
        if "webpages" in raw_data:
            for item in raw_data["webpages"]:
                results.append(
                    self.create_link_result(
                        title=item.get("title"),
                        url=item.get("link"),
                        snippet=item.get("snippet"),
                        site_name=None,
                        published_time=item.get("date")
                    )
                )
        else:
            print(f"Mita解析失败: {raw_data}")
        
        return results


# 兼容旧接口
async def get_links(
    query: str,
    scope: str = "webpage",
    include_summary: bool = False,
    size: int = 10,
    include_raw_content: bool = False,
    concise_snippet: bool = False,
) -> dict:
    async with MitaSource() as source:
        return await source.get_links(
            query,
            scope=scope,
            include_summary=include_summary,
            size=size,
            include_raw_content=include_raw_content,
            concise_snippet=concise_snippet,
        )


# ------------------ 测试 ------------------
if __name__ == "__main__":
    async def main():
        result = await get_links("python")
        print(f"耗时: {result['data']['duration']:.2f}s")
        for link in result["data"]["links"]:
            print(link)

    asyncio.run(main())
