package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/crumbyte/noxdir/drive"
)

func parseSizeLimit() (drive.FileInfoFilter, error) {
	sizeLimit = strings.TrimSpace(sizeLimit)

	// We don't want to include an empty filter that still will be checked for
	// each entry.
	//nolint:nilnil // I'm ok with that
	if len(sizeLimit) == 0 {
		return nil, nil
	}

	limits := strings.Split(sizeLimit, ":")
	if len(limits) != 2 {
		return nil, fmt.Errorf("check the usage example: %s", sizeLimit)
	}

	multiplier := map[string]int{"pb": 40, "tb": 30, "gb": 20, "mb": 10, "kb": 0}

	// parse a single size raw value. If the part is empty a 0 limit will be
	// returned.
	readLimit := func(rawValue string) (int64, error) {
		if rawValue = strings.ToLower(rawValue); len(rawValue) == 0 {
			return 0, nil
		}

		if len(rawValue) < 3 {
			return 0, fmt.Errorf("invalid size-limit value: %s", rawValue)
		}

		minLimit, err := strconv.ParseInt(rawValue[:len(rawValue)-2], 10, 64)
		if err != nil {
			return 0, errors.New("unknown size unit")
		}

		offset, ok := multiplier[rawValue[len(rawValue)-2:]]
		if !ok {
			return 0, errors.New("unknown size unit")
		}

		return minLimit * 1024 << offset, nil
	}

	minLimit, err := readLimit(limits[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse min limit: %w", err)
	}

	maxLimit, err := readLimit(limits[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse max limit: %w", err)
	}

	if maxLimit != 0 && minLimit > maxLimit {
		return nil, errors.New("min value is bigger than max value")
	}

	return drive.NewSizeFilter(minLimit, maxLimit).Filter, nil
}
