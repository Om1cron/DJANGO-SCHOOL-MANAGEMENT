
package cryptonym

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"sort"
	"strconv"
	"sync"
	"time"
)

const spaces = "                                                     " // placeholder for pubkeys

var RefreshRequestsChan = make(chan bool)

func RequestContent(reqContent chan fyne.CanvasObject, refresh chan bool) {
	content, err := GetPending(refresh, Account, Api)
	if err != nil {
		panic(err)
	}
	reqContent <- content
	go func() {
		for {
			select {
			case <-refresh:
				content, err := GetPending(refresh, Account, Api)
				if err != nil {
					errs.ErrChan <- err.Error()
					continue
				}
				reqContent <- content
			}
		}
	}()
}

func GetPending(refreshChan chan bool, account *fio.Account, api *fio.API) (form fyne.CanvasObject, err error) {
	sendNew := widget.NewButtonWithIcon("Request Funds", theme.DocumentCreateIcon(), func() {
		closed := make(chan interface{})
		d := dialog.NewCustom(
			"Send a new funds request",
			"Cancel",
			NewRequest(account, api),
			Win,
		)
		go func() {
			<-closed
			d.Hide()
		}()
		d.SetOnClosed(func() {
			refreshChan <- true
		})
		d.Show()
	})