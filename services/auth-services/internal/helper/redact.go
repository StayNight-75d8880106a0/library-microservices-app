package helper

import (
	"encoding/json"
	"strings"
)

const RedactedValue = "***!REDACTED!***"

var sensitiveKeys = []string{
	"password",
	"phoneNumber",
	"password",
	"refreshToken",
	"accessToken",
	"tokenType",
	"accessTokenExpiresIn",
	"refreshTokenExpiresIn",
}

func RedactSensitiveData(raw []byte) []byte {

	if len(raw) == 0 || len(raw) > MaxAuditBodySize || !json.Valid(raw) {
		return raw
	}

	var parsed interface{}

	errUnmarshal := json.Unmarshal(raw, &parsed)

	if errUnmarshal != nil {
		return raw
	}

	result, errMarshal := json.Marshal(redactValue(parsed))

	if errMarshal != nil {
		return raw
	}

	return result

}

func redactValue(value interface{}) interface{} {

	switch data := value.(type) {

	case map[string]interface{}:
		for key, item := range data {
			if isSensitiveKey(key) {
				data[key] = RedactedValue
				continue
			}
			data[key] = redactValue(item)
		}
		return data

	case []interface{}:
		for index, item := range data {
			data[index] = redactValue(item)
		}
		return data

	default:
		return value

	}

}

func isSensitiveKey(key string) bool {

	for _, sensitive := range sensitiveKeys {
		if strings.EqualFold(key, sensitive) {
			return true
		}
	}

	return false

}
