#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$ROOT_DIR/web"

mkdir -p "$WEB_DIR"

echo "=== Ensuring wasm32-unknown-unknown target is installed ==="
if command -v rustup >/dev/null 2>&1; then
    rustup target add wasm32-unknown-unknown || true
fi

echo "=== Building Rust WebAssembly Binary (wasm32-unknown-unknown) ==="
cd "$ROOT_DIR"
cargo build --release --target wasm32-unknown-unknown --bin missile-command
cp "$ROOT_DIR/target/wasm32-unknown-unknown/release/missile-command.wasm" "$WEB_DIR/missile-command.wasm"

if [ ! -f "$WEB_DIR/mq_js_bundle.js" ]; then
    CARGO_HOME_DIR="${CARGO_HOME:-$HOME/.cargo}"
    MQ_JS="$(find "$CARGO_HOME_DIR/registry/src" -name "mq_js_bundle.js" 2>/dev/null | head -n 1 || true)"
    if [ -n "$MQ_JS" ] && [ -f "$MQ_JS" ]; then
        cp "$MQ_JS" "$WEB_DIR/mq_js_bundle.js"
    else
        curl -sSL "https://not-fl3.github.io/miniquad-samples/mq_js_bundle.js" -o "$WEB_DIR/mq_js_bundle.js" || true
    fi
fi

echo "=== WASM Build Complete ==="
ls -lh "$WEB_DIR/missile-command.wasm" "$WEB_DIR/mq_js_bundle.js" "$WEB_DIR/index.html"
echo ""
echo "To run locally:"
echo "  python3 -m http.server 8080 -d web"
echo "Then navigate to http://localhost:8080"
