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
		return
	}
	errLog, e := os.OpenFile(fmt.Sprintf("%s%c%s%cerror.log", d, os.PathSeparator, settingsDir, os.PathSeparator), os.O_CREATE|os.O_