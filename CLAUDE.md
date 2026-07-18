# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A download manager written in Go with **two front-ends over one shared engine**: a terminal
UI (Bubble Tea) that is the default, and a desktop window (giu / Dear ImGui). The command
`godownloader` opens the TUI; `godownloader --gui` opens the window. The module is named
`simple-gui`; internal imports use that prefix (e.g. `simple-gui/core`).

## Commands

```bash
go run .            # run the TUI (dev)
go run . --gui      # run the giu desktop window (dev)
go test ./...       # run the test suite (core + tui pure logic)
go test -race ./core/   # engine is concurrent — race-check it
go vet ./...
make install-cli    # build and copy `godownloader` to $(go env GOPATH)/bin
make app            # build the signed macOS .app (double-click → GUI)
make install        # copy the .app into /Applications
```

giu depends on OpenGL/GLFW via CGO, so a C toolchain is required to build. The Charm libraries
require the `go` directive in `go.mod` (currently 1.24.x).

## Architecture

`main.go` parses a `--gui` flag, constructs one `core.Manager`, calls `LoadPersisted()`, then
hands the manager to either `tui.Run(m)` or `gui.Run(m)`.

- **`core/`** — the engine and the single source of truth for download state.
  - **`core.Manager`** (`manager.go`) owns `[]*Job` plus the download/state dirs (folds the old
    dir-setup). API: `Add`, `Pause`, `Resume`, `Remove`, `OpenFolder`, `Snapshot`, `LoadPersisted`,
    `PauseAll`, `Close`. A 500ms ticker samples an EWMA transfer speed per running job.
  - **`core.Job`** (`job.go`) is one download. `run(ctx, tempDir)` downloads N sections
    concurrently (via `Range`) into `section-<i>-<file>.tmp`, then merges. **Bytes-downloaded is
    derived from temp file sizes** — the disk is authoritative, so nothing mutable is persisted.
    `ctxReader` makes `io.Copy` abort promptly on pause. Non-segmented downloads (no `Accept-Ranges`
    or unknown size) use `singleStream` and are not resumable.
  - **Pause/resume**: `Pause` cancels the job context (state flips to `Paused` only after `run`
    returns, so temp sizes are an accurate resume point); `Resume` re-enters `run`, which requests
    only each section's missing bytes (`remainingRange`).
  - **Restart survival**: `<id>.meta.json` (`meta.go`) is written once on `Add`. `LoadPersisted`
    rebuilds jobs from those files as `Paused` (progress re-measured from temp files); the user
    resumes them. Nothing auto-resumes.
  - **`core.Controller`** (`controller.go`) is the interface the front-ends depend on.
  - `download.go` now holds only shared HTTP helpers + `GetFileDetails` (single HEAD probe →
    `FileDetails{Name, Size, AcceptRanges}`). `sections.go` has the pure math (`computeSections`,
    `remainingRange`, `ewma`, `etaSeconds`). Paths: `core/consts.go` +
    `GetDownloadPath`/`GetTempPath` (`~/shyn-dl-manager/out/{downloads,tmp}`).

- **`tui/`** — Bubble Tea front-end. `Model`/`Update` (`model.go`) hold no download state; they
  render `Controller.Snapshot()` on a ~150ms tick and forward key actions
  (`a` add, `p/r` pause/resume, `d` delete, `o` open, `q` quit→`PauseAll`). `view.go` is Lip Gloss
  rendering (pure `bar`/`formatETA` helpers). Logic is tested with a fake controller, no terminal.

- **`gui/`** — giu front-end, rewired onto `core.Manager` (`gui.Run(m)` stores it in a package
  var). Immediate-mode: `loop()` re-renders every frame from `manager.Snapshot()`, with a
  `g.ProgressBar` column and pause/resume/open/remove buttons. `states.go` now holds only
  UI-local view flags (no download state).

- **`theme/`** — shared color palette (`color.RGBA` constants) used by both `gui` (giu styles)
  and `tui` (Lip Gloss). **`helper/`** — `IsValidUrl` and `HumanBytes`.

## Things to know before changing code

- **`core.Manager` is the concurrency seam.** Jobs are mutated by download/sample goroutines and
  read by `Snapshot`; guard job fields with `Job.mu` and manager fields with `Manager.mu`. Run
  `go test -race ./core/` after any change here.
- **Front-ends must stay thin.** They render snapshots and call Manager methods — do not add
  download state to `tui` or `gui`.
- **Platform-specific**: `OpenFolder` shells out to the macOS `open` command.
- **Not resumable** when the server lacks `Accept-Ranges` or the size is unknown: pause cancels,
  resume restarts from zero. Surface that (`JobStatus.Resumable`) rather than implying otherwise.
- The TUI needs a real TTY; its rendering can't be unit-tested (logic is, via `Model.Update`).
