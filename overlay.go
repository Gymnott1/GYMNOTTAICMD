package main

// overlay.go - always-on-top floating input + response window

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var overlayWin *gtk.Window

var prefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_prefs"
var textExtractPrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_text_extract_pref"
var cropPrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_crop_pref"
var tooltipModePrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_tooltip_pref"
var tooltipTimeoutPrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_tooltip_timeout_pref"
var agenticModePrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_agentic_pref"
var continuousModePrefsFile = os.Getenv("HOME") + "/.config/gymnott_ai_continuous_pref"
var tooltipModeCheck *gtk.CheckButton
var cropCheckGlobal *gtk.CheckButton
var agenticModeCheck *gtk.CheckButton
var continuousModeCheck *gtk.CheckButton

func loadBoolPref(path string, defaultValue bool) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultValue
	}
	return strings.TrimSpace(string(data)) == "1"
}

func saveBoolPref(path string, value bool) {
	if value {
		os.WriteFile(path, []byte("1"), 0644)
		return
	}
	os.WriteFile(path, []byte("0"), 0644)
}

func loadIntPref(path string, defaultValue int) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultValue
	}
	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return defaultValue
	}
	return value
}

func saveIntPref(path string, value int) {
	os.WriteFile(path, []byte(strconv.Itoa(value)), 0644)
}

func getTooltipMode() bool {
	if tooltipModeCheck == nil {
		return loadTooltipModePref()
	}
	return tooltipModeCheck.GetActive()
}

func getScreenshotPrefs() (withShot, crop bool) {
	if cropCheckGlobal != nil {
		return loadScreenshotPref(), cropCheckGlobal.GetActive()
	}
	return loadScreenshotPref(), loadCropPref()
}

func loadScreenshotPref() bool {
	return loadBoolPref(prefsFile, true)
}

func saveScreenshotPref(v bool) {
	saveBoolPref(prefsFile, v)
}

func loadTextExtractPref() bool {
	return loadBoolPref(textExtractPrefsFile, true)
}

func saveTextExtractPref(v bool) {
	saveBoolPref(textExtractPrefsFile, v)
}

func loadCropPref() bool {
	return loadBoolPref(cropPrefsFile, false)
}

func saveCropPref(v bool) {
	saveBoolPref(cropPrefsFile, v)
}

func loadTooltipModePref() bool {
	return loadBoolPref(tooltipModePrefsFile, false)
}

func saveTooltipModePref(v bool) {
	saveBoolPref(tooltipModePrefsFile, v)
}

func loadTooltipTimeoutPref() int {
	return loadIntPref(tooltipTimeoutPrefsFile, 30)
}

func saveTooltipTimeoutPref(v int) {
	saveIntPref(tooltipTimeoutPrefsFile, v)
}

func loadAgenticModePref() bool {
	return loadBoolPref(agenticModePrefsFile, false)
}

func saveAgenticModePref(v bool) {
	saveBoolPref(agenticModePrefsFile, v)
}

func getAgenticMode() bool {
	if agenticModeCheck == nil {
		return loadAgenticModePref()
	}
	return agenticModeCheck.GetActive()
}

func loadContinuousModePref() bool {
	return loadBoolPref(continuousModePrefsFile, false)
}

func saveContinuousModePref(v bool) {
	saveBoolPref(continuousModePrefsFile, v)
}

func getContinuousMode() bool {
	if continuousModeCheck == nil {
		return loadContinuousModePref()
	}
	return continuousModeCheck.GetActive()
}

func showResponseInTooltip(response string) {
	tooltipTimeoutSecs = loadTooltipTimeoutPref()
	showFollowerTooltip(response)
}

func runQuickTooltipAsk() {
	if !getTooltipMode() {
		scheduleOnMain(showOverlay)
		return
	}

	_, crop := getScreenshotPrefs()
	textExtract := loadTextExtractPref()
	tooltipTimeoutSecs = loadTooltipTimeoutPref()
	setWaiting(true)

	go func() {
		scheduleOnMain(func() {
			if overlayWin != nil {
				overlayWin.Hide()
			}
		})
		time.Sleep(300 * time.Millisecond)

		response := askAI(defaultQuickAskPrompt, true, crop, textExtract, getAgenticMode())
		scheduleOnMain(func() {
			setWaiting(false)
			showResponseInTooltip(response)
		})
	}()
}

// showAgenticRunDialog presents a confirmation dialog listing extracted shell
// blocks. If the user clicks Run, onRun is called.
func showAgenticRunDialog(parent *gtk.Window, commands []string, continuous bool, onRun func()) {
	dlg, _ := gtk.DialogNew()
	dlg.SetTitle("▶ Run commands?")
	dlg.SetTransientFor(parent)
	dlg.SetModal(true)
	dlg.SetDefaultSize(600, 340)

	content, _ := dlg.GetContentArea()
	content.SetSpacing(0)

	dlgCss, _ := gtk.CssProviderNew()
	dlgCss.LoadFromData(`
		window { background-color: #1a1a2e; }
		label { color: #e0e0f0; font-size: 12px; margin: 8px 14px 4px 14px; }
		textview { background-color: #0d0d1a; color: #c8ffc8; font-family: monospace; font-size: 12px; padding: 8px; }
		textview text { background-color: #0d0d1a; color: #c8ffc8; }
		button { font-size: 12px; padding: 5px 16px; }
	`)
	dlgScreen := dlg.GetScreen()
	gtk.AddProviderForScreen(dlgScreen, dlgCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	modeText := "off"
	if continuous {
		modeText = "on"
	}
	lbl, _ := gtk.LabelNew(fmt.Sprintf("Agentic mode detected %d command block(s). Review and run? Continuous mode: %s", len(commands), modeText))
	lbl.SetXAlign(0)
	content.PackStart(lbl, false, false, 0)

	cmdText := ""
	for i, c := range commands {
		if i > 0 {
			cmdText += "\n\n# ── block " + fmt.Sprintf("%d", i+1) + " ──\n"
		}
		cmdText += c
	}

	tv, _ := gtk.TextViewNew()
	tv.SetEditable(false)
	tv.SetMonospace(true)
	tv.SetLeftMargin(8)
	tv.SetRightMargin(8)
	tv.SetTopMargin(6)
	tv.SetBottomMargin(6)
	tbuf, _ := tv.GetBuffer()
	tbuf.SetText(cmdText)

	scroll, _ := gtk.ScrolledWindowNew(nil, nil)
	scroll.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)
	scroll.Add(tv)
	scroll.SetSizeRequest(-1, 220)
	scroll.SetMarginStart(14)
	scroll.SetMarginEnd(14)
	scroll.SetMarginBottom(8)
	content.PackStart(scroll, true, true, 0)

	dlg.AddButton("Cancel", gtk.RESPONSE_CANCEL)
	dlg.AddButton("▶ Run all", gtk.RESPONSE_ACCEPT)
	dlg.ShowAll()

	resp := dlg.Run()
	dlg.Destroy()

	if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
		onRun()
	}
}

func executeCommandBlocks(round int, commands []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n\n─── Execution output (round %d) ───\n", round))
	for _, cmd := range commands {
		firstLine := strings.SplitN(cmd, "\n", 2)[0]
		sb.WriteString(fmt.Sprintf("\n$ %s\n", firstLine))
		if strings.Contains(cmd, "\n") {
			sb.WriteString("(+ more lines)\n")
		}
		sb.WriteString(runShellBlock(cmd))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func runAgenticLoop(query string, commands []string, continuous bool, onChunk func(chunk string), onStatus func(status string), onDone func()) {
	go func() {
		const maxRounds = 5
		current := commands

		for round := 1; round <= maxRounds; round++ {
			onStatus(fmt.Sprintf("Running agentic commands (round %d/%d)…", round, maxRounds))
			roundOutput := executeCommandBlocks(round, current)
			onChunk(roundOutput)

			if !continuous {
				onDone()
				return
			}

			onStatus(fmt.Sprintf("Analyzing results (round %d/%d)…", round, maxRounds))
			followPrompt := fmt.Sprintf(`Original user goal:
%s

Execution output from round %d:
%s

If the goal is achieved, start your response with TARGET_ACHIEVED and include a short summary.
If the goal is not achieved, provide only the next required actions and include shell commands in fenced bash/sh/shell/zsh blocks.`, query, round, roundOutput)

			nextResponse := askAI(followPrompt, false, false, false, true)
			onChunk(fmt.Sprintf("\n\n─── Agentic follow-up (round %d) ───\n%s\n", round, nextResponse))

			if strings.Contains(strings.ToUpper(nextResponse), "TARGET_ACHIEVED") {
				onChunk("\n✅ Continuous mode target achieved.\n")
				onDone()
				return
			}

			nextCommands := extractShellBlocks(nextResponse)
			if len(nextCommands) == 0 {
				onChunk("\n⚠ Continuous mode stopped: no shell blocks returned for the next step.\n")
				onDone()
				return
			}

			current = nextCommands
		}

		onChunk("\n⚠ Continuous mode stopped after max rounds (5). Review output and continue manually if needed.\n")
		onDone()
	}()
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

	tooltipTimeoutSecs = loadTooltipTimeoutPref()

	applyCSS()

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Gymnott AI")
	win.SetDefaultSize(700, 500)
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
	optionsBar, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 8)
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
	cropCheck.SetActive(loadCropPref())
	cropCheck.SetSensitive(loadScreenshotPref())
	cropCheck.Connect("toggled", func() {
		saveCropPref(cropCheck.GetActive())
	})
	screenshotCheck.Connect("toggled", func() {
		active := screenshotCheck.GetActive()
		saveScreenshotPref(active)
		cropCheck.SetSensitive(active)
	})

	textExtractCheck, _ := gtk.CheckButtonNewWithLabel("Text Extract")
	textExtractCheck.SetSizeRequest(115, -1)
	textExtractCheck.SetActive(loadTextExtractPref())
	textExtractCheck.SetSensitive(loadScreenshotPref())
	textExtractCheck.SetTooltipText("Extract text from the screenshot and ask Gemini with that text")
	textExtractCheck.Connect("toggled", func() {
		saveTextExtractPref(textExtractCheck.GetActive())
	})
	screenshotCheck.Connect("toggled", func() {
		active := screenshotCheck.GetActive()
		textExtractCheck.SetSensitive(active)
	})

	tooltipModeCheck, _ = gtk.CheckButtonNewWithLabel("🔔 Tooltip mode")
	tooltipModeCheck.SetActive(loadTooltipModePref())
	tooltipCss, _ := gtk.CssProviderNew()
	tooltipCss.LoadFromData(`checkbutton:checked { color: #f0a500; }`)
	tooltipCtx, _ := tooltipModeCheck.GetStyleContext()
	tooltipCtx.AddProvider(tooltipCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	agenticModeCheck, _ = gtk.CheckButtonNewWithLabel("🤖 Agentic")
	agenticModeCheck.SetActive(loadAgenticModePref())
	agenticModeCheck.SetTooltipText("Action-oriented mode: asks AI for safe, executable steps and commands")
	agenticModeCheck.Connect("toggled", func() {
		saveAgenticModePref(agenticModeCheck.GetActive())
	})

	continuousModeCheck, _ = gtk.CheckButtonNewWithLabel("♻ Continuous")
	continuousModeCheck.SetActive(loadContinuousModePref())
	continuousModeCheck.SetTooltipText("When Agentic is on, keep iterating commands until TARGET_ACHIEVED or max rounds")
	continuousModeCheck.Connect("toggled", func() {
		saveContinuousModePref(continuousModeCheck.GetActive())
	})

	// Timeout spin button (seconds)
	tooltipTimeoutSpin, _ := gtk.SpinButtonNewWithRange(5, 300, 5)
	tooltipTimeoutSpin.SetValue(float64(tooltipTimeoutSecs))
	tooltipTimeoutSpin.SetTooltipText("Tooltip auto-hide (seconds)")
	tooltipTimeoutSpin.SetSizeRequest(60, -1)
	tooltipTimeoutSpin.Connect("value-changed", func() {
		tooltipTimeoutSecs = tooltipTimeoutSpin.GetValueAsInt()
		saveTooltipTimeoutPref(tooltipTimeoutSecs)
	})
	tooltipModeCheck.Connect("toggled", func() {
		saveTooltipModePref(tooltipModeCheck.GetActive())
		tooltipTimeoutSpin.SetSensitive(tooltipModeCheck.GetActive())
	})
	tooltipTimeoutSpin.SetSensitive(tooltipModeCheck.GetActive())

	secsLbl, _ := gtk.LabelNew("s")

	sendBtn, _ := gtk.ButtonNewWithLabel("Ask ↵")
	sendBtnCss, _ := gtk.CssProviderNew()
	sendBtnCss.LoadFromData(`button { padding: 5px 18px; font-size: 13px; }`)
	sendBtnCtx, _ := sendBtn.GetStyleContext()
	sendBtnCtx.AddProvider(sendBtnCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	optionsBar.PackStart(screenshotCheck, false, false, 0)
	optionsBar.PackStart(cropCheck, false, false, 0)
	optionsBar.PackStart(textExtractCheck, false, false, 0)
	optionsBar.PackStart(tooltipModeCheck, false, false, 0)
	optionsBar.PackStart(agenticModeCheck, false, false, 0)
	optionsBar.PackStart(continuousModeCheck, false, false, 0)
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
		textExtract := textExtractCheck.GetActive()
		if withShot {
			if crop {
				statusLabel.SetText("Select area to capture…")
			} else if textExtract {
				statusLabel.SetText("Taking screenshot for text extract…")
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
			agentic := getAgenticMode()
			continuous := getContinuousMode()
			response := askAI(query, withShot, crop, textExtract, agentic)
			scheduleOnMain(func() {
				setWaiting(false)
				win.ShowAll()
				win.Present()
				renderMarkdown(buf, t, responseView, response)
				if getTooltipMode() {
					showResponseInTooltip(response)
				}
				statusLabel.SetText("Done  ·  Enter to ask again")
				sendBtn.SetSensitive(true)
				adj := responseScroll.GetVAdjustment()
				adj.SetValue(adj.GetUpper())

				// Full agentic: extract shell blocks and offer to run them
				if agentic {
					if cmds := extractShellBlocks(response); len(cmds) > 0 {
						showAgenticRunDialog(win, cmds, continuous, func() {
							runAgenticLoop(query, cmds, continuous, func(chunk string) {
								scheduleOnMain(func() {
									end := buf.GetEndIter()
									buf.Insert(end, chunk)
									adj2 := responseScroll.GetVAdjustment()
									adj2.SetValue(adj2.GetUpper())
								})
							}, func(status string) {
								scheduleOnMain(func() {
									statusLabel.SetText(status)
								})
							}, func() {
								scheduleOnMain(func() {
									statusLabel.SetText("Done  ·  Enter to ask again")
								})
							})
						})
					}
				}
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
