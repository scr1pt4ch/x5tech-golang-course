package httpapi

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type page struct {
	limit  int
	offset int
}

func parsePage(rawQuery string) (page, error) {
	p := page{limit: defaultLimit, offset: 0}

	values, err := url.ParseQuery(rawQuery)
	if nil != err {
		return page{}, errors.New("query string is malformed")
	}

	for name, list := range values {
		if 1 < len(list) {
			return page{}, fmt.Errorf("parameter %q is repeated", name)
		}

		switch name {
		case "limit":
			n, ok := parseDigits(list[0])
			if !ok || n < 1 || maxLimit < n {
				return page{}, fmt.Errorf("limit must be an integer from 1 to %d", maxLimit)
			}
			p.limit = n
		case "offset":
			n, ok := parseDigits(list[0])
			if !ok {
				return page{}, errors.New("offset must be a non-negative integer")
			}
			p.offset = n
		default:
			return page{}, fmt.Errorf("unknown parameter %q", name)
		}
	}

	return p, nil
}

func parseDigits(value string) (int, bool) {
	if "" == value {
		return 0, false
	}
	for _, ch := range value {
		if ch < '0' || '9' < ch {
			return 0, false
		}
	}

	n, err := strconv.Atoi(value)
	if nil != err {
		return 0, false
	}
	return n, true
}
