
package cryptonym

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go/eos"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

type ServerInfo struct {
	Info *eos.InfoResp
	Uri  string
}

type prodInfo struct {
	address string
	url     *url.URL
}

func InitServerInfo(info chan ServerInfo, reconnected chan bool) fyne.CanvasObject {
	knownChains := map[string]string{
		"b20901380af44ef59c5918439a1f9a41d83669020319a80574b804a5f95cbd7e": "FIO Testnet",
		"21dcae42c0182200e93f954a074011f9048a7624c6fe81d3c9541a614a88bd1c": "FIO Mainnet",
		"e143d39294a14616dbbee394f1c159a4eb71b656b9ca1094ebf924dc3714d7ae": "Dapix Development Chain",
	}

	prods := make(map[string]*prodInfo)

	uriLabel := widget.NewLabel(Uri)
	rows := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Server", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), layout.NewSpacer()),
			uriLabel,
		),
		layout.NewSpacer(),
	}

	versionLabel := widget.NewLabel("")
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Server Version", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), layout.NewSpacer()),
			versionLabel,
		),
		layout.NewSpacer(),
	}...)

	chainIdKnownLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	chainIdIcon := canvas.NewImageFromResource(theme.WarningIcon())
	chainIdIcon.Hide()
	chainIdLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	chainIdBox := widget.NewHBox(
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), chainIdIcon),
		chainIdKnownLabel,
		chainIdLabel,
	)
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Chain ID", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		chainIdBox,
		layout.NewSpacer(),
	}...)

	headTimeLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	headTimeLagLabel := widget.NewLabel("")
	headTimeLagIcon := canvas.NewImageFromResource(theme.WarningIcon())
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Head Block Time", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), headTimeLagIcon),
			headTimeLagLabel,
			headTimeLabel,
		),
		layout.NewSpacer(),
	}...)

	headBlockLabel := widget.NewLabel("")
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Head Block", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), layout.NewSpacer()),
			headBlockLabel,
		),
		layout.NewSpacer(),
	}...)

	libLabel := widget.NewLabel("")
	libWarnIcon := canvas.NewImageFromResource(theme.WarningIcon())
	libWarnIcon.Hide()
	libWarnLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Last Irreversible Block", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), libWarnIcon),
			libLabel,
			libWarnLabel,
		),
		layout.NewSpacer(),
	}...)

	prodLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})
	prodAddrLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	prodUrl := widget.NewHyperlinkWithStyle("", &url.URL{}, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	rows = append(rows, []fyne.CanvasObject{
		widget.NewLabelWithStyle("Current Producer", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewHBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), layout.NewSpacer()),
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize((W*40)/100, 35)),
				fyne.NewContainerWithLayout(layout.NewGridLayout(3),
					prodAddrLabel,
					prodLabel,
					prodUrl,
				)),
		),
		layout.NewSpacer(),
	}...)

	histApiLabel := widget.NewLabelWithStyle("History API", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	histApiLabel.Hide()
	histApiSp := layout.NewSpacer()
	histApiSp.Hide()
	histApiBox := widget.NewHBox(
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(25, 25)), canvas.NewImageFromResource(theme.ConfirmIcon())),
		widget.NewLabelWithStyle("History API is available", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
	)
	histApiBox.Hide()
	rows = append(rows, []fyne.CanvasObject{
		histApiLabel,
		histApiBox,
		histApiSp,
	}...)

	sizeLabel := widget.NewLabelWithStyle("DB Size", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	sizeLabel.Hide()
	sizeBytesLabel := widget.NewLabel("")
	sizeBytesLabel.Hide()
	sizeWarnIcon := canvas.NewImageFromResource(theme.WarningIcon())
	sizeWarnIcon.Hide()
	sizeWarnLabel := widget.NewLabel("Over 75% of RAM is used!")
	sizeWarnLabel.Hide()
	sizeSp := layout.NewSpacer()