---
name: unittest
description: Run unit tests.
---

# Unit Test

1. Ensure GNU Make is installed by executing `make --version`
2. Ensure go compiler is installed by executing `go version`
3. Ensure redis server is available by calling `redis-cli ping` - it should return `PONG`.
4. Run unit tests via `make test`
5. Show test results to operator
6. Try to fix test failures by editing the code
7. Make diff via `git diff`
8. Show differences to operator
