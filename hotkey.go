package main

// hotkey.go - listens for Ctrl+Space globally using XGrabKey

/*
#cgo pkg-config: x11 xtst
#include <X11/Xlib.h>
#include <X11/keysym.h>
#include <X11/extensions/XTest.h>
#include <stdlib.h>

typedef struct { int type; } XAnyEvent_t;

// Returns 1 when Ctrl+Space is pressed, 2 when Escape is pressed, 3 when Tab is pressed
int wait_for_hotkey() {
    Display *dpy = XOpenDisplay(NULL);
    if (!dpy) return 0;

    Window root = DefaultRootWindow(dpy);
    KeyCode space = XKeysymToKeycode(dpy, XK_space);
    KeyCode esc   = XKeysymToKeycode(dpy, XK_Escape);
    KeyCode tab   = XKeysymToKeycode(dpy, XK_Tab);

    XGrabKey(dpy, space, ControlMask, root, True, GrabModeAsync, GrabModeAsync);
    XGrabKey(dpy, esc,   AnyModifier, root, True, GrabModeAsync, GrabModeAsync);
    XGrabKey(dpy, tab,   AnyModifier, root, True, GrabModeAsync, GrabModeAsync);
    XSelectInput(dpy, root, KeyPressMask);

    XEvent ev;
    while (1) {
        XNextEvent(dpy, &ev);
        if (ev.type == KeyPress) {
            XKeyEvent *ke = (XKeyEvent*)&ev;
            if (ke->keycode == space && (ke->state & ControlMask)) {
                XUngrabKey(dpy, space, ControlMask, root);
                XUngrabKey(dpy, esc,   AnyModifier, root);
                XUngrabKey(dpy, tab,   AnyModifier, root);
                XCloseDisplay(dpy);
                return 1;
            }
            if (ke->keycode == esc) {
                XUngrabKey(dpy, space, ControlMask, root);
                XUngrabKey(dpy, esc,   AnyModifier, root);
                XUngrabKey(dpy, tab,   AnyModifier, root);
                XCloseDisplay(dpy);
                return 2;
            }
            if (ke->keycode == tab) {
                XUngrabKey(dpy, space, ControlMask, root);
                XUngrabKey(dpy, esc,   AnyModifier, root);
                XUngrabKey(dpy, tab,   AnyModifier, root);
                XCloseDisplay(dpy);
                return 3;
            }
        }
    }
}
*/
import "C"

import "time"

func listenHotkey() {
	for {
		ev := int(C.wait_for_hotkey())
		if ev == 2 {
			scheduleOnMain(hideFollowerTooltip)
			continue
		}
		if ev == 3 {
			pasteTooltipText()
			continue
		}
		scheduleOnMain(func() {
			if getTooltipMode() {
				withShot, crop := getScreenshotPrefs()
				if !withShot {
					// no screenshot configured — need a query, open overlay
					showOverlay()
					return
				}
				setWaiting(true)
				go func() {
					time.Sleep(150 * time.Millisecond)
					response := askAI("Look at my screen. Give only the exact commands or complete code fixes needed — no explanations, no intros. Write every command in full, never truncate.", withShot, crop)
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
