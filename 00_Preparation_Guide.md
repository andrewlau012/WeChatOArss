# 前期准备与部署说明 (Preparation Guide)

## 1. 环境准备
在开始部署前，请确保您拥有以下资源：

### 1.1 硬件/服务器
- **配置**：最低 1核 CPU, 1GB 内存 (推荐 2核 2GB 以保证全文抓取时的处理性能)。
- **网络**：能够访问互联网。如果服务器在海外，访问微信接口可能较慢但通常不封锁；国内服务器速度更佳，但需注意IP风控。
- **存储**：至少 10GB 磁盘空间（用于数据库和日志，如果大量存储HTML快照建议更多）。

### 1.2 软件依赖
- **Docker** & **Docker Compose** (核心依赖)
- **Git** (用于拉取代码)

### 1.3 微信账号
- 准备 **2个** 正常使用的个人微信号。
- 建议这两个账号已经关注了目标公众号（可通过微信“通讯录-公众号”查看）。
- **注意**：不要使用新注册的小号（极易封号），建议使用注册半年以上、有正常社交行为的账号。

## 2. 部署步骤 (Docker Compose)

### 步骤 1: 创建目录与配置文件
在服务器上创建项目目录：
```bash
mkdir -p wechat-rss/data
cd wechat-rss
```

创建 `docker-compose.yml` 文件：
```yaml
version: '3.8'

services:
  backend:
    image: ghcr.io/your-repo/wechat-rss-backend:latest
    container_name: wechat_rss_backend
    restart: always
    volumes:
      - ./data:/app/data
    environment:
      - DB_URL=sqlite:////app/data/db.sqlite
      - CRON_JOBS=15 8 * * *,30 12 * * *
      - FETCH_DELAY=10  # 抓取间隔(秒)
      - MAX_THREADS=2
    ports:
      - "8000:8000"

  frontend:
    image: ghcr.io/your-repo/wechat-rss-frontend:latest
    container_name: wechat_rss_frontend
    restart: always
    ports:
      - "8080:80"
    depends_on:
      - backend
```

### 步骤 2: 启动服务
```bash
docker-compose up -d
```

### 步骤 3: 初始化配置
1. 访问 `http://your-server-ip:8080`。
2. 进入“账号管理”，点击“添加账号”。
3. 使用微信扫描屏幕上的二维码。
4. 确认登录成功。
5. 等待系统首次同步公众号列表（约1-5分钟）。

## 3. 维护说明
- **Cookie有效期**：微信读书Cookie通常有效期较长（数周至数月），但仍需定期检查。系统会在Cookie失效时在仪表盘提示。
- **数据备份**：定期备份 `./data/db.sqlite` 文件。
- **日志查看**：
  ```bash
  docker logs -f wechat_rss_backend
  ```

## 4. 常见问题 (FAQ)
- **Q: 为什么生成的RSS没有图片？**
  A: 微信图片有防盗链保护。请在阅读器中配置Referer修改插件，或在系统设置中开启“图片代理”功能。
- **Q: 抓取速度太慢？**
  A: 为保护账号安全，我们强制限制了抓取速率。请勿随意调低间隔，否则可能导致IP或账号被封禁。
