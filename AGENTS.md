# Miniflux Fork — AGENTS.md

## Fork 新增模块（非上游）

### 1. AI 摘要系统 (`internal/integration/ai/`)
- OpenAI-compatible API，用户级配置（integrations 表）
- 去重：仅对新 entry 调用，已有 `ai_summary` 的跳过
- DB: migration #73, entries 表 `ai_summary/ai_score/ai_summarized_at`, integrations 表 `ai_enabled/ai_provider_url/ai_api_key/ai_model`
- UI: entry 详情页摘要折叠、列表页评分 badge、`/ai-digest` 页面

### 2. Web Scraper 引擎 (`internal/reader/webscraper/`)
- 替代 RSS 解析：CSS 选择器 (goquery) + JSON gjson 路径提取
- `feed_source_type='web_scraper'` 时跳过 RSS 解析，直接抓取网页
- 分页：HTML 用正则提取下一页 URL，JSON 用 gjson path
- DB: migration #74, feeds 表 `feed_source_type/ws_*` 字段
- 订阅时 web_scraper 类型跳过 subscription discovery，直接创建 feed

### 3. Pinchtab JS 渲染 (`internal/reader/pinchtab/`)
- 系统级配置：`PINCHTAB_ENABLED/PINCHTAB_URL/PINCHTAB_BINARY_PATH`
- 子进程生命周期：`daemon.go` 启动时 `StartIfEnabled()`，关闭时 `Stop()`
- 每次渲染创建独立 instance（Chrome 进程），支持并发
- 在 `processor.go` 中 `UseJSRender && PinchTabEnabled` 时优先用 pinchtab，失败 fallback 到内置 scraper

## CI/CD
- `test.yml`: push main/tags + PR → go vet + build + unit test + integration test (PostgreSQL 17)
- `release.yml`: Test 通过 + tag `2.*` → per-arch native runner Docker 构建 (amd64/arm64) → merge manifest → GitHub Release
- 镜像推送到 `ghcr.io/naiba-forks/miniflux`

## 关键约定
- 参数编号：`storage/feed.go` UpdateFeed 用到 $1-$49, CreateFeed 用到 $1-$39
- 参数编号：`storage/integration.go` UpdateIntegration AI 字段在 $122-$126
- gjson 依赖：`github.com/tidwall/gjson`（新增，goquery 是上游已有）
