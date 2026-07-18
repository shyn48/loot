#!/bin/bash
# Bundle entry point: open the godownloader TUI in a new Terminal.app window.
# (The TUI needs a real TTY, which Terminal provides.)

# Prefer the PATH-installed binary (updatable via `make install-cli`); fall back
# to the copy shipped inside the bundle.
BIN="$HOME/go/bin/godownloader"
if [ ! -x "$BIN" ]; then
    BIN="$(cd "$(dirname "$0")" && pwd)/godownloader-bin"
fi

# `exec` replaces the Terminal shell with godownloader so the window belongs to
# it; the path is single-quoted for the shell in case it contains spaces.
osascript \
    -e "tell application \"Terminal\" to do script \"exec '$BIN'\"" \
    -e 'tell application "Terminal" to activate'
