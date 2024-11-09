package utils

import (
	"fmt"
	"godin/pkg/godinerrors"
	"regexp"
)

func ToPtr[T any](input T) *T {
	return &input
}

func ValidateColumns(state map[string]interface{}, columns []string) error {
	missingColumns := []string{}
	for _, c := range columns {
		if _, ok := state[c]; !ok {
			missingColumns = append(missingColumns, c)
		}
	}
	if len(missingColumns) != 0 {
		return fmt.Errorf("error validating columns, missing columns: %v", missingColumns)
	}
	return nil
}

func IsMissingColumnError(err error) bool {
	readError, ok := err.(godinerrors.ReadError)
	if !ok {
		return false
	}
	return readError.Code == godinerrors.MissingColumnError
}

func ExtractSteamId(action string) (string, error) {
	r := regexp.MustCompile(`(Got connection SteamID|Closing socket) (\d{17})`)
	steamidbytes := r.FindAllStringSubmatch(action, 999)
	if steamidbytes == nil {
		return "", fmt.Errorf("error finding steam id")
	}
	return string(steamidbytes[0][2]), nil
}
