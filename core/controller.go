package core

// Controller is the behavior the front-ends depend on, so the TUI and GUI (and
// their tests) can be written against an interface rather than the concrete
// Manager.
type Controller interface {
	Add(url string) (string, error)
	Pause(id string)
	Resume(id string)
	Remove(id string)
	OpenFolder() error
	Snapshot() []JobStatus
	ClearCompleted()
	PauseAll()
}

var _ Controller = (*Manager)(nil)
