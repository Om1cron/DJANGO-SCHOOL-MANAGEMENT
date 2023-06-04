
package cryptonym

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	fioassets "github.com/blockpane/cryptonym/assets"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
	"image"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type existVotes struct {
	Producers []string `json:"producers"`
}

type prodRow struct {
	FioAddress string `json:"fio_address"`
}

func GetCurrentVotes(actor string, api *fio.API) (votes string) {
	getVote, err := api.GetTableRows(eos.GetTableRowsRequest{
		Code:  "eosio",
		Scope: "eosio",
		Table: "voters",

		Index:      "3",
		LowerBound: actor,
		UpperBound: actor,
		Limit:      1,
		KeyType:    "name",
		JSON:       true,
	})
	if err != nil {
		return
	}
	v := make([]*existVotes, 0)
	err = json.Unmarshal(getVote.Rows, &v)
	if err != nil {
		return
	}
	if len(v) == 0 {
		return
	}
	votedFor := make([]string, 0)
	for _, row := range v[0].Producers {
		if row == "" {
			continue
		}
		gtr, err := api.GetTableRows(eos.GetTableRowsRequest{
			Code:       "eosio",
			Scope:      "eosio",
			Table:      "producers",
			LowerBound: row,
			UpperBound: row,
			KeyType:    "name",
			Index:      "4",
			JSON:       true,
		})
		if err != nil {
			continue
		}
		p := make([]*prodRow, 0)
		err = json.Unmarshal(gtr.Rows, &p)
		if err != nil {
			continue
		}
		if len(p) == 1 && p[0].FioAddress != "" {
			votedFor = append(votedFor, p[0].FioAddress)
		}
	}
	if len(votedFor) == 0 {
		return
	}
	return strings.Join(votedFor, ", ")
}

type bpInfo struct {
	CurrentVotes float64
	FioAddress   string
	Actor        string
	BpJson       *fio.BpJson
	VoteFor      bool
	OrigVoteFor  bool
	Url          string
	Img          *canvas.Image
	Top21        bool
	Tied         bool
}

var bpInfoCache = make(map[string]*bpInfo)

func getBpInfo(actor string, api *fio.API) ([]bpInfo, error) {
	bpi := make([]bpInfo, 0)
	p, err := api.GetFioProducers()
	if err != nil {
		return bpi, err
	}

	prods := &fio.Producers{
		Producers: make([]fio.Producer, 0),
	}
	// little hack for testnet to cleanup crap from various test scripts:
	for i := range p.Producers {
		if strings.Contains(p.Producers[i].Url, "dapix.io") {
			continue
		}
		prods.Producers = append(prods.Producers, p.Producers[i])
	}

	curVotes := strings.Split(GetCurrentVotes(actor, api), ",")
	for i := range curVotes {
		curVotes[i] = strings.TrimSpace(curVotes[i])
	}
	hasVoted := func(s string) bool {
		for _, v := range curVotes {
			if s == v {
				return true
			}
		}
		return false
	}
	isTopProd := make(map[string]bool)
	sched, _ := api.GetProducerSchedule()
	for _, tp := range sched.Active.Producers {
		isTopProd[string(tp.AccountName)] = true
	}
	voteTies := make(map[string]int)
	for _, bp := range prods.Producers {
		voteTies[bp.TotalVotes] += 1