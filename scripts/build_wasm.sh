#!/usr/bin/env bash
# Build the solver for the browser and copy Go's loader shim next to it.
#
# Both outputs are build artifacts, not sources: solver.wasm is compiled from
# cmd/wasmsolver, and wasm_exec.js must match the toolchain that produced it, so
# copying it from GOROOT is the only way to keep the two in step. Both are
# gitignored and rebuilt by the Pages workflow.
set -euo pipefail

cd "$(dirname "$0")/.."

# -s -w drops the symbol table and DWARF info. The visitor downloads this.
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/solver.wasm ./cmd/wasmsolver

shim="$(go env GOROOT)/lib/wasm/wasm_exec.js"
[ -f "$shim" ] || shim="$(go env GOROOT)/misc/wasm/wasm_exec.js"
cp "$shim" web/wasm_exec.js

# Actual file size and the compressed size a visitor downloads. du would report
# disk blocks, which understates both.
report() {
  bytes=$(wc -c < "$1")
  gz=$(gzip -9 -c "$1" | wc -c)
  printf '%-18s %6.2f MB   %4.0f KB gzipped\n' "$1" "$(echo "$bytes" | awk '{print $1/1048576}')" "$(echo "$gz" | awk '{print $1/1024}')"
}
report web/solver.wasm
report web/wasm_exec.js
