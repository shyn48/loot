package gui

import (
	g "github.com/AllenDang/giu"
)

func Start() {
	// Set a slightly larger, more readable base font before the window is created.
	g.SetDefaultFontSize(15)

	wnd := g.NewMasterWindow("Shyn Download Manager", 900, 620, g.MasterWindowFlagsNotResizable)
	wnd.SetBgColor(colorBackground)
	wnd.Run(loop)
}
