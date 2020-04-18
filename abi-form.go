
package cryptonym

import (
	"encoding/json"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"gopkg.in/yaml.v3"
	"math"
	"regexp"
)

func GetAbiViewer(w int, h int, api *fio.API) (tab *widget.Box, ok bool) {
	structs := widget.NewMultiLineEntry()
	actions := widget.NewMultiLineEntry()
	tables := widget.NewMultiLineEntry()
	asJson := &widget.Check{}
	scrollViews := &fyne.Container{}
	layoutStructs := &widget.TabItem{}
	layoutActions := &widget.TabItem{}
	layoutTables := &widget.TabItem{}
	r := regexp.MustCompile("(?m)^-")

	getAbi := func(s string) {
		if s == "" {
			errs.ErrChan <- "queried for empty abi"
			return
		}
		errs.ErrChan <- "getting abi for " + s
		abiOut, err := api.GetABI(eos.AccountName(s))
		if err != nil {
			errs.ErrChan <- err.Error()
			return
		}

		var yStruct []byte
		if asJson.Checked {
			yStruct, err = json.MarshalIndent(abiOut.ABI.Structs, "", "  ")
		} else {
			yStruct, err = yaml.Marshal(abiOut.ABI.Structs)
		}
		if err != nil {
			errs.ErrChan <- err.Error()
			return
		}
		txt := r.ReplaceAllString(string(yStruct), "\n-")
		func(s string) {
			structs.SetText(s)
			structs.OnChanged = func(string) {
				structs.SetText(s)
			}
		}(txt) // deref
		structs.SetText(txt)

		var yActions []byte
		if asJson.Checked {
			yActions, err = json.MarshalIndent(abiOut.ABI.Actions, "", "  ")
		} else {
			yActions, err = yaml.Marshal(abiOut.ABI.Actions)
		}
		if err != nil {
			errs.ErrChan <- err.Error()
			return
		}
		txt = r.ReplaceAllString(string(yActions), "\n-")
		func(s string) {
			actions.OnChanged = func(string) {
				actions.SetText(s)
			}
		}(txt)