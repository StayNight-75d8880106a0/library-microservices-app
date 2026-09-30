package helper

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

func EncodeCursor(occurredAt time.Time, ID string) string {

	raw := occurredAt.UTC().Format(time.RFC3339Nano) + "|" + ID

	result := base64.URLEncoding.EncodeToString([]byte(raw))

	return result

}

func DecodeCursor(cursor string) (*time.Time, *string, error) {

	if cursor == "" {
		return nil, nil, nil
	}

	decoded, errDecode := base64.URLEncoding.DecodeString(cursor)

	if errDecode != nil {
		return nil, nil, errDecode
	}

	parts := strings.Split(string(decoded), "|")

	if len(parts) != 2 {
		return nil, nil, errors.New("cursor format is not valid")
	}

	occurredAt, errParse := time.Parse(time.RFC3339Nano, parts[0])

	if errParse != nil {
		return nil, nil, errParse
	}

	ID := parts[1]

	return &occurredAt, &ID, nil

}
