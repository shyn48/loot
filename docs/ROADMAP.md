# godownloader — Roadmap / Deferred Improvements

Ideas captured for later. **TUI polish is being done first** (see the reference in
`docs/tui-reference.png` intent). This project is intentionally **macOS-only** — cross-platform
support is explicitly out of scope.

## Engine / robustness
- ✅ **Section retry with backoff** — done. Each section retries up to 4× with exponential
  backoff (200ms→5s cap), recomputing the remaining range each attempt so retries resume
  within the section. Single-stream downloads get a restart-retry.
- ✅ **Concurrency queue** — done. Manager caps active downloads at `maxActive` (default 3); extras
  wait in `StateQueued` and auto-promote when a slot frees. Lights up the TUI "Queued" column.
  (Making `maxActive` user-configurable is part of the config-file item below.)
- **Bandwidth throttle** — per-download and global.
- ✅ **Adaptive section count** — done. ~1 section per 2 MiB (Manager.bytesPerSection), capped at
  20; small files use a single stream.
- ✅ **`Content-Disposition` filename** — done. Uses the server-provided filename when present,
  falling back to URL-based guessing.
- **`ETag`/`Last-Modified` resume validation** — restart from zero if the file changed server-side.
- **Disk-space precheck** before starting a download.
- **Checksum verification** — server `Content-MD5` or a user-supplied hash.

## Features
- ✅ **Config file** — done. `~/.config/godownloader/config.toml` sets `download_dir` (default
  `~/Downloads`), `max_active`, and `section_size_mb`. Written with defaults on first run.
  (Bandwidth throttle setting still TODO.)
- **Clipboard add** — grab a URL from the clipboard with a keypress.
- **Batch add** — multiple URLs at once / import from a text file.
- **macOS completion notifications** — via `osascript` / `terminal-notifier`.
- **History management** — ⚠️ meta files persist after completion, so completed downloads accumulate
  forever and reload on every launch. Add auto-clear on completion or a "clear completed" action.
- **Categories / scheduler** — deferred from the original IDM brainstorm.

## GUI (giu) parity
- Add **speed/ETA columns** to the desktop window so it matches the TUI.

## Distribution
- **GoReleaser + Homebrew tap** (`brew install shyn48/tap/godownloader`).
- **Notarized universal `.app`** for sharing the GUI with other Macs.
- **README with a TUI GIF**.
- **Rename the Go module** from the leftover `simple-gui` to `github.com/shyn48/gownloader`.

## Explicitly out of scope
- Cross-platform (Linux/Windows). macOS only, by choice.
