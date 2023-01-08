package cryptonym

import (
	"fmt"
	"github.com/fioprotocol/fio-go"
	"testing"
)

func TestGetAccountActions(t *testing.T) {
	api, _, err := fio.NewConnection(nil, "http: