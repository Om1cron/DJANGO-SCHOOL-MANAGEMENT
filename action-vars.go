
package cryptonym

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/blockpane/cryptonym/fuzzer"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func abiSelectTypes(mustExist string) []string {
	types := []string{
		"authority",
		"bool",
		"byte",
		"byte[]",
		"checksum256",
		"float128",
		"float32",
		"float64",
		"hex_bytes",
		"int128",
		"int16",
		"int32",
		"int64",
		"name",
		"public_key",
		"signature",
		"string",
		"string[]",
		"symbol",
		"time",
		"timestamp",
		"uint128",
		"uint16",
		"uint32",
		"uint64",
		"varint32",
		"varuint32",
	}
	for _, t := range types {
		if t == mustExist {
			return types
		}
	}
	types = append(types, mustExist)
	return types
}

var sendAsSelectTypes = []string{
	"form value",
	"actor",
	"pub key",
	"fio types",
	"number",
	"bytes/string",
	//"load file",
}

var bytesVar = []string{
	"bytes",
	"bytes: base64 encoded",
	"bytes: hex encoded",
	"random checksum",
	"string",
}

var bytesLen = []string{
	"random length",
	"8",
	"12",
	"16",
	"32",
	"64",
	"128",
	"256",
	"512",
	"2,048",
	"4,096",
	//"8,192",
	//"16,384",
	//"32,768",
	//"65,536",
	//"131,072",
	//"262,144",
	//"524,288",
	//"1,048,576",
	//"2,097,152",
	//"4,194,304",
	//"8,388,608",
	//"16,777,216",
}

var formVar = []string{
	"as is",
	"FIO -> suf",
	"json -> struct",
	"base64 -> byte[]",
	"checksum256",
	"fio address@ (invalid)",
	"fio address@ (valid)",
	"fio address@ (valid, max size)",
	"hex -> byte[]",
	"signature",
}

var actorVar = []string{
	"mine",
	"random",
}

var numericVar = []string{
	"incrementing float",
	"incrementing int",
	"random float",
	"random int",
	"overflow int",
	"random number (mixed)",
	"max int",
}

var maxIntVar = []string{
	"int8",
	"uint8",
	"int16",
	"uint16",
	"int32",
	"uint32",
	"int64",
	"uint64",
}

var fioVar = []string{
	"invalid fio domain",
	"valid fio domain",
	"valid fio domain (max size)",
	"max length: newfundsreq.content",
	"max length: recordobt.content",
	"max length: regproducer.url",
	"max length: voteproducer.producers",
	"max length: addaddress.public_addresses",
	"variable length: addaddress.public_addresses",
	//TODO:
	//"string[] of existing fio address",
}

//TODO: "string[] of existing fio address"....
var addressLen = []string{
	"2",
	"4",
	"8",
	"16",
	"32",
}

var floatLen = []string{
	"32",
	"64",
}

var intLen = []string{
	"8",
	"16",
	"32",
	"64",
	"128",
}

var overflowLen = []string{
	"8",
	"16",
	"32",
}

var numAddresses = []string{
	"1",
	"2",
	"3",
	"4",
	"5",
	"10",
	"50",
	"100",
	"1000",
}

func sendAsVariant(kind string) (options []string, selected string) {
	switch kind {
	case "form value":
		return formVar, "as is"
	case "actor":
		return actorVar, "mine"
	case "pub key":
		return actorVar, "mine"
	case "number":
		return numericVar, "random int"
	case "bytes/string":
		return bytesVar, "string"
	case "fio types":
		return fioVar, "invalid fio domain"
	}
	return []string{}, "--"
}

func getLength(what string) (show bool, values []string, selected string) {
	switch {
	case what == "random float":
		return true, floatLen, "32"
	case what == "variable length addaddress.public_addresses":
		return true, numAddresses, "1"
	case what == "random int":
		return true, intLen, "32"
	case what == "overflow int":
		return true, overflowLen, "16"
	case what == "max int":
		return true, maxIntVar, "int32"
	case what == "random number (mixed)":
		return false, []string{""}, ""
	case strings.HasPrefix(what, "string") ||
		strings.HasPrefix(what, "bytes") ||
		strings.HasPrefix(what, "nop") ||
		strings.HasPrefix(what, "many"):
		return true, bytesLen, "64"
	}
	return
}

func defaultValues(contract string, action string, fieldName string, fieldType string, account *fio.Account, api *fio.API) string {
	var returnValue string
	switch {
	case fieldName == "amount":
		return "1,000.00"
	case fieldName == "bundled_transactions":
		return "100"
	case fieldName == "max_fee":
		api2, _, err := fio.NewConnection(nil, api.BaseURL)
		if err != nil {
			return "0"
		}
		api2.Header.Set("User-Agent", "fio-cryptonym-wallet")
		fio.UpdateMaxFees(api2)
		fee := fio.GetMaxFeeByAction(action)
		if fee == 0 {
			// as expensive as it gets ... pretty safe to return
			fee = fio.GetMaxFee("register_fio_domain")
		}
		returnValue = p.Sprintf("%.9f", fee)
	case fieldName == "can_vote":
		returnValue = "1"
	case fieldName == "is_public":
		returnValue = "1"
	case fieldType == "tokenpubaddr[]":
		a, t := fuzzer.NewPubAddress(account)
		returnValue = fmt.Sprintf(`[{
    "token_code": "%s",
    "chain_code": "%s",
    "public_address": "%s"
}]`, t, t, a)
	case fieldName == "url":
		returnValue = "https://fioprotocol.io"
	case fieldName == "location":
		returnValue = "80"
	case fieldName == "fio_domain":
		returnValue = "cryptonym"
	case fieldType == "bool":
		returnValue = "true"
	case fieldType == "authority":
		returnValue = `{
    "threshold": 2,
    "keys": [],
    "waits": [],
    "accounts": [
      {
        "permission": {
          "actor": "npe3obkgoteh",
          "permission": "active"
        },
        "weight": 1
      },
      {
        "permission": {
          "actor": "extjnqh3j3gt",
          "permission": "active"
        },
        "weight": 1
      }
    ]
  }`
	case strings.HasSuffix(fieldType, "int128"):
		i28 := eos.Uint128{
			Lo: uint64(rand.Int63n(math.MaxInt64)),
			Hi: uint64(rand.Int63n(math.MaxInt64)),
		}
		j, _ := json.Marshal(i28)
		returnValue = strings.Trim(string(j), `"`)
	case strings.HasPrefix(fieldType, "uint") || strings.HasPrefix(fieldType, "int"):
		returnValue = strconv.Itoa(rand.Intn(256))
	case strings.HasPrefix(fieldType, "float"):
		returnValue = "3.14159265359"
	case fieldName == "owner" || fieldName == "account" || fieldName == "actor" || fieldType == "authority" || fieldName == "proxy":
		actor, _ := fio.ActorFromPub(account.PubKey)
		returnValue = string(actor)
	case strings.Contains(fieldName, "public") || strings.HasSuffix(fieldName, "_key"):
		returnValue = account.PubKey
	case fieldName == "tpid":
		returnValue = Settings.Tpid
	case strings.HasSuffix(fieldName, "_address") || strings.HasPrefix(fieldName, "pay"):
		returnValue = DefaultFioAddress
	case fieldName == "authority" || (fieldName == "permission" && fieldType == "name"):
		returnValue = "active"
	case fieldName == "producers":
		returnValue = GetCurrentVotes(string(account.Actor), api)
	case fieldType == "transaction":
		returnValue = `{
  "context_free_actions": [],
  "actions": [
    {
      "signatures": [
        "SIG_K1_..."
      ],
      "compression": "none",
      "packed_context_free_data": "",
      "packed_trx": "b474345e54..."
    }
  ],
  "transaction_extensions": []
}`
	case fieldType == "asset":
		returnValue = "100000.000000000 FIO"
	case (fieldName == "to" || fieldName == "from") && fieldType == "name":
		returnValue = string(account.Actor)
	case fieldType == "permission_level":
		returnValue = fmt.Sprintf(`{
    "actor":"%s",
    "permission":"active"
}`, account.Actor)
	case fieldName == "periods":
		returnValue = `[
    {
        "duration": 86400,
        "percent": 1.2