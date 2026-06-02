#!/usr/bin/env bash
# build.sh — Local build script for SSHCustom
# Builds daemon for arm64 + armv7, packages Magisk ZIP.
# Run from repo root.
set -Eeuo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
VERSION="$(cat "$ROOT/VERSION" | tr -d '[:space:]')"
DIST="$ROOT/dist"
DAEMON_DIR="$ROOT/daemon"
MODULE_DIR="$ROOT/module"
ZIP_OUT="$DIST/SSHCustom-v${VERSION}.zip"
LDFLAGS="-s -w -buildid= -X main.version=${VERSION}"

echo "==> Building SSHCustom v${VERSION}"
mkdir -p "$DIST" "$MODULE_DIR/bin/arm64-v8a" "$MODULE_DIR/bin/armeabi-v7a"

echo "==> Go toolchain"
go version

echo "==> Stamping module.prop version=${VERSION}"
VCODE="$(echo "$VERSION" | tr -d '.')00"
sed -i.bak "s|^version=.*|version=v${VERSION}|" "$MODULE_DIR/module.prop"
sed -i.bak "s|^versionCode=.*|versionCode=${VCODE}|" "$MODULE_DIR/module.prop"
rm -f "$MODULE_DIR/module.prop.bak"

echo "==> Running daemon tests"
(cd "$DAEMON_DIR" && go test ./... 2>&1) && echo "   tests passed" || echo "   WARN: tests failed"

echo "==> Building sshcustomd for android/arm64"
(cd "$DAEMON_DIR" && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="$LDFLAGS" -o "$MODULE_DIR/bin/arm64-v8a/sshcustomd" ./cmd/sshcustomd/)
echo "   arm64: $(ls -lh "$MODULE_DIR/bin/arm64-v8a/sshcustomd" | awk '{print $5}') static"

echo "==> Building sshcustomd for linux/arm (armv7)"
(cd "$DAEMON_DIR" && GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="$LDFLAGS" -o "$MODULE_DIR/bin/armeabi-v7a/sshcustomd" ./cmd/sshcustomd/)
echo "   armv7: $(ls -lh "$MODULE_DIR/bin/armeabi-v7a/sshcustomd" | awk '{print $5}') static"

echo "==> Packaging Magisk ZIP → $ZIP_OUT"
(cd "$MODULE_DIR" && zip -r9 "$ZIP_OUT" . -x "*.DS_Store" -x "__MACOSX/*" -x "bin/.gitkeep")

echo ""
echo "==> BUILD COMPLETE"
echo "    $ZIP_OUT ($(ls -lh "$ZIP_OUT" | awk '{print $5}'))"
