#!/bin/bash
# autostart.sh - install gymnott_ai to run on every login

BINARY_PATH="$(cd "$(dirname "$0")" && pwd)/gymnott_ai"
AUTOSTART_DIR="$HOME/.config/autostart"
DESKTOP_FILE="$AUTOSTART_DIR/gymnott_ai.desktop"
ENV_FILE="$HOME/.config/gymnott_ai.env"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: binary not found at $BINARY_PATH"
    echo "Run: go build -o gymnott_ai . first"
    exit 1
fi

mkdir -p "$AUTOSTART_DIR"

# Prompt for API key if not already saved
if [ ! -f "$ENV_FILE" ]; then
    read -rp "Enter your GROQ_API_KEY: " apikey
    echo "GROQ_API_KEY=$apikey" > "$ENV_FILE"
    chmod 600 "$ENV_FILE"
    echo "Saved to $ENV_FILE"
fi

cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=Gymnott AI
Comment=AI assistant with mouse follower
Exec=bash -c 'source $ENV_FILE && $BINARY_PATH'
Icon=utilities-terminal
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
EOF

echo "✅ Autostart installed: $DESKTOP_FILE"
echo "   Will launch on next login."
echo ""
echo "To start NOW without rebooting:"
echo "   source $ENV_FILE && $BINARY_PATH &"
echo ""
echo "To remove autostart:"
echo "   rm $DESKTOP_FILE"
