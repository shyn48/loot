// Package theme holds the shared color palette used by both front-ends (the
// giu desktop GUI and the Bubble Tea TUI), so the app looks consistent
// regardless of how it is launched.
package theme

import "image/color"

// Dark, cohesive palette with a single blue accent.
var (
	Background   = color.RGBA{R: 0x14, G: 0x17, B: 0x1f, A: 0xff} // near-black slate
	Surface      = color.RGBA{R: 0x1c, G: 0x21, B: 0x2c, A: 0xff} // panels / inputs
	SurfaceHi    = color.RGBA{R: 0x26, G: 0x2d, B: 0x3b, A: 0xff} // hovered surface
	Border       = color.RGBA{R: 0x2f, G: 0x37, B: 0x47, A: 0xff}
	Accent       = color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff} // blue-500
	AccentHover  = color.RGBA{R: 0x60, G: 0x9a, B: 0xf8, A: 0xff}
	AccentActive = color.RGBA{R: 0x2b, G: 0x6c, B: 0xd6, A: 0xff}
	Text         = color.RGBA{R: 0xe6, G: 0xea, B: 0xf2, A: 0xff}
	TextMuted    = color.RGBA{R: 0x8b, G: 0x95, B: 0xa7, A: 0xff}
	Success      = color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 0xff} // emerald for DONE
	Danger       = color.RGBA{R: 0xf8, G: 0x71, B: 0x71, A: 0xff} // red for errors
	Warning      = color.RGBA{R: 0xf5, G: 0xc5, B: 0x18, A: 0xff} // amber for PAUSED
	Purple       = color.RGBA{R: 0xb5, G: 0x7e, B: 0xdc, A: 0xff} // for QUEUED
	HeaderBg     = color.RGBA{R: 0x22, G: 0x29, B: 0x36, A: 0xff}
)
