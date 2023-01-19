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
		} `jso