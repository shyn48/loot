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

# do script returns the tab immediately — before the process has started — so we
# first wait for the tab to become busy, THEN wait for it to finish, and only
# then close its window. (Closing on the first `busy` read would slam the window
# shut right after it opened.)
osascript <<APPLESCRIPT
tell application "Terminal"
    activate
    set theTab to do script "exec '$BIN'"
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
