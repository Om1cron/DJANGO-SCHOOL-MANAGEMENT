
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