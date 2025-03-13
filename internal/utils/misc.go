package utils

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetCpuCount() int {
	return 0
}

// parseHexByte converts a hex string to a byte
func ParseHexByte(s string) (byte, error) {
	val, err := strconv.ParseUint(s, 16, 8)

	if err != nil {
		return 0, fmt.Errorf("invalid byte: %s", s)
	}

	return byte(val), nil
}

func HexadecimalsToBytes(s string) ([]byte, error) {
	var ret []byte

	parts := strings.Fields(s)

	for k, v := range parts {
		if len(v) != 2 {
			return ret, fmt.Errorf("hexadecimal at index %d isn't valid", k)
		}

		b, err := hex.DecodeString(v)

		if err != nil {
			return ret, err
		}

		ret = append(ret, b[0])
	}

	return ret, nil
}

func GetRandInt(min int, max int, rng *rand.Rand) int {
	return rng.Intn(max - min + 1)
}

func GenRandBytes(min int, max int, rng *rand.Rand) []byte {
	n := min + rng.Intn(max-min+1)
	b := make([]byte, n)

	for i := range b {
		b[i] = byte(rng.Intn(256))
	}

	return b
}

func GenRandBytesSingle(length int, rng *rand.Rand) []byte {
	b := make([]byte, length)

	for i := range b {
		b[i] = byte(rng.Intn(256))
	}

	return b
}

func ReadFileAndStoreBytes(p string) ([]byte, error) {
	ret, err := os.ReadFile(p)

	return ret, err
}

func SleepMicro(t uint64) {
	if t > 0 {
		time.Sleep(time.Duration(t) * time.Microsecond)
	}
}
