package wattpilot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"runtime"

	"time"
)

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
	randomSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	if _, err := randomSource.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[1 : n+1]
}

// func getCurrentFunctionName() string {
// 	// Skip GetCurrentFunctionName
// 	return getFrame(1).Function
// }

func getCallerFunctionName() string {
	// Skip GetCallerFunctionName and the function to get the caller of
	return getFrame(2).Function
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
