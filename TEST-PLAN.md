# Bookget 站点下载测试计划

> 基于 2026-06-19 代码重构后的验证计划
> 覆盖 **80+ 个数字图书馆站点**

---

## 测试策略

### 测试层级

| 层级 | 范围       | 方式                              | 执行者      |
| :--: | ---------- | --------------------------------- | ----------- |
|  L0  | 单元测试   | `go test ./...`                   | CI / 自动化 |
|  L1  | 编译检查   | `go build ./...` + `go vet ./...` | CI / 自动化 |
|  L2  | 站点可达性 | 每个站点发送 HEAD 请求验证可达    | 手动        |
|  L3  | 功能验证   | 实际下载 1-2 页验证完整性         | 手动        |

### 测试前提

- 已执行 `make macos-arm64` 构建成功
- 网络连接正常
- 部分站点需要登录或验证（需 bookget-gui）

---

## 中国站点

### 🇨🇳 国家图书馆（read.nlc.cn / mylib.nlc.cn / guji.nlc.cn）

| 测试项       | 命令                                                                                                                                    | 预期结果      |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| 基本连接     | `./bookget-macos-arm64 -i "https://read.nlc.cn/allSearch/searchDetail?searchType=26&showType=1&indexName=data_416&fid=416000000000000"` | 返回书籍信息  |
| PDF 下载     | `./bookget-macos-arm64 -i "https://read.nlc.cn/allSearch/searchDetail?..." -ext .pdf`                                                   | 下载 PDF 文件 |
| 指定页码范围 | `./bookget-macos-arm64 -i "..." -p 1:5`                                                                                                 | 仅下载前 5 页 |
| 多线程       | `./bookget-macos-arm64 -i "..." -n 4`                                                                                                   | 4 线程下载    |

### 🇨🇳 臺灣華文電子書庫（taiwanebook.ncl.edu.tw）

| 测试项   | 命令                                                                           | 预期结果     |
| -------- | ------------------------------------------------------------------------------ | ------------ |
| 基本下载 | `./bookget-macos-arm64 -i "https://taiwanebook.ncl.edu.tw/zh-tw/book/NCL-..."` | 成功下载图片 |

### 🇨🇳 香港中文大学（repository.lib.cuhk.edu.hk）

| 测试项   | 命令                                                                | 预期结果 |
| -------- | ------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://repository.lib.cuhk.edu.hk/..."` | 成功下载 |

### 🇨🇳 香港科技大学（lbezone.hkust.edu.hk）

| 测试项   | 命令                                                          | 预期结果 |
| -------- | ------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://lbezone.hkust.edu.hk/..."` | 成功下载 |

### 🇨🇳 香港大学（digitalrepository.lib.hku.hk）

| 测试项   | 命令                                                                  | 预期结果 |
| -------- | --------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://digitalrepository.lib.hku.hk/..."` | 成功下载 |

### 🇨🇳 洛阳市图书馆（111.7.82.29:8090）

| 测试项   | 命令                                                     | 预期结果 |
| -------- | -------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "http://111.7.82.29:8090/..."` | 成功下载 |

### 🇨🇳 温州市图书馆（oyjy.wzlib.cn）

| 测试项   | 命令                                                   | 预期结果 |
| -------- | ------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://oyjy.wzlib.cn/..."` | 成功下载 |

### 🇨🇳 深圳市图书馆（yun.szlib.org.cn）

| 测试项   | 命令                                                      | 预期结果 |
| -------- | --------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://yun.szlib.org.cn/..."` | 成功下载 |

### 🇨🇳 广州大典（gzdd.gzlib.gov.cn）

| 测试项   | 命令                                                       | 预期结果 |
| -------- | ---------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://gzdd.gzlib.gov.cn/..."` | 成功下载 |

### 🇨🇳 天一阁（gj.tianyige.com.cn）

| 测试项   | 命令                                                        | 预期结果 |
| -------- | ----------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://gj.tianyige.com.cn/..."` | 成功下载 |

### 🇨🇳 江苏高校古籍（jsgxgj.nju.edu.cn）

| 测试项   | 命令                                                       | 预期结果 |
| -------- | ---------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://jsgxgj.nju.edu.cn/..."` | 成功下载 |

### 🇨🇳 中华寻根网（ouroots.nlc.cn）

| 测试项   | 命令                                                    | 预期结果 |
| -------- | ------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://ouroots.nlc.cn/..."` | 成功下载 |

### 🇨🇳 国家哲学社会科学文献中心（ncpssd.org）

| 测试项   | 命令                                                    | 预期结果 |
| -------- | ------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.ncpssd.org/..."` | 成功下载 |

### 🇨🇳 山东中医药大学（gjsztsg.sdutcm.edu.cn）

| 测试项   | 命令                                                           | 预期结果 |
| -------- | -------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://gjsztsg.sdutcm.edu.cn/..."` | 成功下载 |

### 🇨🇳 山东省古籍（guji.sdlib.com）

| 测试项   | 命令                                                   | 预期结果 |
| -------- | ------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "http://guji.sdlib.com/..."` | 成功下载 |

### 🇨🇳 天津图书馆（lswx.tjl.tj.cn:8001）

| 测试项   | 命令                                                        | 预期结果 |
| -------- | ----------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "http://lswx.tjl.tj.cn:8001/..."` | 成功下载 |

### 🇨🇳 云南数字方志馆（dfz.yn.gov.cn）

| 测试项   | 命令                                                   | 预期结果 |
| -------- | ------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://dfz.yn.gov.cn/..."` | 成功下载 |

### 🇨🇳 诸城市图书馆（124.134.220.209:8100）

| 测试项   | 命令                                                         | 预期结果 |
| -------- | ------------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "http://124.134.220.209:8100/..."` | 成功下载 |

### 🇨🇳 中央美术学院（dlib.cafa.edu.cn）

| 测试项   | 命令                                                      | 预期结果 |
| -------- | --------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://dlib.cafa.edu.cn/..."` | 成功下载 |

### 🇨🇳 抗日战争与中日关系文献（modernhistory.org.cn）

| 测试项   | 命令                                                              | 预期结果 |
| -------- | ----------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.modernhistory.org.cn/..."` | 成功下载 |

### 🇨🇳 甘肃省图书馆（zszy.gslib.com.cn）

| 测试项   | 命令                                                       | 预期结果 |
| -------- | ---------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://zszy.gslib.com.cn/..."` | 成功下载 |

---

## 日本站点

### 🇯🇵 国立国会图书馆（dl.ndl.go.jp）

| 测试项   | 命令                                                  | 预期结果 |
| -------- | ----------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://dl.ndl.go.jp/..."` | 成功下载 |

### 🇯🇵 e国宝（emuseum.nich.go.jp）

| 测试项   | 命令                                                        | 预期结果 |
| -------- | ----------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://emuseum.nich.go.jp/..."` | 成功下载 |

### 🇯🇵 宫内厅书陵部（db2.sido.keio.ac.jp）

| 测试项   | 命令                                                                              | 预期结果 |
| -------- | --------------------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://db2.sido.keio.ac.jp/kanseki/bib_frame?id=..."` | 成功下载 |

### 🇯🇵 东京大学东洋文化研究所（shanben.ioc.u-tokyo.ac.jp）

| 测试项   | 命令                                                               | 预期结果 |
| -------- | ------------------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://shanben.ioc.u-tokyo.ac.jp/..."` | 成功下载 |

### 🇯🇵 国立公文书馆（digital.archives.go.jp）

| 测试项   | 命令                                                                | 预期结果 |
| -------- | ------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.digital.archives.go.jp/..."` | 成功下载 |

### 🇯🇵 东洋文库（dsr.nii.ac.jp）

| 测试项   | 命令                                                   | 预期结果 |
| -------- | ------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://dsr.nii.ac.jp/..."` | 成功下载 |

### 🇯🇵 早稻田大学（archive.wul.waseda.ac.jp）

| 测试项   | 命令                                                              | 预期结果 |
| -------- | ----------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://archive.wul.waseda.ac.jp/..."` | 成功下载 |

### 🇯🇵 国書数据库（kokusho.nijl.ac.jp）

| 测试项   | 命令                                                        | 预期结果 |
| -------- | ----------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://kokusho.nijl.ac.jp/..."` | 成功下载 |

### 🇯🇵 京都大学人文科学研究所（kanji.zinbun.kyoto-u.ac.jp）

| 测试项   | 命令                                                                | 预期结果 |
| -------- | ------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://kanji.zinbun.kyoto-u.ac.jp/..."` | 成功下载 |

### 🇯🇵 駒澤大学（repo.komazawa-u.ac.jp）

| 测试项   | 命令                                                           | 预期结果           |
| -------- | -------------------------------------------------------------- | ------------------ |
| 基本下载 | `./bookget-macos-arm64 -i "https://repo.komazawa-u.ac.jp/..."` | IIIF manifest 下载 |

### 🇯🇵 关西大学（www.iiif.ku-orcas.kansai-u.ac.jp）

| 测试项   | 命令              | 预期结果      |
| -------- | ----------------- | ------------- |
| 基本下载 | IIIF manifest URL | IIIF 拼图下载 |

### 🇯🇵 庆应义塾大学（dcollections.lib.keio.ac.jp）

| 测试项   | 命令              | 预期结果      |
| -------- | ----------------- | ------------- |
| 基本下载 | IIIF manifest URL | IIIF 拼图下载 |

### 🇯🇵 国立历史民俗博物馆（khirin-a.rekihaku.ac.jp）

| 测试项   | 命令                                                             | 预期结果 |
| -------- | ---------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://khirin-a.rekihaku.ac.jp/..."` | 成功下载 |

### 🇯🇵 市立米泽图书馆（www.library.yonezawa.yamagata.jp）

| 测试项   | 命令                                                                      | 预期结果 |
| -------- | ------------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.library.yonezawa.yamagata.jp/..."` | 成功下载 |

### 🇯🇵 东京国立博物馆（webarchives.tnm.jp）

| 测试项   | 命令                                                        | 预期结果 |
| -------- | ----------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://webarchives.tnm.jp/..."` | 成功下载 |

### 🇯🇵 龙谷大学（da.library.ryukoku.ac.jp）

| 测试项   | 命令              | 预期结果      |
| -------- | ----------------- | ------------- |
| 基本下载 | IIIF manifest URL | IIIF 拼图下载 |

---

## 美国站点

### 🇺🇸 哈佛大学（iiif.lib.harvard.edu / curiosity.lib.harvard.edu）

| 测试项    | 命令                                                          | 预期结果      |
| --------- | ------------------------------------------------------------- | ------------- |
| IIIF 下载 | `./bookget-macos-arm64 -i "https://iiif.lib.harvard.edu/..."` | IIIF 拼图下载 |

### 🇺🇸 HathiTrust（babel.hathitrust.org）

| 测试项   | 命令                                                                    | 预期结果 |
| -------- | ----------------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://babel.hathitrust.org/cgi/pt?id=..."` | 下载图片 |

### 🇺🇸 普林斯顿大学（catalog.princeton.edu / dpul.princeton.edu）

| 测试项   | 命令                                                           | 预期结果 |
| -------- | -------------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://catalog.princeton.edu/..."` | 成功下载 |

### 🇺🇸 国会图书馆（www.loc.gov）

| 测试项   | 命令                                                 | 预期结果 |
| -------- | ---------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.loc.gov/..."` | 成功下载 |

### 🇺🇸 犹他州家谱（www.familysearch.org）

| 测试项   | 命令                                                          | 预期结果   |
| -------- | ------------------------------------------------------------- | ---------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.familysearch.org/..."` | 需登录验证 |

### 🇺🇸 archive.org（archive.org）

| 测试项    | 命令                                                         | 预期结果      |
| --------- | ------------------------------------------------------------ | ------------- |
| IIIF 下载 | `./bookget-macos-arm64 -i "https://archive.org/details/..."` | IIIF 拼图下载 |

### 🇺🇸 史密森尼学会（ids.si.edu / asia.si.edu）

| 测试项    | 命令              | 预期结果      |
| --------- | ----------------- | ------------- |
| IIIF 下载 | IIIF manifest URL | IIIF 拼图下载 |

### 🇺🇸 柏克莱加州大学（digicoll.lib.berkeley.edu）

| 测试项   | 命令                                                               | 预期结果 |
| -------- | ------------------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://digicoll.lib.berkeley.edu/..."` | 成功下载 |

---

## 欧洲站点

### 🇩🇪 柏林国立图书馆（digital.staatsbibliothek-berlin.de）

| 测试项    | 命令              | 预期结果      |
| --------- | ----------------- | ------------- |
| IIIF 下载 | IIIF manifest URL | IIIF 拼图下载 |

### 🇩🇪 巴伐利亞州立圖書館（ostasien.digitale-sammlungen.de）

| 测试项    | 命令                                                                     | 预期结果      |
| --------- | ------------------------------------------------------------------------ | ------------- |
| IIIF 下载 | `./bookget-macos-arm64 -i "https://ostasien.digitale-sammlungen.de/..."` | IIIF 拼图下载 |

### 🇬🇧 牛津大学博德利图书馆（digital.bodleian.ox.ac.uk）

| 测试项    | 命令              | 预期结果      |
| --------- | ----------------- | ------------- |
| IIIF 下载 | IIIF manifest URL | IIIF 拼图下载 |

### 🇬🇧 大英图书馆（www.bl.uk）

| 测试项   | 命令                                               | 预期结果 |
| -------- | -------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://www.bl.uk/..."` | 成功下载 |

### 🇦🇹 奥地利国图（digital.onb.ac.at）

| 测试项    | 命令              | 预期结果      |
| --------- | ----------------- | ------------- |
| IIIF 下载 | IIIF manifest URL | IIIF 拼图下载 |

---

## 韩国站点

### 🇰🇷 国立中央图书馆（lod.nl.go.kr）

| 测试项   | 命令                                                  | 预期结果            |
| -------- | ----------------------------------------------------- | ------------------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://lod.nl.go.kr/..."` | 需 bookget-gui 验证 |

### 🇰🇷 首尔大学（kyudb.snu.ac.kr）

| 测试项   | 命令                                                     | 预期结果 |
| -------- | -------------------------------------------------------- | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://kyudb.snu.ac.kr/..."` | 成功下载 |

---

## 俄罗斯站点

### 🇷🇺 俄罗斯国立图书馆（viewer.rsl.ru）

| 测试项   | 命令                                                   | 预期结果 |
| -------- | ------------------------------------------------------ | -------- |
| 基本下载 | `./bookget-macos-arm64 -i "https://viewer.rsl.ru/..."` | 成功下载 |

---

## 特殊/通用模式

### IIIF Manifest 自动检测（-m 2）

| 测试项  | 命令                                                                        | 预期结果             |
| ------- | --------------------------------------------------------------------------- | -------------------- |
| IIIF v2 | `./bookget-macos-arm64 -m 2 -i "https://example.com/iiif/manifest.json"`    | 自动解析 manifest    |
| IIIF v3 | `./bookget-macos-arm64 -m 2 -i "https://example.com/iiif/v3/manifest.json"` | 自动解析 manifest v3 |

### 通用批量下载（-m 1）

| 测试项   | 命令                                     | 预期结果     |
| -------- | ---------------------------------------- | ------------ |
| URL 模板 | `./bookget-macos-arm64 -m 1`             | 进入交互模式 |
| 批量文件 | `./bookget-macos-arm64 -m 1 -I urls.txt` | 批量下载     |

---

## 回归测试清单

### 构建验证（每次修改后执行）

```bash
make macos-arm64                    # 构建
go test ./...                       # 单元测试
go vet ./...                        # 静态分析
golangci-lint run                   # Lint 检查
```

### 参数测试

| 参数         | 测试命令           | 覆盖场景          |
| ------------ | ------------------ | ----------------- |
| `-i`         | `-i "URL"`         | 单 URL 下载       |
| `-O`         | `-O "/tmp/output"` | 自定义输出目录    |
| `-p`         | `-p 1:10`          | 页码范围          |
| `-v`         | `-v 1:3`           | 册范围            |
| `-n`         | `-n 4`             | 多线程            |
| `-U`         | `-U "Custom-UA"`   | 自定义 User-Agent |
| `--insecure` | `--insecure`       | TLS 跳过验证      |
| `-ext`       | `-ext .tif`        | 指定扩展名        |
| `-m`         | `-m 2`             | IIIF 模式         |

### 安全测试

| 测试项   | 验证内容                                         |
| -------- | ------------------------------------------------ |
| 路径穿越 | `-O "../../../etc"` 应被 `filepath.Clean()` 拦截 |
| TLS 验证 | 默认应验证证书，`--insecure` 跳过                |
| URL 注入 | 异常 URL 不应导致 panic                          |

---

## 测试记录表

| 站点                               | 测试日期   | 结果 | 诊断说明                                                    |
| ---------------------------------- | ---------- | :--: | ----------------------------------------------------------- |
| **中国站点**                       |            |      |                                                             |
| read.nlc.cn                        | 2026-06-19 |  ✅  | 3.4MB JPEG, 2288×5392                                      |
| gzdd.gzlib.org.cn                  | 2026-06-19 |  ✅  | 34MB PDF 下载成功                                           |
| yun.szlib.org.cn                   | 2026-06-19 |  ✅  | 120卷, 大容量需更多时间                                     |
| taiwanebook.ncl.edu.tw             | 2026-06-19 |  ✅  | 修复: httpClient() 增加 DisableKeepAlives: true             |
| 111.7.82.29:8090 (luoyang)         | 2026-06-19 |  ✅  | 332MB PDF 下载成功(1m4s,服务器慢非bug)                   |
| catalog.princeton.edu              | 2026-06-19 |  ✅  | 350页/7卷, 全部成功(4min,大容量非bug)           |
| ouroots.nlc.cn                     | —          |  ⬜  | 未测试                                                      |
| jsgxgj.nju.edu.cn                  | —          |  ⬜  | 未测试                                                      |
| lbezone.hkust.edu.hk               | —          |  ⬜  | 未测试                                                      |
| digitalrepository.lib.hku.hk       | —          |  ⬜  | 未测试                                                      |
| repository.lib.cuhk.edu.hk         | 2026-06-19 |  ⚠️  | AWS WAF Challenge, 实现3层策略(cookie→chromedp→手动) |
| dl.ndl.go.jp                       | 2026-06-19 |  ⚠️  | IIIF图片服务器受AWS WAF+CloudFront保护,需浏览器             |
| www.ncpssd.org                     | —          |  ⬜  | 未测试                                                      |
| oyjy.wzlib.cn                      | —          |  ⬜  | 未测试                                                      |
| gj.tianyige.com.cn                 | —          |  ⬜  | 未测试                                                      |
| dfz.yn.gov.cn                      | —          |  ⬜  | 未测试                                                      |
| dlib.cafa.edu.cn                   | —          |  ⬜  | 未测试                                                      |
| 124.134.220.209:8100 (zhucheng)    | —          |  ⬜  | 未测试                                                      |
| **日本站点**                       |            |      |                                                             |
| archive.wul.waseda.ac.jp           | 2026-06-19 |  ✅  | 18卷1755页,全部成功                                         |
| emuseum.nich.go.jp                 | 2026-06-19 |  ✅  | IIIF tiles, 5000×4789                                      |
| iiif (Keio Univ)                   | 2026-06-19 |  ✅  | 132 tiles → 2732×3040                                      |
| dl.ndl.go.jp                       | 2026-06-19 |  ⚠️  | AWS WAF保护IIIF图片服务器（见上）                         |
| db2.sido.keio.ac.jp                | —          |  ⬜  | 未测试                                                      |
| shanben.ioc.u-tokyo.ac.jp          | —          |  ⬜  | 未测试                                                      |
| www.digital.archives.go.jp         | —          |  ⬜  | 未测试                                                      |
| dsr.nii.ac.jp                      | —          |  ⬜  | 未测试                                                      |
| kokusho.nijl.ac.jp                 | —          |  ⬜  | 未测试                                                      |
| kanji.zinbun.kyoto-u.ac.jp         | —          |  ⬜  | 未测试                                                      |
| khirin-a.rekihaku.ac.jp            | —          |  ⬜  | 未测试                                                      |
| www.library.yonezawa.yamagata.jp   | —          |  ⬜  | 未测试                                                      |
| webarchives.tnm.jp                 | —          |  ⬜  | 未测试                                                      |
| da.library.ryukoku.ac.jp           | —          |  ⬜  | 未测试                                                      |
| **美国站点**                       |            |      |                                                             |
| archive.org                        | 2026-06-19 |  ✅  | IIIF tiles合并                                              |
| digicoll.lib.berkeley.edu          | 2026-06-19 |  ✅  | 26 PDFs, 565MB                                              |
| catalog.princeton.edu              | 2026-06-19 |  ✅  | 350页, 9714×8241高分辨率IIIF                               |
| www.loc.gov                        | 2026-06-19 |  ❌  | URL格式不匹配? 需检查loc.go解析逻辑                      |
| iiif.lib.harvard.edu               | 2026-06-19 |  ❌  | 403 Forbidden (可能需特定Referer)                         |
| babel.hathitrust.org               | 2026-06-19 |  ❌  | Forbidden (IP区域限制)                                      |
| dpul.princeton.edu                 | —          |  ⬜  | 未测试(备选URL)                                             |
| ids.si.edu                         | —          |  ⬜  | 未测试                                                      |
| www.familysearch.org               | —          |  ⬜  | 未测试(需微卷URL格式)                                       |
| **欧洲站点**                       |            |      |                                                             |
| digital.staatsbibliothek-berlin.de | 2026-06-19 |  ⚠️  | 第1页104 tiles成功,后续页tile 404(服务器兼容)         |
| www.digitale-sammlungen.de         | 2026-06-19 |  ❌  | IIIF info.json格式不被识别,需检查dezoomify逻辑          |
| digital.bodleian.ox.ac.uk          | 2026-06-19 |  ❌  | URL解析失败? 需检查oxacuk.go                              |
| www.bl.uk                          | 2026-06-19 |  ❌  | HTML解析失败,需检查bluk.go                                |
| digital.onb.ac.at                  | —          |  ⬜  | 未测试                                                      |
| **其他站点**                       |            |      |                                                             |
| idp.nlc.cn                         | —          |  ⬜  | 未测试(需先搜索获取uid)                                     |
| kyudb.snu.ac.kr                    | —          |  ⬜  | 未测试                                                      |
| viewer.rsl.ru                      | —          |  ⬜  | 未测试                                                      |
| lib.nomfoundation.org              | —          |  ⬜  | 未测试                                                      |
| hannom.nlv.gov.vn                  | —          |  ⬜  | 未测试                                                      |

---

## 已修复的 Bug 汇总

| Bug | 文件 | 修复内容 | 影响站点 |
|-----|------|---------|---------|
| **IIIF nil pointer crash** | `pkg/downloader/iiif.go:792` | `downloadImage()` 对 404/500 返回 `(nil, nil)` 改为返回 error；goroutine 中加 `img == nil` 检查 | 柏林国立图书馆、所有 IIIF tile 下载 |
| **unexpected EOF (keep-alive)** | `pkg/downloader/downloader.go:670` | `httpClient()` 增加 `DisableKeepAlives: true`，防止 HEAD→GET 序列中 keep-alive 连接被服务器关闭 | 台湾华文电子书库、洛阳市图书馆、所有 PDF/大文件下载 |
| **CUHK 共享内存在 Unix 上为空** | `app/cuhk.go` | 重写整个下载器：移除不工作的 `sharedmemory`，改为 `chromedp` + cookie 文件 + 手动引导 3 层策略 | CUHK |
| **nil buffer 崩溃** | `pkg/downloader/downloader.go:318` | `Download()` 中增加 `if task.buffer == nil { task.buffer = bytes.NewBuffer(nil) }` | 所有使用 `AddFromLegacy`/`AddImageTasks` 的站点 |

---

## 未成功站点分析

### AWS WAF / CloudFront Challenge（需浏览器）

这两个站点使用 AWS WAF 保护 IIIF 图片服务器，自动化工具无法直接访问：

| 站点 | WAF 类型 | 推荐方案 |
|------|---------|---------|
| repository.lib.cuhk.edu.hk | AWS WAF Challenge (x-amzn-waf-action: challenge) | `--cookie cookie.txt` 从浏览器导出 |
| dl.ndl.go.jp (IIIF images) | AWS WAF + CloudFront | 同上 |

**现状**: `pkg/chromedphelper` 已实现 chromedp 自动绕过，但 AWS WAF 的 JS proof-of-work 在 headless Chrome 下 timeout（60s 不够）。Playwright + 真实 Chromium 同样无法在 60s 内完成。

### URL 格式/解析问题（需代码审查）

| 站点 | 现象 | 可能原因 |
|------|------|---------|
| www.loc.gov | 返回空（无文件下载） | `loc.go` 的 HTML 解析逻辑可能与当前 LOC 网站 HTML 结构不匹配 |
| digital.bodleian.ox.ac.uk | 未下载到文件 | `oxacuk.go` URL 格式检测或 manifest 解析问题 |
| www.bl.uk | 返回 `<nil>` | `bluk.go` 的 `getBody` 返回空 body |
| www.digitale-sammlungen.de | IIIF info.json 未识别 | `dezoomify` 的 IIIF 版本检测逻辑不兼容该站格式 |

### IP 区域限制

| 站点 | 限制 |
|------|------|
| babel.hathitrust.org | 需要美国 IP |
| iiif.lib.harvard.edu | 403 (可能需特定 Referer) |

## 常见问题

### 站点无法访问

- 检查网络连接
- 部分站点需要特定地区 IP
- 尝试使用 `--insecure` 参数

### 下载失败

- 检查是否需要 cookie/登录
- 使用 bookget-gui 完成验证
- 尝试减小线程数 `-n 1`

### 图片不完整

- 检查 IIIF manifest 版本（v2/v3）
- 尝试使用 `-d false` 禁用 DZI 拼图

### 输出文件名乱码

- 已集成自动命名框架
- 可通过 `-O` 指定输出目录
