package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func randomHexString(n int) string {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {

		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return strings.ToUpper(hex.EncodeToString(b))
}

func GenerateBorrowingCode() string {
	borrowingCode := fmt.Sprintf("BORROWING/%s/%s", time.Now().Format("20060102"), randomHexString(3))

	return borrowingCode
}
