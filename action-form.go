package cryptonym

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/blockpane/cryptonym/fuzzer"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	FormState    = NewAbi(0)
	bombsAway    = &widget.Button{}
	txWindowOpts = &txResultOpts{
		gone:   true,
	}
)

func ResetTxResult() {
	if txWindowOpts.window != nil {
		txWindowOpts.window.Hide()
		txWindowOpts.window.Close()
	}
	txWindowOpts.window = App.NewWindow("Tx Results")
	txWindowOpts.gone = true
	txWindowOpts.window.SetContent(layout.NewSpacer())
	txWindowOpts.window.Show()
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			if txWindowOpts.window.Content().Visible() {
				txWindowOpts.window.Hide()
				return
			}
		}
	}()
}

// GetAbiForm returns the fyne form for editing the request, it also handles state tracking via
// the FormState which is later used to build the transaction.
func GetAbiForm(action string, account *fio.Account, api *fio.API, opts *fio.TxOptions) (fyne.CanvasObject, error) {
	if api.HttpClient == nil {
		return widget.NewVBox(), nil
	}
	accountAction := strings.Split(action, "::")
	if len(accountAction) != 2 {
		e := "couldn't parse account and action for " + action
		errs.ErrChan <- e
		return nil, errors.New(e)
	}
	abi, err := api.GetABI(eos.AccountName(accountAction[0]))
	if err != nil {
		errs.ErrChan <- err.Error()
		return nil, err
	}
	abiStruct := abi.ABI.StructForName(accountAction[1])
	form := widget.NewForm()

	abiState := NewAbi(len(abiStruct.Fields))
	abiState.Contract = accountAction[0]
	abiState.Action = accountAction[1]
	for i, deRef := range abiStruct.Fields {
		fieldRef := &deRef
		f