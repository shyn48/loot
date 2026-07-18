package theme

import "testing"

func TestPaletteOpaque(t *testing.T) {
	for name, c := range map[string]struct{ R, G, B, A uint8 }{
		"bg":     {Background.R, Background.G, Background.B, Background.A},
		"accent": {Accent.R, Accent.G, Accent.B, Accent.A},
	} {
		if c.A != 0xff {
			t.Fatalf("%s not opaque: A=%x", name, c.A)
		}
	}
}
