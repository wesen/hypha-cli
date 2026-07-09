# hypha-js (xgoja-generated binary)

This directory holds the `xgoja.yaml` that generates `hypha-js`, a standalone
binary that exposes `require("hypha")` (the fluent builder JS module) via the
built-in `run`/`eval`/`repl` commands.

## Doctor

```bash
xgoja doctor -f examples/xgoja/hypha-js/xgoja.yaml
```

## Build

The generated module needs local replaces for both `go-go-goja` and
`hypha-cli` during development (neither is on the public proxy with the
provider package yet). Use `--xgoja-replace` for go-go-goja, then add the
hypha-cli replace manually:

```bash
xgoja build -f examples/xgoja/hypha-js/xgoja.yaml \
  --xgoja-replace /home/manuel/code/wesen/go-go-golems/go-go-goja \
  --keep-work
# then in the printed build workspace:
cd <build-workspace>
go mod edit -replace github.com/go-go-golems/hypha-cli=/home/manuel/code/wesen/go-go-golems/hypha-cli
go mod tidy && go build -buildvcs=false .
./hypha-js run examples/xgoja/hypha-js/scripts/smoke.js
```

Once `hypha-cli` (with `pkg/xgoja/providers/hypha`) is published, the manual
replace is unnecessary and a plain `xgoja build` works.

## Run

```bash
HYPHA_BASE_URL=https://hyphahypha.club HYPHA_PAT=hh_pat_... \
  ./hypha-js run examples/xgoja/hypha-js/scripts/smoke.js
```
