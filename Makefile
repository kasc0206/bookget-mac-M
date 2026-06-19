# 编译配置
GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w
TARGET ?= bookget
DIST_DIR ?= dist
PREFIX ?= $(HOME)/.local

MACOS_ARM64_DIR := $(DIST_DIR)/darwin-arm64
MACOS_ARM64_BIN := $(MACOS_ARM64_DIR)/bookget-macos-arm64
MACOS_ARM64_PACKAGE := $(DIST_DIR)/bookget-macos-arm64.tar.gz

# 构建目标
.PHONY: build
build:
	@echo "Building $(TARGET) for $(GOOS)-$(GOARCH)"
	@mkdir -p $(DIST_DIR)/$(GOOS)-$(GOARCH)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(GOOS)-$(GOARCH)/$(TARGET)$(SUFFIX) ./cmd/

# 跨平台构建（调用示例）
.PHONY: release
release: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: macos-arm64
macos-arm64: darwin-arm64

.PHONY: package-macos-arm64
package-macos-arm64: macos-arm64
	@echo "Packaging $(MACOS_ARM64_PACKAGE)"
	@cp README-FORK.md $(MACOS_ARM64_DIR)/README-FORK.md
	@cp LICENSE $(MACOS_ARM64_DIR)/LICENSE
	@tar -czf $(MACOS_ARM64_PACKAGE) -C $(MACOS_ARM64_DIR) .

.PHONY: install-macos-arm64
install-macos-arm64: macos-arm64
	@mkdir -p $(PREFIX)/bin
	@cp $(MACOS_ARM64_BIN) $(PREFIX)/bin/bookget
	@chmod +x $(PREFIX)/bin/bookget
	@echo "Installed $(PREFIX)/bin/bookget"

linux-amd64:
	@$(MAKE) build GOOS=linux GOARCH=amd64 SUFFIX=-linux

linux-arm64:
	@$(MAKE) build GOOS=linux GOARCH=arm64 SUFFIX=-linux-arm64

darwin-amd64:
	@$(MAKE) build GOOS=darwin GOARCH=amd64 SUFFIX=-macos

darwin-arm64:
	@$(MAKE) build GOOS=darwin GOARCH=arm64 SUFFIX=-macos-arm64

windows-amd64:
	@$(MAKE) build GOOS=windows GOARCH=amd64 SUFFIX=.exe

# 清理
.PHONY: clean
clean:
	@rm -rf $(DIST_DIR)
