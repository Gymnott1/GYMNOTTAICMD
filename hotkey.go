package main

// hotkey.go - listens for Ctrl+Space globally using XGrabKey

/*
#cgo pkg-config: x11 xtst
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <X11/extensions/XTest.h>
#include <stdlib.h>

typedef struct { int type; } XAnyEvent_t;

// Returns 1 when Ctrl+Space is pressed, 2 when Escape is pressed
int wait_for_hotkey() {
    Display *dpy = XOpenDisplay(NULL);
    if (!dpy) return 0;

    Window root = DefaultRootWindow(dpy);
    KeyCode space = XKeysymToKeycode(dpy, XK_space);
    KeyCode esc   = XKeysymToKeycode(dpy, XK_Escape);

    XGrabKey(dpy, space, ControlMask, root, True, GrabModeAsync, GrabModeAsync);
    XGrabKey(dpy, esc,   AnyModifier, root, True, GrabModeAsync, GrabModeAsync);
    XSelectInput(dpy, root, KeyPressMask);

    XEvent ev;
    while (1) {
        XNextEvent(dpy, &ev);
        if (ev.type == KeyPress) {
            XKeyEvent *ke = (XKeyEvent*)&ev;
            if (ke->keycode == space && (ke->state & ControlMask)) {
                XUngrabKey(dpy, space, ControlMask, root);
                XUngrabKey(dpy, esc,   AnyModifier, root);
                XCloseDisplay(dpy);
                return 1;
            }
            if (ke->keycode == esc) {
                XUngrabKey(dpy, space, ControlMask, root);
                XUngrabKey(dpy, esc,   AnyModifier, root);
                XCloseDisplay(dpy);
                return 2;
            }
        }
    }
}
*/
import "C"

func listenHotkey() {
	for {
		ev := int(C.wait_for_hotkey())
		if ev == 2 {
			scheduleOnMain(hideFollowerTooltip)
			continue
		}
		scheduleOnMain(func() {
			if getTooltipMode() {
				withShot, crop := getScreenshotPrefs()
				setWaiting(true)
				go func() {
					response := askAI("What's on my screen? Give a brief helpful summary or answer.", withShot, crop)
					scheduleOnMain(func() {
						setWaiting(false)
						showFollowerTooltip(response)
					})
				}()
			} else {
				showOverlay()
			}
		})
	}
}
