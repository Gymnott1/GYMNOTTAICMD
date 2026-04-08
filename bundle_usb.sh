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
info "Detecting mounted USB drives..."

# Collect mounted removable drives
USB_LIST=()
while IFS= read -r line; do
    USB_LIST+=("$line")
done < <(lsblk -o NAME,TRAN,MOUNTPOINT -nr | awk '$2=="usb" && $3!="" {print $3}')

if [ ${#USB_LIST[@]} -eq 0 ]; then
    echo "No USB drives detected. Please enter path manually:"
    read -rp "Mount path: " USB_PATH
elif [ ${#USB_LIST[@]} -eq 1 ]; then
    USB_PATH="${USB_LIST[0]}"
    ok "Auto-detected USB: $USB_PATH"
else
    echo "Multiple USB drives found:"
    for i in "${!USB_LIST[@]}"; do
        echo "  [$i] ${USB_LIST[$i]}"
    done
    read -rp "Select number: " idx
    USB_PATH="${USB_LIST[$idx]}"
fi

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
