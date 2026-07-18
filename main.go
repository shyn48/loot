package main

import (
	"flag"
	"fmt"
	"os"

	"simple-gui/core"
	"simple-gui/gui"
	"simple-gui/tui"
)

func main() {
	// `loot add <url>...` enqueues URLs for the running app (or the next
	// launch) via the inbox, then exits. Used by the Automator Quick Action.
	if len(os.Args) > 1 && os.Args[1] == "add" {
		urls := os.Args[2:]
		if len(urls) == 0 {
			fmt.Fprintln(os.Stderr, "usage: loot add <url> [url...]")
			os.Exit(1)
		}
		if err := core.AppendToInbox(urls); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	useGUI := flag.Bool("gui", false, "launch the desktop (giu) window instead of the TUI")
	flag.Parse()

	m, err := core.NewManager()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer m.Close()

	if err := m.LoadPersisted(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	if *useGUI {
		gui.Run(m)
		return
	}
	if err := tui.Run(m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
