package errs

import (
	"fyne.io/fyne/widget"
	"log"
	"strings"
	"time"
)

var (
	ErrChan        = make(chan string)
	DisconnectChan = make(chan bool)
	ErrTxt         = make([]string, 50)
	ErrMsgs        = widget.NewMultiLineEntry()
	RefreshChan    = make(chan bool)
	Connected      bool
)

func init() {
	go func(msg chan string, disconnected chan bool) {
		last := time.Now()
		t := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case d := <-disconnecte