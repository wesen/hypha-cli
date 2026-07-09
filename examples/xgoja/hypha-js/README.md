# hypha-js (xgoja-generated binary)

This directory holds the `xgoja.yaml` that generates `hypha-js`, a standalone
binary that exposes `require("hypha")` (the fluent builder JS module) via the
built-in `run`/`eval`/`repl` commands.

## Doctor

```bash
xgoja doctor -f examples/xgoja/hypha-js/xgoja.yaml
```

## Build

The generated module needs a local replace for `go-go-goja` during
development (go-go-goja is not yet tagged). Use `--xgoja-replace`:

```bash
xgoja build -f examples/xgoja/hypha-js/xgoja.yaml \
  --xgoja-replace /home/manuel/code/wesen/go-go-golems/go-go-goja
```

`hypha-cli` itself is published as `v0.1.0` (tagged on
`go-go-golems/hypha-cli` main), so no local replace is needed for it —
`xgoja build` resolves `github.com/go-go-golems/hypha-cli@v0.1.0` from the
Go module proxy. (If you are developing hypha-cli itself and want uncommitted
changes, add `go mod edit -replace github.com/go-go-golems/hypha-cli=<path>`
in the generated workspace.)

## Run

```bash
HYPHA_BASE_URL=https://hyphahypha.club HYPHA_PAT=hh_pat_... \
  ./hypha-js run examples/xgoja/hypha-js/scripts/smoke.js
```
