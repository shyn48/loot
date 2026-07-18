package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The inbox is a plain text file (one URL per line) in the state directory. The
// `loot add <url>` CLI appends to it from a separate process; a running
// manager drains it, so URLs sent from outside (e.g. an Automator Quick Action)
// show up in the live app.

// AppendToInbox appends URLs to the shared inbox for the running app to pick up.
// Used by the `add` CLI subcommand.
func AppendToInbox(urls []string) error {
	dir, err := GetTempPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return appendToInboxFile(filepath.Join(dir, "inbox"), urls)
}

func appendToInboxFile(path string, urls []string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, u := range urls {
		if u = strings.TrimSpace(u); u != "" {
			fmt.Fprintln(f, u)
		}
	}
	return nil
}

// drainInboxFile reads and deletes the inbox, returning the URLs it held.
func drainInboxFile(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	os.Remove(path)
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if u := strings.TrimSpace(line); u != "" {
			out = append(out, u)
		}
	}
	return out
}

// DrainInbox adds any URLs waiting in the inbox. Safe to call repeatedly.
func (m *Manager) DrainInbox() {
	m.inboxMu.Lock()
	urls := drainInboxFile(filepath.Join(m.stateDir, "inbox"))
	m.inboxMu.Unlock()
	for _, u := range urls {
		m.Add(u) // HEAD probe + enqueue
	}
}

// inboxLoop drains the inbox on startup and then once a second until Close.
func (m *Manager) inboxLoop() {
	m.DrainInbox()
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-t.C:
			m.DrainInbox()
		}
	}
}
