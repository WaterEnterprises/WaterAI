#!/bin/bash
# ==============================================================================
# Water AI — Linux Installer Script
# ==============================================================================
#
# This script is executed by the makeself .run self-extracting installer.
# It installs the Water AI application with full desktop integration:
#   - Binary + launcher + Mesa fallback libs → ~/.local/share/water-ai/
#   - Icons → ~/.local/share/icons/hicolor/*/apps/
#   - .desktop file → ~/.local/share/applications/
#   - Symlink → ~/.local/bin/water
#
# ==============================================================================

set -euo pipefail

APP_NAME="Water"
APP_ID="ai.water.app"
APP_DISPLAY_NAME="Water AI"
APP_COMMENT="Water AI Assistant"
APP_CATEGORIES="Development;Utility;"

# XDG base directories (with sensible defaults)
XDG_DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
LOCAL_BIN="${HOME}/.local/bin"

INSTALL_DIR="${XDG_DATA_HOME}/water-ai"
ICONS_DIR="${XDG_DATA_HOME}/icons/hicolor"
APPLICATIONS_DIR="${XDG_DATA_HOME}/applications"

# The directory where makeself extracted us
EXTRACT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              Water AI — Linux Installer                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Installing to: ${INSTALL_DIR}"
echo ""

# --- Validate extracted contents ---------------------------------------------
echo "--> Validating installer contents..."

REQUIRED_FILES=(
    "bin/${APP_NAME}"
    "${APP_NAME}"
    "assets/logo.png"
    "assets/logo-only.png"
    "assets/vscode.png"
)

for f in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "${EXTRACT_DIR}/${f}" ]; then
        echo "ERROR: Required file missing from installer: ${f}" >&2
        exit 1
    fi
done

echo "    All required files present."

# --- Create directories ------------------------------------------------------
echo "--> Creating directories..."
mkdir -p "${INSTALL_DIR}/bin"
mkdir -p "${INSTALL_DIR}/lib/dri"
mkdir -p "${INSTALL_DIR}/assets"
mkdir -p "${LOCAL_BIN}"
mkdir -p "${APPLICATIONS_DIR}"
mkdir -p "${ICONS_DIR}/256x256/apps"
mkdir -p "${ICONS_DIR}/128x128/apps"
mkdir -p "${ICONS_DIR}/64x64/apps"
mkdir -p "${ICONS_DIR}/48x48/apps"
mkdir -p "${ICONS_DIR}/scalable/apps"

# --- Install application files -----------------------------------------------
echo "--> Installing application files..."

# Binary
cp "${EXTRACT_DIR}/bin/${APP_NAME}" "${INSTALL_DIR}/bin/${APP_NAME}"
chmod +x "${INSTALL_DIR}/bin/${APP_NAME}"

# Launcher script (with Mesa fallback)
cp "${EXTRACT_DIR}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
chmod +x "${INSTALL_DIR}/${APP_NAME}"

# Mesa libraries (if present)
if [ -d "${EXTRACT_DIR}/lib" ]; then
    cp -a "${EXTRACT_DIR}/lib/"* "${INSTALL_DIR}/lib/" 2>/dev/null || true
fi

# Assets
cp "${EXTRACT_DIR}/assets/"* "${INSTALL_DIR}/assets/"

echo "    Application files installed."

# --- Install icons -----------------------------------------------------------
echo "--> Installing icons..."

# Use logo-only.png as the app icon (it's the icon without text, best for menus)
ICON_SRC="${EXTRACT_DIR}/assets/logo-only.png"

# Install at multiple sizes for proper desktop integration
# We install the same PNG at all sizes; the desktop environment will scale as needed
for size in 256x256 128x128 64x64 48x48; do
    cp "${ICON_SRC}" "${ICONS_DIR}/${size}/apps/${APP_ID}.png"
    echo "    Installed icon: ${ICONS_DIR}/${size}/apps/${APP_ID}.png"
done

# Also install as scalable (PNG works here too)
cp "${ICON_SRC}" "${ICONS_DIR}/scalable/apps/${APP_ID}.png"
echo "    Installed icon: ${ICONS_DIR}/scalable/apps/${APP_ID}.png"

# --- Create .desktop file ----------------------------------------------------
echo "--> Creating desktop entry..."

DESKTOP_FILE="${APPLICATIONS_DIR}/${APP_ID}.desktop"

cat > "${DESKTOP_FILE}" <<EOF
[Desktop Entry]
Type=Application
Name=${APP_DISPLAY_NAME}
Comment=${APP_COMMENT}
Exec=${INSTALL_DIR}/${APP_NAME} %U
Icon=${APP_ID}
Terminal=false
Categories=${APP_CATEGORIES}
StartupWMClass=${APP_NAME}
StartupNotify=true
EOF

chmod +x "${DESKTOP_FILE}"
echo "    Desktop entry: ${DESKTOP_FILE}"

# --- Create symlink in ~/.local/bin ------------------------------------------
echo "--> Creating command-line symlink..."

ln -sf "${INSTALL_DIR}/${APP_NAME}" "${LOCAL_BIN}/water"
echo "    Symlink: ${LOCAL_BIN}/water -> ${INSTALL_DIR}/${APP_NAME}"

# --- Update icon cache (if available) ----------------------------------------
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
    echo "--> Updating icon cache..."
    gtk-update-icon-cache -f -t "${ICONS_DIR}" 2>/dev/null || true
fi

# --- Update desktop database (if available) -----------------------------------
if command -v update-desktop-database >/dev/null 2>&1; then
    echo "--> Updating desktop database..."
    update-desktop-database "${APPLICATIONS_DIR}" 2>/dev/null || true
fi

# --- Done ---------------------------------------------------------------------
echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              Installation Complete!                         ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║                                                            ║"
echo "║  Launch from:                                              ║"
echo "║    • Application menu (search for 'Water AI')              ║"
echo "║    • Terminal: water                                       ║"
echo "║    • Direct: ${INSTALL_DIR}/${APP_NAME}"
echo "║                                                            ║"
echo "║  To uninstall:                                             ║"
echo "║    rm -rf ${INSTALL_DIR}"
echo "║    rm -f ${LOCAL_BIN}/water"
echo "║    rm -f ${DESKTOP_FILE}"
echo "║    rm -f ${ICONS_DIR}/*/apps/${APP_ID}.png"
echo "║                                                            ║"
echo "╚══════════════════════════════════════════════════════════════╝"
