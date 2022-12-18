
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
	refr := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		refreshChan <- true
	})
	topDesc := widget.NewLabel("")
	top := widget.NewHBox(
		layout.NewSpacer(),
		topDesc,
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(refr.MinSize()), refr),
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(sendNew.MinSize()), sendNew),
		layout.NewSpacer(),
	)

	pending, has, err := api.GetPendingFioRequests(account.PubKey, 101, 0)
	if err != nil {
		return widget.NewHBox(widget.NewLabel(err.Error())), err
	}
	if !has {
		return widget.NewVBox(top, widget.NewLabel("No pending requests.")), err
	}
	howMany := len(pending.Requests)
	topDesc.SetText(fmt.Sprint(howMany) + " pending requests.")
	if howMany > 100 {
		topDesc.SetText("More than 100 pending requests.")
	}
	sort.Slice(pending.Requests, func(i, j int) bool {
		return pending.Requests[i].FioRequestId < pending.Requests[j].FioRequestId
	})
	if howMany > 25 {
		topDesc.SetText(topDesc.Text + fmt.Sprintf(" (only first 25 displayed.)"))
		pending.Requests = pending.Requests[:25]
	}

	requests := fyne.NewContainerWithLayout(layout.NewGridLayout(5),
		widget.NewLabelWithStyle("Actions", fyne.TextAlignLeading, fyne.TextStyle{}),
		widget.NewLabelWithStyle("ID / Time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("From", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("To", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Summary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	for _, req := range pending.Requests {
		func(req fio.RequestStatus) {
			id := widget.NewLabelWithStyle(fmt.Sprintf("%d | "+req.TimeStamp.Local().Format(time.Stamp), req.FioRequestId), fyne.TextAlignCenter, fyne.TextStyle{})
			payer := req.PayerFioAddress
			if len(payer) > 32 {
				payer = payer[:29] + "..."
			}
			payee := req.PayeeFioAddress
			if len(payee) > 32 {
				payee = payee[:29] + "..."
			}
			fr := widget.NewLabelWithStyle(payee, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
			to := widget.NewLabelWithStyle(payer, fyne.TextAlignLeading, fyne.TextStyle{})
			view := widget.NewButtonWithIcon("View", theme.VisibilityIcon(), func() {
				closed := make(chan interface{})
				d := dialog.NewCustom(
					fmt.Sprintf("FIO Request ID %d (%s)", req.FioRequestId, req.PayeeFioAddress),
					"Close",
					ViewRequest(req.FioRequestId, closed, refreshChan, account, api),
					Win,
				)
				go func() {
					<-closed
					d.Hide()
					refreshChan <- true
				}()
				d.Show()
			})
			rejectBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				_, err := api.SignPushActions(fio.NewRejectFndReq(account.Actor, strconv.FormatUint(req.FioRequestId, 10)))
				if err != nil {
					errs.ErrChan <- err.Error()
					return
				}
				errs.ErrChan <- "rejected request id: " + strconv.FormatUint(req.FioRequestId, 10)
				refreshChan <- true
			})
			rejectBtn.HideShadow = true
			requests.AddObject(widget.NewHBox(view, layout.NewSpacer()))
			requests.AddObject(id)
			requests.AddObject(fr)
			requests.AddObject(to)
			obt, err := fio.DecryptContent(account, req.PayeeFioPublicKey, req.Content, fio.ObtRequestType)
			var summary string
			if err != nil {
				view.Hide()
				summary = "invalid content"
				errs.ErrChan <- err.Error()
			} else {
				summary = obt.Request.ChainCode
				if obt.Request.ChainCode != obt.Request.TokenCode {
					summary += "/" + obt.Request.TokenCode
				}
				summary += fmt.Sprintf(" (%s) %q", obt.Request.Amount, obt.Request.Memo)
				if len(summary) > 32 {
					summary = summary[:29] + "..."
				}
			}
			requests.AddObject(widget.NewHBox(layout.NewSpacer(), widget.NewLabelWithStyle(summary, fyne.TextAlignTrailing, fyne.TextStyle{Italic: true}), rejectBtn))
		}(req)
	}
	form = widget.NewVBox(
		top,
		fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(RWidth(), int(float32(PctHeight())*.68))),
			widget.NewScrollContainer(widget.NewVBox(requests,
				fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.Size{
					Width:  20,
					Height: 50,
				}), layout.NewSpacer()),
			)),
		),
	)
	return
}

func ViewRequest(id uint64, closed chan interface{}, refresh chan bool, account *fio.Account, api *fio.API) fyne.CanvasObject {
	req, err := api.GetFioRequest(id)
	if err != nil {
		return widget.NewLabel(err.Error())
	}
	decrypted, err := fio.DecryptContent(account, req.PayeeKey, req.Content, fio.ObtRequestType)
	if err != nil {
		return widget.NewLabel(err.Error())
	}
	reqData := make([]*widget.FormItem, 0)
	add := func(name string, value string) {
		if len(value) > 0 {
			a := widget.NewEntry()
			a.SetText(value)
			a.OnChanged = func(string) {
				a.SetText(value)