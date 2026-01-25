from typing import Dict, Any, List, Optional
from pydantic import BaseModel


class SearchInput(BaseModel):
    """
    搜索输入参数模型
    """
    query: str  # 搜索关键词
    count: int = 5  # 返回结果数量
    freshness: Optional[str] = "oneWeek"  # 时间范围
    include_summary: bool = False  # 是否包含摘要
    scope: str = "webpage"  # 搜索范围


class SearchResultItem(BaseModel):
    """
    单个搜索结果项
    """
    title: str
    url: str
    snippet: Optional[str] = None
    site_name: Optional[str] = None
    published_time: Optional[str] = None


class SearchOutput(BaseModel):
    """
    搜索输出结果
    """
    code: int  # 0表示成功，非0表示失败
    msg: str  # 结果描述
    data: Dict[str, Any]  # 包含links列表和duration
    
    class Config:
        arbitrary_types_allowed = True
