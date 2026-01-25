import tiktoken
import sys
import os




from database.models.link_result import LinkResult

# 初始化tiktoken分词器
# 选择适合中文的编码，cl100k_base是GPT-4的编码，支持多语言
ENCODING = tiktoken.get_encoding("cl100k_base")

#分词
def hf_tokenize(text: str):
    """
    使用 tiktoken 对文本分词

    Args:
        text (str): 输入文本
    Returns:
        List[str]: 分词后的 token 字符串列表
    """
    if not text:
        return []

    # 使用 tiktoken 分词，返回 token ID 列表，然后转换为字符串
    token_ids = ENCODING.encode(text)
    # 转换为字符串表示（使用解码后的子词）
    tokens = [ENCODING.decode([token_id]) for token_id in token_ids]
    return tokens


# 预处理
from typing import List
from collections import Counter
def preprocess_corpus(corpus: List[str]):
    """
    对语料库进行 BM25 预处理： 1、分词 2、统计美篇文档的词频 3、统计整个语料库的文档频率 4、计算平均文档长度

    Args:
        corpus (List[str]): 输入语料库
    Returns:
        docs_tokens (List[List[str]]): 每篇文档的 token 列表
        doc_len (List[int]): 每篇文档长度
        tf_list (List[Counter]): 每篇文档的词频统计
        df (Dict[str, int]): token 的文档频率
        avgdl (float): 平均文档长度

    """
    doc_tokens=[]
    doc_len=[]
    tf_list=[]
    df_counter=Counter()
    for doc in corpus:
        tokens=hf_tokenize(doc)
        doc_tokens.append(tokens)
        doc_len.append(len(tokens)) 
        
        tf_counter = Counter(tokens)
        tf_list.append(tf_counter)

        for token in tf_counter.keys():
            df_counter[token]+=1
    avgdl = sum(doc_len) / len(doc_len) if doc_len else 0
    df = dict(df_counter)
    return doc_tokens, doc_len, tf_list, df, avgdl

import math 
from typing import Dict 

def compute_idf(df: Dict[str, int], N: int) -> Dict[str, float]:
    """
    计算每个 token 的 idf 值

    Args:
        df: token 的文档频率
        N: 语料库的文档数量
    Returns:
        Dict[str, float]: token 及其对应的 idf 值
    """
    idf = {}
    for token, freq in df.items():
        idf[token] = math.log(1+(N-freq+0.5) / (freq+0.5) )
    return idf  


def bm25_score(query: str, doc_index: int, docs_tokens, tf_list, doc_len, avgdl, idf, k1=1.5, b=0.75):
    """
    计算单篇文档对 query 的 BM25 得分

    Args:
        query (str): 查询文本
        doc_index (int): 文档下标
        docs_tokens (List[List[str]]): 文档分词列表
        tf_list (List[Counter]): 每篇文档词频
        doc_len (List[int]): 每篇文档长度
        avgdl (float): 平均文档长度
        idf (Dict[str, float]): token IDF
        k1 (float): BM25 参数
        b (float): BM25 参数

    Returns:
        float: 文档得分
    """
    query_tokens = hf_tokenize(query)
    # 停用词过滤
    
    score = 0.0
    tf_counter = tf_list[doc_index]
    dl = doc_len[doc_index]

    for token in query_tokens:
        f = tf_counter.get(token, 0)
        token_idf = idf.get(token, 0.0)
        denom = f + k1 * (1 - b + b * dl / avgdl)
        score += token_idf * f * (k1 + 1) / (denom + 1e-6)

    return score


def bm25_rank(query: str, docs_tokens, tf_list, doc_len, avgdl, idf):
    """
    对语料库中所有文档计算 BM25 并返回排序索引
    """
    scores = [bm25_score(query, i, docs_tokens, tf_list, doc_len, avgdl, idf) for i in range(len(docs_tokens))]
    ranked_idx = sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)
    return ranked_idx, scores



from typing import List, Tuple

def bm25_rank_tool(corpus: List[str], query: str, k1=1.5, b=0.75) -> List[Tuple[str, float]]:
    """
    对文本列表进行 BM25 排序
    
    Args:
        corpus (List[str]): 文本列表
        query (str): 查询文本
        k1 (float): BM25 参数
        b (float): BM25 参数
    
    Returns:
        List[Tuple[str, float]]: 排序后的文本列表，每项为 (文本, 得分)
    """
    # 1️⃣ 预处理
    docs_tokens, doc_len, tf_list, df, avgdl = preprocess_corpus(corpus)
    
    # 2️⃣ 计算 IDF
    idf = compute_idf(df, len(corpus))
    
    # 3️⃣ 计算每篇文档 BM25 得分
    scores = [bm25_score(query, i, docs_tokens, tf_list, doc_len, avgdl, idf, k1, b)
              for i in range(len(corpus))]
    
    # 4️⃣ 排序
    ranked_idx = sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)
    
    # 返回排序后的文本和得分
    return [(corpus[i], scores[i]) for i in ranked_idx]


def bm25_rank_links(links: List[LinkResult], query: str, use_snippet=True, k1=1.5, b=0.75) -> List[LinkResult]:
    """
    对 LinkResult 列表进行 BM25 排序，并在对象中记录 score 和 rank
    
    Args:
        links: LinkResult 对象列表
        query: 查询文本
        use_snippet: 是否使用 snippet 进行排序，False 则使用 title
        k1, b: BM25 参数
    
    Returns:
        排序后的 LinkResult 列表，每个对象的 extra_fields 中包含 'bm25_score' 和 'bm25_rank'
    """
    # 提取文本
    corpus = []
    for link in links:
        if use_snippet and link.snippet:
            corpus.append(link.snippet)
        else:
            corpus.append(link.title)
    
    # 预处理
    docs_tokens, doc_len, tf_list, df, avgdl = preprocess_corpus(corpus)
    idf = compute_idf(df, len(corpus))
    
    # 计算 BM25 得分
    scores = [bm25_score(query, i, docs_tokens, tf_list, doc_len, avgdl, idf, k1, b)
              for i in range(len(corpus))]
    
    # 排序索引
    ranked_idx = sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)
    
    # 写入 score 和 rank
    for rank, idx in enumerate(ranked_idx, start=1):
        links[idx].extra_fields['bm25_score'] = scores[idx]
        links[idx].extra_fields['bm25_rank'] = rank
    
    # 返回排序后的列表
    return [links[i] for i in ranked_idx]

if __name__ == "__main__":
    # 示例语料库
    sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    corpus = [
        "生成式AI正在改变世界",
        "人工智能的发展非常迅速",
        "机器学习是AI的核心技术",
        "AI 技术在医疗领域的应用越来越广泛",
        "生成式模型可以生成高质量文本和图像"
    ]

    # 查询文本
    query = "生成式 AI"

    # 调用 BM25 排序接口
    sorted_corpus = bm25_rank_tool(corpus, query)

    # 输出排序结果
    print("BM25 排序结果（文本 -> 得分）:")
    for text, score in sorted_corpus:
        print(f"{text} --> {score:.4f}")
