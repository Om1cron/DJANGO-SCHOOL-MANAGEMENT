
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
						abi.mux.RUnlock()
						FormState.UpdateValue(&i, t, isSlice, true)
						abi.mux.RLock()
						return nil
					}
					v = form.Input.Text
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				case "FIO -> suf":
					unFriendly := strings.ReplaceAll(strings.ReplaceAll(form.Input.Text, ",", ""), "_", "")
					f, e := strconv.ParseFloat(unFriendly, 64)
					if e != nil {
						return errors.New(*form.Name + ": " + e.Error())
					}
					t := uint64(f * 1_000_000_000.0)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, t, isSlice, true)
					abi.mux.RLock()
					return nil
				case "json -> struct":
					v = form.Input.Text
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, false, true)
					abi.mux.RLock()
					return nil
				case "hex -> byte[]":
					h, e := hex.DecodeString(form.Input.Text)
					if e != nil {
						return errors.New(*form.Name + ": " + e.Error())
					}
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, h, isSlice, false)
					abi.mux.RLock()
					return nil
				case "base64 -> byte[]":
					buf := bytes.NewReader([]byte(form.Input.Text))
					b64 := base64.NewDecoder(base64.StdEncoding, buf)
					b := make([]byte, 0)
					_, e := b64.Read(b)
					if e != nil {
						return errors.New(*form.Name + ": " + e.Error())
					}
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, b, isSlice, false)
					abi.mux.RLock()
					return nil
				case "checksum256":
					v = fuzzer.ChecksumOf(form.Input.Text)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				case "signature":
					v = fuzzer.SignatureFor(form.Input.Text, key)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				case "fio address@ (valid)":
					v = fuzzer.FioAddressAt(form.Input.Text)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
				case "fio address@ (valid, max size)":
					v = fuzzer.MaxRandomFioAddressAt(form.Input.Text)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
				case "fio address@ (invalid)":
					v = fuzzer.InvalidFioAddressAt(form.Input.Text)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
				}

			case "actor":
				switch form.Variation.Selected {
				case "mine":
					v = key.Actor
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				case "random":
					v = fuzzer.RandomActor()
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				}

			case "pub key":
				switch form.Variation.Selected {
				case "mine":
					v = key.PubKey
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				case "random":
					v = fuzzer.RandomFioPubKey()
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, isSlice, false)
					abi.mux.RLock()
					return nil
				}

			case "fio types":
				switch form.Variation.Selected {
				// TODO:
				//case "random array of existing fio address":
				case "max length: addaddress.public_addresses":
					v = fuzzer.MaxAddPubAddress()
					fmt.Println(v)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, false, true)
					abi.mux.RLock()
					return nil
				case "max length: voteproducer.producers":
					v = fuzzer.MaxVoteProducers(Uri)
					fmt.Println(v)
					abi.mux.RUnlock()
					FormState.UpdateValueWithConvert(&i, v, false, "string[]", false)
					abi.mux.RLock()
					return nil
				case "variable length: addaddress.public_addresses":
					var payloadLen int
					var err error
					if form.Len.Selected == "" {
						payloadLen = 1
						form.Len.SetSelected("1")
					} else {
						payloadLen, err = strconv.Atoi(form.Len.Selected)
						if err != nil {
							payloadLen = 1
						}
					}
					v = fuzzer.RandomAddAddress(payloadLen)
					abi.mux.RUnlock()
					FormState.UpdateValue(&i, v, false, true)
					abi.mux.RLock()