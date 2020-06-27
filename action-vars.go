
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