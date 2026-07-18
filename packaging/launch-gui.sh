#!/bin/bash
# Bundle entry point: the .app is the desktop GUI, so launch the binary with
# --gui. (The same binary run from a terminal with no flag opens the TUI.)
exec "$(dirname "$0")/godownloader-bin" --gui
