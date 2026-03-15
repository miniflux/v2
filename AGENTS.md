# MinifluxNg — AGENTS.md

## Fork 新增模块（非上游）

### 1. AI 摘要系统 (`internal/integration/ai/`)
- OpenAI-compatible API，用户级配置（integrations 表）
- 去重：仅对新 entry 调用，已有 `ai_summary` 的跳过
- DB: migration #73, entries 表 `ai_summary/ai_score/ai_summarized_at`, integrations 表 `ai_enabled/ai_provider_url/ai_api_key/ai_model`
- UI: entry 详情页摘要折叠、列表页评分 badge、`/ai-digest` 分页页面
- 导航栏仅在 AI 启用时显示 AI Digest 菜单 + 未读计数（`showAIDigest`/`countAIDigest` 在 50+ handler 中设置）
- AI Digest 页面顶部：一键生成本页整体总结 → 一键标记已读（`GeneratePageSummary` + JS fetch）
- 语言感知：`buildSystemPrompt(language)` 根据 `user.Language` 生成对应语言摘要
- 回填：`BackfillAISummaries` + `ForceBackfillAISummaries`，batch 50，`maxConsecutiveErrors=3`
- 回填可停止：`StopBackfill()` 通过 `backfillStopSignals` sync.Map 通知 goroutine 退出
- 回填按钮状态同步：`GET /ai-backfill-status` + JS 轮询 3s + 终止按钮
- `storage/entry.go`: `IsAIEnabled()` 直查 integrations 表，`CountUnreadAIDigestEntries()` 用于导航计数

### 2. Web Scraper 引擎 (`internal/reader/webscraper/`)
- 替代 RSS 解析：CSS 选择器 (goquery) + JSON gjson 路径提取
- `feed_source_type='web_scraper'` 时跳过 RSS 解析，直接抓取网页
- 分页：HTML 用正则提取下一页 URL，JSON 用 gjson path
- DB: migration #74, feeds 表 `feed_source_type/ws_*` 字段
- 订阅时 web_scraper 类型跳过 subscription discovery，直接创建 feed

### 3. Lightpanda 无头浏览器 JS 渲染 (`internal/reader/headless/`)
- 系统级配置：`LIGHTPANDA_ENABLED/LIGHTPANDA_BINARY_PATH`
- 架构：go-rod (CDP 客户端) + Lightpanda (Zig 编写, V8 引擎, 非 Chromium 轻量 headless browser)
- 两阶段内容提取：Lightpanda 渲染页面取 outerHTML → node 子进程 Defuddle (Readability 替代) 提取正文
  - Defuddle 不能直接在 Lightpanda 内运行（缺 `getComputedStyle` 等 API 会导致进程 crash）
  - Defuddle 在 Docker 构建时从 GitHub clone 即时 build，产物安装到 `/usr/share/miniflux/defuddle/`
  - Go 代码通过 `node -e` 内联脚本调用，30 秒超时，失败 fallback 到 `innerText`
- 子进程生命周期：每次渲染启动临时 Lightpanda `serve` 进程 → CDP WebSocket 连接 → SIGTERM 关闭，3 秒无响应 SIGKILL
- 资源回收：`activeProcessCount` 原子计数，`browser.Close()` 加 `recover()` 防 crash 后 panic
- 代理：`proxyURL` 非空时通过 `--http_proxy` 参数传给 Lightpanda
- RSS feed：`processor.go` 中 `UseJSRender && LightpandaEnabled` 时优先用 headless，失败 fallback 到内置 scraper
- Web Scraper feed：`handler.go` 中列表页渲染用 `RenderPageHTML` + `ScrapeRenderedHTML`
- 依赖：`github.com/go-rod/rod`（CDP 客户端），Docker 内 `nodejs`（运行 Defuddle 提取）
- UI：`UseJSRender` checkbox 在 Network Settings 区域（紧跟 Crawler），对 RSS 和 web_scraper 都生效

## CI/CD
- `tests.yml`: push main/tags + PR → go vet + build + unit test + integration test (PostgreSQL 17)
- `release.yml`: Test 通过 + tag `v*` → per-arch native runner Docker 构建 (amd64/arm64) → merge manifest → GitHub Release
- 镜像推送到 `ghcr.io/naiba-forks/miniflux`

## 关键约定
- **版本号**：fork 版本线 `v0.x`，发版前先 `git tag --sort=-v:refname -l 'v*' | head` 确认最新版本号再递增，创建 `v*` tag 触发 CI（上游 release 无 `v` 前缀，不冲突）
- **合并上游**：必须用 `git fetch upstream --no-tags && git merge upstream/main`，禁止拉取上游 tag 污染本地
- **CSP 限制**：`style-src 'nonce-xxx'` 会阻止 HTML 内联 `style=""` 属性，必须用 CSS class（如 `.initially-hidden`）控制初始可见性
- 参数编号：`storage/feed.go` UpdateFeed 用到 $1-$49, CreateFeed 用到 $1-$39
- 参数编号：`storage/integration.go` UpdateIntegration AI 字段在 $122-$125（$126 是 WHERE user_id）
- gjson 依赖：`github.com/tidwall/gjson`（新增，goquery 是上游已有）
- 导航栏 AI 数据：所有 50+ UI handler 必须设置 `showAIDigest` 和 `countAIDigest`（sed 批量添加，后新增 handler 需手动加）
