#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# Gymnott AI — Portable Installer
# Copy this entire folder to a USB. Run ./install.sh on any Linux machine.
# ─────────────────────────────────────────────────────────────────────────────
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="$HOME/.local/share/gymnott_ai"
ENV_FILE="$HOME/.config/gymnott_ai.env"
SERVICE_FILE="$HOME/.config/systemd/user/gymnott_ai.service"
AUTOSTART_FILE="$HOME/.config/autostart/gymnott_ai.desktop"
GO_VERSION="1.22.4"
GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="https://go.dev/dl/${GO_TAR}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✅ $1${NC}"; }
info() { echo -e "${YELLOW}➜  $1${NC}"; }
err()  { echo -e "${RED}❌ $1${NC}"; exit 1; }

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   GYMNOTT AI — Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ── Detect OS ────────────────────────────────────────────────────────────────
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)
info "Detected OS: $OS"

# ── Install system dependencies ───────────────────────────────────────────────
info "Installing system dependencies..."
case "$OS" in
    ubuntu|debian|linuxmint|pop)
        if ! sudo apt-get update -qq; then
            info "apt-get update failed (likely due to a third-party repo key issue)."
            info "Continuing with existing package index and attempting dependency install..."
        fi
        sudo apt-get install -y libgtk-3-dev libx11-dev libxtst-dev libxext-dev scrot xdotool slop wget curl tesseract-ocr
        ;;
    fedora)
        sudo dnf install -y gtk3-devel libX11-devel libXtst-devel libXext-devel scrot xdotool slop wget curl tesseract
        ;;
    arch|manjaro)
        sudo pacman -Sy --noconfirm gtk3 libx11 libxtst libxext scrot xdotool slop wget curl tesseract
        ;;
    opensuse*|sles)
        sudo zypper install -y gtk3-devel libX11-devel libXtst-devel libXext-devel scrot xdotool wget curl tesseract-ocr
        ;;
    *)
        info "Unknown OS — attempting apt-get (may fail)"
        sudo apt-get install -y libgtk-3-dev libx11-dev libxtst-dev libxext-dev scrot xdotool slop wget curl tesseract-ocr || true
        ;;
esac
ok "System dependencies installed"

# ── Install Go ────────────────────────────────────────────────────────────────
if command -v go &>/dev/null; then
    ok "Go already installed: $(go version)"
else
    info "Installing Go ${GO_VERSION}..."
    # Check if we have the tarball on the USB first
    if [ -f "$SCRIPT_DIR/${GO_TAR}" ]; then
        info "Using bundled Go tarball from USB"
        sudo tar -C /usr/local -xzf "$SCRIPT_DIR/${GO_TAR}"
    else
        info "Downloading Go..."
        wget -q --show-progress "$GO_URL" -O "/tmp/${GO_TAR}"
        sudo tar -C /usr/local -xzf "/tmp/${GO_TAR}"
        rm "/tmp/${GO_TAR}"
    fi

    # Add to PATH for this session and permanently
    export PATH=$PATH:/usr/local/go/bin
    grep -q '/usr/local/go/bin' ~/.bashrc || echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    grep -q '/usr/local/go/bin' ~/.profile || echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    ok "Go ${GO_VERSION} installed"
fi

export PATH=$PATH:/usr/local/go/bin

# ── Copy files to install dir ─────────────────────────────────────────────────
info "Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR"/*.go "$INSTALL_DIR/"
cp "$SCRIPT_DIR/go.mod" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/go.sum" "$INSTALL_DIR/" 2>/dev/null || true
ok "Files copied"

# ── Build ─────────────────────────────────────────────────────────────────────
info "Building gymnott_ai..."
cd "$INSTALL_DIR"
go mod tidy
go build -o gymnott_ai .
ok "Build successful"

# ── API Key ───────────────────────────────────────────────────────────────────
if [ -f "$SCRIPT_DIR/gymnott_ai.env" ]; then
    # Bundled key from USB
    cp "$SCRIPT_DIR/gymnott_ai.env" "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    ok "API key loaded from USB"
elif [ -f "$ENV_FILE" ]; then
    ok "API key already exists at $ENV_FILE"
else
    echo ""
    read -rp "Enter your GROQ_API_KEY (get free at console.groq.com): " apikey
    echo "GROQ_API_KEY=$apikey" > "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    ok "API key saved"
fi

if ! grep -q '^GEMINI_API_KEY=' "$ENV_FILE"; then
    echo ""
    read -rp "Enter your GEMINI_API_KEY (optional, press Enter to skip): " gemini_key
    if [ -n "$gemini_key" ]; then
        echo "GEMINI_API_KEY=$gemini_key" >> "$ENV_FILE"
        chmod 600 "$ENV_FILE"
        ok "Gemini API key saved"
    fi
fi

# Load key into current shell and persist
grep -q 'gymnott_ai.env' ~/.bashrc  || echo '[ -f ~/.config/gymnott_ai.env ] && export $(cat ~/.config/gymnott_ai.env | xargs)' >> ~/.bashrc
grep -q 'gymnott_ai.env' ~/.profile || echo '[ -f ~/.config/gymnott_ai.env ] && export $(cat ~/.config/gymnott_ai.env | xargs)' >> ~/.profile

# ── Systemd user service ──────────────────────────────────────────────────────
info "Installing systemd user service..."
mkdir -p "$(dirname "$SERVICE_FILE")"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Gymnott AI Desktop Assistant
After=default.target

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
Environment="DISPLAY=:0"
Environment="XAUTHORITY=${HOME}/.Xauthority"
ExecStart=${INSTALL_DIR}/gymnott_ai
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable gymnott_ai.service
systemctl --user start gymnott_ai.service
loginctl enable-linger "$USER" 2>/dev/null || true
ok "Systemd service enabled and started"

# ── XDG Autostart (fallback for non-systemd DEs) ─────────────────────────────
mkdir -p "$(dirname "$AUTOSTART_FILE")"
cat > "$AUTOSTART_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=Gymnott AI
Comment=AI assistant with mouse follower
Exec=bash -c 'source ${ENV_FILE} && ${INSTALL_DIR}/gymnott_ai'
Icon=utilities-terminal
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF
ok "Autostart entry installed"

# ── Done ──────────────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}  Gymnott AI installed and running!${NC}"
echo ""
echo "  Hotkey:   Ctrl+Space"
echo "  Manage:   systemctl --user status gymnott_ai"
echo "  Logs:     journalctl --user -u gymnott_ai -f"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
