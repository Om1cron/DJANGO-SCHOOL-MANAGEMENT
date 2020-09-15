
package fuzzer

import (
	"encoding/hex"
	"fmt"
	"github.com/fioprotocol/fio-go"
	"math/rand"
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := RandomString(17)
		fmt.Println(s)
		if len(s) < 16 {
			t.Error("too short")
		}
	}
}

func TestRandomActor(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := RandomActor()
		fmt.Println(s)
		if len(s) != 12 {
			t.Error("wrong size")
		}
	}
}

func TestRandomFioPubKey(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := RandomFioPubKey()
		fmt.Println(s)
		if len(s) != 53 {
			t.Error("wrong size")
		}
	}
}

func TestRandomBytes(t *testing.T) {
	for _, enc := range []int8{EncodeBase64, EncodeHexString, EncodeRaw} {
		for i := 0; i < 10; i++ {
			s := RandomBytes(rand.Intn(112)+16, enc)
			fmt.Println(s)
			if s == "" {
				t.Error("empty!")
			}
		}
	}
}

func TestRandomChecksum(t *testing.T) {
	for i := 0; i < 10; i++ {
		s := RandomChecksum()
		fmt.Println(s)
		if len(s) != 64 {
			t.Error("wrong size for checksum256")
		}
	}
}
