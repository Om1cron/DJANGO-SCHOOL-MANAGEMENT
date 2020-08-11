
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	explorer "github.com/blockpane/cryptonym"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type tabs struct {
	Editor      widget.TabItem
	Api         widget.TabItem
	Info        widget.TabItem
	Browser     widget.TabItem
	Abi         widget.TabItem
	AccountInfo widget.TabItem
	KeyGen      widget.TabItem
	Msig        widget.TabItem
	Vote        widget.TabItem
	Requests    widget.TabItem
}

var (
	uri   = &explorer.Uri
	proxy = func() *string {
		s := "127.0.0.1:8080"
		return &s
	}()
	api           = explorer.Api
	opts          = explorer.Opts
	account       = explorer.Account
	balance       float64
	actionsGroup  = &widget.Group{}
	connectButton = &widget.Button{}
	proxyCheck    = &widget.Check{}
	keyContent    = &widget.Box{}
	tabContent    = &widget.TabContainer{}
	tabEntries    = tabs{}
	hostEntry     = explorer.NewClickEntry(connectButton)
	myFioAddress  = widget.NewEntry()
	moneyBags     = widget.NewSelect(moneySlice(), func(s string) {})
	wifEntry      = widget.NewPasswordEntry()
	balanceLabel  = widget.NewLabel("Balance: unknown")
	loadButton    = &widget.Button{}
	importButton  = &widget.Button{}
	balanceButton = &widget.Button{}
	regenButton   = &widget.Button{}
	uriContent    = uriInput(true)
	uriContainer  = &fyne.Container{}
	ready         = false
	connectedChan = make(chan bool)
	p             = message.NewPrinter(language.English)
	keyBox        = &widget.Box{}
	serverInfoCh  = make(chan explorer.ServerInfo)
	serverInfoRef = make(chan bool)
	serverInfoBox = explorer.InitServerInfo(serverInfoCh, serverInfoRef)
)

// ActionButtons is a slice of pointers to our action buttons, this way we can set them to hidden if using
// the filter ....
var (
	ActionButtons = make([]*widget.Button, 0)
	ActionLabels  = make([]*widget.Label, 0)
	filterActions = &widget.Entry{}
	filterCheck   = &widget.Check{}
	prodsCheck    = &widget.Check{}
)

var savedKeys = map[string]string{
	"devnet vote1":   "5JBbUG5SDpLWxvBKihMeXLENinUzdNKNeozLas23Mj6ZNhz3hLS",
	"devnet vote2":   "5KC6Edd4BcKTLnRuGj2c8TRT9oLuuXLd3ZuCGxM9iNngc3D8S93",
	"devnet bp1":     "5KQ6f9ZgUtagD3LZ4wcMKhhvK9qy4BuwL3L1pkm6E2v62HCne2R",
	"devnet locked1": "5HwvMtAEd7kwDPtKhZrwA41eRMdFH5AaBKPRim6KxkTXcg5M9L5",
}

func main() {
	// the MacOS resolver causes serious performance issues, if GODEBUG is empty, then set it to force pure go resolver.
	if runtime.GOOS == "darwin" {
		gdb := os.Getenv("GODEBUG")
		if gdb == "" {
			_ = os.Setenv("GODEBUG", "netdns=go")
		}
	}
	topLayout := &fyne.Container{}
	errs.ErrTxt[0] = fmt.Sprintf("\nEvent Log: started at %s", time.Now().Format(time.Stamp))
	errs.ErrMsgs.SetText(strings.Join(errs.ErrTxt, "\n"))
	keyContent = keyBoxContent()
	myFioAddress.Hide()

	loadButton.Disable()
	balanceButton.Disable()
	regenButton.Disable()

	space := strings.Repeat("  ", 55)
	go func() {
		for {
			select {
			case <-connectedChan:
				time.Sleep(time.Second)
				serverInfoRef <- true
				explorer.Connected = true
				uriContainer.Objects = []fyne.CanvasObject{
					widget.NewVBox(
						widget.NewLabel(" "),
						widget.NewHBox(
							widget.NewLabel(space),
							widget.NewLabel(" nodeos @ "+*uri+" "),
							widget.NewLabel(space),
						),
					),
				}
				loadButton.Enable()
				balanceButton.Enable()
				regenButton.Enable()
				refreshMyName()
			case <-errs.RefreshChan:
				if !ready {
					continue
				}
				refreshNotNil(loadButton)
				refreshNotNil(balanceButton)
				refreshNotNil(regenButton)
				refreshNotNil(actionsGroup)
				refreshNotNil(hostEntry)
				refreshNotNil(uriContent)