# go-repomap v0.0.1

Initial release of go-repomap, a Go library for analyzing and visualizing code repositories using tree-sitter.

## Features

- Code analysis using tree-sitter for accurate parsing
- Support for 26 programming languages including:
  - Bash
  - C/C++
  - C#
  - CSS
  - Dockerfile
  - Elixir
  - Elm
  - Go
  - Groovy
  - HCL
  - HTML
  - Java
  - JavaScript/TypeScript
  - Kotlin
  - OCaml
  - PHP
  - Protocol Buffers
  - Python
  - Ruby
  - Rust
  - Scala
  - SQL
  - Svelte
  - Swift
  - TOML
  - YAML
  - Add your own grammars here!
- Tag-based code navigation
- Graph-based code analysis with PageRank-style importance ranking
- Configurable output size with token limiting
- Memory-efficient incremental parsing

## Dependencies

- Uses tree-sitter for parsing (MIT License)
- Uses go-tree-sitter for Go bindings (MIT License)

## Getting Started

```go
go get github.com/entrepeneur4lyf/go-repomap@v0.0.1
```

See the [README.md](https://github.com/entrepeneur4lyf/go-repomap#usage) for usage examples.

## Platform Support

This library is tested and supported on:
- Linux
- macOS
- Windows

The library uses Go's standard `path/filepath` package for cross-platform path handling and `os` package for file system operations, ensuring compatibility across different operating systems. All file paths are normalized using `filepath.Clean` and `filepath.ToSlash` for consistent behavior.

## Requirements

- Go 1.21 or later
- C compiler (for tree-sitter support)