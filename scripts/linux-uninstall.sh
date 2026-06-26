#!/bin/bash
# ==============================================================================
# Water AI — Linux Uninstaller Script
# ==============================================================================

set -euo pipefail

APP_NAME="Water"
APP_ID="ai.water.app"
INSTALL_DIR="/opt/${APP_NAME}"
ICON_DIR="${HOME}/.local/share/icons/hicolor"
DESKTOP_DIR="${HOME}/.local/share/applications"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

info "Uninstalling ${APP_NAME}..."

# Remove desktop entry
if [ -f "${DESKTOP_DIR}/${APP_ID}.desktop" ]; then
    rm -f "${DESKTOP_DIR}/${APP_ID}.desktop"
    info "Removed desktop entry"
fi

# Remove icons
for size_dir in 48x48 128x128 256x256 scalable; do
    rm -f "${ICON_DIR}/${size_dir}/apps/${APP_ID}.png" 2>/dev/null || true
done

# Update icon cache
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
    gtk-update-icon-cache -f -t "${ICON_DIR}" 2>/dev/null || true
fi
info "Removed icons"

# Remove symlink
SYMLINK_PATH="/usr/local/bin/water"
if [ -L "${SYMLINK_PATH}" ]; then
    if [ -w "$(dirname "${SYMLINK_PATH}")" ]; then
        rm -f "${SYMLINK_PATH}"
    else
        sudo rm -f "${SYMLINK_PATH}" 2>/dev/null || true
    fi
    info "Removed symlink"
fi

# Remove install directory
if [ -d "${INSTALL_DIR}" ]; then
    if [ -w "$(dirname "${INSTALL_DIR}")" ]; then
        rm -rf "${INSTALL_DIR}"
    else
        sudo rm -rf "${INSTALL_DIR}"
    fi
    info "Removed ${INSTALL_DIR}"
fi

# Update desktop database
if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${DESKTOP_DIR}" 2>/dev/null || true
fi

info "${APP_NAME} has been uninstalled."
