
package fuzzer

import (
	"encoding/json"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math/rand"
	"sort"
	"strings"
)

func MaxRandomFioAddressAt(domain string) string {
	if !strings.HasPrefix(domain, "@") {
		domain = "@" + domain
	}
	return RandomString(64-len(domain)) + domain
}

func MaxRandomFioDomain() string {
	return RandomString(62)
}

func MaxAddPubAddress() string {
	addresses := make([]fio.TokenPubAddr, 5)
	for i := range addresses {
		addresses[i].ChainCode = RandomString(10)
		addresses[i].TokenCode = RandomString(10)
		addresses[i].PublicAddress = RandomString(128)
	}
	j, _ := json.Marshal(addresses)
	return string(j)
}

func MaxNewFundsContent() string {
	// end up with 96 bytes of encoding overhead:
	return RandomBytes(296-96, EncodeBase64).(string)
}

func MaxRecObtContent() string {
	// end up with 144 bytes of encoding overhead:
	return RandomBytes(432-144, EncodeBase64).(string)
}

func MaxProducerUrl() string {
	return "http://" + RandomString(500) + ".com"
}

func MaxVoteProducers(url string) interface{} {
	errs.ErrChan <- "for this to be effective, you may need to register a lot of producers with long names"
	errs.ErrChan <- "querying producers table to find up to 30 producers, with the longest fio addresses."
	api, _, err := fio.NewConnection(nil, url)
	if err != nil {
		errs.ErrChan <- err.Error()
		return []string{""}
	}
	api.Header.Set("User-Agent", "fio-cryptonym-wallet")
	producers, err := api.GetFioProducers()
	if err != nil {
		errs.ErrChan <- err.Error()
		return []string{""}
	}
	bpFioNames := make([]string, 0)
	for _, fioAddress := range producers.Producers {
		bpFioNames = append(bpFioNames, string(fioAddress.FioAddress))
	}

	sort.Slice(bpFioNames, func(i, j int) bool {
		return len(bpFioNames[i]) > len(bpFioNames[j])
	})
	if len(producers.Producers) < 30 {
		errs.ErrChan <- fmt.Sprintf("only found %d producers", len(producers.Producers))
		return bpFioNames
	}
	return bpFioNames[:30]
}

type fioNamesResp struct {
	Name string `json:"name"`
}

func RandomExistingFioAddress(url string) string {
	api, _, err := fio.NewConnection(nil, url)
	if err != nil {
		errs.ErrChan <- err.Error()
		return ""
	}
	api.Header.Set("User-Agent", "fio-cryptonym-wallet")