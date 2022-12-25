
package cryptonym

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"net/url"
	"os"
	"strconv"
)

const settingsTitle = "Cryptonym Settings"

var MainnetApi = []string{
	// These allow TLS 1.0, excluding
	//"https://fio.eoscannon.io",
	//"https://fio-mainnet.eosblocksmith.io",
	//"https://fio.eos.barcelona",
	//"https://fio.eosargentina.io",
	//"https://api.fio.services",

	// Does not allow access to get_supported_apis endpoint:
	//"https://fioapi.nodeone.io",
	//"https://fio.maltablock.org",

	"https://fio.eosdac.io",           //ok
	"https://fio.eosphere.io",         //ok
	"https://fio.eosrio.io",           //ok
	"https://fio.eosusa.news",         //ok
	"https://api.fio.alohaeos.com",    //ok
	"https://fio.genereos.io",         //ok
	"https://fio.greymass.com",        //ok
	"https://api.fio.eosdetroit.io",   // ok
	"https://fio.zenblocks.io",        // ok
	"https://api.fio.currencyhub.io",  // ok
	"https://fio.cryptolions.io",      // ok
	"https://fio.eosdublin.io",        // ok
	"https://api.fio.greeneosio.com",  // ok
	"https://api.fiosweden.org",       // ok
	"https://fio.eu.eosamsterdam.net", //ok
	"https://fioapi.ledgerwise.io",    // sort of ok, lots of errors
	"https://fio.acherontrading.com",  //ok
}

func SettingsWindow() {
	if PasswordVisible {
		return
	}
	w := App.NewWindow(settingsTitle)
	w.Resize(fyne.NewSize(600, 800))
	w.SetOnClosed(func() {
		for _, w := range fyne.CurrentApp().Driver().AllWindows() {
			if w.Title() == AppTitle {
				w.RequestFocus()
				return
			}
		}
	})

	if Settings == nil || Settings.Server == "" {
		Settings = DefaultSettings()