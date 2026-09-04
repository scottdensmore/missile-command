#!/bin/bash
set -euo pipefail

APP_NAME="Missile Command"
APP_DIR="${APP_NAME}.app"
CONTENTS_DIR="${APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

VERSION="${1:-1.1.0}"
BIN_SRC="${2:-}"

echo "📦 Bundling ${APP_NAME} v${VERSION} for macOS..."

rm -rf "${APP_DIR}"
mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

# Build or copy binary
if [ -n "${BIN_SRC}" ]; then
    echo "  📋 Installing binary from ${BIN_SRC}..."
    cp "${BIN_SRC}" "${MACOS_DIR}/missile-command"
else
    echo "  🔨 Compiling binary..."
    go build -ldflags="-s -w" -o "${MACOS_DIR}/missile-command" .
fi

# Ensure icon.icns exists
if [ ! -f "assets/icon.icns" ]; then
    echo "  🎨 Generating assets/icon.icns..."
    go run scripts/generate-icons.go
fi

# Copy icon
echo "  🎨 Copying icon..."
cp assets/icon.icns "${RESOURCES_DIR}/icon.icns"

# Create Info.plist
echo "  📝 Creating Info.plist..."
cat <<EOF > "${CONTENTS_DIR}/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>com.scottdensmore.missile-command</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleExecutable</key>
    <string>missile-command</string>
    <key>CFBundleIconFile</key>
    <string>icon.icns</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
</dict>
</plist>
EOF

chmod +x "${MACOS_DIR}/missile-command"
echo "✅ Successfully created '${APP_DIR}'!"
