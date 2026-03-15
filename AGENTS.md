# MinifluxNg — AGENTS.md

## Fork 新增模块（非上游）

### 1. AI 摘要系统 (`internal/integration/ai/`)
- 导航栏 AI 数据：所有 50+ UI handler 必须设置 `showAIDigest` 和 `countAIDigest`（sed 批量添加，后新增 handler 需手动加）
- 回填可停止：`StopBackfill()` 通过 `backfillStopSignals` sync.Map 通知 goroutine 退出
- `storage/entry.go`: `IsAIEnabled()` 直查 integrations 表，`CountUnreadAIDigestEntries()` 用于导航计数

### 2. Web Scraper 引擎 (`internal/reader/webscraper/`)
- `mergeURL`：相对 URL 解析必须用 `url.ResolveReference()`（RFC 3986），手写拼接会把 `archives.html/blog/x.html` 当目录

### 3. Lightpanda 无头浏览器 JS 渲染 (`internal/reader/headless/`)
- 架构：go-rod (CDP 客户端) + Lightpanda (Zig 编写, V8 引擎, 非 Chromium 轻量 headless browser)
- 两阶段内容提取：Lightpanda 渲染页面取 outerHTML → node 子进程 Defuddle (Readability 替代) 提取正文
  - Defuddle 不能直接在 Lightpanda 内运行（缺 `getComputedStyle` 等 API 会导致进程 crash）
  - Defuddle 在 Docker 构建时从 GitHub clone 即时 build，产物安装到 `/usr/share/miniflux/defuddle/`
  - Go 代码通过 `node -e` 内联脚本调用，30 秒超时，失败 fallback 到 `innerText`
- 资源回收：`activeProcessCount` 原子计数，`browser.Close()` 加 `recover()` 防 crash 后 panic
- 不 fallback：JS 渲染启用时，headless 失败不会 fallback 到 HTTP scraper（避免掩盖真实错误），与 `handler.go` 列表页逻辑一致
- RSS feed：`processor.go` 中 `UseJSRender && LightpandaEnabled` 时用 headless，失败直接跳过不 fallback
- Web Scraper feed：`handler.go` 中列表页渲染用 `RenderPageHTML` + `ScrapeRenderedHTML`；条目全文在 `processor.go` 中同样走 headless 不 fallback

## CI/CD
- `release.yml`: Test 通过 + tag `v*` → per-arch native runner Docker 构建 (amd64/arm64) → merge manifest → GitHub Release
- 镜像推送到 `ghcr.io/naiba-forks/miniflux`

## 关键约定
- **版本号**：fork 版本线 `v0.x`，发版前先 `git tag --sort=-v:refname -l 'v*' | head` 确认最新版本号再递增，创建 `v*` tag 触发 CI（上游 release 无 `v` 前缀，不冲突）
- **合并上游**：必须用 `git fetch upstream --no-tags && git merge upstream/main`，禁止拉取上游 tag 污染本地
- **CSP 限制**：`style-src 'nonce-xxx'` 会阻止 HTML 内联 `style=""` 属性，必须用 CSS class（如 `.initially-hidden`）控制初始可见性
- 参数编号：`storage/feed.go` UpdateFeed 用到 $1-$49, CreateFeed 用到 $1-$39
- 参数编号：`storage/integration.go` UpdateIntegration AI 字段在 $122-$125（$126 是 WHERE user_id）
- gjson 依赖：`github.com/tidwall/gjson`（新增，goquery 是上游已有）
