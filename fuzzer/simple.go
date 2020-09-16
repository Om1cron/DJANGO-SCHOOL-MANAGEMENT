
package fuzzer

import (
	"bytes"
	"context"
	cr "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"github.com/fioprotocol/fio-go/eos/ecc"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	EncodeRaw int8 = iota
	EncodeHexString
	EncodeBase64
	badChars = `~!#$%^&*()+=\|{}[]";:?/.>,<@"` + "`"
)

func RandomString(length int) string {
	var payload string
	for i := 0; i < length; i++ {
		payload = payload + string(byte(rand.Intn(26)+97))
	}
	return payload
}

type RandomNumberResult struct {
	abi     string
	value   interface{}
	convert func(s interface{}) interface{}
}

func (rn RandomNumberResult) AbiType() string {
	return rn.abi
}

func (rn RandomNumberResult) String() string {
	return fmt.Sprintf("%v", rn.value)
}

func (rn RandomNumberResult) Interface() interface{} {
	return rn.value
}

func (rn RandomNumberResult) ConvertFunc() func(s interface{}) interface{} {
	return rn.convert
}

func RandomNumber() RandomNumberResult {
	intLens := [...]int{8, 16, 32, 64} // don't send 128 here.
	floatLen := 32

	var result interface{}
	signed := false
	rn := RandomNumberResult{}

	negative := 1
	if rand.Intn(2) > 0 {
		signed = true
		if rand.Intn(2) > 0 {
			negative = -1
		}
	}

	intOrFloat := rand.Intn(2)
	switch intOrFloat {
	case 0:
		l := intLens[rand.Intn(len(intLens))]
		result = RandomInteger(l)
		rn.abi = fmt.Sprintf("int%d", l)
		if !signed {
			rn.abi = fmt.Sprintf("uint%d", l)
		}
		switch l {
		case 8:
			rn.convert = func(s interface{}) interface{} {
				parsed, _ := strconv.ParseInt(fmt.Sprintf("%d", s), 10, 8)
				if !signed {
					return uint8(parsed)
				}
				return int8(parsed * int64(negative))
			}
		case 16:
			rn.convert = func(s interface{}) interface{} {
				parsed, _ := strconv.ParseInt(fmt.Sprintf("%d", s), 10, 16)
				if !signed {
					return uint16(parsed)
				}
				return int16(parsed * int64(negative))
			}
		case 32:
			rn.convert = func(s interface{}) interface{} {
				parsed, _ := strconv.ParseInt(fmt.Sprintf("%d", s), 10, 32)
				if !signed {
					return uint32(parsed)
				}
				return int32(parsed * int64(negative))
			}
		case 64:
			rn.convert = func(s interface{}) interface{} {
				parsed, _ := strconv.ParseInt(fmt.Sprintf("%d", s), 10, 64)
				if !signed {
					return uint64(parsed)
				}
				return parsed * int64(negative)
			}
		}
	case 1:
		fl := floatLen * (rand.Intn(2) + 1)
		if fl == 32 {
			rn.abi = "float32"
			result = float32(RandomFloat(fl) * float64(negative))
		} else {
			rn.abi = "float64"
			result = RandomFloat(fl) * float64(negative)
		}

	}
	rn.value = result
	return rn
}

func MaxInt(size string) uint64 {
	switch size {
	case "int8":
		return uint64(math.MaxInt8)
	case "uint8":
		return uint64(math.MaxUint8)
	case "int16":
		return uint64(math.MaxInt16)
	case "uint16":
		return uint64(math.MaxUint16)
	case "int32":
		return uint64(math.MaxInt32)
	case "uint32":
		return uint64(math.MaxUint32)
	case "int64":
		return uint64(math.MaxInt64)
	case "uint64":
		return math.MaxUint64
	default:
		return uint64(math.MaxInt32)
	}
}

func RandomInteger(size int) RandomNumberResult {
	switch size {
	case 8:
		//return int8(rand.Intn(math.MaxInt8-1) + 1)
		return RandomNumberResult{
			abi:   "int8",
			value: int8(rand.Intn(math.MaxInt8-1) + 1),
		}
	case 16:
		return RandomNumberResult{