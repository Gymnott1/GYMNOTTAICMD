package main

// hotkey.go - listens for Ctrl+Space and Ctrl+Alt+Space globally using XGrabKey

/*
#cgo pkg-config: x11 xtst
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <X11/extensions/XTest.h>
#include <stdlib.h>

static void grab_space(Display *dpy, Window root, KeyCode space, unsigned int mod) {
    unsigned int variants[] = {
        mod,
        mod | LockMask,
        mod | Mod2Mask,
        mod | LockMask | Mod2Mask,
    };
    for (int i = 0; i < 4; i++) {
        XGrabKey(dpy, space, variants[i], root, True, GrabModeAsync, GrabModeAsync);
    }
}

static void ungrab_space(Display *dpy, Window root, KeyCode space, unsigned int mod) {
    unsigned int variants[] = {
        mod,
        mod | LockMask,
        mod | Mod2Mask,
        mod | LockMask | Mod2Mask,
    };
    for (int i = 0; i < 4; i++) {
        XUngrabKey(dpy, space, variants[i], root);
    }
}

static int close_hotkey(Display *dpy, Window root, KeyCode space, int hotkey) {
    ungrab_space(dpy, root, space, ControlMask);
    ungrab_space(dpy, root, space, ControlMask | Mod1Mask);
    XCloseDisplay(dpy);
    return hotkey;
}

int wait_for_hotkey() {
    Display *dpy = XOpenDisplay(NULL);
    if (!dpy) return 0;

    Window root = DefaultRootWindow(dpy);
    KeyCode space = XKeysymToKeycode(dpy, XK_space);

    grab_space(dpy, root, space, ControlMask);
    grab_space(dpy, root, space, ControlMask | Mod1Mask);
    XSelectInput(dpy, root, KeyPressMask);

    XEvent ev;
    while (1) {
        XNextEvent(dpy, &ev);
        if (ev.type != KeyPress) {
            continue;
        }

        XKeyEvent *ke = (XKeyEvent*)&ev;
        unsigned int state = ke->state & (ControlMask | Mod1Mask);
        if (ke->keycode != space) {
            continue;
        }
        if (state == (ControlMask | Mod1Mask)) {
            return close_hotkey(dpy, root, space, 2);
        }
        if (state == ControlMask) {
            return close_hotkey(dpy, root, space, 1);
        }
    }
}
*/
import "C"

func listenHotkey() {
	for {
		switch C.wait_for_hotkey() {
		case 2:
			runQuickTooltipAsk()
		default:
			scheduleOnMain(showOverlay)
		}
	}
}
