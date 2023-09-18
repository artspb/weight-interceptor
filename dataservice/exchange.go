package dataservice

import (
	"fmt"
	"log"
	"time"
	"weight-interceptor-http/dataservice/hash"
	"weight-interceptor-http/dataservice/ntp"
)

const (
	ok            = "A00000000000000001000000000000000000000000000000bec650a1"
	syncOk        = "A5000000000000000100000000000000000000000000000056e5abd9"
	terminationOk = "A20000000000000000000000000000000000000000000000c9950d3f"
)

type responder interface {
	respond() string
}

type syncRequest struct{}

func (syncRequest) respond() string {
	return ok
}

type sync struct{}

func (sync) respond() string {
	return syncOk
}

type syncResponse struct{}

func (syncResponse) respond() string {
	now := time.Now()
	message := fmt.Sprintf("A100000000000000%s000000000000000000000000", ntp.ToHex(now))
	checksum, err := hash.Checksum(message)
	if err != nil {
		log.Println(err)
	}
	return message + checksum
}

type dataTransmission struct{}

func (dataTransmission) respond() string {
	return ok
}

type terminationRequest struct{}

func (terminationRequest) respond() string {
	return terminationOk
}

type termination struct{}

func (termination) respond() string {
	return ""
}
