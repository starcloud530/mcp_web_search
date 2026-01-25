# bm25_api.py
from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Tuple
from utils.bm25 import bm25_rank_tool
import asyncio
from concurrent.futures import ThreadPoolExecutor

app = FastAPI(title="BM25 Ranking API")

# 请求模型
class BM25Request(BaseModel):
    corpus: List[str]
    query: str
    k1: float = 1.5
    b: float = 0.75

# 响应模型
class BM25Result(BaseModel):
    text: str
    score: float

# 异步封装 bm25_rank_tool
async def async_rank_texts(corpus: List[str], query: str, k1: float, b: float) -> List[Tuple[str, float]]:
    loop = asyncio.get_event_loop()
    with ThreadPoolExecutor() as pool:
        ranked = await loop.run_in_executor(pool, bm25_rank_tool, corpus, query, k1, b)
    return ranked

@app.post("/bm25_rank", response_model=List[BM25Result])
async def bm25_rank_api(request: BM25Request):
    """
    BM25 排序 API
    """
    ranked = await async_rank_texts(request.corpus, request.query, request.k1, request.b)
    # 转成响应模型
    return [{"text": text, "score": score} for text, score in ranked]


# 测试
if __name__ == "__main__":
    import uvicorn
    uvicorn.run("utils.bm25_api:app", host="0.0.0.0", port=8003, reload=True)
