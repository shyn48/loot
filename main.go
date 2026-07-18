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
