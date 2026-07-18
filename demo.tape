# VHS tape to record an animated demo of Loot.
#   brew install vhs        # (needs ffmpeg + ttyd)
#   make install-cli        # so `loot` is on your PATH
#   vhs demo.tape           # produces demo.gif
#
# Uses public OVH test files, so it needs a network connection.

Output demo.gif

Set Shell "zsh"
Set FontSize 15
Set Width 1100
Set Height 640
Set Padding 16
Set Theme "Catppuccin Mocha"

# Launch the TUI.
Type "loot"  Enter
Sleep 1.5s

# Add a large download (shows the queue + live progress).
Type "a"  Sleep 700ms
Type "https://proof.ovh.net/files/100Mb.dat"  Enter
Sleep 1s

# Add a second so the concurrency queue is visible.
Type "a"  Sleep 500ms
Type "https://proof.ovh.net/files/10Mb.dat"  Enter
Sleep 4s

# Pause the selected one, then resume it.
Type "p"  Sleep 1.5s
Type "r"  Sleep 3s

# Show the filter.
Type "/"  Sleep 400ms
Type "100"  Sleep 1.5s
Escape  Sleep 800ms

# Quit.
Type "q"
Sleep 500ms
