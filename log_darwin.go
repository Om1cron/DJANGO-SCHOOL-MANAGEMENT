package cryptonym

import (
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"log"
	"os"
	"syscall"
)

func startErrLog() {
	d, e := os.UserConfigDir()
	if e != nil {
		log.Println(e)
		retur