package ntp

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const baseline = 1262300400

func FromHex(hex string) (time.Time, error) {
	if len(hex) != 8 {
		return time.Time{}, errors.New("invalid string length, must be 8")
	}
	dec, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(baseline+dec, 0), nil
}

func ToHex(t time.Time) string {
	dec := t.Unix() - baseline
	return fmt.Sprintf("%08x", dec)
}

func calcBaseline() int64 {
	baseline, err := time.Parse(time.RFC3339, "2010-01-01T00:00:00+01:00")
	if err != nil {
		panic(err)
	}
	return baseline.Unix()
}
