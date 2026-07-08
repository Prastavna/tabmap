#!/usr/bin/env bash
# Per-user desktop integration for TabMap (no root required).
# Installs the built binary, a .desktop launcher and an icon into ~/.local
# so TabMap shows up in your application launcher/dock with its icon.
set -euo pipefail

APP_NAME="tabmap"
DISPLAY_NAME="TabMap"

# Resolve repo root (this script lives in <repo>/scripts).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

BIN_SRC="$REPO_ROOT/build/bin/$APP_NAME"
ICON_SRC="$REPO_ROOT/frontend/public/favicon.svg"

if [[ ! -x "$BIN_SRC" ]]; then
  echo "Error: built binary not found at $BIN_SRC" >&2
  echo "Build it first with: wails build" >&2
  exit 1
fi

BIN_DIR="$HOME/.local/bin"
ICON_DIR="$HOME/.local/share/icons/hicolor/512x512/apps"
DESKTOP_DIR="$HOME/.local/share/applications"

mkdir -p "$BIN_DIR" "$ICON_DIR" "$DESKTOP_DIR"

install -m 755 "$BIN_SRC" "$BIN_DIR/$APP_NAME"

if [[ -f "$ICON_SRC" ]]; then
  install -m 644 "$ICON_SRC" "$ICON_DIR/$APP_NAME.png"
fi

cat > "$DESKTOP_DIR/$APP_NAME.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=$DISPLAY_NAME
Comment=Configure graphics tablet buttons and touch ring
Exec=$BIN_DIR/$APP_NAME
Icon=$APP_NAME
Terminal=false
Categories=Utility;Settings;
EOF

# Refresh caches when the tools are available (ignore failures).
command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$DESKTOP_DIR" || true
command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor" >/dev/null 2>&1 || true

echo "Installed $DISPLAY_NAME. It should now appear in your application launcher."
