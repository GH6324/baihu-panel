# 命令行工具 (CLI)

白虎面板在环境内内置了同名的 `baihu` 命令行工具。如果您在终端内需要执行系统级别的操作，可以使用这些内置命令。

## 常用核心指令

| 命令 | 描述 |
| :--- | :--- |
| `baihu server` | 面板启动指令，运行服务端后台进程。 |
| `baihu reposync` | 供定时任务调用，将远程 Git 仓库的高级特性同步到本地目录中。 |
| `baihu resetpwd` | 交互式重置系统 admin 账号密码（密码丢失时可通过进入终端重置）。 |
| `baihu restore <file>` | 使用本地的 .zip 备份压缩包文件，一条命令直接全量恢复系统数据。 |

---

## 使用场景示例

### 1. 密码重置
您可以进入 Docker 容器或通过 ssh 连入宿主机控制台：
```bash
docker exec -it baihu baihu resetpwd
```
然后根据提示，输入新的管理员密码即可重置成功。

### 2. 手动启动
如果是通过手动部署二进制文件，可以使用 `baihu server` 启动：
```bash
nohup ./baihu server > /dev/null 2>&1 &
```

### 3. 数据恢复
上传备份后的 ZIP 文件至容器目录：
```bash
docker exec -it baihu baihu restore /app/data/backup-2026xxxx.zip
```
该操作会全量覆盖现有数据库和脚本文件，请谨慎操作。

---

## `reposync` 参数详解

`baihu reposync` 是面板核心的同步命令，除了在任务中自动调用外，您也可以通过命令行手动执行。

### 参数列表

| 参数名 | 默认值 | 描述 |
| :--- | :--- | :--- |
| `--source-type` | `git` | 同步源类型，可选 `git`（Git 仓库）或 `url`（文件直链下载）。 |
| `--source-url` | | 同步源地址，Git 仓库地址或下载 URL。 |
| `--target-path` | | 目标保存路径。支持变量替换（如 `$SCRIPTS_DIR$`）。 |
| `--branch` | | Git 分支名。留空时将自动检测远程默认分支（如 `main` 或 `master`）。 |
| `--path` | | 稀疏检出（Sparse checkout）的指定路径，或在单文件模式下的相对路径。 |
| `--single-file` | `false` | 是否开启单文件模式，仅从 Git 提取指定单个文件。 |
| `--proxy` | `none` | Github 加速代理类型，可选 `none`、`ghproxy`、`mirror`、`custom`。 |
| `--proxy-url` | | 自定义代理地址，仅在 `--proxy=custom` 时生效。 |
| `--auth-token` | | 私有仓库或 API 访问使用的鉴权 Token。 |
| `--http-proxy` | | HTTP/HTTPS 代理地址，例如 `http://127.0.0.1:7890`。 |
| `--whitelist-paths`| | 白名单路径（逗号或竖线分隔），同步时受保护不被清理的路径。 |
| `--blacklist` | | 黑名单关键字（竖线 `\|` 分隔），包含该关键字的文件将会被过滤删除。 |
| `--dependence` | | 依赖文件关键字（竖线 `\|` 分隔），这些文件将强制保留。 |
| `--extensions` | | 允许的脚本扩展名（竖线 `\|` 分隔，如 `.js\|.py`），后缀不符的文件将被删除。 |
| `--task-id` | | 内部任务 ID，用于在同步完成后通知调度器刷新增量任务。 |
| `--task-langs` | | 任务配置的语言（JSON格式），用于标记和解析。 |
| `--repo-task-id` | | 原始任务 ID。 |
| `--task-timeout` | `30` | 同步任务的超时时间，单位为分钟。 |
| `--commenttotask` | `false` | 是否启用青龙 (QL) 格式的脚本注释解析（`true`/`false`）。 |

### 使用示例

#### 1. 基础 Git 仓库同步
将指定仓库克隆或拉取到特定目录：
```bash
baihu reposync --source-url https://github.com/example/repo.git --target-path /app/data/scripts/example_repo
```

#### 2. 启用代理的同步
针对 Github 仓库使用加速代理，并限定只保留 `.js` 和 `.py` 脚本：
```bash
baihu reposync --source-url https://github.com/example/repo.git \
  --target-path /app/data/scripts/example_repo \
  --proxy ghproxy \
  --extensions ".js|.py"
```

#### 3. 稀疏检出 (Sparse Checkout)
当仓库庞大时，仅同步特定的子目录或文件：
```bash
baihu reposync --source-url https://github.com/example/repo.git \
  --target-path /app/data/scripts/example_repo \
  --path "scripts/daily"
```

#### 4. 单文件下载模式
如果只需要仓库中的某一个脚本文件：
```bash
baihu reposync --source-url https://github.com/example/repo.git \
  --target-path /app/data/scripts/ \
  --single-file true \
  --path "main_script.py"
```

#### 5. 高级过滤与青龙注释解析
使用黑名单排除特定脚本，并开启青龙格式注释解析以自动生成定时任务：
```bash
baihu reposync --source-url https://github.com/example/repo.git \
  --target-path /app/data/scripts/example_repo \
  --blacklist "test|mock" \
  --dependence "package.json|requirements.txt" \
  --commenttotask "true"
```

---

## 其他帮助
终端内直接执行 `baihu` 即可在控制台直接打印内置支持详细说明和命令列表参数。
