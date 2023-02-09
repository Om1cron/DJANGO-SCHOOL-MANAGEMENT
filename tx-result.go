package cryptonym

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"math"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	Results       = make([]TxResult, 0)
	requestText   = widget.NewMultiLineEntry()
	responseText  = widget.NewMultiLineEntry()
	stopRequested = make(chan bool)
)

type TxResult struct {
	FullResp []byte
	FullReq  []byte
	Resp     []byte
	Req      []byte
	Success  bool
	Index    int
	Summary  string
}

type TxSummary struct {
	TransactionId string `json:"transaction_id" yaml:"Transaction Id"`
	Processed     struct {
		BlockNum  uint32 `json:"block_num" yaml:"Block Number"`
		BlockTime string `json:"block_time" yaml:"Block Time"`
		Receipt   struct {
			Status string `json:"status" yaml:"Status"`
		} `json:"receipt" yaml:"Receipt,omitempty"`
	} `json:"processed" yaml:"Processed,omitempty"`
	ErrorCode  interface{} `json:"error_code" yaml:"Error,omitempty"`                             // is this a string, int, varies on context?
	TotalBytes int         `json:"total_bytes,omitempty" yaml:"TX Size of All Actions,omitempty"` // this is field we calculate later
}

// to get the *real* size of what was transacted, we need to dig into the action traces and look at the length
// of the hex_data field, which is buried in the response.
type txTraces struct {
	Processed struct {
		ActionTraces []struct {
			Act struct {
				HexData string `json:"hex_data"`
			} `json:"act"`
		} `json:"action_traces"`
	} `json:"processed"`
}

func (tt txTraces) size() int {
	if len(tt.Processed.ActionTraces) == 0 {
		return 0
	}
	var sz int
	for _, t := range tt.Processed.ActionTraces {
		sz = sz + (len(t.Act.HexData) / 2)
	}
	return sz
}

type txResultOpts struct {
	repeat      int
	loop        bool
	threads     string
	hideFail    bool
	hideSucc    bool
	window      fyne.Window
	gone        bool
	msig        bool
	msigSigners string
	msigAccount string
	msigName    func() string
	wrap        bool
	wrapActor   string
}

func TxResultsWindow(win *txResultOpts, api *fio.API, opts *fio.TxOptions, account *fio.Account) {
	ResetTxResult()

	// this is a workaround for fyne  sometimes showing blank black windows, resizing fixes
	// but when it happens the window still doesn't work correctly. It will show up, but does not
	// refresh. Beats a black window, and at least the close button works.
	resizeTrigger := make(chan interface{})
	go func() {
		for {
			select {
			case