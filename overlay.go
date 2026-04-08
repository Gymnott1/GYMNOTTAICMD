package main

// overlay.go - always-on-top floating input + response window

import (
	"os"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var overlayWin *gtk.Window

var prefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_prefs"

func loadScreenshotPref() bool {
	data, err := os.ReadFile(prefsFile)
	if err != nil {
		return true // default: on
	}
	return strings.TrimSpace(string(data)) == "1"
}

func saveScreenshotPref(v bool) {
	val := "0"
	if v {
		val = "1"
	}
	os.WriteFile(prefsFile, []byte(val), 0644)
}

func showOverlay() {
	if overlayWin != nil {
		overlayWin.Present()
		return
	}

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("AI Assistant")
	win.SetDefaultSize(520, 460)
	win.SetKeepAbove(true)
	win.SetDecorated(true)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)

	mx, my := getMousePos()
	win.Move(mx+30, my+30)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(8)
	box.SetMarginEnd(8)

	// Input
	inputView, _ := gtk.TextViewNew()
	inputView.SetWrapMode(gtk.WRAP_WORD_CHAR)
	inputView.SetAcceptsTab(false)
	inputView.SetSizeRequest(-1, 60)
	inputScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	inputScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	inputScroll.Add(inputView)

	// Response
	responseView, _ := gtk.TextViewNew()
	responseView.SetEditable(false)
	responseView.SetWrapMode(gtk.WRAP_WORD_CHAR)
	responseView.SetSizeRequest(-1, 300)
	buf, _ := responseView.GetBuffer()
	t := makeTags(buf)
	responseScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	responseScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	responseScroll.Add(responseView)

	// Bottom bar: checkbox + status + send button
	bottomBar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)

	screenshotCheck, _ := gtk.CheckButtonNewWithLabel("📸 Send screenshot")
	screenshotCheck.SetActive(loadScreenshotPref())
	screenshotCheck.Connect("toggled", func() {
		saveScreenshotPref(screenshotCheck.GetActive())
	})

	cropCheck, _ := gtk.CheckButtonNewWithLabel("✂ Crop")
	cropCheck.SetActive(false)
	cropCheck.SetSensitive(loadScreenshotPref())
	screenshotCheck.Connect("toggled", func() {
		active := screenshotCheck.GetActive()
		saveScreenshotPref(active)
		cropCheck.SetSensitive(active)
		if !active {
			cropCheck.SetActive(false)
		}
	})

	statusLabel, _ := gtk.LabelNew("Enter to send · Shift+Enter newline · Esc hide")
	statusLabel.SetXAlign(0)
	statusLabel.SetHExpand(true)

	sendBtn, _ := gtk.ButtonNewWithLabel("Ask ↵")
	newChatBtn, _ := gtk.ButtonNewWithLabel("🗑 New Chat")

	bottomBar.PackStart(screenshotCheck, false, false, 0)
	bottomBar.PackStart(cropCheck, false, false, 0)
	bottomBar.PackStart(statusLabel, true, true, 0)
	bottomBar.PackEnd(sendBtn, false, false, 0)
	bottomBar.PackEnd(newChatBtn, false, false, 0)

	sendFn := func() {
		ibuf, _ := inputView.GetBuffer()
		start, end := ibuf.GetBounds()
		query, _ := ibuf.GetText(start, end, false)
		if strings.TrimSpace(query) == "" {
			return
		}
		ibuf.SetText("")
		withShot := screenshotCheck.GetActive()
		crop := cropCheck.GetActive()
		if withShot {
			if crop {
				statusLabel.SetText("Select area to capture…")
			} else {
				statusLabel.SetText("Taking screenshot…")
			}
		} else {
			statusLabel.SetText("Asking AI…")
		}
		sendBtn.SetSensitive(false)
		setWaiting(true)

		go func() {
			if withShot {
				scheduleOnMain(func() { win.Hide() })
				time.Sleep(300 * time.Millisecond)
			}
			response := askAI(query, withShot, crop)
			scheduleOnMain(func() {
				setWaiting(false)
				win.ShowAll()
				win.Present()
				renderMarkdown(buf, t, responseView, response)
				statusLabel.SetText("Done · Enter to ask again")
				sendBtn.SetSensitive(true)
				adj := responseScroll.GetVAdjustment()
				adj.SetValue(adj.GetUpper())
			})
		}()
	}

	newChatBtn.Connect("clicked", func() {
		clearHistory()
		buf.SetText("")
		statusLabel.SetText("New chat started · Enter to send")
	})

	sendBtn.Connect("clicked", sendFn)

	inputView.Connect("key-press-event", func(_ *gtk.TextView, ev *gdk.Event) bool {
		keyEv := gdk.EventKeyNewFromEvent(ev)
		if keyEv.KeyVal() == gdk.KEY_Return && (keyEv.State()&uint(gdk.SHIFT_MASK)) == 0 {
			sendFn()
			return true
		}
		return false
	})

	win.Connect("key-press-event", func(_ *gtk.Window, ev *gdk.Event) bool {
		keyEv := gdk.EventKeyNewFromEvent(ev)
		if keyEv.KeyVal() == gdk.KEY_Escape {
			win.Hide()
			return true
		}
		return false
	})

	win.Connect("delete-event", func() bool {
		win.Hide()
		return true
	})

	box.PackStart(inputScroll, false, false, 0)
	box.PackStart(bottomBar, false, false, 0)
	box.PackStart(responseScroll, true, true, 0)
	win.Add(box)
	win.ShowAll()

	overlayWin = win
}
