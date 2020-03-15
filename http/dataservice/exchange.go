package dataservice

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	"weight-interceptor-http/dataservice/hash"
	"weight-interceptor-http/dataservice/ntp"
	"weight-interceptor-http/dns"
)

const (
	ok            = "A00000000000000001000000000000000000000000000000bec650a1"
	syncOk        = "A5000000000000000100000000000000000000000000000056e5abd9"
	terminationOk = "A20000000000000000000000000000000000000000000000c9950d3f"
)

type responder interface {
	respond(url *url.URL) string
}

type syncRequest struct{}

func (syncRequest) respond(url *url.URL) string {
	return askServerAndCompare(url, ok)
}

type sync struct{}

func (sync) respond(url *url.URL) string {
	return askServerAndCompare(url, syncOk)
}

type syncResponse struct{}

func (syncResponse) respond(url *url.URL) string {
	return askServerAndCompareTime(url)
}

func askServerAndCompareTime(url *url.URL) string {
	now := time.Now()
	message := fmt.Sprintf("A100000000000000%s000000000000000000000000", ntp.ToHex(now))
	checksum, err := hash.Checksum(message)
	if err != nil {
		log.Println(err)
	}
	expected := message + checksum
	bytes, err := askServer(url)
	if err != nil {
		fmt.Println(err)
		return expected
	}

	response := string(bytes)
	server, err := ntp.FromHex(response[16:24])
	if err != nil {
		fmt.Println(err)
		return response
	}

	if server.Sub(now) > 1*time.Hour {
		fmt.Printf("Server's time differs from local by more than an hour: %s - %s\n", server.String(), now.String())
	}
	return response
}

type dataTransmission struct{}

func (dataTransmission) respond(url *url.URL) string {
	return askServerAndCompare(url, ok)
}

type terminationRequest struct{}

func (terminationRequest) respond(url *url.URL) string {
	return askServerAndCompare(url, terminationOk)
}

type termination struct{}

func (termination) respond(url *url.URL) string {
	return askServerAndCompare(url, "")
}

type unknown struct{}

func (unknown) respond(url *url.URL) string {
	response, err := askServer(url)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(response)
}

func askServerAndCompare(url *url.URL, expected string) string {
	bytes, err := askServer(url)
	if err != nil {
		fmt.Println(err)
		return expected
	}

	response := string(bytes)
	if response != expected {
		fmt.Printf("Server's response doesn't match expected: %s - %s\n", response, ok)
	}
	return response
}

func askServer(url *url.URL) ([]byte, error) {
	host := url.Hostname()
	ip := dns.Lookup(host)
	client := &http.Client{}
	uri := fmt.Sprintf("http://%s%s", ip, url.RequestURI())
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	request.Host = host
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	code := response.StatusCode
	if code != http.StatusOK {
		return nil, fmt.Errorf("invalid response code: %d", code)
	}

	return ioutil.ReadAll(response.Body)
}
