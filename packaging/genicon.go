//go:build ignore

// genicon renders the app icon as a 1024x1024 PNG that matches the GUI theme
// (dark slate background, blue accent "download into tray" glyph). Run it via
// `make icon`, which then turns the PNG into an .icns with sips + iconutil.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const size = 1024

// Theme colors, kept in sync with gui/theme.go.
var (
	bg     = color.RGBA{0x14, 0x17, 0x1f, 0xff}
	accent = color.RGBA{0x3b, 0x82, 0xf6, 0xff}
)

// sdRoundBox is the signed distance from p to a rounded box centered at the
// origin with the given half-size and corner radius (<=0 means inside).
func sdRoundBox(px, py, halfX, halfY, r float64) float64 {
	qx := math.Abs(px) - halfX + r
	qy := math.Abs(py) - halfY + r
	outside := math.Hypot(math.Max(qx, 0), math.Max(qy, 0))
	inside := math.Min(math.Max(qx, qy), 0)
	return outside + inside - r
}

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	const (
		cx = size / 2.0
		cy = size / 2.0
	)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x), float64(y)

			// Rounded-square background (transparent outside).
			if sdRoundBox(fx-cx, fy-cy, 512, 512, 230) > 0 {
				img.Set(x, y, color.RGBA{})
				continue
			}

			c := bg
			if inGlyph(fx, fy) {
				c = accent
			}
			img.Set(x, y, c)
		}
	}

	f, err := os.Create("packaging/AppIcon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

// inGlyph is the "download into a tray" mark: a downward arrow above an open
// tray, both in the accent color.
func inGlyph(x, y float64) bool {
	const cx = size / 2.0

	// Arrow stem.
	if x >= cx-46 && x <= cx+46 && y >= 268 && y <= 566 {
		return true
	}
	// Arrowhead: half-width shrinks linearly to the tip at y=742.
	if y >= 560 && y <= 742 {
		hw := 190 * (742 - y) / (742 - 560)
		if math.Abs(x-cx) <= hw {
			return true
		}
	}
	// Open tray (U-shape) beneath the arrow.
	const (
		trayL, trayR = cx - 232, cx + 232
		trayTop      = 762.0
		trayBot      = 862.0
	)
	if y >= trayBot-64 && y <= trayBot && x >= trayL && x <= trayR { // bottom bar
		return true
	}
	if y >= trayTop && y <= trayBot && x >= trayL && x <= trayL+64 { // left wall
		return true
	}
	if y >= trayTop && y <= trayBot && x >= trayR-64 && x <= trayR { // right wall
		return true
	}
	return false
}
