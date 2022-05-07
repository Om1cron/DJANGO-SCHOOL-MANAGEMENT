package cryptonym

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type signer struct {
	weight *widget.Entry
	actor  *widget.Entry
	index  int
}

var (
	MsigRefreshRequests = make(chan bool)
	MsigLastTab         = 0
	MsigLoaded          bool
)

func UpdateAuthContent(container chan fyne.Container, api *fio.API, opts *fio.TxOptions, account *fio.Account) {
	for !Connected {
		time.Sleep(time.S