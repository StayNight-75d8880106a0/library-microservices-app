package helper

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmptyInput         = errors.New("The phone number cannot be empty")
	ErrInvalidCountryCode = errors.New("The phone number must begin with the country code '62'")
	ErrInvalidFormat      = errors.New("The phone number may only contain numbers after sanitisation")
	ErrInvalidLength      = errors.New("The phone number is formally invalidated by its non-compliance with the requisite 10–15 digit cardinality threshold, thereby violating the fundamental syntactic constraint of the validation protocol.")
	ErrInvalidProvider    = errors.New("Indonesian mobile numbers must be preceded by the digit 8 after the code 62 (e.g. 628xxx)")
)

var indonesianPhoneRegex = regexp.MustCompile(`^628[1-9][0-9]{7,11}$`)

func ValidateIndonesianPhoneNumber(phone string) error {

	if phone == "" {
		return ErrEmptyInput
	}

	cleaned := strings.TrimSpace(phone)
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	if strings.HasPrefix(cleaned, "+62") {
		cleaned = cleaned[1:]
	}

	if cleaned == "" {
		return ErrEmptyInput
	}

	if !strings.HasPrefix(cleaned, "62") {
		return ErrInvalidCountryCode
	}

	if !strings.HasPrefix(cleaned, "628") {
		return ErrInvalidProvider
	}

	if len(cleaned) < 10 || len(cleaned) > 15 {
		return ErrInvalidLength
	}

	if !indonesianPhoneRegex.MatchString(cleaned) {
		return ErrInvalidFormat
	}

	return nil
}
