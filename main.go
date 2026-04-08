package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	go startFollower()
	go listenHotkey()

	// Handle Ctrl+C gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		gtk.MainQuit()
	}()

	gtk.Main()
}
