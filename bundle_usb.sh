#!/bin/bash
# bundle_usb.sh — copy Gymnott AI to a USB drive, ready to install on any machine
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
GO_VERSION="1.22.4"
GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✅ $1${NC}"; }
info() { echo -e "${YELLOW}➜  $1${NC}"; }

# ── Find USB ──────────────────────────────────────────────────────────────────
echo ""
echo "Available removable drives:"
lsblk -o NAME,SIZE,MOUNTPOINT,LABEL | grep -v "^loop"
echo ""
read -rp "Enter USB mount path (e.g. /media/gymnott/MYUSB): " USB_PATH

[ -d "$USB_PATH" ] || { echo "Path not found: $USB_PATH"; exit 1; }

DEST="$USB_PATH/gymnott_ai"
mkdir -p "$DEST"

# ── Copy source files ─────────────────────────────────────────────────────────
info "Copying source files..."
cp "$SCRIPT_DIR"/*.go "$DEST/"
cp "$SCRIPT_DIR/go.mod" "$DEST/"
cp "$SCRIPT_DIR/go.sum" "$DEST/" 2>/dev/null || true
cp "$SCRIPT_DIR/install.sh" "$DEST/"
cp "$SCRIPT_DIR/README.md" "$DEST/" 2>/dev/null || true
chmod +x "$DEST/install.sh"
ok "Source files copied"

# ── Bundle API key ────────────────────────────────────────────────────────────
if [ -f "$HOME/.config/gymnott_ai.env" ]; then
    cp "$HOME/.config/gymnott_ai.env" "$DEST/gymnott_ai.env"
    ok "API key bundled"
else
    info "No API key found — you'll be prompted on install"
fi

# ── Bundle Go tarball (for offline install) ───────────────────────────────────
if [ -f "/tmp/${GO_TAR}" ]; then
    info "Using cached Go tarball..."
    cp "/tmp/${GO_TAR}" "$DEST/"
elif [ -f "/usr/local/go/bin/go" ]; then
    info "Downloading Go tarball for offline use..."
    wget -q --show-progress "https://go.dev/dl/${GO_TAR}" -O "$DEST/${GO_TAR}"
else
    info "Downloading Go tarball for offline use..."
    wget -q --show-progress "https://go.dev/dl/${GO_TAR}" -O "$DEST/${GO_TAR}"
fi
ok "Go tarball bundled"

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}  USB ready at: $DEST${NC}"
echo ""
echo "  On the new machine:"
echo "    cd $DEST"  
echo "    chmod +x install.sh && ./install.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
