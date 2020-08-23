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
	ErrTxt         = make([]st