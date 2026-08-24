package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GenerateLibraryCardNumber() string {
	libraryCardNumber := fmt.Sprintf("LIBRARY-%s-%s", time.Now().Format("20060102"), RandomHexString(7))
	return libraryCardNumber
}

func RandomHexString(n int) string {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {

		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return strings.ToUpper(hex.EncodeToString(b))
}
