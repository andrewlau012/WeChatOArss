# API 接口规范 (API Specification)

本接口文档参考 RESTful 规范，基础路径为 `/api/v1`。

## 1. 认证与账号 (Auth)

### 获取登录二维码
- **GET** `/auth/qrcode`
- **Response**:
  ```json
  {
    "uuid": "xxxxx",
    "qrcode_url": "https://weread.qq.com/login/...",
    "expire_seconds": 300
  }
  ```

### 检查登录状态
- **GET** `/auth/status?uuid=xxxxx`
- **Response**:
  ```json
  {
    "status": "waiting", // waiting, scanned, confirmed, expired
    "user_info": { // 仅 confirmed 状态返回
      "vid": "123456",
      "name": "UserA"
    }
  }
  ```

### 获取账号列表
- **GET** `/auth/accounts`
- **Response**:
  ```json
  [
    {
      "vid": "123456",
      "name": "UserA",
      "status": "active", // active (正常), blocked (风控/限制), expired (失效)
      "last_check": "2023-10-27T10:00:00Z" // 最后一次状态检测时间
    }
  ]
  ```

## 2. 公众号源 (Feeds)

### 添加公众号 (搜索/链接)
- **POST** `/feeds/add`
- **Body**:
  ```json
  {
    "type": "search_select", // 或 "link_parse"
    "value": "机器之心", // 关键词 或 文章URL
    "biz_id": "MzI..." // search_select 时必填
  }
  ```
- **Response**: `{"id": "biz_...", "name": "机器之心", "status": "added"}`

### 搜索公众号 (辅助接口)
- **GET** `/feeds/search?keyword=机器之心`
- **Response**:
  ```json
  [
    {
      "biz_id": "MzI...",
      "name": "机器之心",
      "head_img": "http://..."
    }
  ]
  ```

### 获取关注列表
- **GET** `/feeds/list`
- **Parameters**:
  - `page`: 1
  - `size`: 20
- **Response**:
  ```json
  {
    "total": 305,
    "items": [
      {
        "id": "biz_MzI...",
        "name": "机器之心",
        "unread_count": 2,
        "latest_update": "2023-10-27T08:00:00Z"
      }
    ]
  }
  ```

### 强制刷新/触发抓取
- **POST** `/feeds/refresh`
- **Body**: `{"target_id": "all"}` (或指定biz_id)
- **Response**: `{"task_id": "job_123"}`

## 3. RSS 输出 (Public)

### 获取特定公众号 RSS
- **GET** `/rss/{biz_id}.xml`
- **Parameters**:
  - `limit`: 20 (默认20，最大可配)
- **Description**: 返回标准 XML 格式 RSS 2.0 数据。
- **Content-Type**: `application/xml`

### 获取聚合 RSS
- **GET** `/rss/all.xml`
- **Description**: 返回所有订阅源的最新文章聚合。

### 获取OPML导出
- **GET** `/rss/opml`
- **Description**: 导出订阅列表，供其他阅读器导入。

## 4. 文章 (Articles)

### 获取文章详情 (API模式)
- **GET** `/articles/{article_id}`
- **Response**:
  ```json
  {
    "title": "DeepMind发布新模型...",
    "content_html": "<div>...</div>", // 经过清洗的HTML
    "url": "http://mp.weixin.qq.com/s/...",
    "pub_date": "2023-10-27T08:00:00Z",
    "author": "机器之心"
  }
  ```

## 5. 系统 (System)

### 获取系统状态
- **GET** `/system/status`
- **Response**:
  ```json
  {
    "scheduler": "running",
    "next_run": "2023-10-27T12:30:00Z",
    "account_count": 2,
    "feed_count": 305
  }
  ```
