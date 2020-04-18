
package cryptonym

import (
	"context"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"strings"
	"time"
)

func NewAccountSearchTab(box chan fyne.Container, account *fio.Account) {

	accountOuput := &widget.Box{}
	accountInput := NewClickEntry(&widget.Button{})
	accountSelect := widget.NewSelect(accountSearchType, func(s string) {
		accountInput.Refresh()
	})
	accountSelect.SetSelected(accountSearchType[0])
	accountInput.SetText(account.PubKey)
	accountInput.OnChanged = func(s string) {
		selected := accountSelect.Selected
		switch {
		case len(s) == 53 && strings.HasPrefix(s, "FIO"):
			accountSelect.SetSelected("Public Key")
		case len(s) == 51 && strings.HasPrefix(s, "5"):
			accountSelect.SetSelected("Private Key")
		case strings.Contains(s, "@"):
			accountSelect.SetSelected("Fio Address")
		case len(s) == 12:
			accountSelect.SetSelected("Actor/Account")
		case selected != "Fio Domain":
			accountSelect.SetSelected("Fio Domain")
		}
		accountInput.SetText(s)
		go func() {
			time.Sleep(100 * time.Millisecond)
			accountInput.Refresh()
		}()
	}

	accountSubmit := &widget.Button{}
	emptyBox := func() *widget.Box {
		return widget.NewVBox(
			fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(RWidth(), 40)),