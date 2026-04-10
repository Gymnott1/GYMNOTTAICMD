package main

// follower.go - draws a small always-on-top X11 window that mirrors the real cursor

/*
#cgo pkg-config: x11
#cgo LDFLAGS: -lXext
#include <X11/Xlib.h>
#include <X11/extensions/shape.h>

void set_input_passthrough(unsigned long xid) {
    Display *dpy = XOpenDisplay(NULL);
    if (!dpy) return;
    Region region = XCreateRegion();
    XShapeCombineRegion(dpy, (Window)xid, ShapeInput, 0, 0, region, ShapeSet);
    XDestroyRegion(region);
    XCloseDisplay(dpy);
}
*/
import "C"

import (
	"math"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

const (
	followerSize   = 14
	followerOffset = 10
)

// waiting is set to 1 while AI is processing, 0 otherwise
var waiting int32

// tooltipWin is the floating tooltip that shows AI response near the cursor
var tooltipWin *gtk.Window
var tooltipLabel *gtk.Label
var tooltipText string
var tooltipTimeoutSecs = 30

func pasteTooltipText() {
	if tooltipText == "" {
		return
	}
	hideFollowerTooltip()
	text := tooltipText
	go func() {
		time.Sleep(150 * time.Millisecond)
		for _, line := range strings.Split(text, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			exec.Command("xdotool", "type", "--clearmodifiers", "--", line).Run()
			exec.Command("xdotool", "key", "Return").Run()
		}
	}()
}

func showFollowerTooltip(text string) {
	scheduleOnMain(func() {
		if tooltipWin == nil {
			win, _ := gtk.WindowNew(gtk.WINDOW_POPUP)
			win.SetDecorated(false)
			win.SetKeepAbove(true)
			win.SetSkipTaskbarHint(true)
			win.SetDefaultSize(400, -1)
			win.SetAppPaintable(true)

			screen := win.GetScreen()
			visual, _ := screen.GetRGBAVisual()
			if visual != nil {
				win.SetVisual(visual)
			}

			css, _ := gtk.CssProviderNew()
			css.LoadFromData(`
				window { background-color: #1e1e2e; border-radius: 8px; border: 1px solid #3a3a5a; }
				label { color: #e0e0f0; font-size: 12px; }
			`)
			gtk.AddProviderForScreen(screen, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

			box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
			box.SetMarginTop(10)
			box.SetMarginBottom(10)
			box.SetMarginStart(12)
			box.SetMarginEnd(12)

			lbl, _ := gtk.LabelNew("")
			lbl.SetLineWrap(true)
			lbl.SetXAlign(0)
			lbl.SetSelectable(true)
			lbl.SetMaxWidthChars(55)
			box.PackStart(lbl, true, true, 0)

			hintLbl, _ := gtk.LabelNew("Tab to paste · Esc to dismiss")
			hintLbl.SetXAlign(1)
			hintCss, _ := gtk.CssProviderNew()
			hintCss.LoadFromData(`label { color: #555577; font-size: 10px; }`)
			hintCtx, _ := hintLbl.GetStyleContext()
			hintCtx.AddProvider(hintCss, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
			box.PackEnd(hintLbl, false, false, 0)

			win.Add(box)

			tooltipLabel = lbl
			tooltipWin = win
		}
		tooltipText = text
		tooltipLabel.SetText(text)
		x, y := getMousePos()
		tooltipWin.Move(x+followerOffset+followerSize+2, y+followerOffset)
		tooltipWin.ShowAll()

		// auto-hide after timeout
		secs := tooltipTimeoutSecs
		go func() {
			time.Sleep(time.Duration(secs) * time.Second)
			hideFollowerTooltip()
		}()
	})
}

func hideFollowerTooltip() {
	scheduleOnMain(func() {
		if tooltipWin != nil {
			tooltipWin.Hide()
		}
	})
}

func setWaiting(v bool) {
	if v {
		atomic.StoreInt32(&waiting, 1)
	} else {
		atomic.StoreInt32(&waiting, 0)
	}
}

func isWaiting() bool {
	return atomic.LoadInt32(&waiting) == 1
}

func startFollower() {
	var win *gtk.Window
	var da *gtk.DrawingArea

	// spinAngle advances each frame when waiting
	spinAngle := 0.0

	setup := func() {
		win, _ = gtk.WindowNew(gtk.WINDOW_POPUP)
		win.SetDefaultSize(followerSize, followerSize)
		win.SetDecorated(false)
		win.SetSkipTaskbarHint(true)
		win.SetKeepAbove(true)
		win.SetAppPaintable(true)

		screen := win.GetScreen()
		visual, _ := screen.GetRGBAVisual()
		if visual != nil {
			win.SetVisual(visual)
		}

		da, _ = gtk.DrawingAreaNew()
		da.Connect("draw", func(_ *gtk.DrawingArea, cr *cairo.Context) {
			cr.SetSourceRGBA(0, 0, 0, 0)
			cr.SetOperator(cairo.OPERATOR_SOURCE)
			cr.Paint()

			if isWaiting() {
				// Spinning arc
				cx := followerSize / 2.0
				cy := followerSize / 2.0
				r := followerSize/2.0 - 2

				// Dark background circle
				cr.SetSourceRGBA(0.1, 0.1, 0.1, 0.7)
				cr.Arc(cx, cy, r+1, 0, 2*math.Pi)
				cr.Fill()

				// Spinning arc (orange/yellow)
				cr.SetSourceRGBA(1, 0.65, 0, 1)
				cr.SetLineWidth(2.5)
				cr.Arc(cx, cy, r, spinAngle, spinAngle+1.8*math.Pi)
				cr.Stroke()

				// Small dot at tip
				tipX := cx + r*math.Cos(spinAngle+1.8*math.Pi)
				tipY := cy + r*math.Sin(spinAngle+1.8*math.Pi)
				cr.SetSourceRGBA(1, 0.9, 0.2, 1)
				cr.Arc(tipX, tipY, 2, 0, 2*math.Pi)
				cr.Fill()
			} else {
				// Red dot
				cx := followerSize / 2.0
				cy := followerSize / 2.0
				r := followerSize/2.0 - 1.5

				// Soft shadow
				cr.SetSourceRGBA(0, 0, 0, 0.25)
				cr.Arc(cx+1, cy+1, r, 0, 2*math.Pi)
				cr.Fill()

				// Red fill
				cr.SetSourceRGBA(0.95, 0.15, 0.15, 0.92)
				cr.Arc(cx, cy, r, 0, 2*math.Pi)
				cr.Fill()

				// White highlight
				cr.SetSourceRGBA(1, 1, 1, 0.35)
				cr.Arc(cx-r*0.25, cy-r*0.25, r*0.4, 0, 2*math.Pi)
				cr.Fill()
			}
		})

		win.Add(da)
		win.ShowAll()

		gdkWin, _ := win.GetWindow()
		if gdkWin != nil {
			xid := gdkWin.GetXID()
			C.set_input_passthrough(C.ulong(xid))
		}
	}

	scheduleOnMain(setup)

	for {
		time.Sleep(16 * time.Millisecond)
		if win == nil {
			continue
		}
		if isWaiting() {
			spinAngle += 0.15 // rotation speed
		}
		x, y := getMousePos()
		scheduleOnMain(func() {
			win.Move(x+followerOffset, y+followerOffset)
			if tooltipWin != nil && tooltipWin.IsVisible() {
				tooltipWin.Move(x+followerOffset+followerSize+2, y+followerOffset)
			}
			if da != nil {
				da.QueueDraw()
			}
		})
	}
}
