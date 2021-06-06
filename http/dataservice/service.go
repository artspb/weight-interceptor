package dataservice

import (
	"fmt"
	"log"
	"net/http"
	"weight-interceptor-http/dataprocessor"
	"weight-interceptor-http/dataservice/hash"
	"weight-interceptor-http/storage"
)

var responders = map[string]responder{
	"25": syncRequest{},
	"28": sync{},
	"21": syncResponse{},
	"24": dataTransmission{},
	"22": terminationRequest{},
	"29": termination{},
}

var processors = [2]dataprocessor.Processor{
	dataprocessor.FitProcessor{},
	dataprocessor.RawProcessor{},
}

func StartService() {
	go dataprocessor.RetryAll()
	defer func() {
		err := storage.Shutdown()
		if err != nil {
			log.Fatal(err)
		}
	}()
	fmt.Println("Starting service...")
	http.HandleFunc("/devicedataservice/dataservice", dataService)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func dataService(writer http.ResponseWriter, request *http.Request) {
	data := extractData(request)
	checksum, err := hash.Checksum(data[:len(data)-8])
	if err != nil {
		log.Println(err)
	}
	if checksum != data[len(data)-8:] {
		log.Printf("Invalid CRC32 of %s: expected %s but got %s\n", data[:len(data)-8], checksum, data[len(data)-8:])
	}
	code := ""
	if len(data) > 1 {
		code = data[:2]
	}
	if code == "24" {
		go func() {
			for _, processor := range processors {
				if err := processor.Process(data); err != nil {
					log.Printf("Stopped processing due to error: %v\n", err)
					break
				}
			}
		}()
	}
	responder := responders[code]
	if responder == nil {
		log.Printf("Uknown request code: %s\n", request.URL.RequestURI())
		responder = responders[""]
	}
	response := responder.respond()
	_, err = writer.Write([]byte(response))
	if err != nil {
		log.Println(err)
	}
}

func extractData(request *http.Request) string {
	data, ok := request.URL.Query()["data"]
	if !ok || len(data) != 1 {
		return ""
	}
	return data[0]
}
