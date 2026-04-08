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
	"sync/atomic"
	"time"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gtk"
)

const (
	followerSize   = 20 // slightly bigger so spinner is visible
	followerOffset = 14
)

// waiting is set to 1 while AI is processing, 0 otherwise
var waiting int32

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
				// Normal red arrow
				cr.SetSourceRGBA(1, 0.2, 0.2, 0.9)
				cr.MoveTo(0, 0)
				cr.LineTo(0, followerSize)
				cr.LineTo(followerSize*0.35, followerSize*0.65)
				cr.LineTo(followerSize*0.5, followerSize)
				cr.LineTo(followerSize*0.65, followerSize*0.9)
				cr.LineTo(followerSize*0.45, followerSize*0.6)
				cr.LineTo(followerSize, followerSize*0.6)
				cr.ClosePath()
				cr.Fill()

				cr.SetSourceRGBA(1, 1, 1, 0.8)
				cr.SetLineWidth(0.8)
				cr.MoveTo(0, 0)
				cr.LineTo(0, followerSize)
				cr.LineTo(followerSize*0.35, followerSize*0.65)
				cr.LineTo(followerSize*0.5, followerSize)
				cr.LineTo(followerSize*0.65, followerSize*0.9)
				cr.LineTo(followerSize*0.45, followerSize*0.6)
				cr.LineTo(followerSize, followerSize*0.6)
				cr.ClosePath()
				cr.Stroke()
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
			if da != nil {
				da.QueueDraw()
			}
		})
	}
}
