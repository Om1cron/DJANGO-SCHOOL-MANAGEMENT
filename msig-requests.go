
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
	deny := widget.NewButtonWithIcon(p.Sprintf("Un-Approve %s %g", fio.FioSymbol, dFee), theme.ContentUndoIcon(), func() {
		_, tx, err := api.SignTransaction(
			fio.NewTransaction([]*fio.Action{
				fio.NewMsigUnapprove(eos.AccountName(proposer), requests[index].ProposalName, account.Actor),
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
		j, _ := json.MarshalIndent(res, "", "    ")
		errs.ErrChan <- fmt.Sprintf("withdrawing approval for proposal '%s' proposed by %s", requests[index].ProposalName, proposer)
		refresh()
		proposalWindow.SetContent(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize((W*85)/100, (H*85)/100)),
				requestBox(proposer, requests, index, proposalWindow, api, opts, account),
			))
		proposalWindow.Content().Refresh()
		resultPopup(string(j), proposalWindow)
	})
	deny.Hide()
	cancel := widget.NewButtonWithIcon(p.Sprintf("Cancel %s %g", fio.FioSymbol, cFee), theme.DeleteIcon(), func() {
		_, tx, err := api.SignTransaction(
			fio.NewTransaction([]*fio.Action{
				fio.NewMsigCancel(eos.AccountName(proposer), requests[index].ProposalName, account.Actor),
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
		j, _ := json.MarshalIndent(res, "", "    ")
		errs.ErrChan <- fmt.Sprintf("cancel proposal '%s'", requests[index].ProposalName)
		resultPopup(string(j), proposalWindow)
	})
	cancel.Hide()
	execute := widget.NewButtonWithIcon(p.Sprintf("Execute %s %g", fio.FioSymbol, eFee), fioassets.NewFioLogoResource(), func() {
		_, tx, err := api.SignTransaction(
			fio.NewTransaction([]*fio.Action{
				fio.NewMsigExec(eos.AccountName(proposer), requests[index].ProposalName, fio.Tokens(eFee), account.Actor),
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
		j, _ := json.MarshalIndent(res, "", "    ")
		errs.ErrChan <- fmt.Sprintf("executing proposal '%s' proposed by %s", requests[index].ProposalName, proposer)
		refresh()
		proposalWindow.SetContent(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize((W*85)/100, (H*85)/100)),
				requestBox(proposer, requests, index, proposalWindow, api, opts, account),
			))
		proposalWindow.Content().Refresh()
		resultPopup(string(j), proposalWindow)
	})
	execute.Hide()
	if proposer == string(account.Actor) {
		cancel.Show()
	}
	if len(requests) <= index {
		return widget.NewHBox(
			layout.NewSpacer(),
			widget.NewLabelWithStyle("Requests table has changed, please refresh and try again.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
		)
	}
	if requests[index].HasApproved(account.Actor) {
		deny.Show()
	}
	if requests[index].HasRequested(account.Actor) && !requests[index].HasApproved(account.Actor) {
		approve.Show()
	}
	approvers := make(map[string]bool)
	approvalWeightLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{})
	proposalTitle := widget.NewHBox(
		layout.NewSpacer(),
		widget.NewLabel("Proposal Name: "),
		widget.NewLabelWithStyle(string(requests[index].ProposalName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			Win.Clipboard().SetContent(string(requests[index].ProposalName))
		}),
		layout.NewSpacer(),
	)
	proposalAuthor := widget.NewHBox(
		layout.NewSpacer(),
		widget.NewLabel("Proposal Author: "),
		widget.NewLabelWithStyle(proposer, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			Win.Clipboard().SetContent(proposer)
		}),
		layout.NewSpacer(),
	)
	approversRows := widget.NewVBox(proposalTitle, proposalAuthor, approvalWeightLabel)
	producers, pErr := api.GetProducerSchedule()
	var top21Count int
	isProd := func(name eos.AccountName) string {
		result := make([]string, 0)
		if pErr != nil || producers == nil {
			return ""
		}
		for _, p := range producers.Active.Producers {
			if p.AccountName == name {
				result = append(result, "Top 21 Producer")
				break
			}
		}
		if account.Actor == name {
			result = append(result, "Current Actor")
		}
		if string(name) == proposer {
			result = append(result, "Proposal Author")
		}
		return strings.Join(result, " ~ ")
	}

	for _, approver := range requests[index].RequestedApprovals {
		approvers[string(approver.Level.Actor)] = false
	}
	for _, approver := range requests[index].ProvidedApprovals {
		approvers[string(approver.Level.Actor)] = true
	}
	approversRows.Append(widget.NewHBox(
		layout.NewSpacer(), approve, deny, cancel, execute, layout.NewSpacer(),
	))
	approversRows.Append(
		fyne.NewContainerWithLayout(layout.NewGridLayout(3),
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),
				widget.NewLabelWithStyle("Approved", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle("Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			),
			widget.NewLabelWithStyle("Account", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
		),