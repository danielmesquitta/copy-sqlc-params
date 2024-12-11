# Copy SQLC Params

Copy SQLC params to a different package

I like using the repository pattern, but unfortunately, in Go, you often have to duplicate code to maintain clean architecture. Fortunately, with the [jinzhu/copier](https://github.com/jinzhu/copier) package and this CLI code generation tool, I can streamline this process effectively. If you're curious about my approach, feel free to explore my [api-finance-manager](https://github.com/danielmesquitta/api-finance-manager) repository to see the implementation in action.

## Installation

```bash
go install github.com/danielmesquitta/copy-sqlc-params@latest
```

## Usage

```bash
copy-sqlc-params --input ./path/to/sqlc/out --output ./path/to/output/dir
```
