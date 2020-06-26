
package cryptonym

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"reflect"
	"strings"
	"sync"
)

type AbiFormItem struct {
	Contract     string
	Action       string
	Order        int
	Name         *string
	Type         *widget.Select
	SendAs       *widget.Select
	Variation    *widget.Select
	Len          *widget.Select
	Input        *widget.Entry
	Value        *interface{}
	IsSlice      bool
	SliceValue   []*interface{}
	convert      func(s interface{}) interface{}
	typeOverride string
	noJsonEscape bool // if true uses fmt, otherwise json-encodes the value ... fmt is useful for some numeric values
}

type Abi struct {
	mux sync.RWMutex

	lookUp   map[string]int
	Rows     []AbiFormItem
	Action   string
	Contract string
}

func NewAbi(length int) *Abi {
	return &Abi{
		Rows:   make([]AbiFormItem, length),
		lookUp: make(map[string]int),
	}
}

func (abi *Abi) AppendRow(myName string, account *fio.Account, form *widget.Form) {
	go func() {
		if myName == "" {
			myName = fmt.Sprintf("new_row_%d", len(abi.Rows))
		}
		if abi.Rows == nil {
			abi.Rows = make([]AbiFormItem, 0)
		}
		typeSelect := &widget.Select{}