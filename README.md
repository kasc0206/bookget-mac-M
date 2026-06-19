# bookget macOS Apple Silicon fork

`bookget` is a digital ancient-book downloader that supports many library and IIIF sources. This repository is a fork of [deweizhu/bookget](https://github.com/deweizhu/bookget), tuned for MacBook and Mac desktop machines with Apple Silicon chips.

For the original project documentation and wiki, see:

- Original repository: [deweizhu/bookget](https://github.com/deweizhu/bookget)
- Original wiki: [bookget wiki](https://github.com/deweizhu/bookget/wiki)

## Quick Start On MacBook M-Series

Install Go:

```bash
brew install go
```

Build:

```bash
make macos-arm64
```

Run:

```bash
./dist/darwin-arm64/bookget-macos-arm64 --help
```

Install as `bookget`:

```bash
make install-macos-arm64
```

Use it:

```bash
bookget -O "$HOME/Downloads/bookget" 'https://example.com/book/url'
```

## Fork Features

- Apple Silicon binary target: `dist/darwin-arm64/bookget-macos-arm64`
- GitHub Actions release package: `bookget-macos-arm64.tar.gz`
- readable automatic filenames for selected sites
- upstream `bookget` downloader support retained
- `archive.org` and IIIF workflows available from upstream

More details are in [README-FORK.md](README-FORK.md).

## Common Commands

```bash
make macos-arm64
make install-macos-arm64
make package-macos-arm64
go test ./...
```

## Notes

Some downloaders require cookies, custom headers, login, or access from an allowed region. This project is for learning and research use; please respect each library's terms and copyright rules.

## License

GPLv3, same as the original project.
