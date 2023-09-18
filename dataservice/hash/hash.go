package hash

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

func Checksum(hex string) (string, error) {
	var bytes []byte
	for i := 0; i < len(hex); i += 2 {
		result, err := strconv.ParseInt(hex[i:i+2], 16, 16)
		if err != nil {
			return "00000000", err
		}
		bytes = append(bytes, byte(result))
	}
	checksum := crc32.ChecksumIEEE(bytes)
	return fmt.Sprintf("%08x", int64(checksum)), nil
}
