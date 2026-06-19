# Bookget for macOS Apple Silicon

This fork of [deweizhu/bookget](https://github.com/deweizhu/bookget) is maintained for MacBook and Mac desktop machines with Apple Silicon chips: M1, M2, M3, M4, and newer.

The focus is a small, dependable command-line build for `darwin/arm64`, plus fork-specific filename improvements for several digital library downloaders.

## What This Fork Adds

- macOS Apple Silicon build target: `make macos-arm64`
- local install target: `make install-macos-arm64`
- release package target: `make package-macos-arm64`
- automatic, readable output names for selected sites
- upstream updates through `deweizhu/bookget` main, including archive.org support and Chinese-path handling

## Build On A MacBook M-Series

Install Go with Homebrew:

```bash
brew install go
```

Build the binary:

```bash
make macos-arm64
```

The binary is written to:

```text
dist/darwin-arm64/bookget-macos-arm64
```

Install it into `~/.local/bin/bookget`:

```bash
make install-macos-arm64
```

If `~/.local/bin` is not already on your PATH, add this to your shell profile:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Then restart the terminal or run:

```bash
source ~/.zshrc
```

## Usage

Download one URL:

```bash
bookget 'https://example.com/book/url'
```

Choose an output directory:

```bash
bookget -O "$HOME/Downloads/bookget" 'https://example.com/book/url'
```

Download a page range:

```bash
bookget -p 1:20 'https://example.com/book/url'
```

Use IIIF manifest mode:

```bash
bookget -m 2 'https://example.com/iiif/manifest.json'
```

Show all options:

```bash
bookget --help
```

## Release Package

Create a tarball for GitHub Releases:

```bash
make package-macos-arm64
```

The package is written to:

```text
dist/bookget-macos-arm64.tar.gz
```

## Auto-Naming Enhancements

This fork adds a shared filename helper in `app/template.go`:

- `NormalizeNamePart`
- `BuildOutputFileName`
- `ExtractHTMLTitle`

The helper is used by selected downloaders such as NLC, Wzlib, University of Tokyo, Luoyang Library, Guangzhou Library, and Seoul National University.

## Platform Scope

Supported by this fork:

- macOS 11+ on Apple Silicon
- Go 1.23.x
- CLI usage from Terminal

Not the focus of this fork:

- Windows GUI builds
- Intel Mac release artifacts
- Linux release artifacts

For those platforms, use the original project unless you plan to maintain the extra builds yourself.

## Validation

Run the local checks:

```bash
go test ./...
make macos-arm64
```

## License

This fork keeps the original GPLv3 license. See `LICENSE`.
