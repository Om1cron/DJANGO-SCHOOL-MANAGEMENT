
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