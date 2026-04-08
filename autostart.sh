#!/bin/bash
set -e

BINARY="$(cd "$(dirname "$0")" && pwd)/gymnott_ai"
ENV_FILE="$HOME/.config/gymnott_ai.env"
SERVICE_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SERVICE_DIR/gymnott_ai.service"
AUTOSTART_DIR="$HOME/.config/autostart"
DESKTOP_FILE="$AUTOSTART_DIR/gymnott_ai.desktop"

# ── 1. Check binary ──────────────────────────────────────────────────────────
if [ ! -f "$BINARY" ]; then
    echo "Binary not found. Building..."
    export PATH=$PATH:/usr/local/go/bin
    go build -o gymnott_ai .
fi

# ── 2. Save API key ──────────────────────────────────────────────────────────
if [ ! -f "$ENV_FILE" ]; then
    read -rp "Enter your GROQ_API_KEY: " apikey
    echo "GROQ_API_KEY=$apikey" > "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    echo "✅ API key saved to $ENV_FILE"
else
    echo "✅ API key already saved at $ENV_FILE"
fi

# ── 3. Systemd user service ──────────────────────────────────────────────────
mkdir -p "$SERVICE_DIR"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Gymnott AI Desktop Assistant
After=graphical-session.target
PartOf=graphical-session.target

[Service]
Type=simple
EnvironmentFile=$ENV_FILE
ExecStart=$BINARY
Restart=on-failure
RestartSec=3

[Install]
WantedBy=graphical-session.target
EOF

systemctl --user daemon-reload
systemctl --user enable gymnott_ai.service
systemctl --user start gymnott_ai.service
echo "✅ Systemd user service enabled and started"

# ── 4. XDG autostart fallback (for Cinnamon session pickup) ─────────────────
mkdir -p "$AUTOSTART_DIR"
cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=Gymnott AI
Comment=AI assistant with mouse follower
Exec=bash -c 'source $ENV_FILE && $BINARY'
Icon=utilities-terminal
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF
echo "✅ Autostart desktop entry installed"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Gymnott AI is now running in the background"
echo "  It will auto-start on every login"
echo ""
echo "  Hotkey:  Ctrl+Space"
echo ""
echo "  Manage:"
echo "    systemctl --user status gymnott_ai"
echo "    systemctl --user stop gymnott_ai"
echo "    systemctl --user restart gymnott_ai"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
