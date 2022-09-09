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
				layout.NewSpacer(),
				widget.NewLabelWithStyle("Actor "+strconv.Itoa(signerSlice[index].index+1)+": ", fyne.TextAlignTrailing, fyne.TextStyle{}),
				signerSlice[index].actor,
				widget.NewLabelWithStyle("Vote Weight: ", fyne.TextAlignTrailing, fyne.TextStyle{}),
				signerSlice[index].weight,
				layout.NewSpacer(),
			)
		}
		signerGroup := widget.NewGroup(" Signers ", newSigner(string(account.Actor)))
		addSigner := widget.NewButtonWithIcon("Add Signer", theme.ContentAddIcon(), func() {
			signerGroup.Append(newSigner(""))
			signerGroup.Refresh()
			update.Content.Refresh()
		})

		resetSigners := widget.NewButtonWithIcon("Reset", theme.ContentClearIcon(), func() {
			MsigLastTab = 1
			authTab()
			go func() {
				time.Sleep(100 * time.Millisecond)
				for _, a := range fyne.CurrentApp().Driver().AllWindows() {
					a.Content().Refresh()
				}
			}()
		})

		submitButton := &widget.Button{}
		submitButton = widget.NewButtonWithIcon("Submit", fioassets.NewFioLogoResource(), func() {
			submitButton.Disable()
			ok, _, msg := checkSigners(signerSlice, "active")
			if !ok {
				dialog.ShowError(msg, Win)
				return
			}
			defer submitButton.Enable()
			if newRandCheck.Checked {
				if ok, err := fundRandMsig(newAccount, account, len(signerSlice), api, opts); !ok {
					errs.ErrChan <- "new msig account was not created!"
					dialog.ShowError(err, Win)
					return
				}
			}
			acc := &fio.Account{}
			acc = account
			if newRandCheck.Checked {
				acc = newAccount
			}
			t, err := strconv.Atoi(threshEntry.Text)
			if err != nil {
				errs.ErrChan <- "Invalid threshold, refusing to continue"
				return
			}
			ok, info, err := updateAuthResult(acc, signerSlice, t)
			if ok {
				dialog.ShowCustom("Success", "OK", info, Win)
				return
			}
			dialog.ShowError(err, Win)
		})

		update = widget.NewTabItem("Update Auth",
			widget.NewScrollContainer(
				widget.NewVBox(
					widget.NewHBox(
						widget.NewLabelWithStyle("Account: ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						newRandCheck,
						accountEntry,
						widget.NewLabelWithStyle("Threshold: ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						threshEntry,
					),
					signerGroup,
					layout.NewSpacer(),
					widget.NewHBox(layout.NewSpacer(), addSigner, resetSigners, layout.NewSpacer(), warning, fee, submitButton, layout.NewSpacer()),
					widget.NewLabel(""),
				),
			))
		tabs := widget.NewTabContainer(MsigRequestsContent(api, opts, account), update)
		tabs.SelectTabIndex(MsigLastTab)
		container <- *fyne.NewContainerWithLayout(layout.NewMaxLayout(), tabs)
	}
	go func() {
		for {
			select {
			case r := <-MsigRefreshRequests:
				if r {
					a := *Api
					api = &a
					o := *Opts
					opts = &o
					u := *Account
					account = &u
					authTab()
				}
				// do we ever 