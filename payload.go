
package cryptonym

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/blockpane/cryptonym/fuzzer"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// TODO: rethink this, instead of calling this every request, maybe pass pointers to functions?
func (abi *Abi) GeneratePayloads(key *fio.Account) error {
	abi.mux.RLock()
	defer abi.mux.RUnlock()

	for i := 0; i < len(abi.Rows); i++ {
		form := &abi.Rows[i]
		err := func() error {
			var v interface{}
			v = nil
			isSlice := false
			if strings.HasSuffix(form.Type.Selected, `[]`) {
				isSlice = true
			}
			switch form.SendAs.Selected {

			case "form value":
				switch form.Variation.Selected {
				case "as is":
					if strings.Contains(form.Type.Selected, "int") {
						t, e := strconv.ParseInt(form.Input.Text, 10, 64)
						if e != nil {
							return errors.New(*form.Name + ": " + e.Error())
						}
						abi.mux.RUnlock()
						FormState.UpdateValue(&i, t, isSlice, true)
						abi.mux.RLock()
						return nil
					} else if strings.Contains(form.Type.Selected, "float") {
						t, e := strconv.ParseFloat(form.Input.Text, 64)
						if e != nil {
							return errors.New(*form.Name + ": " + e.Error())
						}