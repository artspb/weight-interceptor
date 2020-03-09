package dataservice

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
	"weight-interceptor-http/dns"
	"weight-interceptor-http/storage"
)

func StartService() {
	defer func() {
		dns.Shutdown()
		err := storage.Shutdown()
		if err != nil {
			log.Fatal(err)
		}
	}()
	http.HandleFunc("/devicedataservice/dataservice", dataService)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func dataService(writer http.ResponseWriter, request *http.Request) {
	timestamp, err := extractData(request)
	response, err := findResponse(request.URL, timestamp)
	_, err = writer.Write(response)
	if err != nil {
		http.Error(writer, "", http.StatusInternalServerError)
	}
}

func extractData(request *http.Request) (string, error) {
	data, ok := request.URL.Query()["data"]
	if !ok || len(data) != 1 {
		return "", errors.New("invalid data in request")
	}
	timestamp := time.Now().Format(time.RFC3339)
	return timestamp, storage.Request(timestamp, []byte(data[0]))
}

func findResponse(url *url.URL, timestamp string) ([]byte, error) {
	host := url.Hostname()
	ip := dns.Lookup(host)
	client := &http.Client{}
	uri := fmt.Sprintf("http://%s%s", ip.String(), url.RequestURI())
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	request.Host = host
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	code := response.StatusCode
	if code != http.StatusOK {
		panic(fmt.Sprintf("invalid response code: %d", code))
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	err = storage.Response(timestamp, data)
	return data, err
}
