#!/bin/sh
# Installs Resumator's desktop entry and icon so Linux docks/taskbars show
# the app icon. The window is matched to this entry via its app-id
# (StartupWMClass=resumator, set through ProgramName in main.go).
#
# Usage: ./install-linux.sh [PREFIX]
#   PREFIX defaults to /usr/local when writable, otherwise ~/.local
# The resumator binary itself must be on PATH (e.g. $PREFIX/bin/resumator).
set -eu

prefix=${1:-}
if [ -z "$prefix" ]; then
	if [ -w /usr/local/share/applications ] 2>/dev/null; then
		prefix=/usr/local
	else
		prefix="$HOME/.local"
	fi
fi

here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
icon_dir="$prefix/share/icons/hicolor/512x512/apps"
app_dir="$prefix/share/applications"

mkdir -p "$icon_dir" "$app_dir" "$prefix/bin"
install -m 0644 "$here/resumator.png" "$icon_dir/resumator.png"
install -m 0644 "$here/resumator.desktop" "$app_dir/resumator.desktop"

# Refresh caches when the tools exist; not fatal if they don't.
update-desktop-database "$app_dir" 2>/dev/null || true
gtk-update-icon-cache "$prefix/share/icons/hicolor" 2>/dev/null || true

echo "Installed desktop entry and icon under $prefix"
echo "If the dock does not pick it up immediately, log out and back in."
