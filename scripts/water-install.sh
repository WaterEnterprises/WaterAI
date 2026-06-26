#!/bin/bash
# ==============================================================================
# Water AI — Linux Installer Script
# ==============================================================================
#
# This script is executed by the makeself .run self-extracting installer.
# It installs the Water application to the user's home directory following
# XDG conventions:
#
#   ~/.local/bin/Water              — main launcher script
#   ~/.local/bin/Water.bin          — actual binary
#   ~/.local/lib/water/             — bundled Mesa libraries (fallback)
#   ~/.local/share/icons/water.png  — application icon
#   ~/.local/share/applications/water.desktop — desktop entry
#
# ==============================================================================

set -euo pipefail

APP_NAME="Water"
APP_ID="ai.water.app"

# --- Resolve install paths (XDG-compliant, user-local) -----------------------
BIN_DIR="${HOME}/.local/bin"
LIB_DIR="${HOME}/.local/lib/water"
ICON_DIR="${HOME}/.local/share/icons"
DESKTOP_DIR="${HOME}/.local/share/applications"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              WATER AI — LINUX INSTALLER                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Installing ${APP_NAME} to:"
echo "  Binary:   ${BIN_DIR}/${APP_NAME}"
echo "  Libs:     ${LIB_DIR}/"
echo "  Icon:     ${ICON_DIR}/water.png"
echo "  Desktop:  ${DESKTOP_DIR}/water.desktop"
echo ""

# --- Create directories ------------------------------------------------------
mkdir -p "${BIN_DIR}" "${LIB_DIR}/dri" "${ICON_DIR}" "${DESKTOP_DIR}"

# --- Install binary -----------------------------------------------------------
echo "--> Installing binary..."
cp bin/${APP_NAME} "${BIN_DIR}/${APP_NAME}.bin"
chmod +x "${BIN_DIR}/${APP_NAME}.bin"

# --- Install launcher script --------------------------------------------------
echo "--> Installing launcher..."
# Rewrite the launcher to point to the installed paths
cat > "${BIN_DIR}/${APP_NAME}" << 'LAUNCHER_EOF'
#!/bin/bash
# Water AI — Installed Launcher
set -euo pipefail

BIN_DIR="${HOME}/.local/bin"
LIB_DIR="${HOME}/.local/lib/water"
BINARY="${BIN_DIR}/Water.bin"

if [ ! -x "$BINARY" ]; then
    echo "ERROR: Water binary not found at $BINARY" >&2
    exit 1
fi

# Check if the system has a working libGL
HAS_SYSTEM_GL=true
if ! ldconfig -p 2>/dev/null | grep -q 'libGL\.so'; then
    HAS_SYSTEM_GL=false
fi

if [ "$HAS_SYSTEM_GL" = true ]; then
    exec "$BINARY" "$@"
else
    echo "INFO: No system OpenGL driver detected. Using bundled Mesa software renderer." >&2
    if [ ! -d "$LIB_DIR" ]; then
        echo "ERROR: Bundled Mesa libraries not found at $LIB_DIR" >&2
        exit 1
    fi
    export LD_LIBRARY_PATH="${LIB_DIR}${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
    export LIBGL_ALWAYS_SOFTWARE=1
    export LIBGL_DRIVERS_PATH="${LIB_DIR}/dri"
    export GALLIUM_DRIVER=llvmpipe
    exec "$BINARY" "$@"
fi
LAUNCHER_EOF
chmod +x "${BIN_DIR}/${APP_NAME}"

# --- Install Mesa fallback libraries ------------------------------------------
echo "--> Installing Mesa fallback libraries..."
if [ -d lib ] && [ "$(ls -A lib 2>/dev/null)" ]; then
    cp -a lib/* "${LIB_DIR}/"
    echo "    Mesa libraries installed to ${LIB_DIR}/"
else
    echo "    No Mesa libraries found in payload (skipping)"
fi

# --- Install icon -------------------------------------------------------------
echo "--> Installing icon..."
if [ -f icon.png ]; then
    cp icon.png "${ICON_DIR}/water.png"
    echo "    Icon installed to ${ICON_DIR}/water.png"
else
    echo "    WARN: icon.png not found in payload"
fi

# --- Install desktop entry ----------------------------------------------------
echo "--> Installing desktop entry..."
if [ -f water.desktop ]; then
    # Substitute actual paths into the desktop file
    sed -e "s|@BIN_DIR@|${BIN_DIR}|g" \
        -e "s|@ICON_DIR@|${ICON_DIR}|g" \
        water.desktop > "${DESKTOP_DIR}/water.desktop"
    chmod +x "${DESKTOP_DIR}/water.desktop"
    echo "    Desktop entry installed to ${DESKTOP_DIR}/water.desktop"
else
    echo "    WARN: water.desktop not found in payload"
fi

# --- Update desktop database (if available) -----------------------------------
if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${DESKTOP_DIR}" 2>/dev/null || true
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              INSTALLATION COMPLETE!                         ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║                                                            ║"
echo "║  Run Water from the terminal:                              ║"
echo "║    $ Water                                                 ║"
echo "║                                                            ║"
echo "║  Or find it in your application launcher.                  ║"
echo "║                                                            ║"
echo "║  Make sure ~/.local/bin is in your PATH:                   ║"
echo "║    export PATH=\"\$HOME/.local/bin:\$PATH\"                    ║"
echo "║                                                            ║"
echo "╚══════════════════════════════════════════════════════════════╝"
