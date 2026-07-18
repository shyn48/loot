# loot — Roadmap / Deferred Improvements

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
- ✅ **Config file** — done. `~/.config/loot/config.toml` sets `download_dir` (default
  `~/Downloads`), `max_active`, and `section_size_mb`. Written with defaults on first run.
  (Bandwidth throttle setting still TODO.)
- ✅ **Send to Loot (Automator Quick Action)** — done. `loot add <url>` appends to
  an inbox the running app drains every second; ships an Automator Quick Action to right-click any
  URL. See `docs/QUICK-ACTION.md`. (A background daemon for headless downloads is still out of scope.)
- ✅ **Clipboard add** — done. Pressing `a` prefills the input from the clipboard when it's a URL.
- ✅ **Batch add** — done. The add input accepts multiple whitespace/newline-separated URLs.
  (Import-from-file could be a follow-up.)
- ✅ **macOS completion notifications** — done. `osascript` notification on completion, toggled by
  `notifications` in the config.
- ✅ **History management** — done. `c` clears completed downloads (removes them + their metadata,
  keeps the files) via `Manager.ClearCompleted`.
- **Categories / scheduler** — deferred from the original IDM brainstorm.

## GUI (giu) parity
- Add **speed/ETA columns** to the desktop window so it matches the TUI.

## Distribution
- **GoReleaser + Homebrew tap** (`brew install shyn48/tap/loot`).
- **Notarized universal `.app`** for sharing the GUI with other Macs.
- **README with a TUI GIF**.
- **Rename the Go module** from the leftover `simple-gui` to `github.com/shyn48/loot`.

## Explicitly out of scope
- Cross-platform (Linux/Windows). macOS only, by choice.
