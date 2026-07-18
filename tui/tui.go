// Package tui is the terminal front-end (Bubble Tea) for the download manager.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"simple-gui/core"
)

// Run launches the TUI against the given controller and blocks until the user
// quits.
func Run(ctrl core.Controller) error {
	p := tea.NewProgram(newModel(ctrl), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
