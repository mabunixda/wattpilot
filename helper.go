package wattpilot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func hasKey(data map[string]interface{}, key string) bool {
	_, isKnown := data[key]
	return isKnown
}

func sha256sum(data string) string {
	bs := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", bs)
}
func randomHexString(n int) string {
	b := make([]byte, (n+2)/2) // can be simplified to n/2 if n is always even

	if _, err := randomSource.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[1 : n+1]
}
