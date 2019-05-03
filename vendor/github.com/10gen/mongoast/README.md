# mongoast

mongoast represents a MongoDB aggregation framework pipeline as an abstract syntax tree (AST). It supports parsing query pipelines from BSON, deparsing into BSON, and a visitor implementation.

## Getting Started

Once the project is cloned, just do `go run build.go` from the root directory to list the tasks available. `go run build.go verify` will run unit tests and static analysis.

### Requirements

* Go 1.11
