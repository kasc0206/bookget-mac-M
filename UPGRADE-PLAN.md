# Bookget 升级计划

> 基于 2026-06-19 代码审查结果制定

---

## P0 — 紧急修复（建议立即执行）

### P0.1 修复 `go vet` nil 指针风险

**文件**: `pkg/gohttp/multithread.go`

**问题**: 第 112 行和第 246 行在检查 `err` 之前就解引用 `resp`，当请求失败时 `resp` 可能为 nil，导致 panic。

**修复方式**:

- 第 112 行: 将 `defer _resp.Body.Close()` 移到错误检查之后
- 第 246 行: 将 `defer resp.Body.Close()` 移到错误检查之后

```go
// 修改前 (line 112)
resp, err := r.cli.Do(r.req)
defer _resp.Body.Close()

// 修改后
resp, err := r.cli.Do(r.req)
if err != nil {
    return nil, err
}
defer _resp.Body.Close()
```

**预计耗时**: 10 分钟

---

### P0.2 修复 `Luoyang.postBody()` panic

**文件**: `app/luoyang.go`

**问题**: `postBody()` 方法直接 `panic("implement me")`，生产代码中不应有 panic。

**修复方式**: 实现该方法或返回 `errors.New("not implemented")`。

**预计耗时**: 10 分钟

---

## P1 — 高优先级（建议本轮完成）

### P1.1 为核心模块添加单元测试

**覆盖目标**: 按优先级排序

| 模块 | 文件                           | 测试重点                     | 估算行数 |
| ---- | ------------------------------ | ---------------------------- | -------- |
| 1    | `pkg/downloader/downloader.go` | 任务队列、并发控制、重试逻辑 | ~150 行  |
| 2    | `pkg/downloader/iiif.go`       | info.json 解析、拼图参数计算 | ~100 行  |
| 3    | `app/iiif.go`                  | manifest.json 解析 (v2/v3)   | ~100 行  |
| 4    | `app/image_downloader.go`      | URL 模板替换、占位符解析     | ~80 行   |
| 5    | `config/init.go`               | 页码范围、册范围逻辑         | ~60 行   |
| 6    | `router/interface.go`          | 域名路由匹配                 | ~50 行   |

**建议方案**:

- 使用 `net/http/httptest` 模拟 HTTP 服务器
- 使用本地 JSON fixture 文件测试 manifest 解析
- 优先测试纯逻辑函数（解析、计算、验证），再测试网络相关

**预计耗时**: 4-6 小时

---

### P1.2 统一 User-Agent 常量

**文件冲突**:

| 文件                     | 当前值                                              |
| ------------------------ | --------------------------------------------------- |
| `config/constant.go`     | Chrome 136 (`Mozilla/5.0 ... Chrome/136.0.0.0 ...`) |
| `pkg/downloader/base.go` | Firefox 139 (`Mozilla/5.0 ... Firefox/139.0`)       |

**修复方式**:

1. 在 `config/constant.go` 中保留默认 UA
2. 删除 `pkg/downloader/base.go` 中的 `userAgent` 常量
3. 所有代码统一引用 `config.Conf.UserAgent`

**预计耗时**: 20 分钟

---

## P2 — 中优先级（建议下一轮）

### P2.1 将 TLS 验证改为可配置

**当前状态**: 几乎所有 `NewXxx()` 构造函数都硬编码了 `InsecureSkipVerify: true`。

**改造方案**:

1. 在 `config.Input` 结构体中新增字段:

```go
SkipVerify bool // 跳过 TLS 证书验证，默认 false
```

2. 新增命令行参数:

```
--insecure    跳过 TLS 证书验证（访问自签名证书站点时使用）
```

3. 创建工具函数替代各下载器的重复代码:

```go
func NewInsecureClient(timeout time.Duration) *http.Client {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: config.Conf.SkipVerify,
        },
    }
    jar, _ := cookiejar.New(nil)
    return &http.Client{Timeout: timeout, Jar: jar, Transport: tr}
}
```

**涉及文件**: 所有 `app/*.go` 中的 `NewXxx()` 函数（约 40+ 个文件需要修改）

**预计耗时**: 4-6 小时

---

### P2.2 统一下载器架构（逐步废弃旧模式）

**现状**: 项目中存在两套下载体系共存

| 模式                     | 代表文件                       | 特征                                               |
| ------------------------ | ------------------------------ | -------------------------------------------------- |
| 旧模式 `DownloadTask`    | `app/template.go`              | 使用 `gohttp` + `QueueLimit`，各下载器自己管理并发 |
| 新模式 `DownloadManager` | `pkg/downloader/downloader.go` | 集中式任务管理，统一进度条、重试、并发控制         |

**迁移计划**:

```
阶段 1（P2.2a）: 为 DownloadManager 添加缺失功能
  - 单任务进度条（基于字节数）
  - 暂停/恢复能力
  - 回调通知

阶段 2（P2.2b）: 挑选 2-3 个典型下载器迁移
  - 目标: app/waseda.go, app/ndljp.go, app/harvard.go
  - 验证新架构的通用性

阶段 3（P2.2c）: 批量迁移剩余下载器
  - 可分批进行，每批 5-10 个
```

**预计耗时**: 8-12 小时（全阶段）

---

### P2.3 集成 golangci-lint CI 检查

**步骤**:

1. 创建 `.golangci.yml` 配置文件:

```yaml
linters:
  enable:
    - govet
    - gofmt
    - errcheck
    - staticcheck
    - ineffassign
    - unconvert
```

2. 更新 `.github/workflows/go.yml` 在 CI 中运行:

```yaml
- name: Lint
  uses: golangci/golangci-lint-action@v3
```

**预计耗时**: 1 小时

---

## P3 — 低优先级（建议后续迭代）

### P3.1 提取通用下载逻辑到基础类型

**现状**: 每个下载器独立实现 `getBody()`、`postBody()`、header 构造等，存在大量重复代码。

**建议**: 创建一个基础类型（如 `BaseDownloader`），封装公共方法:

```go
type BaseDownloader struct {
    Client  *http.Client
    Jar     *cookiejar.Jar
    Headers map[string]string
    ctx     context.Context
    cancel  context.CancelFunc
}

func (b *BaseDownloader) NewClient(timeout time.Duration) { ... }
func (b *BaseDownloader) GetBody(url string) ([]byte, error) { ... }
func (b *BaseDownloader) PostBody(url string, data []byte) ([]byte, error) { ... }
func (b *BaseDownloader) BuildRequestHeader() map[string]string { ... }
func (b *BaseDownloader) GetBookId(url string, pattern *regexp.Regexp) string { ... }
```

各站点头下载器通过嵌入 `BaseDownloader` 获得公共能力。

**预计耗时**: 6-8 小时

---

### P3.2 添加路径穿越防护

**文件**: `config/conf.go` — `Conf.Directory` 直接传递给 `os.Mkdir`

**修复方式**:

```go
// 在 init 阶段验证路径
safeDir := filepath.Clean(Conf.Directory)
if _, err := os.Stat(safeDir); os.IsNotExist(err) {
    if err := os.MkdirAll(safeDir, 0755); err != nil {
        log.Fatalf("无法创建目录: %v", err)
    }
}
```

**预计耗时**: 15 分钟

---

### P3.3 错误处理强化

**目标文件**: 各 `app/*.go` 中的 `getBody()` 调用

**当前问题**: 很多地方在 `resp.GetBody()` 返回 nil 时仅记录部分错误信息，未传递原始 HTTP 状态码。

**改进**: 统一使用 `downloader.DownloadManager` 的错误处理机制，或封装一个标准错误类型:

```go
type HTTPError struct {
    StatusCode int
    URL        string
    Message    string
}
```

**预计耗时**: 2-3 小时

---

## P4 — 清理（可选，低风险/噪声）

### P4.1 errcheck — 未检查的返回值（~172 处）

**说明**: 大量 `r.do()`、`bar.Add()`、`gohttp.FastGet()` 等调用的返回值未检查。绝大多数是设计上故意忽略的（进度条更新、日志等），修复成本高、收益低。

**建议**:

- 在 `.golangci.yml` 中禁用 `errcheck` linter，或添加 `//nolint:errcheck` 注释
- 无需逐处修复

**预计耗时**: 10 分钟（配置调整）

### P4.2 unused — 未使用的声明（~128 处）

**说明**: 主要包括：

1. **Stub 方法** — 为满足接口而实现但未实际调用的方法（如 `getVolumes`、`postBody` 等）
2. **未使用的常量/变量** — `defaultTimeout`、`defaultFormat` 等
3. **未使用的结构体字段** — 各下载器中的 `bufBuilder`、`urlsFile` 等

**建议**:

- Stub 方法：上游代码需要，不可删除
- 未使用常量：可删除
- 未使用字段：可清理

**预计耗时**: 2-3 小时（选择性清理）

---

## 时间线建议

```
已完成: P0 (紧急修复), P1.2 (User-Agent), P2.3 (CI Lint), P3.2 (路径防护)
Week 1: P1.1 (添加测试)
Week 2: P2.1 (TLS 配置化)
Week 3: P2.2a (DownloadManager 增强)
Week 4-5: P2.2b (试点迁移 2-3 个下载器)
Week 6: P3.1 (基础类型设计) + P3.3 (错误处理)
Week 7+: P4 (清理 - 可选)
```

---

## 影响评估

| 任务                 | 风险等级 | 向后兼容      | 说明                                    |
| -------------------- | -------- | ------------- | --------------------------------------- |
| P0.1 nil 指针修复    | 🟢 低    | ✅ 完全兼容   | 纯 bug 修复                             |
| P0.2 panic 修复      | 🟢 低    | ✅ 完全兼容   | 纯 bug 修复                             |
| P1.1 添加测试        | 🟢 低    | ✅ 完全兼容   | 只增不删                                |
| P1.2 User-Agent 统一 | 🟢 低    | ✅ 完全兼容   | 行为不变                                |
| P2.1 TLS 配置化      | 🟡 中    | ⚠️ 需调整配置 | 默认改为安全，现有用法需加 `--insecure` |
| P2.2 架构统一        | 🟡 中    | ⚠️ 需充分测试 | 重构需配套测试保障                      |
| P3.1 基础类型        | 🟡 中    | ✅ 完全兼容   | 逐步替换，可并行存在                    |
