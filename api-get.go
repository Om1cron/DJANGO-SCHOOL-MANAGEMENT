
package cryptonym

import (
	"encoding/json"
	"errors"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"io/ioutil"
	"strings"
	"sync"
)

type SupportedApis struct {
	Apis []string `json:"apis"`
}

func (apiList *SupportedApis) Update(url string, filter bool) error {
	api, _, err := fio.NewConnection(nil, url)
	if api.HttpClient == nil || err != nil {
		errMsg := "attempted to retrieve api information, but not connected "
		if err != nil {
			errMsg = errMsg + err.Error()
		}
		errs.ErrChan <- "fetchApis: " + errMsg
		return errors.New(errMsg)
	}
	resp, err := api.HttpClient.Post(api.BaseURL+"/v1/node/get_supported_apis", "application/json", nil)
	if err != nil {
		errs.ErrChan <- "fetchApis: " + err.Error()
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errs.ErrChan <- "fetchApis: " + err.Error()
		return err
	}
	supported := &SupportedApis{}
	err = json.Unmarshal(body, supported)
	if err != nil {
		errs.ErrChan <- "fetchApis: " + err.Error()
		return err
	}
	supported.Apis = append(supported.Apis, "/v1/node/get_supported_apis")
	if filter {
		newList := make([]string, 0)
		for _, a := range supported.Apis {
			if strings.Contains(a, "get") || strings.Contains(a, "abi") ||
				strings.Contains(a, "net") || strings.Contains(a, "json") ||
				strings.Contains(a, "check") {
				continue
			}
			newList = append(newList, a)
		}
		apiList.Apis = newList
		return nil
	}
	apiList.Apis = supported.Apis
	return nil
}

func DefaultJsonFor(endpoint string) string {
	defaultApiJsonMux.RLock()
	defer defaultApiJsonMux.RUnlock()
	//return defaultApiJson[endpoint]
	switch endpoint {
	case "/v1/chain/get_transaction_id":
		return `{
  "transaction": {
    "actions": [
      {
        "account": "fio.token",
        "name": "trnsfiopubky",
        "authorization": [
          {
            "actor": "` + defaultActor() + `",
            "permission": "active"
          }
        ],
        "data": "00"
      }
    ]
  }
}`
	case "/v1/chain/get_table_by_scope":
		return `{
  "code": "eosio.msig",
  "table": "proposal",
  "lower_bound": "111111111111",
  "upper_bound": "zzzzzzzzzzzz",
  "limit": 10
}`
	case "/v1/chain/get_required_keys":
		return `{
  "transaction": {
    "actions": [{
        "account": "fio.token",
        "name": "trnsfiopubky",
        "authorization": [{
            "actor": "` + defaultActor() + `",
            "permission": "active"
          }
        ],
        "data": "00"
      }
    ],
    "transaction_extensions": []
  },
  "available_keys": [
    "` + defaultPub() + `"
  ]
}`
	case "/v1/history/get_actions":
		return `{
  "account_name": "` + defaultActor() + `",
  "pos": -1
}`
	case "/v1/history/get_block_txids":
		return `{"block_num": 123}`
	case "/v1/history/get_key_accounts":
		return `{
  "public_key": "` + defaultPub() + `"
}`
	case "/v1/history/get_controlled_accounts":
		return `{
  "controlling_account": "` + defaultActor() + `"
}`
	case "/v1/history/get_transaction":
		return `{
  "id": "1111111111111111111111111111111111111111111111111111111111111111"
}`
	case "/v1/chain/get_scheduled_transactions":
		return `{"limit":1, "json": true}`
	// TODO: update *_whitelist when API is defined.
	case "/v1/chain/check_whitelist":
		return `{"fio_address":"` + defaultAddress() + `"}`
	case "/v1/chain/get_whitelist":
		return `{"fio_address":"` + defaultAddress() + `"}`

	case "/v1/chain/avail_check":
		return `{
  "fio_name": "` + defaultAddress() + `"
}`
	case "/v1/chain/get_abi":
		return `{
  "account_name": "fio.address"
}`
	case "/v1/chain/get_account":
		return `{
  "account_name": "` + defaultActor() + `"
}`
	case "/v1/chain/get_block":
		return `{
  "block_num_or_id": "123"
}`
	case "/v1/chain/get_block_header_state":
		return `{
  "block_num_or_id": "123"
}`
	case "/v1/chain/get_currency_balance":
		return `{
  "account": "` + defaultActor() + `",
  "code": "fio.token",
  "symbol": "FIO"
}`
	case "/v1/chain/get_currency_stats":
		return `{
  "json": false,
  "code": "fio.token",
  "symbol": "FIO"
}`
	case "/v1/chain/get_fee":
		return `{
  "end_point": "add_pub_address",
  "fio_address": "` + defaultAddress() + `"
}`
	case "/v1/chain/get_fio_balance":
		return `{
  "fio_public_key": "` + defaultPub() + `"
}`
	case "/v1/chain/get_actor":
		return `{
  "fio_public_key": "` + defaultPub() + `"
}`
	case "/v1/chain/get_fio_addresses":
		return `{
  "fio_public_key": "` + defaultPub() + `"
}`
	case "/v1/chain/get_fio_domains":
		return `{
  "fio_public_key": "` + defaultPub() + `"
}`
	case "/v1/chain/get_fio_names":
		return `{
  "fio_public_key": "` + defaultPub() + `"
}`
	case "/v1/chain/get_obt_data":
		return `{
  "fio_public_key": "` + defaultPub() + `",
  "limit": 100,
  "offset": 0
}`
	case "/v1/chain/get_pending_fio_requests":
		return `{
  "fio_public_key": "` + defaultPub() + `",
  "limit": 100,
  "offset": 0
}`
	case "/v1/chain/get_cancelled_fio_requests":
		return `{
  "fio_public_key": "` + defaultPub() + `",
  "limit": 100,
  "offset": 0
}`
	case "/v1/chain/get_raw_abi":
		return `{
  "account_name": "fio.token"
}`
	case "/v1/chain/get_pub_address":
		return `{
  "fio_address": "` + defaultAddress() + `",
  "token_code": "FIO",