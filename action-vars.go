
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