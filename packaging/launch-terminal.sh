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

# do script returns the tab; we wait until it's no longer busy (the TUI has
# exited) and then close its window. The single quotes handle a path with spaces.
osascript <<APPLESCRIPT
tell application "Terminal"
    activate
    set theTab to do script "exec '$BIN'"
    try
        repeat while theTab is busy
            delay 0.3
        end repeat
        close (first window whose tabs contains theTab) saving no
    end try
end tell
APPLESCRIPT
