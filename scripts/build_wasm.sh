#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
mkdir -p "$DIR/web/static"

echo "Compiling Gorithm WebAssembly client..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "$DIR/web/static/app.wasm" "$DIR/web/wasm"
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
  cp "$GOROOT/lib/wasm/wasm_exec.js" "$DIR/web/static/wasm_exec.js"
else
  cp "$GOROOT/misc/wasm/wasm_exec.js" "$DIR/web/static/wasm_exec.js"
fi
echo "WebAssembly binary built successfully: web/static/app.wasm ($(du -h "$DIR/web/static/app.wasm" | cut -f1))"
