
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
	return darkGrey
}

func (FioCustomTheme) ButtonColor() color.Color {
	return darkerGrey
}

func (FioCustomTheme) DisabledButtonColor() color.Color {
	//return darkestGrey
	return darkGrey
}

func (FioCustomTheme) HyperlinkColor() color.Color {
	return fioTertiary
}

func (FioCustomTheme) TextColor() color.Color {
	return lightestGrey
}

func (FioCustomTheme) DisabledTextColor() color.Color {
	return lightGrey
}

func (FioCustomTheme) IconColor() color.Color {
	return fioTertiary
}

func (FioCustomTheme) DisabledIconColor() color.Color {
	return grey
}

func (FioCustomTheme) PlaceHolderColor() color.Color {
	return fioPrimary
}

func (FioCustomTheme) PrimaryColor() color.Color {
	return fioPrimary
}

func (FioCustomTheme) HoverColor() color.Color {
	return fioSecondary
}

func (FioCustomTheme) FocusColor() color.Color {
	return &color.RGBA{R: 93, G: 93, B: 93, A: 124}
}

func (FioCustomTheme) ScrollBarColor() color.Color {
	//return greyBorder
	//return fioPrimary
	return &color.RGBA{R: 26, G: 20, B: 60, A: 128}
}

func (FioCustomTheme) ShadowColor() color.Color {
	return &color.RGBA{R: 2, G: 0, B: 4, A: 166}
}

func (FioCustomTheme) TextSize() int {
	return 14
}

func (FioCustomTheme) TextFont() fyne.Resource {
	return theme.DefaultTextFont()
}

func (FioCustomTheme) TextBoldFont() fyne.Resource {
	return theme.DefaultTextBoldFont()
}

func (FioCustomTheme) TextItalicFont() fyne.Resource {
	return theme.DefaultTextBoldItalicFont()
}

func (FioCustomTheme) TextBoldItalicFont() fyne.Resource {
	return theme.DefaultTextBoldItalicFont()
}

func (FioCustomTheme) TextMonospaceFont() fyne.Resource {
	return theme.DefaultTextMonospaceFont()
}

func (FioCustomTheme) Padding() int {
	return 3
}

func (FioCustomTheme) IconInlineSize() int {
	return 20
}

func (FioCustomTheme) ScrollBarSize() int {
	return 12
}

func (FioCustomTheme) ScrollBarSmallSize() int {
	return 4
}

func CustomTheme() fyne.Theme {
	return &FioCustomTheme{}
}

func FioLogoCanvas() fyne.CanvasObject {
	i, _, err := fioassets.NewFioLogo()
	if err != nil {
		return nil
	}
	image := canvas.NewImageFromImage(i)
	return fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(55, 55)), layout.NewSpacer(), image)
}

type ClickEntry struct {
	widget.Entry
	Button *widget.Button
}

func (e *ClickEntry) onEnter() {
	e.Button.Tapped(&fyne.PointEvent{})
}

func NewClickEntry(b *widget.Button) *ClickEntry {
	entry := &ClickEntry{
		Entry:  widget.Entry{},
		Button: b,
	}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *ClickEntry) KeyDown(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyReturn:
		e.onEnter()
	default:
		e.Entry.KeyDown(key)