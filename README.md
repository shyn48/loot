# Loot

A fast, resumable **terminal download manager for macOS** — segmented multi-connection
downloads with pause/resume, a live `btop`-style TUI, and a desktop window when you want it.

![Loot demo](demo.gif)

```text
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│ Loot  │ Active: 2 │ Total Speed: 3.31 MB/s │ Completed: 2 │ Queued: 1 │ Errors: 0     10:42:31 │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│   Name                        Progress                  Size      Speed       ETA Status       │
│ ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌ │
│ ▸ ubuntu-24.04-desktop-amd6…  ■■■■■■■■■■■···  78%    4.38 GB  2.23 MB/s  00:00:52 downloading  │
│   fedora-workstation-40.iso   ■■■■■■········  45%    5.22 GB  1.08 MB/s  00:01:37 downloading  │
│   archlinux-2024.05.01-x86_…  ■■■■··········  32%    1.10 GB          —         — paused       │
│   linux-firmware-20240501.t…  □□□□□□□□□□□□□□   0%     604 MB          —         — queued       │
│ ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌ │
│   neovim-linux-x86_64.appim…  ■■■■■■■■■■■■■■ 100%    14.9 MB          —  00:00:00 completed    │
│   curl-8.7.1.tar.xz           ■■■■■■■■■■■■■■ 100%    3.13 MB          —  00:00:00 completed    │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
╭────────────────────────────────────────────────────────────────────────────────────────────────╮
│ File: ubuntu-24.04-desktop-a… │ Speed (MB/s)                  │ Progress ───────────────────── │
│ URL : https://releases.ubunt…    3.2    ▂▄▆█         ▁▃▅▇       ■■■■■■■■■■■■■■■■■■■·····  78%  │
│ Path: /Users/you/Downloads/u…        ▃▅█████     ▁▃▅▇████       Started: 17:52:55              │
│ Connections: 20                  1.6 ███████  ▂▄▆████████  ▂▄   Elapsed: 00:02:19              │
│ Downloaded: 3.42 GB / 4.38 GB    0.0 ███████▆████████████▅███   ETA:     00:00:52              │
╰────────────────────────────────────────────────────────────────────────────────────────────────╯
 a add │ p pause │ r resume │ d delete │ o open │ / filter │ ? help │ q quit
```

> A static snapshot — colors and the live speed sparkline animate in the GIF above (and when you run it).

## Features

- **Segmented downloads** — up to 20 parallel connections per file (adaptive to size).
- **Pause / resume** that survives quitting and relaunching — partial pieces are kept on disk.
- **Resume safety** — if a file changes on the server (ETag/Last-Modified) between pause and
  resume, the stale partial is discarded and the download restarts, so you never get a corrupt mix.
- **Retry with backoff** — a flaky connection retries the affected section instead of failing.
- **Concurrency queue** — a cap on simultaneous downloads; the rest wait as `queued`.
- **Live TUI** — progress bars, transfer speed, ETA, a per-download speed graph, and filtering.
- **Two front-ends, one engine** — the terminal UI by default, or a desktop window with `--gui`.
- **Quality-of-life** — clipboard add, batch add, completion notifications, "Send to Loot"
  right-click Quick Action, and a config file.

## Install

Requires Go and a C toolchain (the desktop window links OpenGL/GLFW via cgo).

```bash
git clone <this repo> && cd loot
make install-cli          # build and put `loot` on your PATH (~/go/bin)
make install              # build the macOS .app into /Applications (opens the TUI in Terminal)
make install-quick-action # optional: right-click any URL → Services → Send to Loot
```

## Usage

Run `loot` in a terminal (or double-click **Loot** in Applications). Keys:

| Key | Action |
|-----|--------|
| `a` | Add a download (prefills a URL from the clipboard; paste several for batch add) |
| `p` / `r` | Pause / resume the selected download |
| `d` | Delete the selected row |
| `c` | Clear completed downloads (keeps the files) |
| `o` | Open the downloads folder |
| `/` | Filter the list by name |
| `j`/`k` or ↑/↓ | Move the cursor · `?` help · `q` quit |

**Adding from anywhere:** with Loot running, right-click a URL in any app →
**Services ▸ Send to Loot**, or run `loot add <url>` from a terminal.

**Desktop window:** `loot --gui` opens the same downloads in a giu window.

## Configuration

`~/.config/loot/config.toml` (created on first run):

```toml
download_dir = "~/Downloads"
max_active = 3          # simultaneous downloads
section_size_mb = 2     # ~1 parallel section per N MB (max 20 sections)
notifications = true    # macOS notification when a download completes
```

## How it works

`main.go` builds one `core.Manager` — the single source of truth for all download state — and
hands it to either the TUI (`tui/`, Bubble Tea) or the giu GUI (`gui/`). Each download runs in N
sections streamed to temp files; **bytes-downloaded is derived from the temp file sizes**, so the
disk is authoritative and resume is just "request the missing ranges." A once-written
`<id>.meta.json` lets Loot rebuild in-flight downloads after a restart. The front-ends are thin —
they render `Manager.Snapshot()` and forward key/click actions.

## Build & test

```bash
make run             # run the TUI from source
make gui             # run the desktop window from source
go test ./...        # unit + integration tests (httptest-based engine tests)
go test -race ./core/ # the engine is concurrent — race-check it
```

## Notes

- **macOS only**, by design.
- The `.app` is **ad-hoc signed** for your machine (no Apple Developer account). To share it with
  other Macs it would need notarizing.
- Roadmap and deferred ideas: [`docs/ROADMAP.md`](docs/ROADMAP.md).
  Right-click integration details: [`docs/QUICK-ACTION.md`](docs/QUICK-ACTION.md).
