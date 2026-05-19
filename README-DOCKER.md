# 部署指南

## 方式一：Hugging Face Spaces（免费，推荐）

此方式适合团队共享，无需自购服务器。

### 1. 创建 Dataset（用于数据持久化）

1. 打开 https://huggingface.co/new-dataset
2. 填写：
   - **Dataset Name**: `trpg-note-data`
   - **Type**: 勾选 `Public`
3. 点击 **Create dataset**

### 2. 创建 Space

1. 打开 https://huggingface.co/new-space
2. 填写：
   - **Space Name**: `trpg-note`
   - **Space SDK**: 选择 **Docker**
   - **Docker Template**: 选择 **Blank**
   - **Space Hardware**: 选择 **CPU basic (free)**
   - 勾选 `Public`
3. 点击 **Create space**

### 3. 配置 Variable 和 Token

在 Space 的 **Settings** → **Repository Secrets** 中：

1. **Variables**（新增）：
   - `BTR_PORT` = `7860`
   - `BTR_DB_PATH` = `/app/data/storage.db`
   - `HF_DATASET_REPO` = `Astraea42/trpg-note-data`（你刚创建的dataset完整名字）

2. **Secrets**（新增）：
   - `HF_SYNC_TOKEN` = 你的 Hugging Face Token（在 https://huggingface.co/settings/tokens 创建，需要 `write` 权限）

### 4. 关联 GitHub 仓库

在 Space 的 **Settings** → **Repository secrets** 下方找到：
- **Git** → 填入 GitHub 仓库地址：`https://github.com/Astraea42/TRPG_note.git`
- 设置自动同步

或者你也可以直接在 Space 的 **Files** 页面新建 `Dockerfile` 并将本项目的内容复制进去，点击 **Commit**。

> 如果使用自动同步，需要保证你的 GitHub 仓库是 `public` 或 Hugging Face 可以访问的。

### 5. 访问

Space 构建完成后，访问：`https://Astraea42-trpg-note.hf.space`

数据会自动同步到 `Astraea42/trpg-note-data` Dataset，容器重启后数据不丢失。

---

## 方式二：Docker / 自托管

```bash
docker build -t trpg-note:latest .

docker run -d \
  --name trpg-note \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  trpg-note:latest
```

启动后访问：http://localhost:8080

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BTR_PORT` | `7860` | 服务端口（HF Spaces 默认 7860） |
| `BTR_DB_PATH` | `/app/data/storage.db` | SQLite 数据库路径 |
| `HF_DATASET_REPO` | — | Hugging Face Dataset 名称（启用自动持久化时需要） |
| `HF_SYNC_TOKEN` | — | Hugging Face Token，有 write 权限（也可用 `HF_TOKEN`） |
