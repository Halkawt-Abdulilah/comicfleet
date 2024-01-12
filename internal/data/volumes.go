package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidVolumeFormat = errors.New("invalid format for field 'Volumes'")

type Volumes int32

func (v Volumes) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d Volumes", v)

	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

func (v *Volumes) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidVolumeFormat
	}

	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[1] != "Volumes" {
		return ErrInvalidVolumeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidVolumeFormat
	}

	*v = Volumes(i)

	return nil
}
