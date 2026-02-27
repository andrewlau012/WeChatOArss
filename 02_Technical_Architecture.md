# 微信公众号RSS服务 - 技术方案设计

## 1. 总体架构
采用 **前后端分离** 架构，后端负责核心业务逻辑与爬虫调度，前端提供可视化管理界面。系统设计优先考虑容器化部署。

```mermaid
graph TD
    User[用户/RSS阅读器] -->|HTTP/RSS| Nginx[反向代理/网关]
    Nginx --> Frontend[前端 (Vue3/React)]
    Nginx --> Backend[后端 (Python FastAPI)]
    
    subgraph "后端服务"
        Backend --> Scheduler[定时任务调度器]
        Scheduler --> TaskQueue[任务队列 (Memory/Redis)]
        TaskQueue --> Workers[爬虫Worker池]
        Workers -->|1. 获取列表| WeReadAPI[微信读书API]
        Workers -->|2. 获取全文| MP_Weixin[微信文章页面]
        
        Backend --> DB[(SQLite/PostgreSQL)]
        Backend --> Cache[缓存 (Local/Redis)]
    end
```

## 2. 技术栈选择

### 2.1 后端 (Backend)
- **语言**：Python 3.10+ (因爬虫库生态丰富)
- **框架**：FastAPI (高性能，自动生成OpenAPI文档)
- **爬虫/请求**：
  - `requests` / `httpx`：基础HTTP请求。
  - `playwright` (可选)：如果遇到强JS渲染或滑块验证，作为备选方案。
- **调度**：`APScheduler` (轻量级定时任务)。
- **RSS生成**：`rfeed` 或 `feedgen`。

### 2.2 前端 (Frontend)
- **框架**：Vue 3 + Vite
- **UI库**：Element Plus 或 Naive UI
- **主要页面**：仪表盘、账号管理（扫码）、公众号列表、系统设置。

### 2.3 数据存储
- **元数据**：SQLite (默认，单文件易备份)，支持切换PostgreSQL。
- **文件存储**：本地文件系统 (存储HTML快照，可选)。

## 3. 核心流程设计

### 3.1 登录流程 (Cookie注入/扫码)
1.  后端调用微信读书登录接口，获取二维码URL。
2.  前端展示二维码。
3.  用户微信扫码确认。
4.  后端轮询/回调获取 `wr_vid`, `wr_skey` 等关键Cookie。
5.  验证Cookie有效性，存入数据库。

### 3.2 增量抓取流程 (Task Flow)
1.  **触发**：Cron触发 (08:15, 12:30)。
2.  **获取列表**：使用轮换账号的Cookie，调用微信读书接口获取“书架/关注”列表更新。
3.  **解析文章**：
    - 遍历新文章列表。
    - 检查DB是否存在（去重）。
    - 访问文章URL（需通过代理或控制速率，模拟正常浏览器UA）。
    - 提取 `#js_content` 内容，处理 `data-src` 图片懒加载属性。
4.  **数据清理**：
    - 每日任务结束后，执行清理逻辑：`DELETE FROM articles WHERE pub_date < date('now', '-60 days')`。
5.  **存储**：保存标题、摘要、作者、发布时间、HTML正文。
6.  **RSS生成**：
    - 动态生成：`/rss/{biz_id}.xml` 接口查询 DB，限制 `LIMIT 20`。

### 3.3 全文获取与反爬策略
- **账号状态检测**：
    - 定期（如每小时）调用轻量级接口（如获取书架信息），检查 HTTP 状态码及返回内容。
    - 若返回特定错误码（如 401 或业务错误码），标记账号为“风控/失效”，停止分配任务并告警。
- **并发控制**：300+公众号，假设每日更新500篇。
- **速率限制**：单IP建议 5-10秒/请求。

### 3.4 新增公众号流程
1.  **关键词搜索**：调用 WeRead 搜索接口 `search/books` 或 `search/mp`，返回匹配列表供用户选择。
2.  **链接解析**：
    - 用户输入文章 URL。
    - 后端请求该 URL。
    - 正则提取 `__biz` 参数及 `nickname` (公众号名称)。
    - 自动将该 `__biz` 添加到监控列表。
- **User-Agent轮换**：使用常见浏览器UA库。
- **图片防盗链**：
  - 方案A：前端渲染时使用 `https://images.weserv.nl/?url=` 等公共图片代理。
  - 方案B：后端下载图片并本地存储（占用空间大，不推荐）。
  - 方案C：后端提供 `/proxy/image?url=xxx` 接口，转发请求并修改Referer。

## 4. 部署方案
- **Docker Compose**：
  ```yaml
  version: '3'
  services:
    app:
      image: wechat-rss-app:latest
      volumes:
        - ./data:/app/data
      ports:
        - "4000:4000"
      environment:
        - CRON_EXPRESSION="15 8,30 12 * * *"
  ```
