package gui

import (
	"image/color"

	g "github.com/AllenDang/giu"
)

// Dark, cohesive palette with a single blue accent. Kept in one place so the
// whole UI stays consistent and is easy to retune.
var (
	colorBackground   = color.RGBA{R: 0x14, G: 0x17, B: 0x1f, A: 0xff} // near-black slate
	colorSurface      = color.RGBA{R: 0x1c, G: 0x21, B: 0x2c, A: 0xff} // panels / inputs
	colorSurfaceHi    = color.RGBA{R: 0x26, G: 0x2d, B: 0x3b, A: 0xff} // hovered surface
	colorBorder       = color.RGBA{R: 0x2f, G: 0x37, B: 0x47, A: 0xff}
	colorAccent       = color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff} // blue-500
	colorAccentHover  = color.RGBA{R: 0x60, G: 0x9a, B: 0xf8, A: 0xff}
	colorAccentActive = color.RGBA{R: 0x2b, G: 0x6c, B: 0xd6, A: 0xff}
	colorText         = color.RGBA{R: 0xe6, G: 0xea, B: 0xf2, A: 0xff}
	colorTextMuted    = color.RGBA{R: 0x8b, G: 0x95, B: 0xa7, A: 0xff}
	colorSuccess      = color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 0xff} // emerald for DONE
	colorDanger       = color.RGBA{R: 0xf8, G: 0x71, B: 0x71, A: 0xff} // red for errors
	colorHeaderBg     = color.RGBA{R: 0x22, G: 0x29, B: 0x36, A: 0xff}
)

// baseStyle returns a StyleSetter configured with the app theme. Wrap the
// window content in baseStyle().To(...) so every frame re-applies the theme
// (immediate-mode: styles must be pushed each frame).
func baseStyle() *g.StyleSetter {
	return g.Style().
		SetColor(g.StyleColorWindowBg, colorBackground).
		SetColor(g.StyleColorChildBg, colorSurface).
		SetColor(g.StyleColorPopupBg, colorSurface).
		SetColor(g.StyleColorText, colorText).
		SetColor(g.StyleColorTextDisabled, colorTextMuted).
		SetColor(g.StyleColorBorder, colorBorder).
		SetColor(g.StyleColorFrameBg, colorSurface).
		SetColor(g.StyleColorFrameBgHovered, colorSurfaceHi).
		SetColor(g.StyleColorFrameBgActive, colorSurfaceHi).
		SetColor(g.StyleColorButton, colorSurfaceHi).
		SetColor(g.StyleColorButtonHovered, colorBorder).
		SetColor(g.StyleColorButtonActive, colorAccentActive).
		SetColor(g.StyleColorHeader, colorSurfaceHi).
		SetColor(g.StyleColorHeaderHovered, colorSurfaceHi).
		SetColor(g.StyleColorHeaderActive, colorAccentActive).
		SetColor(g.StyleColorSeparator, colorBorder).
		SetColor(g.StyleColorTableHeaderBg, colorHeaderBg).
		SetColor(g.StyleColorTableBorderStrong, colorBorder).
		SetColor(g.StyleColorTableBorderLight, colorBorder).
		SetColor(g.StyleColorTableRowBg, colorBackground).
		SetColor(g.StyleColorTableRowBgAlt, colorSurface).
		SetColor(g.StyleColorPlotHistogram, colorAccent).
		SetColor(g.StyleColorTitleBg, colorHeaderBg).
		SetColor(g.StyleColorTitleBgActive, colorHeaderBg).
		SetColor(g.StyleColorScrollbarBg, colorBackground).
		SetColor(g.StyleColorScrollbarGrab, colorBorder).
		SetColor(g.StyleColorScrollbarGrabHovered, colorSurfaceHi).
		SetStyleFloat(g.StyleVarWindowRounding, 8).
		SetStyleFloat(g.StyleVarChildRounding, 8).
		SetStyleFloat(g.StyleVarPopupRounding, 8).
		SetStyleFloat(g.StyleVarFrameRounding, 6).
		SetStyleFloat(g.StyleVarGrabRounding, 6).
		SetStyleFloat(g.StyleVarScrollbarRounding, 6).
		SetStyleFloat(g.StyleVarFrameBorderSize, 1).
		SetStyle(g.StyleVarWindowPadding, 20, 18).
		SetStyle(g.StyleVarFramePadding, 12, 8).
		SetStyle(g.StyleVarItemSpacing, 10, 10).
		SetStyle(g.StyleVarItemInnerSpacing, 8, 6)
}

// primaryButton renders an accent-colored call-to-action button.
func primaryButton(label string, onClick func()) g.Widget {
	return g.Style().
		SetColor(g.StyleColorButton, colorAccent).
		SetColor(g.StyleColorButtonHovered, colorAccentHover).
		SetColor(g.StyleColorButtonActive, colorAccentActive).
		SetColor(g.StyleColorText, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}).
		To(g.Button(label).OnClick(onClick))
}

// coloredLabel renders a label in a specific color.
func coloredLabel(text string, col color.Color) g.Widget {
	return g.Style().SetColor(g.StyleColorText, col).To(g.Label(text))
}
