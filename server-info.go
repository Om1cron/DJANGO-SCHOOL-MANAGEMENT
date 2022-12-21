
package cryptonym

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go/eos"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)