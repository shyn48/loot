package gui

import (
	g "github.com/AllenDang/giu"

	"simple-gui/core"
)

// manager is the shared download engine this front-end renders. Set by Run.
var manager *core.Manager

func Run(m *core.Manager) {
	manager = m

	// Set a slightly larger, more readable base font before the window is created.
	g.SetDefaultFontSize(15)

	wnd := g.NewMasterWindow("Loot", 900, 620, g.MasterWindowFlagsNotResizable)
	wnd.SetBgColor(colorBackground)
	wnd.Run(loop)
}
