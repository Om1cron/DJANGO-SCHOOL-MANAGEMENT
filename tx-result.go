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
			case <-resizeTrigger:
				if win.window == nil || !win.window.Content().Visible() {
					continue
				}
				win.window.Resize(fyne.NewSize(txW, txH))
				time.Sleep(100 * time.Millisecond)
				win.window.Resize(win.window.Content().MinSize())
			}
		}
	}()

	workers, e := strconv.Atoi(win.threads)
	if e != nil {
		workers = 1
	}

	var (
		grid              *fyne.Container
		b                 *widget.Button
		stopButton        *widget.Button
		closeRow          *widget.Group
		running           bool
		exit              bool
		fullResponseIndex int
	)

	successLabel := widget.NewLabel("")
	failedLabel := widget.NewLabel("")
	successChan := make(chan bool)
	failedChan := make(chan bool)
	go func(s chan bool, f chan bool) {
		time.Sleep(100 * time.Millisecond)
		BalanceChan <- true
		tick := time.NewTicker(time.Second)
		update := false
		updateBalance := false
		successCount := 0
		failedCount := 0
		for {
			select {
			case <-tick.C:
				if updateBalance {
					BalanceChan <- true
					updateBalance = false
				}
				if update {
					successLabel.SetText(p.Sprintf("%d", successCount))
					failedLabel.SetText(p.Sprintf("%d", failedCount))
					successLabel.Refresh()
					failedLabel.Refresh()
					update = false
				}
			case <-f:
				update = true
				failedCount = failedCount + 1
			case <-s:
				update = true
				updateBalance = true
				successCount = successCount + 1
			}
		}
	}(successChan, failedChan)

	run := func() {}
	mux := sync.Mutex{}
	Results = make([]TxResult, 0)

	summaryGroup := widget.NewGroupWithScroller("Transaction Result")
	showFullResponseButton := widget.NewButtonWithIcon("Show Response Details", theme.VisibilityIcon(), func() {
		// avoid nil pointer
		if len(Results) <= fullResponseIndex {
			errs.ErrChan <- "could not show full response: invalid result index - this shouldn't happen!"
			return
		}
		if len(Results[fullResponseIndex].FullResp) == 0 {
			errs.ErrChan <- "could not show full response: empty string"
			return
		}
		ShowFullResponse(Results[fullResponseIndex].FullResp, win.window)
	})
	showFullRequestButton := widget.NewButtonWithIcon("Show Request JSON", theme.VisibilityIcon(), func() {
		// avoid nil pointer
		if len(Results) <= fullResponseIndex {
			errs.ErrChan <- "could not show full request: invalid result index - this shouldn't happen!"
			return
		}
		if len(Results[fullResponseIndex].FullReq) == 0 {
			errs.ErrChan <- "could not show full request: empty string"
			return
		}
		ShowFullRequest(Results[fullResponseIndex].FullReq, win.window)
	})

	textUpdateDone := make(chan interface{})
	textUpdateReq := make(chan string)
	textUpdateResp := make(chan string)
	go func() {
		for {
			select {
			case <-textUpdateDone:
				return
			case s := <-textUpdateReq:
				requestText.OnChanged = func(string) {
					requestText.SetText(s)
				}
				requestText.SetText(s)
			case s := <-textUpdateResp:
				responseText.OnChanged = func(string) {
					responseText.SetText(s)
				}
				responseText.SetText(s)
			}
		}
	}()

	setGrid := func() {
		grid = fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			fyne.NewContainerWithLayout(layout.NewGridLayoutWithRows(1),
				closeRow,
				fyne.NewContainerWithLayout(layout.NewMaxLayout(),
					summaryGroup,
				),
			),
			widget.NewVBox(
				fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(txW, 30)),
					fyne.NewContainerWithLayout(layout.NewGridLayout(2), showFullResponseButton, showFullRequestButton),
				),
				widget.NewLabel("Request:"),
				requestText,
				widget.NewLabel("Response Summary:"),
				responseText,
			),
		)

		win.window.SetContent(grid)
		win.window.Resize(win.window.Content().MinSize())
	}

	clear := func() {
		Results = make([]TxResult, 0)
		summaryGroup = widget.NewGroupWithScroller("Transaction Result")
		summaryGroup.Refresh()
		textUpdateResp <- ""
		textUpdateReq <- ""
		setGrid()
	}

	closeButton := widget.NewButtonWithIcon(
		"close",
		theme.DeleteIcon(),
		func() {
			go func() {
				if running {
					stopRequested <- true
				}
				win.gone = true
				win.window.Hide()
				// this causes a segfault on linux, but on darwin if not closed it leaves a window hanging around.
				if runtime.GOOS == "darwin" {
					win.window.Close()
				}
			}()
		},
	)
	resendButton := widget.NewButtonWithIcon("resend", theme.ViewRefreshIcon(), func() {
		if running {
			return
		}
		exit = false
		go run()
	})
	stopButton = widget.NewButtonWithIcon("stop", theme.CancelIcon(), func() {
		if running {
			stopRequested <- true
		}
	})

	clearButton := widget.NewButtonWithIcon("clear results", theme.ContentRemoveIcon(), func() {
		clear()
	})
	closeRow = widget.NewGroup(" Control ",
		stopButton,
		resendButton,
		clearButton,
		closeButton,
		layout.NewSpacer(),
		BalanceLabel,
		widget.NewLabel("Successful Requests:"),
		successLabel,
		widget.NewLabel("Failed Requests:"),
		failedLabel,
	)
	closeRow.Show()

	reqChan := make(chan string)
	respChan := make(chan string)
	fullRespChan := make(chan int)

	trimDisplayed := func(s string) string {
		re := regexp.MustCompile(`[[^:ascii:]]`)
		var displayed string
		s = s + "\n"
		reader := strings.NewReader(s)
		buf := bufio.NewReader(reader)
		var lines int
		for {
			lines = lines + 1
			line, err := buf.ReadString('\n')
			if err != nil {
				break
			}
			line, _ = strconv.Unquote(strconv.QuoteToASCII(line))
			if len(line) > 128+21 {
				line = fmt.Sprintf("%s ... trimmed %d chars ...\n", line[:128], len(line)-128)
			}
			displayed = displayed + line
			if lines > 31 {
				displayed = displayed + "\n ... too many lines to display ..."
				break
			}
		}
		return re.ReplaceAllString(displayed, "?")
	}

	go func(rq chan string, rs chan string, frs chan int) {
		for {
			select {
			case q := <-rq:
				textUpdateReq <- trimDisplayed(q)
			case s := <-rs:
				textUpdateResp <- trimDisplayed(s)
			case fullResponseIndex = <-frs:
			}
		}
	}(reqChan, respChan, fullRespChan)
	reqChan <- ""
	respChan <- ""

	repaint := func() {
		mux.Lock()
		closeRow.Refresh()
		summaryGroup.Refresh()
		responseText.Refresh()
		requestText.Refresh()
		if grid != nil {
			grid.Refresh()
		}
		mux.Unlock()
	}

	newButton := func(title string, index int, failed bool) {
		if failed {
			failedChan <- false
		} else {
			successChan <- true
		}
		if (!failed && win.hideSucc) || (failed && win.hideFail) {
			return
		}
		// possible race while clearing the screen
		if index > len(Results) {
			return
		}
		deRef := &index
		i := *deRef
		if i-1 > len(Results) || len(Results) == 0 {
			return
		}
		if len(Results) > 256 {
			clear()
		}
		icon := theme.ConfirmIcon()
		if failed {
			icon = theme.CancelIcon()
		}

		b = widget.NewButtonWithIcon(title, icon, func() {
			if i >= len(Results) {
				return
			}
			reqChan <- string(Results[i].Req)
			respChan <- string(Results[i].Resp)
			fullRespChan <- i
		})
		summaryGroup.Append(b)
		repaint()
	}

	run = func() {
		defer func() {
			if running {
				stopRequested <- true
			}
		}()
		// give each thread it's own http client pool:
		workerApi, workerOpts, err := fio.NewConnection(account.KeyBag, api.BaseURL)
		if err != nil {
			errs.ErrChan <- err.Error()
			errs.ErrChan <- "ERROR: could not get new client connection"
			return
		}
		workerApi.Header.Set("User-Agent", "fio-cryptonym-wallet")
		running = true
		stopButton.Enable()
		bombsAway.Disable()
		resendButton.Disable()
		closeButton.Disable()

		defer func() {
			running = false
			stopButton.Disable()
			bombsAway.Enable()
			resendButton.Enable()
			closeButton.Enable()
		}()

		var end int
		switch {
		case win.loop:
			end = math.MaxInt32
		case win.repeat > 1:
			end = win.repeat
		default:
			end = 1
		}
		finished := make(chan bool)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer func() {
				running = false
				finished <- true
				wg.Done()
			}()
			for i := 0; i < end; i++ {
				if exit {
					return
				}
				output := TxResult{
					Summary: fmt.Sprintf("%s", time.Now().Format("15:04:05.000")),
					Index:   i,
				}
				e := FormState.GeneratePayloads(account)
				if e != nil {
					errs.ErrChan <- e.Error()
					errs.ErrChan <- "there was a problem generating dynamic payloads"
					output.Resp = []byte(e.Error())
					Results = append(Results, output)
					newButton(output.Summary, len(Results)-1, true)
					continue
				}
				if exit {
					return
				}
				raw, tx, err := FormState.PackAndSign(workerApi, workerOpts, account, win.msig)
				if tx == nil || tx.PackedTransaction == nil {
					errs.ErrChan <- "sending a signed transaction with null action data"
					empty := fio.NewAction(eos.AccountName(F