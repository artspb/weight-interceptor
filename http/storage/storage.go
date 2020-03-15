package storage

import (
	"log"
	"os"
)

var data *os.File

func init() {
	err := os.MkdirAll("data", 0600)
	if err != nil {
		log.Fatal(err)
	}
	data, err = os.OpenFile("data/data.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0700)
	if err != nil {
		log.Fatal(err)
	}
}

func Request(request []byte) error {
	_, err := data.Write(request)
	if err != nil {
		return err
	}
	_, err = data.WriteString("\n")
	return err
}

func Shutdown() error {
	return data.Close()
}
