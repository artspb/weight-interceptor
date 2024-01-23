package dataprocessor

import (
	"errors"
	"fmt"
	"log"
	"weight-interceptor-http/dataprocessor/fit"
	"weight-interceptor-http/parser"
	"weight-interceptor-http/storage"
)

type Processor interface {
	IsAvailable() bool
	Process(data string) error
}

type RawProcessor struct{}

func (RawProcessor) IsAvailable() bool {
	return true
}

func (RawProcessor) Process(data string) error {
	err := storage.Request([]byte(data))
	if err != nil {
		log.Printf("Unable to store data: %v\n", err)
	}
	return nil
}

type FitProcessor struct{}

func (FitProcessor) IsAvailable() bool {
	return fit.IsAvailable()
}

func (FitProcessor) Process(data string) error {
	defer RetryAll()

	weight, err := parser.ParseData(data)
	if err != nil {
		log.Printf("Unable to parse data: %v\n", err)
		return nil
	}
	err = fit.AddWeight(weight)
	if err != nil && !errors.Is(err, storage.UnknownUser) {
		log.Printf("Unable to submit data: %v\n", err)
		err := storage.AddRetry([]byte(data))
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func RetryAll() {
	success, failure, err := storage.RetryAll(func(request string) error {
		weight, err := parser.ParseData(request)
		if err != nil {
			log.Printf("Unable to parse data during retry: %v\n", err)
			return err
		}
		err = fit.AddWeight(weight)
		if err != nil {
			log.Printf("Unable to submit data during retry: %v\n", err)
		}
		return err
	})
	if err != nil {
		log.Printf("Retry failed (%d sent, %d postponed): %v\n", success, failure, err)
	} else if success != 0 || failure != 0 {
		fmt.Printf("Retry has been finished successfully (%d sent, %d postponed)\n", success, failure)
	}
}

type CsvProcessor struct{}

func (CsvProcessor) IsAvailable() bool {
	return true
}

func (CsvProcessor) Process(data string) error {
	weight, err := parser.ParseData(data)
	if err != nil {
		log.Printf("Unable to parse data: %v\n", err)
		return nil
	}
	err = storage.StoreWeightToCsv(weight)
	if err != nil {
		log.Printf("Unable to store data: %v\n", err)
	}
	return nil
}
