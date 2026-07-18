# "Send to Loot" — right-click any URL

A macOS Automator **Quick Action** that sends a selected/link URL to loot
from anywhere (browser, notes, Finder text selection, etc.).

## How it works

loot's TUI holds its state in memory, so an external process can't add to
it directly. Instead:

- `loot add <url>...` appends the URL(s) to an **inbox** file
  (`~/.loot/tmp/inbox`) and exits immediately.
- A **running** loot drains the inbox every second (and on launch) and
  starts each download (respecting the concurrency queue).
- If loot **isn't running**, the URL waits in the inbox and starts the
  next time you launch it.

The Quick Action is a "Run Shell Script" workflow that appends the URL straight
to the inbox file (`~/.loot/tmp/inbox`) — it deliberately does **not** launch the
`loot` binary, so there's no launch/permission failure mode.

## Install

```bash
make install-cli            # ensure `loot` is on your PATH (~/go/bin)
make install-quick-action   # copy the workflow into ~/Library/Services
```

Then right-click a URL or selected text → **Services ▸ Send to Loot**.
If it doesn't appear, enable it in **System Settings ▸ Keyboard ▸ Keyboard
Shortcuts ▸ Services ▸ Text** (or Files/URLs), then re-open the app you're using.

## Verify from a terminal

```bash
loot add "https://proof.ovh.net/files/1Mb.dat"
# → appears in a running loot within ~1s, or on next launch
```

## Build it yourself in Automator (reliable fallback)

If the prebuilt workflow doesn't show up, make it in ~30 seconds:

1. **Automator ▸ New ▸ Quick Action**.
2. "Workflow receives current" → **text** (or URLs) in **any application**.
3. Add a **Run Shell Script** action; set **Pass input: as arguments**.
4. Shell `/bin/zsh`, script:
   ```sh
   INBOX="$HOME/.loot/tmp"; mkdir -p "$INBOX"
   for u in "$@"; do printf '%s\n' "$u" >> "$INBOX/inbox"; done
   ```
5. Save as **Send to Loot**.

## Limitation

This queues URLs into the app; it does not run downloads headlessly in the
background. For downloads that start without the TUI/GUI open, a background daemon
would be needed (not built).
