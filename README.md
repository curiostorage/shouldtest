# shouldtest

Skip expensive CI test jobs when a pull request does not touch the dependency closure of a Go package.

## Usage

From a Go module root (with `.github/shouldtest.json` and per-package `shouldtest.json`):

```bash
go run github.com/curiostorage/shouldtest/cmd/shouldtest --ci --verbose ./lib/proof
```

Flags: `--ci`, `--verbose`, `--json`, `--coverage-dir`.

## Configuration

- **Repository:** `.github/shouldtest.json` — default branch, workflows to scan, `ciInvoke` string for registry checks.
- **Package:** `<dir>/shouldtest.json` — `buildTags`, `extraPaths`, `coverageFile`.

## Library

Import `github.com/curiostorage/shouldtest` for `Check`, `RunCI`, and profile helpers. Use `github.com/curiostorage/shouldtest/registry` to validate that every package profile is referenced in CI workflows.
