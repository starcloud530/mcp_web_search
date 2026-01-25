from typing import Optional
from dataclasses import dataclass
from dataclasses import field
@dataclass
class LinkResult:
    """
    用于封装搜索结果的网页信息
    """
    title: str                  # 网页标题
    url: str                    # 网页链接
    snippet: Optional[str] = None   # 搜索摘要
    site_name: Optional[str] = None # 来源网站名
    published_time: Optional[str] = None  # 发布时间
    extra_fields: dict = field(default_factory=dict)  # 可存任意额外字段

    def __repr__(self) -> str:
        return f"<LinkResult title='{self.title}' url='{self.url}'>"

    def to_dict(self) -> dict:
        """
        转换为 dict，方便序列化
        """
        return {
            "title": self.title,
            "url": self.url,
            "snippet": self.snippet,
            "site_name": self.site_name,
            "published_time": self.published_time
        }
