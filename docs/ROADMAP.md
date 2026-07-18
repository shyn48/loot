# godownloader — Roadmap / Deferred Improvements

Ideas captured for later. **TUI polish is being done first** (see the reference in
`docs/tui-reference.png` intent). This project is intentionally **macOS-only** — cross-platform
support is explicitly out of scope.

## Engine / robustness
- **Section retry with backoff** — one flaky section shouldn't fail the whole download.
- ✅ **Concurrency queue** — done. Manager caps active downloads at `maxActive` (default 3); extras
  wait in `StateQueued` and auto-promote when a slot frees. Lights up the TUI "Queued" column.
  (Making `maxActive` user-configurable is part of the config-file item below.)
- **Bandwidth throttle** — per-download and global.
- **Adaptive section count** — don't split a 50 KB file into 20 sections.
- **`Content-Disposition` filename** — use the server-provided name instead of guessing from the URL.
- **`ETag`/`Last-Modified` resume validation** — restart from zero if the file changed server-side.
- **Disk-space precheck** before starting a download.
- **Checksum verification** — server `Content-MD5` or a user-supplied hash.

## Features
- **Config file** (`~/.config/godownloader/config.toml`): download dir, max concurrency, default
  sections, throttle — replacing today's hardcoded `~/shyn-dl-manager` and `TotalSection = 20`.
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
