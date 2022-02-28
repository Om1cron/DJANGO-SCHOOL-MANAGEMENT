
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