
package cryptonym

import (
	"bytes"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"github.com/fioprotocol/fio-go/eos/ecc"
	"github.com/skip2/go-qrcode"
	"image"
	"image/color"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var imageSize = func() (size int) {
	size = 196
	scale := os.Getenv("FYNE_SCALE")
	if scale != "" {
		new, err := strconv.Atoi(scale)
		if err != nil {
			return
		}
		size = ((new * 10) * size) / 10
	}
	return
}()

var RefreshQr = make(chan bool)

func KeyGenTab() *widget.Box {

	keyStr := ""
	waitMsg := "      ... waiting for entropy from operating system ...              "
	entry := widget.NewEntry()
	var key *ecc.PrivateKey
	var showingPriv bool
	var err error

	vanityQuit := make(chan bool)
	vanityOpt := &vanityOptions{}
	vanityOpt.threads = runtime.NumCPU()
	vanitySearch := widget.NewSelect([]string{"Actor", "Pubkey", "Either"}, func(s string) {
		switch s {
		case "Actor":
			vanityOpt.actor = true
			vanityOpt.pub = false
		case "Pubkey":
			vanityOpt.actor = false
			vanityOpt.pub = true
		default:
			vanityOpt.actor = true
			vanityOpt.pub = true
		}
	})
	vanitySearch.SetSelected("Actor")
	vanitySearch.Hide()
	vanityMatch := widget.NewCheck("match anywhere", func(b bool) {
		if b {
			vanityOpt.anywhere = true
			return
		}
		vanityOpt.anywhere = false
	})
	vanityMatch.Hide()
	vanityLabel := widget.NewLabel(" ")
	vanityEntry := NewClickEntry(&widget.Button{})
	vanityEntry.SetPlaceHolder("enter string to search for")
	vanityEntry.OnChanged = func(s string) {
		vanityLabel.SetText(" ")
		if len(s) >= 6 {
			vanityLabel.SetText("Note: searching for 6 or more characters can take a very long time.")
		}
		vanityOpt.word = strings.ToLower(s)
		vanityLabel.Refresh()
	}
	vanityEntry.Hide()
	vanityStopButton := &widget.Button{}
	vanityStopButton = widget.NewButtonWithIcon("Stop Searching", theme.CancelIcon(), func() {
		vanityQuit <- true
		entry.SetText("Vanity key generation cancelled")
		vanityStopButton.Hide()
	})
	vanityStopButton.Hide()
	vanityCheck := widget.NewCheck("Generate Vanity Address", func(b bool) {
		if b {
			waitMsg = "Please wait, generating vanity key"
			vanitySearch.Show()
			vanityMatch.Show()
			vanityEntry.Show()
			return
		}
		waitMsg = "      ... waiting for entropy from operating system ...              "
		vanitySearch.Hide()
		vanityMatch.Hide()
		vanityEntry.Hide()
		vanityStopButton.Hide()
	})
	vanityBox := widget.NewVBox(
		widget.NewHBox(
			layout.NewSpacer(),
			vanityCheck,
			vanitySearch,
			vanityMatch,
			vanityEntry,
			layout.NewSpacer(),
		),
		widget.NewHBox(
			layout.NewSpacer(),
			vanityLabel,
			layout.NewSpacer(),
		),
	)

	emptyQr := disabledImage(imageSize, imageSize)
	qrImage := canvas.NewImageFromImage(emptyQr)
	qrImage.FillMode = canvas.ImageFillOriginal
	newQrPub := image.Image(emptyQr)
	newQrPriv := image.Image(emptyQr)
	copyToClip := widget.NewButton("", nil)
	swapQrButton := widget.NewButton("", nil)

	setWait := func(s string) {
		keyStr = s
		swapQrButton.Disable()
		copyToClip.Disable()
		qrImage.Image = emptyQr
		entry.SetText(keyStr)
		copyToClip.Refresh()
		swapQrButton.Refresh()
		qrImage.Refresh()
		entry.Refresh()
	}
	setWait(waitMsg)

	qrPriv := make([]byte, 0)
	qrPub := make([]byte, 0)
	qrLabel := widget.NewLabel("Public Key:")
	qrLabel.Alignment = fyne.TextAlignCenter

	swapQr := func() {
		switch showingPriv {
		case false:
			swapQrButton.Text = "Show Pub Key QR Code"
			qrLabel.SetText("Private Key:")
			qrImage.Image = newQrPriv
			showingPriv = true
		case true:
			swapQrButton.Text = "Show Priv Key QR Code"
			qrLabel.SetText("Public Key:")
			qrImage.Image = newQrPub
			showingPriv = false
		}
		qrLabel.Refresh()
		qrImage.Refresh()
		swapQrButton.Refresh()
	}
	swapQrButton = widget.NewButtonWithIcon("Show Private Key QR Code", theme.VisibilityIcon(), swapQr)

	regenButton := &widget.Button{}
	newKey := true
	var setBusy bool
	setKey := func() {
		time.Sleep(20 * time.Millisecond) // lame, but prevents a double event on darwin?!?