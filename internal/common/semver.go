package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func MajorVersion(v string) (int, error) {
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return 0, errors.New("invalid version format")
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid major version format: %w", err)
	}

	return major, nil
}
