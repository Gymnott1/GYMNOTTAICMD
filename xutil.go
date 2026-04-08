package main

// xutil.go - X11 mouse position + schedule on GTK main thread

/*
#cgo pkg-config: x11
#include <X11/Xlib.h>

void get_mouse_pos(int *x, int *y) {
    Display *dpy = XOpenDisplay(NULL);
    if (!dpy) return;
    Window root = DefaultRootWindow(dpy);
    Window rroot, child;
    int rx, ry, wx, wy;
    unsigned int mask;
    XQueryPointer(dpy, root, &rroot, &child, &rx, &ry, &wx, &wy, &mask);
    *x = rx;
    *y = ry;
    XCloseDisplay(dpy);
}
*/
import "C"

import "github.com/gotk3/gotk3/glib"

func getMousePos() (int, int) {
	var x, y C.int
	C.get_mouse_pos(&x, &y)
	return int(x), int(y)
}

func scheduleOnMain(fn func()) {
	glib.IdleAdd(fn)
}
