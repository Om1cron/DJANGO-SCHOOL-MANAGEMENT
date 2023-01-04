
package cryptonym

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"sort"
	"strconv"
	"sync"
)

type FioActions struct {
	sync.RWMutex
	Index   []string
	Actions map[string][]string
}

type contracts struct {
	Owner string `json:"owner"`
}

func GetAccountSummary(api *fio.API) (*FioActions, error) {

	table, err := api.GetTableRows(eos.GetTableRowsRequest{
		Code:  "eosio",
		Scope: "eosio",
		Table: "abihash",
		Limit: uint32(math.MaxUint32),
		JSON:  true,
	})
	if err != nil {
		return nil, err
	}
	result := make([]contracts, 0)
	err = json.Unmarshal(table.Rows, &result)
	if err != nil {
		return nil, err
	}
	// FIXME: reading the abihash table isn't returning everything because of how the chain is boostrapped.
	// for now, appending a list of known contracts if not found :(
	defaults := []contracts{
		{Owner: "eosio"},
		//{Owner: "eosio.bios"},
		{Owner: "eosio.msig"},
		{Owner: "eosio.wrap"},
		{Owner: "fio.address"},
		//{Owner: "fio.common"},
		{Owner: "fio.fee"},
		{Owner: "fio.foundatn"},
		{Owner: "fio.reqobt"},
		//{Owner: "fio.system"},
		{Owner: "fio.token"},
		{Owner: "fio.tpid"},
		{Owner: "fio.treasury"},
		//{Owner: "fio.whitelst"},
	}
	for _, def := range defaults {
		func() {
			for _, found := range result {
				if def.Owner == found.Owner {
					return
				}
			}
			result = append(result, def)
		}()
	}

	actions := FioActions{
		Actions: make(map[string][]string),
	}

	// sort by account name
	func() {
		sorted := make([]string, 0)
		resultTemp := make([]contracts, 0)
		sortMap := make(map[string]int)
		for i, c := range result {
			sorted = append(sorted, c.Owner)
			sortMap[c.Owner] = i
		}
		sort.Strings(sorted)
		for _, ascending := range sorted {
			resultTemp = append(resultTemp, result[sortMap[ascending]])
		}
		result = resultTemp
	}()

	for _, name := range result {
		bi, err := api.GetABI(eos.AccountName(name.Owner))
		if err != nil {
			errs.ErrChan <- "problem while loading abi: " + err.Error()
			continue
		}
		actionList := make(map[string]bool, 0)
		for _, name := range bi.ABI.Actions {
			actionList[string(name.Name)] = true
		}
		if actions.Actions[name.Owner] == nil {
			actions.Actions[name.Owner] = make([]string, 0)
			actions.Index = append(actions.Index, name.Owner)
		}
		for a := range actionList {
			actions.Actions[name.Owner] = append(actions.Actions[name.Owner], a)
		}
		func() {
			tableList := make([]string, 0)
			for _, table := range bi.ABI.Tables {
				tableList = append(tableList, string(table.Name))
			}
			if len(tableList) == 0 {
				return
			}
			TableIndex.Add(name.Owner, tableList)
		}()
	}
	return &actions, nil
}

type TableBrowserIndex struct {
	mux     sync.RWMutex
	tables  map[string][]string
	created bool
}

func NewTableIndex() *TableBrowserIndex {
	return &TableBrowserIndex{tables: make(map[string][]string)}
}

func (tb *TableBrowserIndex) IsCreated() bool {
	return tb.created
}

func (tb *TableBrowserIndex) SetCreated(b bool) {
	tb.created = b
}

func (tb *TableBrowserIndex) Add(contract string, tables []string) (ok bool) {
	if contract == "" || len(tables) == 0 {
		return false
	}
	tb.mux.Lock()
	defer tb.mux.Unlock()
	sort.Strings(tables)
	tb.tables[contract] = tables
	return true
}
