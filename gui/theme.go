package gui

import (
	"image/color"

	g "github.com/AllenDang/giu"

	"simple-gui/theme"
)

// Palette aliases so the rest of the giu layer keeps its short local names while
// the actual values live in the shared theme package (also used by the TUI).
var (
	colorBackground   = theme.Background
	colorSurface      = theme.Surface
	colorSurfaceHi    = theme.SurfaceHi
	colorBorder       = theme.Border
	colorAccent       = theme.Accent
	colorAccentHover  = theme.AccentHover
	colorAccentActive = theme.AccentActive
	colorText         = theme.Text
	colorTextMuted    = theme.TextMuted
	colorSuccess      = theme.Success
	colorDanger       = theme.Danger
	colorHeaderBg     = theme.HeaderBg
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
