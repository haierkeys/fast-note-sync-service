#!/usr/bin/env bash
# Build fast-note-sync-service into a fnOS .fpk package.
#
# Builds the Go binary from source (frontend is embedded via //go:embed),
# stamps the manifest version from internal/app/version.go, downloads the
# official fnpack tool, and produces a .fpk for the requested arch.
#
# Usage:  packaging/fnos/build.sh [amd64|arm64]   (default: amd64)
# Run from the repository root.
set -eu

ARCH="${1:-amd64}"
case "$ARCH" in
  amd64) PLATFORM=x86;  GOARCH=amd64 ;;
  arm64) PLATFORM=arm;  GOARCH=arm64 ;;
  *) echo "unknown arch: $ARCH (want amd64|arm64)" >&2; exit 1 ;;
esac

PKG_DIR="packaging/fnos"
BIN_NAME="fast-note-sync-service"
FNPACK_VERSION="1.2.3"
FNPACK_SHA256="54b97fa7b70968c4d05c79840f5daeff508957d0bb2062fdb0376d00d9615c93"

# Resolve version from source (same field the release process uses).
VERSION=$(grep -E 'Version\s+string' internal/app/version.go | awk -F '"' '{print $2}')
if [ -z "${VERSION}" ]; then
  echo "could not read Version from internal/app/version.go" >&2
  exit 1
fi
echo ">>> version=${VERSION} arch=${ARCH} platform=${PLATFORM}"

# 1. Build the binary from source (CGO disabled -> static, portable).
echo ">>> building binary"
mkdir -p "${PKG_DIR}/app"
CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH}" go build -trimpath \
  -o "${PKG_DIR}/app/${BIN_NAME}" .

# 2. Stamp manifest (version + platform).
echo ">>> stamping manifest"
sed -i "s|^version=.*|version=\"${VERSION}\"|" "${PKG_DIR}/manifest"
sed -i "s|^platform=.*|platform=\"${PLATFORM}\"|"     "${PKG_DIR}/manifest"

# 3. Download fnpack (pinned by SHA256).
echo ">>> downloading fnpack"
curl -fL --retry 3 -o "${PKG_DIR}/fnpack" \
  "https://static2.fnnas.com/fnpack/fnpack-${FNPACK_VERSION}-linux-amd64"
echo "${FNPACK_SHA256}  ${PKG_DIR}/fnpack" | sha256sum -c -
chmod +x "${PKG_DIR}/fnpack"

# 4. Package.
echo ">>> building .fpk"
( cd "${PKG_DIR}" && ./fnpack build --directory . )

# 5. Rename + checksum.
OUT="${PKG_DIR}/fastnotesync-${VERSION}-${PLATFORM}.fpk"
mv "${PKG_DIR}/fastnotesync.fpk" "${OUT}"
( cd "${PKG_DIR}" && sha256sum "$(basename "${OUT}")" > "$(basename "${OUT}").sha256" )

# 6. Clean intermediate tool binary (keep the app binary out of git too).
rm -f "${PKG_DIR}/fnpack"

echo ">>> done: ${OUT}"
