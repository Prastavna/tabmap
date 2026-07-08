#!/usr/bin/env bash
# Removes the per-user TabMap desktop integration installed by install.sh.
set -euo pipefail

APP_NAME="tabmap"

rm -f "$HOME/.local/bin/$APP_NAME"
rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/$APP_NAME.png"
rm -f "$HOME/.local/share/applications/$APP_NAME.desktop"

command -v update-desktop-database >/dev/null 2>&1 && update-desktop-database "$HOME/.local/share/applications" || true
command -v gtk-update-icon-cache >/dev/null 2>&1 && gtk-update-icon-cache -f -t "$HOME/.local/share/icons/hicolor" >/dev/null 2>&1 || true

echo "Removed TabMap desktop entry."
