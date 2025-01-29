# go-repomap
**Version**: v0.0.1

A Go implementation of repository mapping functionality using tree-sitter for code analysis. This tool helps analyze and understand codebases by creating a graph representation of code definitions and references.

## Credits

This project uses the following open-source libraries:

- [tree-sitter](https://tree-sitter.github.io/tree-sitter/) - A parser generator tool and incremental parsing library
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter) - Golang bindings for tree-sitter and language grammars

Special thanks to the contributors of these projects for their excellent work.

## Features

- Code analysis using tree-sitter
- Support for multiple programming languages:
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
- Tag-based code navigation
- Graph-based code analysis
- Ranked tag generation

## Installation

```bash
go get github.com/entrepeneur4lyf/go-repomap
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/entrepeneur4lyf/go-repomap"
)

func main() {
    // Create a new tag index for your repository
    tagIndex := repomap.NewTagIndex("/path/to/repo")

    // Create an analyzer
    analyzer := repomap.NewTagAnalyzer(tagIndex)

    // Get ranked tags
    rankedTags := analyzer.GetRankedTags()
    fmt.Printf("Found %d tags\n", len(rankedTags))

    // Create a repo map with custom token limit
    rm := repomap.NewRepoMap().WithMapTokens(2048)
    repomap, err := rm.GetRepoMap(tagIndex)
    if err != nil {
        panic(err)
    }

    fmt.Println(repomap)
}
```

The generated map provides a concise overview of your codebase's structure, highlighting important code definitions and their relationships. The output is optimized to stay within the specified token limit while preserving the most relevant information.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.