#!/bin/bash
# Bundle entry point: open the Loot TUI in a new Terminal.app window and close
# that window when the TUI exits (user presses q). The TUI needs a real TTY,
# which Terminal provides.

# Prefer the PATH-installed binary (updatable via `make install-cli`); fall back
# to the copy shipped inside the bundle.
BIN="$HOME/go/bin/loot"
if [ ! -x "$BIN" ]; then
    BIN="$(cd "$(dirname "$0")" && pwd)/loot-bin"
fi

# Run loot as a normal foreground child of the shell (NOT via exec) so Terminal's
# `busy` flag stays true while it runs — with exec the shell is replaced and busy
# reads false immediately, closing the window the instant the TUI appears.
# Wait for the tab to become busy, then for it to finish, then close its window.
osascript <<APPLESCRIPT
tell application "Terminal"
    activate
    set theTab to do script "'$BIN'"
    try
        set waited to 0.0
        repeat until (theTab is busy) or (waited > 5.0)
            delay 0.1
            set waited to waited + 0.1
        end repeat
        repeat while theTab is busy
            delay 0.3
        end repeat
        close (first window whose tabs contains theTab) saving no
    end try
end tell
APPLESCRIPT
