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
var tooltipModeCheck *gtk.CheckButton
var cropCheckGlobal *gtk.CheckButton

func getTooltipMode() bool {
	if tooltipModeCheck == nil {
		return false
	}
	return tooltipModeCheck.GetActive()
}

func getScreenshotPrefs() (withShot, crop bool) {
	if cropCheckGlobal != nil {
		return loadScreenshotPref(), cropCheckGlobal.GetActive()
	}
	return loadScreenshotPref(), false
}

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

func applyCSS() {
	css, _ := gtk.CssProviderNew()
	css.LoadFromData(`
		window {
			background-color: #1a1a2e;
			border-radius: 12px;
			border: 1px solid #2a2a4a;
		}
		textview {
			background-color: #12122a;
			color: #e0e0f0;
			font-family: "JetBrains Mono", "Fira Code", monospace;
			font-size: 13px;
			border-radius: 8px;
			padding: 6px;
		}
		textview text {
			background-color: #12122a;
			color: #e0e0f0;
		}
		scrolledwindow {
			border-radius: 8px;
			border: 1px solid #2a2a4a;
		}
		button {
			background: linear-gradient(135deg, #6c63ff, #4a90d9);
			color: #ffffff;
			border: none;
			border-radius: 8px;
			padding: 6px 14px;
			font-weight: bold;
			font-size: 12px;
		}
		button:hover {
			background: linear-gradient(135deg, #7c73ff, #5aa0e9);
		}
		button:disabled {
			background: #2a2a4a;
			color: #555577;
		}
		checkbutton {
			color: #a0a0c0;
			font-size: 12px;
		}
		checkbutton:checked {
			color: #6c63ff;
		}
		label {
			color: #606080;
			font-size: 11px;
		}
	`)
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func showOverlay() {
	if overlayWin != nil {
		overlayWin.Present()
		return
	}

	applyCSS()

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Gymnott AI")
	win.SetDefaultSize(560, 500)
	win.SetKeepAbove(true)
	win.SetDecorated(true)
	win.SetTypeHint(gdk.WINDOW_TYPE_HINT_DIALOG)
	win.SetResizable(true)

	// Enable RGBA for rounded corners
	screen, _ := gdk.ScreenGetDefault()
	visual, _ := screen.GetRGBAVisual()
	if visual != nil {
		win.SetVisual(visual)
	}
	win.SetAppPaintable(true)

	mx, my := getMousePos()
	win.Move(mx+30, my+30)

	outer, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	// ── Title bar ──
	titleBar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	titleBar.SetMarginTop(10)
	titleBar.SetMarginBottom(6)
	titleBar.SetMarginStart(14)
	titleBar.SetMarginEnd(14)
	titleCss, _ := gtk.CssProviderNew()
	titleCss.LoadFromData(`box { background-color: #1a1a2e; }`)

	titleLbl, _ := gtk.LabelNew("✦ Gymnott AI")
	titleLbl.SetXAlign(0)
	titleLbl.SetHExpand(true)
	titleLblCss, _ := gtk.CssProviderNew()
	titleLblCss.LoadFromData(`label { color: #6c63ff; font-weight: bold; font-size: 13px; }`)
	titleLblCtx, _ := titleLbl.GetStyleContext()
	titleLblCtx.AddProvider(titleLblCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	newChatBtn, _ := gtk.ButtonNewWithLabel("🗑 New Chat")
	newChatBtnCss, _ := gtk.CssProviderNew()
	newChatBtnCss.LoadFromData(`button { background: #2a2a4a; color: #a0a0c0; font-size: 11px; padding: 4px 10px; }`)
	newChatBtnCtx, _ := newChatBtn.GetStyleContext()
	newChatBtnCtx.AddProvider(newChatBtnCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	titleBar.PackStart(titleLbl, true, true, 0)
	titleBar.PackEnd(newChatBtn, false, false, 0)

	// ── Input area ──
	inputView, _ := gtk.TextViewNew()
	inputView.SetWrapMode(gtk.WRAP_WORD_CHAR)
	inputView.SetAcceptsTab(false)
	inputView.SetSizeRequest(-1, 70)
	inputView.SetLeftMargin(8)
	inputView.SetRightMargin(8)
	inputView.SetTopMargin(6)
	inputView.SetBottomMargin(6)
	inputScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	inputScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	inputScroll.Add(inputView)
	inputScroll.SetMarginStart(12)
	inputScroll.SetMarginEnd(12)
	inputScroll.SetMarginBottom(6)

	// ── Options bar ──
	optionsBar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 12)
	optionsBar.SetMarginStart(14)
	optionsBar.SetMarginEnd(14)
	optionsBar.SetMarginBottom(6)

	screenshotCheck, _ := gtk.CheckButtonNewWithLabel("📸 Screenshot")
	screenshotCheck.SetActive(loadScreenshotPref())
	screenshotCheck.Connect("toggled", func() {
		saveScreenshotPref(screenshotCheck.GetActive())
	})

	cropCheck, _ := gtk.CheckButtonNewWithLabel("✂ Crop")
	cropCheckGlobal = cropCheck
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

	tooltipModeCheck, _ = gtk.CheckButtonNewWithLabel("🔔 Tooltip mode")
	tooltipModeCheck.SetActive(false)
	tooltipCss, _ := gtk.CssProviderNew()
	tooltipCss.LoadFromData(`checkbutton:checked { color: #f0a500; }`)
	tooltipCtx, _ := tooltipModeCheck.GetStyleContext()
	tooltipCtx.AddProvider(tooltipCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	// Timeout spin button (seconds)
	tooltipTimeoutSpin, _ := gtk.SpinButtonNewWithRange(5, 300, 5)
	tooltipTimeoutSpin.SetValue(float64(tooltipTimeoutSecs))
	tooltipTimeoutSpin.SetTooltipText("Tooltip auto-hide (seconds)")
	tooltipTimeoutSpin.SetSizeRequest(60, -1)
	tooltipTimeoutSpin.Connect("value-changed", func() {
		tooltipTimeoutSecs = tooltipTimeoutSpin.GetValueAsInt()
	})
	tooltipModeCheck.Connect("toggled", func() {
		tooltipTimeoutSpin.SetSensitive(tooltipModeCheck.GetActive())
	})
	tooltipTimeoutSpin.SetSensitive(false)

	secsLbl, _ := gtk.LabelNew("s")

	sendBtn, _ := gtk.ButtonNewWithLabel("Ask ↵")
	sendBtnCss, _ := gtk.CssProviderNew()
	sendBtnCss.LoadFromData(`button { padding: 5px 18px; font-size: 13px; }`)
	sendBtnCtx, _ := sendBtn.GetStyleContext()
	sendBtnCtx.AddProvider(sendBtnCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	optionsBar.PackStart(screenshotCheck, false, false, 0)
	optionsBar.PackStart(cropCheck, false, false, 0)
	optionsBar.PackStart(tooltipModeCheck, false, false, 0)
	optionsBar.PackStart(tooltipTimeoutSpin, false, false, 0)
	optionsBar.PackStart(secsLbl, false, false, 0)
	optionsBar.PackEnd(sendBtn, false, false, 0)

	// ── Status label ──
	statusLabel, _ := gtk.LabelNew("Enter to send  ·  Shift+Enter newline  ·  Esc hide")
	statusLabel.SetXAlign(0)
	statusLabel.SetMarginStart(14)
	statusLabel.SetMarginEnd(14)
	statusLabel.SetMarginBottom(4)

	// ── Response area ──
	responseView, _ := gtk.TextViewNew()
	responseView.SetEditable(false)
	responseView.SetWrapMode(gtk.WRAP_WORD_CHAR)
	responseView.SetLeftMargin(10)
	responseView.SetRightMargin(10)
	responseView.SetTopMargin(8)
	responseView.SetBottomMargin(8)
	buf, _ := responseView.GetBuffer()
	t := makeTags(buf)
	responseScroll, _ := gtk.ScrolledWindowNew(nil, nil)
	responseScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	responseScroll.Add(responseView)
	responseScroll.SetMarginStart(12)
	responseScroll.SetMarginEnd(12)
	responseScroll.SetMarginBottom(10)

	// ── Separator ──
	sep, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	sepCss, _ := gtk.CssProviderNew()
	sepCss.LoadFromData(`separator { background-color: #2a2a4a; min-height: 1px; margin: 2px 12px; }`)
	sepCtx, _ := sep.GetStyleContext()
	sepCtx.AddProvider(sepCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	outer.PackStart(titleBar, false, false, 0)
	outer.PackStart(inputScroll, false, false, 0)
	outer.PackStart(optionsBar, false, false, 0)
	outer.PackStart(statusLabel, false, false, 0)
	outer.PackStart(sep, false, false, 0)
	outer.PackStart(responseScroll, true, true, 0)
	win.Add(outer)

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
				statusLabel.SetText("Done  ·  Enter to ask again")
				sendBtn.SetSensitive(true)
				adj := responseScroll.GetVAdjustment()
				adj.SetValue(adj.GetUpper())
			})
		}()
	}

	newChatBtn.Connect("clicked", func() {
		clearHistory()
		buf.SetText("")
		statusLabel.SetText("New chat started  ·  Enter to send")
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

	win.ShowAll()
	overlayWin = win
}
