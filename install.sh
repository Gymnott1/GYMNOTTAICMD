#!/bin/bash
set -e

echo "=== Installing system dependencies ==="
sudo apt-get install -y libgtk-3-dev libx11-dev libxtst-dev libxext-dev scrot xdotool slop

echo "=== Installing Go (if not present) ==="
if ! command -v go &>/dev/null; then
    GO_VER="1.22.4"
    wget -q "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    echo "Go installed. Run: source ~/.bashrc"
fi

echo "=== Fetching Go modules ==="
go mod tidy

echo "=== Building ==="
go build -o gymnott_ai .

echo ""
echo "Done! Run with:"
echo "  export GROQ_API_KEY=your_key_here"
echo "  ./gymnott_ai"
echo ""
echo "Hotkey: Ctrl+Space  → opens AI overlay"
echo "Escape              → hides overlay"
