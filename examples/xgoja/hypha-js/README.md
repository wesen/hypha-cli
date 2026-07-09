# hypha-js (xgoja-generated binary)

This directory holds the `xgoja.yaml` that generates `hypha-js`, a standalone
binary that exposes `require("hypha")` (the fluent builder JS module) via the
built-in `run`/`eval`/`repl` commands.

## Doctor

```bash
xgoja doctor -f examples/xgoja/hypha-js/xgoja.yaml
```

## Build

Both `go-go-goja` (v0.10.1) and `hypha-cli` (v0.1.0) are tagged, so the
build needs **no local replaces** — just tell xgoja the go-go-goja version to
require (the default `v0.0.0` is a placeholder that does not resolve):

```bash
xgoja build -f examples/xgoja/hypha-js/xgoja.yaml --xgoja-version v0.10.1
```

This resolves `github.com/go-go-golems/hypha-cli@v0.1.0` and
`github.com/go-go-golems/go-go-goja@v0.10.1` from the Go module proxy and
emits `dist/hypha-js`. (If you are developing hypha-cli or go-go-goja locally
and want uncommitted changes, use `--xgoja-replace <path>` instead of
`--xgoja-version`.)

Type declarations work the same way:

```bash
xgoja gen-dts -f examples/xgoja/hypha-js/xgoja.yaml --out hypha.d.ts --xgoja-version v0.10.1
```

## Run

```bash
HYPHA_BASE_URL=https://hyphahypha.club HYPHA_PAT=hh_pat_... \
  ./hypha-js run examples/xgoja/hypha-js/scripts/smoke.js
```
