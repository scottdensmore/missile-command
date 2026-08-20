#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$ROOT_DIR/web"

mkdir -p "$WEB_DIR"

echo "=== Locating Go wasm_exec.js ==="
GOROOT="$(go env GOROOT)"
WASM_EXEC=""

if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    WASM_EXEC="$GOROOT/lib/wasm/wasm_exec.js"
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    WASM_EXEC="$GOROOT/misc/wasm/wasm_exec.js"
fi

if [ -n "$WASM_EXEC" ]; then
    echo "Copying $WASM_EXEC to $WEB_DIR/wasm_exec.js"
    cp "$WASM_EXEC" "$WEB_DIR/wasm_exec.js"
else
    echo "Warning: wasm_exec.js not found in GOROOT ($GOROOT). Attempting download from standard Go repository..."
    curl -sSL "https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js" -o "$WEB_DIR/wasm_exec.js" || true
fi

echo "=== Building WebAssembly Binary (GOOS=js GOARCH=wasm) ==="
cd "$ROOT_DIR"
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$WEB_DIR/missile-command.wasm" .

echo "=== WASM Build Complete ==="
ls -lh "$WEB_DIR/missile-command.wasm" "$WEB_DIR/wasm_exec.js" "$WEB_DIR/index.html"
echo ""
echo "To run locally:"
echo "  python3 -m http.server 8080 -d web"
echo "Then navigate to http://localhost:8080"
