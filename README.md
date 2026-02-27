# WeChatOArss

微信公众号RSS服务 - 长期稳定可用的自部署方案

## 特性

- ✅ 微信扫码登录
- ✅ 公众号搜索订阅
- ✅ 标准RSS输出（RSS 2.0 / JSON Feed）
- ✅ Web阅读界面
- ✅ Docker一键部署
- ✅ 数据本地存储

## 快速开始

### 1. 下载项目

```bash
git clone https://github.com/yourrepo/WeChatOArss.git
cd WeChatOArss
```

### 2. 配置环境

复制配置文件并修改 Token：

```bash
cp config/config.yaml config.yaml
# 编辑 config.yaml，修改 RSS_TOKEN
```

### 3. 启动服务

```bash
docker-compose up -d
```

### 4. 访问服务

- Web界面: http://localhost:8080
- API版本: http://localhost:8080/version

## 使用说明

### 登录

首次访问会显示微信登录二维码，请使用微信扫码登录。

### 添加公众号

1. 进入"公众号"页面
2. 点击"添加公众号"
3. 可以通过文章链接或名称搜索添加

### RSS订阅

每个公众号都有独立的RSS地址：
- XML格式: `/feed/{biz_id}.xml`
- JSON格式: `/feed/{biz_id}.json`

全量订阅：
- `/feed/all.xml`
- `/feed/all.json`

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| RSS_HOST | 服务地址（用于生成RSS链接） | localhost:8080 |
| RSS_TOKEN | API访问密码 | - |
| SCHEDULER_TIMES | 定时抓取时间 | 07:00,12:00,20:00 |
| RSS_MAX_ITEM_COUNT | RSS最大文章数 | 20 |

## 目录结构

```
WeChatOArss/
├── cmd/server/          # Go后端入口
├── internal/            # 内部包
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP处理器
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   └── store/           # 数据库
├── web/                 # Vue3前端
│   ├── src/
│   │   ├── views/      # 页面组件
│   │   ├── router/      # 路由
│   │   └── assets/      # 静态资源
│   └── dist/            # 构建输出
├── config/             # 配置文件
├── docker/             # Docker配置
└── docker-compose.yml  # Docker编排
```

## 技术栈

- **后端**: Go + Gin
- **数据库**: SQLite
- **前端**: Vue 3 + Vite
- **部署**: Docker + Docker Compose

## 许可证

MIT License
