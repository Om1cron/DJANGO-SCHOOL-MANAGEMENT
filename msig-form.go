package cryptonym

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type signer struct {
	weight *widget.Entry
	actor  *widget.Entry
	index  int
}

var (
	MsigRefreshRequests = make(chan bool)
	MsigLastTab         = 0
	MsigLoaded          bool
)

func UpdateAuthContent(container chan fyne.Container, api *fio.API, opts *fio.TxOptions, account *fio.Account) {
	for !Connected {
		time.Sleep(time.Second)
	}
	authTab := func() {} //recursive
	authTab = func() {
		accountEntry := widget.NewEntry()
		newAccount := &fio.Account{}
		update := &widget.TabItem{}
		fee := widget.NewLabelWithStyle(p.Sprintf("Required Fee: %s %G", fio.FioSymbol, fio.GetMaxFee(fio.FeeAuthUpdate)*2.0), fyne.TextAlignTrailing, fyne.TextStyle{})
		warning := widget.NewHBox(
			widget.NewIcon(theme.WarningIcon()),
			widget.NewLabelWithStyle("Warning: converting active account to multi-sig!", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		)
		warning.Hide()
		newRandCheck := widget.NewCheck("Create New and Burn", func(b bool) {
			if b {
				newAccount, _ = fio.NewRandomAccount()
				accountEntry.SetText(string(newAccount.Actor))
				fee.SetText(p.Sprintf("Required Fee: %s%G", fio.FioSymbol, fio.GetMaxFe