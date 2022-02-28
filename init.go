
package cryptonym

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/fioprotocol/fio-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"runtime"
)

const (
	AppTitle = "Cryptonym"
)

var (
	WinSettings             = getSavedWindowSettings()
	W                       = WinSettings.W
	H                       = WinSettings.H
	txW                     = 1200
	txH                     = 500
	ActionW                 = 220 // width of action buttons on left side
	WidthReduce             = 26  // trim down size of right window this much to account for padding
	App                     = app.NewWithID("explorer")
	Win                     = App.NewWindow(AppTitle)
	BalanceChan             = make(chan bool)
	BalanceLabel            = widget.NewLabel("")
	DefaultFioAddress       = ""
	TableIndex              = NewTableIndex()
	delayTxSec              = 1000000000
	EndPoints               = SupportedApis{Apis: []string{"/v1/chain/push_transaction"}}
	actionEndPointActive    = "/v1/chain/push_transaction"
	apiEndPointActive       = "/v1/chain/get_info"
	p                       = message.NewPrinter(language.English)
	RepaintChan             = make(chan bool)
	PasswordVisible         bool
	SettingsLoaded          = make(chan *FioSettings)
	Settings                = DefaultSettings()
	TxResultBalanceChan     = make(chan string)
	TxResultBalanceChanOpen = false
	useZlib                 = false
	deferTx                 = false
	Connected               bool
	Uri                     = ""
	Api                     = &fio.API{}
	Opts                    = &fio.TxOptions{}
	Account                 = func() *fio.Account {
		a, _ := fio.NewAccountFromWif("5JBbUG5SDpLWxvBKihMeXLENinUzdNKNeozLas23Mj6ZNhz3hLS") // vote1@dapixdev
		return a
	}()
)

func init() {
	txW = (W * 65) / 100
	txH = (H * 85) / 100
	go func() {
		TxResultBalanceChanOpen = true
		defer func() {
			TxResultBalanceChanOpen = false
		}()
		for {
			select {
			case bal := <-TxResultBalanceChan:
				BalanceLabel.SetText(bal)
				BalanceLabel.Refresh()
			}