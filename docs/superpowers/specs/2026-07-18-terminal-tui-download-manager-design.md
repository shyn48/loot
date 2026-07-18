# Terminal TUI Download Manager — Design

**Date:** 2026-07-18
**Status:** ✅ Implemented (and extended). This is the original, historical design; the app is
now named **Loot** and has grown well past this spec (concurrency queue, retry/backoff, config
file, ETag resume validation, clipboard/batch add, notifications, Quick Action, etc.). See
[`docs/ROADMAP.md`](../../ROADMAP.md) for current status and what remains.

## Goal

Add a terminal UI (TUI) as the primary front-end for the download manager, launched by
running `godownloader` in a terminal. Keep the existing giu desktop GUI working via
`godownloader --gui`. Deliver an IDM-style experience: a live download list with per-item
progress bar, percentage, transfer speed, and ETA, plus **pause/resume** that survives an
app restart.

## Scope

**In scope (v1):**
- New Bubble Tea TUI: download list, add-URL input, pause/resume/delete/open-folder, help.
- Engine rewrite around a shared `core.Manager` that owns download state and exposes
  progress, so both front-ends render from one source of truth.
- Live progress reporting: % done, smoothed transfer speed, ETA.
- Pause/resume of segmented downloads, including resume after the process is killed and
  relaunched (restart survival).
- Rewire the giu GUI to read from `core.Manager` (earns it progress bars too).
- `godownloader` binary installable onto `PATH`; `--gui` flag selects the window front-end.
- Unit tests for the engine (sectioning, resume-range math, metadata round-trip,
  pause→resume, no-Range fallback) and for TUI `Update` logic.

**Out of scope (v1):**
- Categories sidebar, scheduler, browser integration, clipboard URL auto-catch.
- Notarized/universal distribution (unchanged from current ad-hoc local build).
- True resume for servers that do **not** support HTTP `Range` (see Limitations).

## Architecture

```
main.go ── parse flags ──▶ default → tui.Run(manager)
                           --gui   → gui.Run(manager)
                                        │
                        ┌───────────────┴───────────────┐
                   tui/ (Bubble Tea)              gui/ (giu)
                        └───────────────┬───────────────┘
                                  core.Manager
                                  (owns []*Job, thread-safe)
                                        │
                                    core.Job
```

- **`core.Manager`** — single owner of all download state. Thread-safe (guarded by a mutex).
  Methods:
  - `Add(url string) (jobID string, err error)` — HEAD-probes, creates a `Job`, persists
    its metadata, starts downloading, returns immediately.
  - `Pause(id string)` / `Resume(id string)`
  - `Remove(id string)` — cancels if running, deletes temp files + metadata.
  - `OpenFolder()` — reveal the downloads directory.
  - `Snapshot() []JobStatus` — an immutable copy of current state for rendering.
  - `LoadPersisted()` — on startup, scan the state dir and rebuild jobs as **Paused**.
- **`core.Job`** — one download. Holds URL, filename, target path, total size,
  `acceptRanges`, section boundaries, status, a `context.CancelFunc`, and a progress
  sampler. Owns its goroutines.
- **Front-ends** are thin: they call Manager methods on key/click events and render
  `Snapshot()`. No download state lives in the front-ends anymore.

`JobStatus` (the render DTO) fields: `ID, Name, Size int64, Downloaded int64, Percent float64,
SpeedBytesPerSec float64, ETASeconds int, State, Resumable bool, Err string`.

## Front-end selection (`main.go`)

```
manager := core.NewManager()
manager.LoadPersisted()
if guiFlag { gui.Run(manager) } else { tui.Run(manager) }
```

`core.Start()` (directory setup) is folded into `core.NewManager()`.

## Engine: progress, pause/resume, persistence

### States

`Queued → Downloading → (Paused ⇄ Downloading) → Merging → Done | Failed`

### Sections and the disk-as-truth rule

- A download with known size **and** `Accept-Ranges: bytes` splits into N sections
  (N = 20, existing default, but 1 for small files), each streaming to
  `section-<i>-<file>.tmp`.
- **Bytes downloaded for section i = current size of its temp file.** No counter needs to be
  persisted for progress; the temp files are authoritative. Total downloaded = sum of temp
  file sizes.
- Downloads without a known size or without Range support run as a single stream to one
  temp file (or directly to target) and are marked **not resumable**.

### Metadata (`<id>.meta.json` in the state dir)

Written once on `Add`. Fields: `id, url, filename, targetPath, size, totalSection,
sections [][2]int64, acceptRanges, createdAt`. Never rewritten during download — everything
mutable is derived from temp file sizes on disk. This is what makes restart survival cheap.

### Pause

`Job.Pause()` calls its `context.CancelFunc`. Each section goroutine is doing
`io.Copy(tempFile, resp.Body)` with a context-aware reader; on cancel it stops, flushes,
and closes its temp file. Partial temp files remain. State → `Paused`.

### Resume

For each section i: `have = size(section tmp file)`; if `have < (end-start+1)`, issue
`GET` with `Range: bytes=(start+have)-(end)`, verify `206`, and **append** (`O_APPEND`) the
body to the temp file. When all sections are full → merge → `Done`.

### Restart survival

`LoadPersisted()` scans the state dir for `*.meta.json`, and for each: if the final target
exists and no temp files remain → `Done` (skip); otherwise reconstruct the `Job` from the
metadata, compute progress from existing temp file sizes, and set state `Paused`. The user
presses `r` to continue. (Auto-resume on launch is deliberately not done in v1 — the user
decides.)

### Progress / speed / ETA

A per-Manager ticker samples total downloaded bytes for each running job every ~500 ms.
`speed = EWMA(Δbytes / Δt)` (smoothing factor ~0.3 to avoid jitter). `eta = remaining / speed`
(shown as `—` when speed is ~0). Percent = downloaded / size.

### Concurrency safety

- Manager mutates its job map under a mutex; `Snapshot()` returns copies.
- Per-section temp files are written by exactly one goroutine each (no shared writer).
- Speed sampling reads temp file sizes / atomic counters; never touches front-end state.
- Cancellation is context-based; no goroutine is force-killed mid-write.

## TUI (Bubble Tea + Lip Gloss + Bubbles)

Elm architecture: `Model{ manager, rows []JobStatus, cursor int, adding bool, input textinput,
help }`, `Update(msg)`, `View()`.

- A `tea.Tick` (~10 Hz) triggers `Model.rows = manager.Snapshot()` and a re-render.
- Components: `bubbles/progress` for the bars, `bubbles/textinput` for the add-URL line,
  `bubbles/help` + `bubbles/key` for the key hints.
- Styling via Lip Gloss. The palette color values currently in `gui/theme.go` move to a new
  standalone `theme` package (plain `color.RGBA`/hex constants); the giu `StyleSetter` stays
  in `gui` and references them, and the TUI's Lip Gloss styles reference the same constants,
  so both front-ends share one palette.

Layout:

```
 godownloader                                    3 active · 12.4 MB/s
 ────────────────────────────────────────────────────────────────
  NAME              PROGRESS                 SIZE    SPEED     ETA
 ▸ ubuntu.iso       ███████████░░░░  78%     1.4 GB  9.2 MB/s  0:12
   video.mp4        ███░░░░░░░░░░░░  24%     820 MB  1.1 MB/s  1:40
   song.mp3         ██████████████ done      4.1 MB  —         —
   backup.zip       ██████░░░░░░░░  52%  ⏸   2.0 GB  paused    —
 ────────────────────────────────────────────────────────────────
  a add   p pause   r resume   d delete   o open   ? help   q quit
```

Keys: `a` add (opens input; `enter` confirm, `esc` cancel), `p`/`r` pause/resume selected,
`d` delete selected, `o` open downloads folder, `j`/`k` or arrows navigate, `?` toggle help,
`q`/`ctrl+c` quit. On quit, v1 pauses all in-flight downloads so their partial temp files and
metadata persist and are resumable on next launch (this matches the restart-survival story).

## Dependencies

Add: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`,
`github.com/charmbracelet/bubbles`. Bump the `go` directive in `go.mod` as required by the
Charm libraries (toolchain is 1.25). The giu/cgo dependency stays (GUI front-end).

## Binary and install

- Module stays `simple-gui` internally; the built binary is named `godownloader`.
- `Makefile`:
  - `make build` → `dist/godownloader`
  - `make install-cli` → copy `godownloader` to `$(go env GOPATH)/bin` (on PATH) —
    the primary "run `godownloader` in the terminal" path.
  - `make app` / `make install` (existing) still build the giu `.app`, whose executable now
    launches with `--gui`.
- Running `godownloader` with no args opens the TUI.

## Testing

- **Engine unit tests** against `net/http/httptest`:
  - Section boundary math for various sizes (incl. size < N, size not divisible by N).
  - Full segmented download equals source bytes (checksum).
  - Resume: download part, pause, resume, assert final bytes == source and only the missing
    ranges were re-requested (Range header assertions).
  - Restart survival: build metadata + partial temp files on disk, `LoadPersisted()`,
    resume, assert completion.
  - No-Range server: falls back to single stream, marked non-resumable.
  - `206` enforcement: a server that ignores Range yields an error, not corruption.
- **TUI tests:** drive `Model.Update` with synthetic key/tick messages; assert cursor
  movement, add-mode toggling, and that key actions call the right Manager methods (Manager
  behind a small interface so tests use a fake).

## Limitations (called out honestly)

- Servers without `Accept-Ranges` / unknown Content-Length: not resumable; pause = cancel,
  resume = restart from zero. Shown as non-resumable in the UI.
- No integrity verification (checksums) of completed files beyond size.
- Single machine, single user; no download queue limits/throttling in v1.

## Risks

- **Pause-mid-write correctness** is the highest-risk area: a section must stop cleanly so
  its temp file size is an accurate resume point. Mitigation: context-aware copy, flush +
  close on cancel, and the resume tests above.
- **GUI rewire** touches working code. Mitigation: mechanical change to a stable
  `core.Manager` API; GUI behavior verified after.
- **Charm/go version**: bumping the `go` directive could interact with the giu cgo build.
  Mitigation: verify `go build ./...` for both front-ends after the bump.
