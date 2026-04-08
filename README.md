# GYMNOTT AI — Linux Desktop AI Assistant

> Ask AI anything on your screen with a single hotkey. Powered by Groq's ultra-fast LLM inference.

![Linux](https://img.shields.io/badge/Linux-Mint%20%2F%20Ubuntu-green)
![Go](https://img.shields.io/badge/Go-1.22+-blue)
![Groq](https://img.shields.io/badge/AI-Groq%20API-orange)
![License](https://img.shields.io/badge/license-MIT-blue)
![Stars](https://img.shields.io/github/stars/Gymnott1/GYMNOTTAICMD?style=social)

A lightweight, always-on Linux desktop AI assistant built in Go. Press `Ctrl+Space` from anywhere on your desktop to instantly ask AI about what's on your screen — errors, configs, code, logs, anything.

---

## ✨ Features

- 🔴 **Visual mouse follower** — a small red dot that shadows your real cursor (click-through, always on top), morphs into a **spinning loader** while the AI is thinking
- ⚡ **Global hotkey** — `Ctrl+Space` opens the AI overlay instantly from any app
- 📸 **Screenshot mode** — hides the overlay, captures the active window, and sends it to the AI vision model
- ✂️ **Crop mode** — drag to select a specific region of the screen instead of the full window
- 💬 **Multi-turn chat** — full conversation history kept across messages in the same session
- 🗑️ **New Chat** — wipe history and start a completely fresh conversation
- 📝 **Markdown rendering** — headings, bold, italic, inline code, fenced code blocks with a **📋 Copy** button per block
- 💾 **Persistent preferences** — screenshot checkbox state saved and restored across restarts
- 🚀 **Auto-start on login** — one command installs it as a desktop autostart entry

---

## 🖥️ Demo

```
Ctrl+Space  →  type your question  →  Enter
```

The AI sees your screen and responds with commands, configs, and explanations — no copy-pasting errors into a browser.

---

## 📦 Install

```bash
# 1. Install system dependencies
sudo apt-get install -y libgtk-3-dev libx11-dev libxtst-dev libxext-dev scrot xdotool slop

# 2. Install Go (if not already installed)
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 3. Clone and build
git clone https://github.com/Gymnott1/GYMNOTTAICMD.git
cd GYMNOTTAICMD
go mod tidy
go build -o gymnott_ai .
```

Or use the one-command install script:

```bash
./install.sh
```

---

## 🔑 Setup

Get a **free** API key from [console.groq.com](https://console.groq.com), then:

```bash
export GROQ_API_KEY=gsk_your_key_here
./gymnott_ai
```

To make the key permanent:

```bash
echo 'export GROQ_API_KEY=gsk_your_key_here' >> ~/.bashrc
source ~/.bashrc
```

---

## 🚀 Auto-start on Login

```bash
./autostart.sh
```

Saves your API key to `~/.config/gymnott_ai.env` (chmod 600) and installs a `.desktop` entry in `~/.config/autostart/` — launches automatically on every login.

```bash
# Start immediately without rebooting
source ~/.config/gymnott_ai.env && ./gymnott_ai &

# Remove autostart
rm ~/.config/autostart/gymnott_ai.desktop
```

---

## ⌨️ Usage

| Hotkey | Action |
|--------|--------|
| `Ctrl+Space` | Open / focus AI overlay |
| `Enter` | Send message |
| `Shift+Enter` | New line in input |
| `Escape` | Hide overlay |

### Screenshot options

| Checkbox | Behaviour |
|----------|-----------|
| 📸 Send screenshot | Hides the overlay, captures the active window, sends it with your message |
| ✂️ Crop | Drag-select a region instead of the full window (requires 📸 checked) |

### Chat history

The AI remembers previous messages in the same session. Click **🗑 New Chat** to start fresh with no prior context.

---

## 🏗️ Architecture

| File | Purpose |
|------|---------|
| `main.go` | GTK init, starts goroutines |
| `follower.go` | Transparent X11 window — red dot / spinning loader |
| `hotkey.go` | XGrabKey global `Ctrl+Space` listener |
| `xutil.go` | CGo X11 mouse position + GTK main-thread scheduler |
| `overlay.go` | Always-on-top input/response GTK window |
| `ai.go` | Screenshot, Groq API via curl, markdown renderer, chat history |

---

## 🤖 Models

| Mode | Model |
|------|-------|
| Vision (screenshot) | `meta-llama/llama-4-scout-17b-16e-instruct` |
| Text only | `llama-3.3-70b-versatile` |

Both served by [Groq](https://groq.com) — the fastest LLM inference available.

---

## 📚 Dependencies

- [gotk3](https://github.com/gotk3/gotk3) — GTK3 bindings for Go
- [scrot](https://github.com/resurrecting-open-source-projects/scrot) — screenshot capture
- [slop](https://github.com/naelstrof/slop) — interactive screen region selection
- [Groq API](https://console.groq.com) — LLM inference (free tier available)

---

## 🔍 Keywords

`linux ai assistant` · `desktop ai linux` · `groq linux` · `ai hotkey linux` · `screenshot ai linux` · `llm desktop app` · `linux ai overlay` · `ctrl space ai` · `groq api linux app` · `linux copilot` · `ai terminal helper` · `linux ai tool` · `open source ai desktop`

---

## 📄 License

MIT — free to use, modify, and distribute.

---

> Built for Linux power users who want AI at their fingertips without leaving their workflow.
