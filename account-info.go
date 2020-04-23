
package cryptonym

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type AccountInformation struct {
	*sync.Mutex

	Actor      string   `json:"actor"`
	PubKey     string   `json:"pub_key"`
	PrivKey    string   `json:"priv_key"`
	Balance    int64    `json:"balance"`
	BundleCred int      `json:"bundle_cred"`
	MsigOwners []string `json:"msig_owners"`
	MsigThresh uint32   `json:"msig_thresh"`
	RamUsed    int64    `json:"ram_used"`
	fioNames   []string
	FioNames   []FioAddressStruct `json:"fio_names"`
	fioDomains []string
	FioDomains []FioDomainStruct `json:"fio_domains"`
	PublicKeys []AddressesList   `json:"public_keys"`
	api        *fio.API
	Producer   *ProducerInfo `json:"producer"`
}

type FioAddressStruct struct {
	Id           int             `json:"id"`
	Name         string          `json:"name"`
	NameHash     string          `json:"namehash"`
	Domain       string          `json:"domain"`
	DomainHash   string          `json:"domainhash"`
	Expiration   int64           `json:"expiration"`
	OwnerAccount string          `json:"owner_account"`
	Addresses    []AddressesList `json:"addresses"`
	BundleCount  uint64          `json:"bundleeligiblecountdown"`
}

type FioDomainStruct struct {
	Name       string          `json:"name"`
	IsPublic   uint8           `json:"is_public"`
	Expiration int64           `json:"expiration"`
	Account    eos.AccountName `json:"account"`
}

type AddressesList struct {
	TokenCode     string `json:"token_code"`
	ChainCode     string `json:"chain_code"`
	PublicAddress string `json:"public_address"`
}

type ProducerInfo struct {
	Owner             string    `json:"owner"`
	FioAddress        string    `json:"fio_address"`
	TotalVotes        float64   `json:"total_votes"`
	ProducerPublicKey string    `json:"producer_public_key"`
	IsActive          bool      `json:"is_active"`
	Url               string    `json:"url"`
	UnpaidBlocks      int       `json:"unpaid_blocks"`
	LastClaimTime     time.Time `json:"last_claim_time"`
	Location          int       `json:"location"`
}

var bpLocationMux sync.RWMutex
var bpLocationMap = map[int]string{
	10: "East Asia",
	20: "Australia",
	30: "West Asia",
	40: "Africa",
	50: "Europe",
	60: "East North America",
	70: "South America",
	80: "West North America",
}

var accountSearchType = []string{
	"Public Key",
	"Fio Address",
	"Private Key",
	"Actor/Account",
	"Fio Domain", // TODO: how is index derived on fio.address domains table?
}

func GetLocation(i int) string {
	bpLocationMux.RLock()
	defer bpLocationMux.RUnlock()
	loc := bpLocationMap[i]
	if loc == "" {
		return "Invalid Location"
	}
	return loc
}

func AccountSearch(searchFor string, value string) (as *AccountInformation, err error) {
	as = &AccountInformation{}
	as.api, _, err = fio.NewConnection(nil, Uri)
	if err != nil {
		return nil, err
	}
	as.api.Header.Set("User-Agent", "fio-cryptonym-wallet")
	switch searchFor {
	case "Actor/Account":
		return as, as.searchForActor(value)
	case "Public Key":