# MCP Web Search Server

MCP网页搜索服务器，支持多源搜索引擎。

## 功能特性

- 🔍 **智能搜索**: 通过多个搜索引擎获取相关链接
- 🌐 **网页爬取**: 自动爬取网页内容并进行文本切块处理
- 🤖 **MCP 集成**: 提供标准 MCP 工具接口，可与各种 AI 助手集成
- ⚡ **高性能**: 支持并发处理和批量操作
- 🔧 **Pipeline 架构**: 可复用的处理管道，支持灵活扩展
- 🎯 **智能选择**: 根据 URL 自动选择合适的处理 pipeline

## Pipeline 架构

### 核心设计

项目采用 **Pipeline 模式** 处理不同类型的网页内容，通过统一的接口规范实现高度复用和灵活扩展：

```go
// Pipeline 接口定义
type Pipeline interface {
    Process(Type) (Type, error)  // 处理方法
    Match(url string) bool       // URL 匹配方法
}
```

### 架构优势

1. **组件化设计**：每个 pipeline 都是独立的处理单元
2. **可复用性**：爬虫、清洗、分块组件可灵活组合
3. **易于扩展**：新增 pipeline 只需实现统一接口
4. **智能选择**：根据 URL 自动匹配合适的处理 pipeline

### 内置 Pipeline

| Pipeline | 优先级 | 适用场景 |
|----------|--------|----------|
| PDF Pipeline | 25 | PDF 文档处理 |
| Markdown Pipeline | 20 | Markdown 文件 |
| GitHub Pipeline | 15 | GitHub 仓库和文件 |
| Colly Pipeline | 10 | 通用网页爬取（默认） |

### Pipeline 复用机制

```go
// 示例：组合不同组件创建自定义 pipeline
type CustomPipeline struct {
    Crawler types.Crawler  // 爬虫组件
    Cleaner types.Cleaner  // 清洗组件
    Chunker types.Chunker  // 分块组件
}

func NewCustomPipeline() *CustomPipeline {
    return &CustomPipeline{
        Crawer: NewCustomCrawler(),
        Cleaner: NewAdvancedCleaner(),
        Chunker: NewScoredChunker(0.8),
    }
}
```

### 扩展新的 Pipeline

```go
// 1. 实现 Pipeline 接口
type MyPipeline struct{}

func (p *MyPipeline) Process(input types.Type) (types.Type, error) {
    // 实现处理逻辑
    return result, nil
}

func (p *MyPipeline) Match(url string) bool {
    // 实现 URL 匹配逻辑
    return strings.HasSuffix(url, ".myext")
}

// 2. 注册到系统
func init() {
    core.RegisterPipeline(30, "myext", NewMyPipeline())
}
```

## 项目结构

```
mcp-web-search-server/
├── server.py              # MCP 服务器主文件 (端口: 8006)
├── client_test.py         # 客户端测试文件
├── config.yaml           # 统一配置文件
├── config.yaml.example   # 配置文件模板
├── requirements.txt      # Python 依赖列表
├── start.sh              # 一键启动所有服务
├── stop.sh               # 一键停止所有服务
├── context_crawl/         # 网页爬取服务 (Go)
│   ├── main.go           # 爬取服务入口
│   ├── app/              # HTTP 路由处理
│   │   ├── route.go      # 路由注册
│   │   └── v1/
│   │       └── api.go    # API 端点定义
│   ├── core/             # Pipeline 核心管理
│   │   ├── register.go   # Pipeline 注册机制
│   │   └── choice.go     # Pipeline 选择器
│   ├── base/             # 基础 Pipeline 实现
│   │   └── colly/        # 通用 Colly Pipeline
│   │       ├── crawl.go  # 爬虫组件
│   │       ├── clean.go  # 清洗组件
│   │       ├── chunk.go  # 分块组件
│   │       └── pipeline.go
│   ├── custom/           # 专用 Pipeline 实现
│   │   ├── github/       # GitHub 专用 Pipeline
│   │   ├── md/           # Markdown 专用 Pipeline
│   │   └── pdf/          # PDF 专用 Pipeline
│   ├── types/            # 类型定义和接口
│   │   ├── pipeline.go   # Pipeline 接口定义
│   │   ├── crawler.go    # 爬虫接口
│   │   ├── cleaner.go    # 清洗接口
│   │   └── chunker.go    # 分块接口
│   ├── handler/          # 请求处理
│   │   ├── url_handler.go
│   │   └── models/
│   │       ├── request.go
│   │       └── response.go
│   ├── service/          # 业务逻辑
│   │   └── url_service.go
│   └── utils/            # 工具函数
│       └── config.go
└── links_search/         # 链接搜索服务 (Python)
    ├── main.py          # 搜索服务入口  
    ├── utils/           # Python 工具模块
    │   ├── config.py    # Python 配置工具
    │   └── bm25.py      # BM25 排序算法
    ├── source/          # 搜索引擎实现
    │   ├── base.py      # 搜索源基类
    │   └── ...          # 各搜索引擎实现
    └── ...             # 其他 Python 模块文件
```

## 快速开始

### 前置要求

- Python 3.8+
- Go 1.18+ (用于 context_crawl 服务)

### 安装依赖

```bash
# 安装 Python 依赖
pip install -r requirements.txt

# 安装 Go 依赖 (在 context_crawl 目录下)
cd context_crawl
go mod tidy
cd ..
```

### 启动服务

#### 方式一：一键启动所有服务 (推荐)

```bash
./start.sh
```

此脚本会自动启动以下服务：
- 链接搜索服务 (端口: 8004)
- 网页爬取服务 (端口: 8003)
- MCP 服务器 (端口: 8006)

#### 方式二：分别启动服务

1. **启动链接搜索服务** (端口: 8004)
   ```bash
   cd links_search
   python main.py
   ```

2. **启动网页爬取服务** (端口: 8003)
   ```bash
   cd context_crawl
   go run main.go
   ```

3. **启动 MCP 服务器** (端口: 8006)
   ```bash
   python server.py
   ```

### 停止服务

```bash
./stop.sh
```

此脚本会停止所有已启动的服务。

### 测试服务

运行测试客户端验证服务是否正常工作:
```bash
python client_test.py
```

## MCP 工具说明

### get_links 工具

获取与查询相关的搜索链接列表。

**参数:**
- `query` (str): 搜索关键词，例如 "python 爬虫 教程"
- `count` (int, 可选): 需要返回的链接数量，默认 5

**返回值:**
格式化后的链接信息，包含 URL、标题和摘要。

**示例:**
```python
await session.call_tool("get_links", {
    "query": "中国 人工智能 行业报告",
    "count": 3
})
```

### get_page_content 工具

爬取指定 URL 的网页完整内容。

**参数:**
- `urls` (List[str]): URL 列表，例如 ["https://www.example.com"]

**返回值:**
网页的完整文本内容，包含分块处理和相关性评分。

**示例:**
```python
await session.call_tool("get_page_content", {
    "urls": ["https://www.yicai.com/news/102246615.html"]
})
```

## API 接口

### 链接搜索服务 (端口: 8004)

**POST /get_links**
```json
{
  "query": "搜索关键词",
  "count": 5
}
```

### 网页爬取服务 (端口: 8003)

**POST /crawl**
```json
{
  "urls": ["https://example.com/page1", "https://example.com/page2"]
}
```

## 配置说明

### 配置文件设置

**安全提醒：** `config.yaml` 文件包含敏感信息（如 API keys），请勿提交到版本控制系统！

1. **从模板创建配置文件：**
   ```bash
   cp config.yaml.example config.yaml
   ```

2. **编辑配置文件：**
   ```bash
   vim config.yaml
   ```

3. **配置内容说明：**

所有服务的配置都集中在项目根目录的 `config.yaml` 文件中：

```yaml
project:
  name: "mcp_web_search"
  description: "MCP Web Search Server - 整合链接搜索和网页爬取服务"
  version: "1.0.0"

# MCP 服务配置
mcp:
  host: 0.0.0.0
  port: 8006

# 链接搜索服务配置
links_search:
  host: 0.0.0.0
  port: 8004
  sources:
    bocha:
      enabled: true
      url: "https://api.bochaai.com/v1/web-search"
      api_key: "your-api-key-here"
    mita:
      enabled: true
      url: "https://metaso.cn/api/v1/search"
      api_key: "your-api-key-here"
    duckgo:
      enabled: false

# 网页爬取服务配置
context_crawl:
  host: 0.0.0.0
  port: 8003
```

### 配置工具

- **Python 配置工具**: `links_search/utils/config.py`
- **Go 配置工具**: `context_crawl/utils/config.go`

这些工具负责加载和解析统一配置文件，为各自服务提供配置访问接口。

## 开发说明

### Pipeline 开发

#### 1. 定义 Pipeline 接口

在 `context_crawl/types/` 目录下定义新的接口（如果需要）：

```go
// types/my_feature.go
package types

// MyFeaturePipeline 处理特定类型的内容
type MyFeaturePipeline interface {
    Pipeline // 嵌入 Pipeline 接口
    
    // 添加特定方法
    Configure(options map[string]interface{}) error
}
```

#### 2. 实现 Pipeline

在 `context_crawl/custom/` 或 `context_crawl/base/` 目录下创建新的 pipeline：

```go
// custom/myfeature/pipeline.go
package myfeature

import (
    "context_crawl/types"
)

type MyFeaturePipeline struct {
    Crawler types.Crawler
    Cleaner types.Cleaner
    Chunker types.Chunker
}

func NewMyFeaturePipeline() *MyFeaturePipeline {
    return &MyFeaturePipeline{
        Crawler: NewMyFeatureCrawler(),
        Cleaner: NewBasicCleaner(),
        Chunker: NewScoredChunker(0.5),
    }
}

func (p *MyFeaturePipeline) Process(input types.Type) (types.Type, error) {
    // 实现处理逻辑
    pageResult, err := p.Crawler.Crawl(input)
    if err != nil {
        return types.Type{}, err
    }
    
    cleanResult, err := p.Cleaner.Clean(pageResult)
    if err != nil {
        return types.Type{}, err
    }
    
    chunkResult, err := p.Chunker.Chunk(cleanResult)
    if err != nil {
        return types.Type{}, err
    }
    
    return chunkResult, nil
}

func (p *MyFeaturePipeline) Match(url string) bool {
    // 实现 URL 匹配逻辑
    return strings.Contains(url, "myfeature.com")
}
```

#### 3. 注册 Pipeline

在 `context_crawl/core/register.go` 中注册新的 pipeline：

```go
// 4️⃣ 注册 MyFeature pipeline
RegisterPipeline(30, "myfeature", myfeature.NewMyFeaturePipeline())
```

#### 4. 组件复用

复用现有的组件来构建新的 pipeline：

```go
type CustomPipeline struct {
    // 复用现有的爬虫组件
    Crawler types.Crawler
    // 复用现有的清洗组件
    Cleaner types.Cleaner
    // 复用现有的分块组件
    Chunker types.Chunker
}
```

### 添加新的搜索引擎

在 `links_search/source/` 目录下添加新的搜索引擎模块，继承自 `base.py` 中定义的 `SearchSource` 基类，实现统一的接口规范。

```python
from source.base import SearchSource

class NewSearchSource(SearchSource):
    async def get_links(self, query: str, **kwargs) -> Dict[str, Any]:
        # 实现获取搜索链接的逻辑
        pass
    
    def parse_result(self, raw_data: Any) -> List:
        # 实现解析搜索结果的逻辑
        pass
```


### 扩展 MCP 工具

在 `server.py` 中添加新的 `@mcp.tool()` 装饰器函数来提供更多功能。

## 故障排除

### 端口冲突
如果出现端口被占用错误，请修改对应服务的 config.yaml 文件中的端口号。

### 依赖问题
确保所有必要的依赖都已正确安装:
```bash
# Python 依赖
pip install -r requirements.txt

# Go 依赖 (在 context_crawl 目录下)
go mod tidy
```

### 连接问题
确保三个服务都正常启动，并且能够相互访问。

## 许可证

MIT License


## 后续优化
- 增加网页图片抓取、解析
- 增加文件内容抓取、解析