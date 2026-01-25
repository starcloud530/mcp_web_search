import abc
import httpx
import asyncio
from typing import Dict, Any, List, Optional
from database.models.link_result import LinkResult


class SearchSource(abc.ABC):
    """
    搜索源基类，定义统一接口
    """
    
    def __init__(self):
        # 异步客户端，默认使用
        self.client = None
        self.use_async_client = True
    
    async def __aenter__(self):
        if self.use_async_client:
            self.client = httpx.AsyncClient(timeout=5)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.client:
            await self.client.aclose()
    
    @abc.abstractmethod
    async def get_links(self, query: str, **kwargs) -> Dict[str, Any]:
        """
        获取搜索链接
        
        Args:
            query: 搜索关键词
            **kwargs: 其他参数
            
        Returns:
            包含code、msg和data的字典
            data包含links列表和duration
        """
        pass
    
    @abc.abstractmethod
    def parse_result(self, raw_data: Any) -> List:
        """
        解析原始搜索结果
        
        Args:
            raw_data: 原始搜索结果
            
        Returns:
            LinkResult列表
        """
        pass
    
    def create_link_result(self, title: str, url: str, snippet: Optional[str] = None,
                          site_name: Optional[str] = None, published_time: Optional[str] = None,
                          **kwargs) -> LinkResult:
        """
        创建LinkResult对象，统一处理换行符
        """
        # 删除摘要中的换行符
        if snippet:
            snippet = snippet.replace('\n', '')
        
        return LinkResult(
            title=title,
            url=url,
            snippet=snippet,
            site_name=site_name,
            published_time=published_time,
            **kwargs
        )
    
    async def run_sync_function(self, func, *args, **kwargs):
        """
        在异步上下文中运行同步函数
        
        Args:
            func: 同步函数
            *args: 函数参数
            **kwargs: 函数关键字参数
            
        Returns:
            函数返回值
        """
        return await asyncio.to_thread(func, *args, **kwargs)

