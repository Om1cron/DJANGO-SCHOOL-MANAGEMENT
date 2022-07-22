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
				fee.SetText(p.Sprintf("Required Fee: %s%G", fio.FioSymbol, fio.GetMaxFee(fio.FeeAuthUpdate)*2.0+fio.GetMaxFee(fio.FeeTransferTokensPubKey)))
				fee.Refresh()
				warning.Hide()
			} else {
				accountEntry.SetText(string(account.Actor))
				fee.SetText(p.Sprintf("Required Fee: %s%G", fio.FioSymbol, fio.GetMaxFee(fio.FeeAuthUpdate)*2.0))
				fee.Refresh()
				warning.Show()
			}
		})
		accountEntry.SetText(string(account.Actor))
		newRandCheck.SetChecked(true)

		threshEntry := widget.NewEntry()
		threshEntry.SetText("2")
		tMux := sync.Mutex{}
		threshEntry.OnChanged = func(s string) {
			tMux.Lock()
			time.Sleep(300 * time.Millisecond)
			if _, e := strconv.Atoi(s); e != nil {
				tMux.Unlock()
				threshEntry.SetText("2")
				return
			}
			tMux.Unlock()
		}

		signerSlice := make([]signer, 0) // keeps order correct when adding rows, and is sorted when submitting tx
		newSigner := func(s string) *fyne.Container {
			if s == "" {
				for i := 0; i < 12; i++ {
					b := []byte{uint8(len(signerSlice) + 96)} // assuming we will start with 1 by default
					s = s + string(b)
				}
			}
			w := widget.NewEntry()
			w.SetText("1")
			a := widget.NewEntry()
			a.SetText(s)
			index := len(signerSlice)
			shouldAppend := func() bool {
				for _, sc := range signerSlice {
					if sc.actor == a {
						return false
					}
				}
				return true
			}
			if shouldAppend() {
				signerSlice = append(signerSlice, signer{
					weight: w,
					actor:  a,
					index:  index,
				})
			} else {
				return nil
			}
			threshEntry.SetText(fmt.Sprintf("%d", 1+len(signerSlice)/2))
			threshEntry.Refresh()
			return fyne.NewContainerWithLayout(layout.NewGridLayoutWithColumns(6),
				layout.NewS