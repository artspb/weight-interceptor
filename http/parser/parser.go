package parser

import (
	"fmt"
	"strconv"
	"time"
	"weight-interceptor-http/dataservice/hash"
	"weight-interceptor-http/dataservice/ntp"
)

const (
	length = 68
	code   = "24"
)

type Weight struct {
	Uid    string
	Time   time.Time
	Weight int
}

func ParseData(request string) (Weight, error) {
	if len(request) != length {
		return Weight{}, fmt.Errorf("expected length is %d, got %d", length, len(request))
	}
	if code != request[0:2] {
		return Weight{}, fmt.Errorf("expected code is %s, got %s", code, request[0:2])
	}
	crc, err := hash.Checksum(request[0:60])
	if err != nil {
		return Weight{}, fmt.Errorf("unable to compute checksun: %w", err)
	}
	if crc != request[60:68] {
		return Weight{}, fmt.Errorf("expected crc is %s, got %s", crc, request[60:68])
	}

	uid := request[2:26]
	data := request[26:42]
	hex := data[4:12]
	t, err := ntp.FromHex(hex)
	if err != nil {
		return Weight{}, fmt.Errorf("unable to parse time: %w", err)
	}
	w := data[12:16]
	i, err := strconv.ParseInt(w, 16, 16)
	if err != nil {
		return Weight{}, fmt.Errorf("unable to parse weight: %w", err)
	}

	return Weight{uid, t, int(i)}, nil
}
