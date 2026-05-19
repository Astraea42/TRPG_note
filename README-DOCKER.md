# Docker 部署指南

## 前提条件

- Docker（推荐 Docker Desktop 24+）
- 或 Podman 等其他容器运行时

## 构建镜像

```bash
# 在项目根目录执行
docker build -t trpg-note:latest .
```

## 启动容器

### 使用 docker-compose（推荐）

```bash
docker compose up -d
```

启动后访问：http://localhost:8080

### 使用 docker run

```bash
docker run -d \
  --name trpg-note \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  trpg-note:latest
```

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BTR_PORT` | `8080` | 服务端口 |
| `BTR_DB_PATH` | `/app/data/storage.db` | SQLite 数据库路径 |

## 数据持久化

数据存储在 `/app/data` 目录中，通过 volume 映射到宿主机。包含：
- `storage.db` — SQLite 数据库
- `graph_assets/` — 上传的图片资源

## 使用官方镜像

本项目也发布在 GitHub Container Registry：

```bash
docker run -d \
  --name trpg-note \
  -p 8080:8080 \
  -v ./trpg-note-data:/app/data \
  ghcr.io/你的GitHub用户名/trpg_note:latest
```
