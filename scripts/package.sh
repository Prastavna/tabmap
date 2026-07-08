#!/usr/bin/env bash
# Build TabMap and produce a distributable Linux package with nfpm.
#
# Usage: ./scripts/package.sh [deb|rpm]   (default: deb)
#
# The package installs the binary to /usr/bin, registers a .desktop entry and
# installs the icon into the system icon theme, so TabMap appears in the
# application launcher with its icon.
set -euo pipefail

PACKAGER="${1:-deb}"
case "$PACKAGER" in
  deb|rpm) ;;
  *) echo "Error: unknown packager '$PACKAGER' (expected 'deb' or 'rpm')" >&2; exit 1 ;;
esac

command -v nfpm >/dev/null 2>&1 || {
  echo "Error: nfpm not found. Install it with:" >&2
  echo "  go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest" >&2
  exit 1
}

APP_NAME="tabmap"
DISPLAY_NAME="TabMap"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# Version from the latest git tag (strip a leading "v"), fall back to 0.0.0.
VERSION="$(git describe --tags 2>/dev/null | sed 's/^v//')"
VERSION="${VERSION:-0.0.0}"
ARCH="$(go env GOARCH)"

# Runtime dependency differs per distro family.
if [[ "$PACKAGER" == "deb" ]]; then
  WACOM_DEP="xserver-xorg-input-wacom"
else
  WACOM_DEP="xorg-x11-drv-wacom"
fi

echo "Building $DISPLAY_NAME $VERSION ($ARCH) ..."
wails build -clean -tags webkit2_41

BIN="$REPO_ROOT/build/bin/$APP_NAME"
[[ -x "$BIN" ]] || { echo "Error: build did not produce $BIN" >&2; exit 1; }

ICON_SRC="$REPO_ROOT/frontend/public/favicon.svg"

# Stage the generated .desktop entry.
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

cat > "$STAGE/$APP_NAME.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=$DISPLAY_NAME
Comment=Configure graphics tablet buttons and touch ring
Exec=$APP_NAME
Icon=$APP_NAME
Terminal=false
Categories=Utility;Settings;
EOF

# Generate the nfpm config.
cat > "$STAGE/nfpm.yaml" <<EOF
name: $APP_NAME
arch: $ARCH
version: "$VERSION"
maintainer: "Sahil Patel <ssahillppatell@gmail.com>"
description: Configure graphics tablet buttons and touch ring with xsetwacom.
homepage: https://github.com/Prastavna/tabmap
license: MIT
section: utils
priority: optional
depends:
  - $WACOM_DEP
contents:
  - src: $BIN
    dst: /usr/bin/$APP_NAME
  - src: $STAGE/$APP_NAME.desktop
    dst: /usr/share/applications/$APP_NAME.desktop
  - src: $ICON_SRC
    dst: /usr/share/icons/hicolor/512x512/apps/$APP_NAME.png
EOF

echo "Packaging ($PACKAGER) ..."
nfpm package --config "$STAGE/nfpm.yaml" --packager "$PACKAGER" --target "$REPO_ROOT/build/bin/"

echo "Done. Package written to build/bin/"
ls -1 "$REPO_ROOT/build/bin/"*."$PACKAGER"
