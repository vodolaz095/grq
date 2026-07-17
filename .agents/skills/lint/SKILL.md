---
name: lint
description: Run linter and fix lint errors.
---

# Lint

1. Ensure GNU Make is installed by executing `make --version`
2. Ensure gofmt, golint and go compiler are installed by executing `gofmt --version`, `golint --version` and `go version` respectively
3. Run linter via `make lint`
4. Try to fix lint errors by editing the code
5. Make diff via `git diff`
6. Show differences to operator
