#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
mkdir -p "$DIR/web/static"

echo "Compiling Gorithm WebAssembly client..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "$DIR/web/static/app.wasm" "$DIR/web/wasm"
echo "WebAssembly binary built successfully: web/static/app.wasm ($(du -h "$DIR/web/static/app.wasm" | cut -f1))"
