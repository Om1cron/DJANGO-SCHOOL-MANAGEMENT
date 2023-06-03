
package cryptonym

import (
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos/ecc"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"strings"
	"sync"
	"time"
)

func vanityKey(o *vanityOptions, quit chan bool) (*fio.Account, error) {
	account := &fio.Account{}
	var hit bool
	errs.ErrChan <- "vanity search starting for " + o.word
	found := func(k *key) {
		errs.ErrChan <- fmt.Sprintf("Vanity generator found a match: %s %s", k.actor, k.pub)