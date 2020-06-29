
package cryptonym

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"gopkg.in/alessio/shellescape.v1"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

func NewApiRequestTab(container chan fyne.Container) {
	apiList := SupportedApis{Apis: []string{"/v1/chain/get_info"}}
	err := apiList.Update(Uri, false)
	if err != nil {
		errs.ErrChan <- "Error updating list of available APIs: " + err.Error()
	}
	inputEntry := widget.NewMultiLineEntry()
	outputEntry := widget.NewMultiLineEntry()
	statusLabel := widget.NewLabel("")
	submit := &widget.Button{}
	inputTab := &widget.TabItem{}
	outputTab := &widget.TabItem{}
	apiTabs := &widget.TabContainer{}

	submit = widget.NewButtonWithIcon("Submit", fioassets.NewFioLogoResource(), func() {
		submit.Disable()
		statusLabel.SetText("")
		outputEntry.SetText("")
		outputEntry.OnChanged = func(string) {}
		apiTabs.SelectTab(outputTab)
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
		done := make(chan bool, 1)
		defer func() {
			submit.Enable()
			outputEntry.Refresh()
			cancel()
		}()
		go func() {
			defer func() {
				done <- true
			}()
			resp, err := http.Post(Uri+apiEndPointActive, "application/json", bytes.NewReader([]byte(inputEntry.Text)))
			if err != nil {
				outputEntry.SetText(err.Error())
				errs.ErrChan <- err.Error()