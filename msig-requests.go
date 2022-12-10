
package cryptonym

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"sort"
	"strings"
)

func MsigRequestsContent(api *fio.API, opts *fio.TxOptions, account *fio.Account) *widget.TabItem {
	pendingTab := widget.NewTabItem("Pending Requests",
		widget.NewScrollContainer(
			ProposalRows(0, 100, api, opts, account),
		),
	)
	return pendingTab
}

func requestBox(proposer string, requests []*fio.MsigApprovalsInfo, index int, proposalWindow fyne.Window, api *fio.API, opts *fio.TxOptions, account *fio.Account) fyne.CanvasObject {
	p := message.NewPrinter(language.AmericanEnglish)
	aFee := fio.GetMaxFee(fio.FeeMsigApprove)
	dFee := fio.GetMaxFee(fio.FeeMsigUnapprove)
	cFee := fio.GetMaxFee(fio.FeeMsigCancel)
	eFee := fio.GetMaxFee(fio.FeeMsigExec)
	proposalHash := eos.Checksum256{}

	refresh := func() {
		_, ai, err := api.GetApprovals(fio.Name(proposer), 10)
		if err != nil {
			errs.ErrChan <- err.Error()
			return
		}
		requests = ai
	}

	approve := widget.NewButtonWithIcon(p.Sprintf("Approve %s %g", fio.FioSymbol, aFee), theme.ConfirmIcon(), func() {
		_, tx, err := api.SignTransaction(
			fio.NewTransaction([]*fio.Action{
				fio.NewMsigApprove(eos.AccountName(proposer), requests[index].ProposalName, account.Actor, proposalHash),
			}, opts),
			opts.ChainID, fio.CompressionNone,
		)
		if err != nil {
			errs.ErrChan <- err.Error()
			resultPopup(err.Error(), proposalWindow)
			return
		}
		res, err := api.PushTransactionRaw(tx)
		if err != nil {
			errs.ErrChan <- err.Error()
			resultPopup(err.Error(), proposalWindow)
			return
		}
		errs.ErrChan <- fmt.Sprintf("sending approval for proposal '%s' proposed by %s", requests[index].ProposalName, proposer)
		j, _ := json.MarshalIndent(res, "", "    ")
		refresh()
		proposalWindow.SetContent(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize((W*85)/100, (H*85)/100)),
				requestBox(proposer, requests, index, proposalWindow, api, opts, account),
			))
		proposalWindow.Content().Refresh()
		resultPopup(string(j), proposalWindow)
	})
	approve.Hide()