package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type StatusCodeRange struct {
	From uint16
	To   uint16
}

func CheckStatusCode(status uint16, acceptedPatterns []string) (bool, error) {
	if len(acceptedPatterns) == 0 {
		return false, errors.New("no accepted status codes provided")
	}

	for _, pattern := range acceptedPatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		from, to, err := parsePattern(pattern)
		if err != nil {
			return false, err
		}

		if status >= from && status <= to {
			return true, nil
		}
	}
	return false, nil
}

func ParseStatusCode(acceptedPatterns []string) ([]StatusCodeRange, error) {
	if len(acceptedPatterns) == 0 {
		return nil, errors.New("no accepted status codes provided")
	}

	var ranges []StatusCodeRange

	for _, pattern := range acceptedPatterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		from, to, err := parsePattern(pattern)
		if err != nil {
			return nil, err
		}

		ranges = append(ranges, StatusCodeRange{From: from, To: to})
	}

	return ranges, nil
}

func parsePattern(pattern string) (from, to uint16, err error) {
	if len(pattern) == 0 {
		return 0, 0, errors.New("empty status code pattern")
	}

	switch {
	case strings.ContainsRune(pattern, '-'):
		return parseRangePattern(pattern)
	case hasWildcard(pattern):
		return expandWildcardPattern(pattern)
	default:
		code, err := parseExactCode(pattern)
		return code, code, err
	}
}

func parseRangePattern(s string) (uint16, uint16, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range: must have exactly one '-'")
	}

	from, err1 := parseExactCode(parts[0])
	to, err2 := parseExactCode(parts[1])

	if err1 != nil {
		return 0, 0, err1
	}
	if err2 != nil {
		return 0, 0, err2
	}
	if from > to {
		return 0, 0, errors.New("range start must not exceed end")
	}

	return from, to, nil
}

func hasWildcard(s string) bool {
	return strings.ContainsAny(s, "xX*")
}

func expandWildcardPattern(s string) (uint16, uint16, error) {
	if len(s) != 3 {
		return 0, 0, fmt.Errorf("wildcard pattern must be 3 characters (got %q)", s)
	}

	s = strings.ToLower(s)
	var fromBuilder, toBuilder strings.Builder
	fromBuilder.Grow(3)
	toBuilder.Grow(3)

	for _, r := range s {
		switch r {
		case 'x', '*':
			fromBuilder.WriteRune('0')
			toBuilder.WriteRune('9')
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			fromBuilder.WriteRune(r)
			toBuilder.WriteRune(r)
		default:
			return 0, 0, fmt.Errorf("invalid character in wildcard pattern: %c", r)
		}
	}

	fromStr := fromBuilder.String()
	toStr := toBuilder.String()

	from, _ := strconv.ParseUint(fromStr, 10, 16)
	to, _ := strconv.ParseUint(toStr, 10, 16)

	return uint16(from), uint16(to), nil
}

func parseExactCode(s string) (uint16, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty status code")
	}

	val, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid status code %q: %w", s, err)
	}

	if val < 100 || val > 599 {
		return 0, fmt.Errorf("status code %d is outside typical HTTP range (100-599)", val)
	}

	return uint16(val), nil
}
