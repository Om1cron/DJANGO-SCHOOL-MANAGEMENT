
package cryptonym

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	"github.com/blockpane/prettyfyne"
	"image/color"
)

func ExLightTheme() prettyfyne.PrettyTheme {
	lt := prettyfyne.ExampleMaterialLight
	lt.TextSize = 14
	lt.IconInlineSize = 20
	lt.FocusColor = lt.HoverColor
	lt.Padding = 3
	lt.FocusColor = &color.RGBA{R: 23, G: 11, B: 64, A: 128}
	return lt
}

func ExGreyTheme() prettyfyne.PrettyTheme {
	lt := prettyfyne.ExampleCubicleLife
	lt.TextSize = 14
	lt.TextColor = &color.RGBA{R: 0, G: 0, B: 0, A: 255}
	lt.IconInlineSize = 20
	lt.FocusColor = &color.RGBA{R: 24, G: 24, B: 24, A: 127}
	lt.Padding = 3
	lt.BackgroundColor = &color.RGBA{R: 210, G: 210, B: 210, A: 255}
	return lt
}

var (
	fioTertiary  = &color.RGBA{R: 46, G: 102, B: 132, A: 255}
	fioPrimary   = &color.RGBA{R: 30, G: 62, B: 97, A: 255}
	fioSecondary = &color.RGBA{R: 0, G: 0, B: 0, A: 162}
	lightestGrey = &color.RGBA{R: 200, G: 200, B: 200, A: 255}
	lightGrey    = &color.RGBA{R: 155, G: 155, B: 155, A: 127}
	grey         = &color.RGBA{R: 99, G: 99, B: 99, A: 255}
	//greyBorder   = &color.RGBA{R: 35, G: 35, B: 35, A: 8}
	darkGrey    = &color.RGBA{R: 28, G: 28, B: 29, A: 255}
	darkerGrey  = &color.RGBA{R: 24, G: 24, B: 24, A: 255}
	darkestGrey = &color.RGBA{R: 15, G: 15, B: 17, A: 255}
)

// FioCustomTheme is a simple demonstration of a bespoke theme loaded by a Fyne app.
type FioCustomTheme struct {
}

func (FioCustomTheme) BackgroundColor() color.Color {