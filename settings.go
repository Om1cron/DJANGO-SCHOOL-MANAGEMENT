
package cryptonym

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	errs "github.com/blockpane/cryptonym/errLog"
	"golang.org/x/crypto/pbkdf2"
	"os"
	"time"
)

const (
	settingsDir      = "com.blockpane.cryptonym"
	settingsFileName = "cryptonym.dat"
)

type FioSettings struct {
	Server string `json:"server"`
	Proxy  string `json:"proxy"`

	DefaultKey     string `json:"default_key"`
	DefaultKeyDesc string `json:"default_key_desc"`
	FavKey2        string `json:"fav_key_2"`
	FavKey2Desc    string `json:"fav_key_2_desc"`
	FavKey3        string `json:"fav_key_3"`
	FavKey3Desc    string `json:"fav_key_3_desc"`
	FavKey4        string `json:"fav_key_4"`
	FavKey4Desc    string `json:"fav_key_4_desc"`

	MsigAccount string `json:"msig_account"`
	Tpid        string `json:"tpid"`

	AdvancedFeatures bool `json:"advanced_features"`

	// future:
	KeosdAddress  string `json:"keosd_address"`
	KeosdPassword string `json:"keosd_password"`
}

func DefaultSettings() *FioSettings {
	return &FioSettings{
		Server:         "http://127.0.0.1:8888",
		Proxy:          "http://127.0.0.1:8080",
		DefaultKey:     "5JBbUG5SDpLWxvBKihMeXLENinUzdNKNeozLas23Mj6ZNhz3hLS",
		DefaultKeyDesc: "devnet - vote 1",
		FavKey2:        "5KC6Edd4BcKTLnRuGj2c8TRT9oLuuXLd3ZuCGxM9iNngc3D8S93",
		FavKey2Desc:    "devnet - vote 2",
		FavKey3:        "5KQ6f9ZgUtagD3LZ4wcMKhhvK9qy4BuwL3L1pkm6E2v62HCne2R",
		FavKey3Desc:    "devnet - bp1",
		FavKey4:        "5HwvMtAEd7kwDPtKhZrwA41eRMdFH5AaBKPRim6KxkTXcg5M9L5",
		FavKey4Desc:    "devnet - locked 1",
		MsigAccount:    "eosio",
		Tpid:           "tpid@blockpane",
	}
}

func EncryptSettings(set *FioSettings, salt []byte, password string) (encrypted []byte, err error) {
	if password == "" {
		return nil, errors.New("invalid password supplied")
	}

	// if a salt isn't supplied, create one, note: using crypto/rand NOT math/rand, has better entropy
	if salt == nil || len(salt) != 12 || bytes.Equal(salt, bytes.Repeat([]byte{0}, 12)) {
		salt = make([]byte, 12)
		if _, e := rand.Read(salt); e != nil {
			errs.ErrChan <- "EncryptSettings: " + e.Error()
			return nil, err
		}
	}

	// prepend the salt to our buffer:
	crypted := bytes.NewBuffer(nil)
	crypted.Write(salt)

	// convert our settings to a binary struct
	data := bytes.NewBuffer(nil)
	g := gob.NewEncoder(data)