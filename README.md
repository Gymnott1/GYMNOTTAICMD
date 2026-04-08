# Gymnott AI

A Linux desktop AI assistant with a visual mouse follower. Press `Ctrl+Space` anywhere to ask AI about anything on your screen.

![Linux](https://img.shields.io/badge/Linux-Mint%20%2F%20Ubuntu-green)
![Go](https://img.shields.io/badge/Go-1.22+-blue)
![Groq](https://img.shields.io/badge/AI-Groq%20API-orange)

## Features

- 🖱️ **Mouse follower** — small red arrow cursor that shadows your real cursor (click-through, always on top), turns into a **spinning loader** while waiting for AI response
- 🤖 **Global hotkey** — `Ctrl+Space` opens the AI overlay from anywhere
- 📸 **Screenshot mode** — captures the active window and sends it to the AI (vision model)
- ✂️ **Crop mode** — drag to select a specific area of the screen to send instead of the full window
- 💬 **Multi-turn chat** — conversation history is kept across messages in the same session
- 🗑️ **New Chat** — clear history and start a fresh conversation instantly
- 📝 **Markdown rendering** — headings, bold, italic, inline code, fenced code blocks with a **📋 Copy button** per block
- ⌨️ `Enter` to send, `Shift+Enter` for newlines, `Escape` to hide
- 💾 **Persistent preferences** — screenshot checkbox state saved across restarts
- 🚀 **Auto-start on login** — one command to install as a desktop autostart entry

## Install

```bash
# 1. Install system dependencies
sudo apt-get install -y libgtk-3-dev libx11-dev libxtst-dev libxext-dev scrot xdotool slop

# 2. Install Go (if not already installed)
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 3. Clone and build
git clone https://github.com/gymnott/gymnott_ai.git
cd gymnott_ai
go mod tidy
go build -o gymnott_ai .
```

Or use the install script which does all of the above:

```bash
./install.sh
```

## Setup

Get a free API key from [console.groq.com](https://console.groq.com), then:

```bash
export GROQ_API_KEY=gsk_your_key_here
./gymnott_ai
```

To make the API key permanent:

```bash
echo 'export GROQ_API_KEY=gsk_your_key_here' >> ~/.bashrc
source ~/.bashrc
```

## Auto-start on Login

```bash
./autostart.sh
```

This saves your API key to `~/.config/gymnott_ai.env` (chmod 600) and installs a `.desktop` entry in `~/.config/autostart/` so it launches automatically on every login.

To start immediately without rebooting:

```bash
source ~/.config/gymnott_ai.env && ./gymnott_ai &
```

To remove autostart:

```bash
rm ~/.config/autostart/gymnott_ai.desktop
```

## Usage

Once running you'll see a small red arrow following your cursor.

| Hotkey | Action |
|--------|--------|
| `Ctrl+Space` | Open / focus AI overlay |
| `Enter` | Send message |
| `Shift+Enter` | New line in input |
| `Escape` | Hide overlay |

### Screenshot options

| Checkbox | Behaviour |
|----------|-----------|
| 📸 Send screenshot | Captures the active window behind the overlay and sends it with your message |
| ✂️ Crop | Lets you drag-select a region of the screen instead of the full window (requires 📸 to be checked) |

When screenshot is enabled the overlay hides itself before capturing so it never appears in the image.

### Chat history

Messages in the same session are sent as a conversation — the AI remembers previous context. Click **🗑 New Chat** to wipe history and start fresh.

## Architecture

| File | Purpose |
|------|---------|
| `main.go` | GTK init, starts goroutines |
| `follower.go` | Transparent X11 window with red arrow / spinning loader |
| `hotkey.go` | XGrabKey global `Ctrl+Space` listener |
| `xutil.go` | CGo X11 mouse position + GTK main-thread scheduler |
| `overlay.go` | Always-on-top input/response GTK window |
| `ai.go` | Screenshot, Groq API via curl, markdown renderer, chat history |

## Models

| Mode | Model |
|------|-------|
| Vision (screenshot) | `meta-llama/llama-4-scout-17b-16e-instruct` |
| Text only | `llama-3.3-70b-versatile` |

## Dependencies

- [gotk3](https://github.com/gotk3/gotk3) — GTK3 bindings for Go
- [scrot](https://github.com/resurrecting-open-source-projects/scrot) — screenshot tool
- [slop](https://github.com/naelstrof/slop) — screen region selection (for crop mode)
- [Groq API](https://console.groq.com) — LLM inference

## License

MIT
# GYMNOTTAICMD
