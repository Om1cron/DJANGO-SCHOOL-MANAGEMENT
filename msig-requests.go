
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
	)

	approversSorted := func() []string {
		s := make([]string, 0)
		for k := range approvers {
			s = append(s, k)
		}
		sort.Strings(s)
		return s
	}()
	var checked int
	for _, k := range approversSorted {
		// actor, fio address, has approved, is produce
		//hasApproved := theme.CancelIcon()
		hasApproved := theme.CheckButtonIcon()
		asterisk := ""
		if approvers[k] {
			//hasApproved = theme.ConfirmIcon()
			hasApproved = theme.CheckButtonCheckedIcon()
			checked += 1
			for _, p := range producers.Active.Producers {
				if p.AccountName == eos.AccountName(k) {
					top21Count += 1
					asterisk = "*"
					break
				}
			}
		}
		top21Label := widget.NewLabel(asterisk)
		var firstName string
		n, ok, _ := api.GetFioNamesForActor(k)
		if ok && len(n.FioAddresses) > 0 {
			firstName = n.FioAddresses[0].FioAddress
		}
		deref := &k
		actor := *deref
		copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			Win.Clipboard().SetContent(actor)
		})
		approversRows.Append(fyne.NewContainerWithLayout(layout.NewGridLayout(3),
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),
				widget.NewHBox(
					layout.NewSpacer(),
					top21Label,
					fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(32, 32)),
						canvas.NewImageFromResource(hasApproved),
					)),
				widget.NewLabel(firstName),
			),
			widget.NewHBox(
				layout.NewSpacer(),
				copyButton,
				widget.NewLabelWithStyle(k, fyne.TextAlignTrailing, fyne.TextStyle{Monospace: true}),
			),
			widget.NewLabelWithStyle(isProd(eos.AccountName(k)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		))

	}

	type actionActor struct {
		Actor string `json:"actor"`
	}
	// will use for counting vote weights ...
	actorMap := make(map[string]bool)
	actions := make([]msigActionInfo, 0)
	actionString := ""
	tx, err := api.GetProposalTransaction(eos.AccountName(proposer), requests[index].ProposalName)
	if err != nil {
		return widget.NewHBox(widget.NewLabelWithStyle(err.Error(), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	} else {
		proposalHash = tx.ProposalHash
		for _, action := range tx.PackedTransaction.Actions {
			abi, err := api.GetABI(action.Account)
			if err != nil {
				errs.ErrChan <- err.Error()
				continue
			}
			decoded, err := abi.ABI.DecodeAction(action.HexData, action.Name)
			if err != nil {
				errs.ErrChan <- err.Error()
				continue
			}
			actions = append(actions, msigActionInfo{
				Action:       decoded,
				Account:      string(action.Account),
				Name:         string(action.Name),
				ProposalHash: tx.ProposalHash,
			})
			aActor := &actionActor{}
			err = json.Unmarshal(decoded, aActor)
			if err == nil && aActor.Actor != "" {
				actorMap[aActor.Actor] = true
			}
		}
		a, err := json.MarshalIndent(actions, "", "  ")
		if err != nil {
			errs.ErrChan <- err.Error()
		} else {
			actionString = string(a)
		}
	}
	hasApprovals := true
	approvalsNeeded := make([]string, 0)
	have := "have"
	var privAction bool
	for a := range PrivilegedActions {
		sa := strings.Split(a, "::")
		if len(sa) != 2 {
			continue
		}
		if string(tx.PackedTransaction.Actions[0].Name) == sa[1] {
			privAction = true
			break
		}
	}
	if privAction {
		if checked == 1 {
			have = "has"
		}
		var required int
		type tp struct {
			Producer string `json:"producer"`
		}
		rows := make([]tp, 0)
		gtr, err := api.GetTableRows(eos.GetTableRowsRequest{
			Code:  "eosio",
			Scope: "eosio",
			Table: "topprods",
			Limit: 21,
			JSON:  true,
		})
		if err != nil {
			errs.ErrChan <- err.Error()
			required = 15
		} else {
			_ = json.Unmarshal(gtr.Rows, &rows)
			if len(rows) == 0 || len(rows) == 21 {
				required = 15
			} else {
				required = (len(rows) / 2) + ((len(rows) / 2) / 2)
			}
		}
		var top21Voted string
		if top21Count > 0 {
			top21Voted = fmt.Sprintf(" - (%d are Top 21 Producers)", top21Count)
		}
		approvalsNeeded = append(approvalsNeeded, fmt.Sprintf("Account %s requires %d approvals, %d %s been provided%s", tx.PackedTransaction.Actions[0].Account, required, checked, have, top21Voted))
		if checked < required {
			hasApprovals = false
		}
	} else {
		for msigAccount := range actorMap {
			needs, has, err := getVoteWeight(msigAccount, requests[index].ProvidedApprovals, api)