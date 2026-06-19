# Changelog

This is a fork of [deweizhu/bookget](https://github.com/deweizhu/bookget) optimized for macOS with Apple Silicon (M-series).

## Changes Made

### 0. macOS Apple Silicon Release Flow
- **Files**: `Makefile`, `.github/workflows/go.yml`, `README.md`, `README-FORK.md`
- **Changes**:
  - Added `make macos-arm64` for Apple Silicon builds
  - Added `make install-macos-arm64` for local installation to `~/.local/bin/bookget`
  - Added `make package-macos-arm64` for GitHub Release tarballs
  - Updated CI to run tests and publish `bookget-macos-arm64.tar.gz`
  - Reworked documentation around MacBook M-series usage
- **Validation**:
  - `go test ./...`
  - `make package-macos-arm64`
  - Verified output binary is `Mach-O 64-bit executable arm64`

### 1. Auto-Naming Framework (Generic Naming System)
- **File**: `app/template.go` (new)
- **Functions Added**:
  - `NormalizeNamePart()`: HTML entity unescape + whitespace normalization
  - `BuildOutputFileName()`: Deduplicates name parts using map, joins with underscore separator
  - `ExtractHTMLTitle()`: Tries 3 regex patterns (og:title, meta title, title tag)
- **Impact**: Replaced sequential numbering with meaningful filenames across all downloader

### 2. Publication Time Support (NLC)
- **File**: `app/nlc.go`
- **Changes**:
  - Added `pubTime` field to metadata
  - New function `extractNlcPublishTime()` parses "出版时间：" from page
  - Filename format: `[BookTitle]_[VolumeNumber]_[PublicationTime].pdf`
  - Example: `直隸省通志稿_0001_民國間[1912-1949].pdf`

### 3. Wzlib Integration (Both Campuses)
- **File**: `app/wzlib.go`
- **Changes**:
  - Added `pdfNames` map for per-volume naming
  - New campus: Extracts title from API response (`resT.Title`)
  - Old campus: Extracts metadata from search results
  - Both use `BuildOutputFileName()` for consistent naming
  - Tested filenames: `遜翁詞賸.pdf`, `青白之广济.pdf`

### 4. Seoul National University (KyudbSnu)
- **File**: `app/kyudbsnu.go`
- **Changes**:
  - Added `pdfNames` map for URL→filename mapping
  - Extracts BOOK_NM/ORI_TIT/VOL_NO from JSON response
  - Uses `BuildOutputFileName()` in `getPdfUrls()` and `doPdf()`
  - Enhanced HTTP headers for better compatibility
  - Improved itemId extraction from HTML pages

### 5. University of Tokyo (Utokyo)
- **File**: `app/utokyo.go`
- **Changes**:
  - Added HTML title extraction via `ExtractHTMLTitle()`
  - Filename format: `[BookTitle]_[SortId].pdf`
  - Example: `周禮註疏四十二卷（十三經註疏所收）_0001.pdf`

### 6. Luoyang Teacher's College Library
- **File**: `app/luoyang.go`
- **Changes**:
  - Added HTML title extraction via `ExtractHTMLTitle()`
  - Filename format: `[BookTitle]_[SortId].pdf`
  - Example: `中州人物考 - 史部 - 洛阳市图书馆_0001.pdf`

### 7. Guangzhou Library (Gzlib)
- **File**: `app/gzlib.go`
- **Changes**:
  - Added `title` field for metadata storage
  - Uses `ExtractHTMLTitle()` for page title extraction
  - Single PDF naming: `BuildOutputFileName(".pdf", title, bookId)`
  - Example: `广州大典_GZDD011525000.pdf`

## Tested & Validated

✓ **NLC (National Library of China)**: Real download test with publication time  
✓ **Wzlib New Campus**: `遜翁詞賸.pdf`  
✓ **Wzlib Old Campus**: `青白之广济.pdf`  
✓ **Guangzhou Library**: `广州大典_GZDD011525000.pdf`  
✓ **University of Tokyo**: `周禮註疏四十二卷（十三經註疏所收）_0001.pdf`  
✓ **Luoyang Library**: `中州人物考 - 史部 - 洛阳市图书馆_0001.pdf`  
✓ **SeoulNUM (Code Review)**: Auto-naming framework fully integrated  

## Platform Support

⚠️ **This fork is optimized for macOS with Apple Silicon (M-series) chips only**
- Compiled binary: `dist/darwin-arm64/bookget-macos-arm64`
- Other platforms: Please refer to the original repository

## Building

```bash
# For macOS M-series
go build -trimpath -ldflags "-s -w" -o dist/darwin-arm64/bookget-macos-arm64 ./cmd/
```

## License

This project maintains the original **GPLv3 License**. All modifications are also licensed under GPLv3.

See [LICENSE](LICENSE) for details.
