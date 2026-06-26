#!/bin/bash
# ==============================================================================
# Water AI — Linux Launcher Script
# ==============================================================================
#
# This script launches the Water binary, automatically falling back to Mesa's
# software OpenGL renderer (llvmpipe/swrast) if the system lacks a hardware
# OpenGL driver.
#
# Directory layout expected (works both in bundle and installed location):
#   Water              — this launcher script (renamed from water-launcher.sh)
#   bin/Water          — the actual binary
#   lib/               — bundled Mesa software rendering libraries
#     libGL.so.1
#     libEGL.so.1
#     libGLX.so.0
#     swrast_dri.so    (or other Mesa DRI drivers)
#   icons/             — application icons
#     logo.png
#     logo-only.png
#     vscode.png
#
# ==============================================================================

set -euo pipefail

# Resolve the real directory of this script (handles symlinks)
SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")")" && pwd)"

BINARY="${SCRIPT_DIR}/bin/Water"
MESA_LIB_DIR="${SCRIPT_DIR}/lib"

if [ ! -x "$BINARY" ]; then
    echo "ERROR: Water binary not found at $BINARY" >&2
    exit 1
fi

# --- Attempt 1: Try launching with the system's native OpenGL driver ---------
# We do a quick check: try to run the binary normally. If it exits immediately
# with a GL-related error, we fall back to the bundled Mesa software renderer.

# First, check if the system has a working libGL at all
HAS_SYSTEM_GL=true
if ! ldconfig -p 2>/dev/null | grep -q 'libGL\.so'; then
    HAS_SYSTEM_GL=false
fi

if [ "$HAS_SYSTEM_GL" = true ]; then
    # System has libGL — try running normally
    exec "$BINARY" "$@"
else
    # --- Attempt 2: Fall back to bundled Mesa software renderer ---------------
    echo "INFO: No system OpenGL driver detected. Using bundled Mesa software renderer." >&2

    if [ ! -d "$MESA_LIB_DIR" ]; then
        echo "ERROR: Bundled Mesa libraries not found at $MESA_LIB_DIR" >&2
        echo "       Please ensure the lib/ directory is present alongside this launcher." >&2
        exit 1
    fi

    export LD_LIBRARY_PATH="${MESA_LIB_DIR}${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
    export LIBGL_ALWAYS_SOFTWARE=1
    export LIBGL_DRIVERS_PATH="${MESA_LIB_DIR}/dri"
    export GALLIUM_DRIVER=llvmpipe

    exec "$BINARY" "$@"
fi
