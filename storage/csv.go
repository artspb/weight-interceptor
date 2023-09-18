package storage

import (
	"fmt"
	"log"
	"os"
	"weight-interceptor-http/parser"
)

func StoreWeightToCsv(weight parser.Weight) error {
	user, err := FindUser(weight.GetWeight())
	if err != nil {
		return err
	}
	path := fmt.Sprintf("data/%s.csv", user)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
			if err != nil {
				return fmt.Errorf("unable to open CSV file: %v", err)
			}
			_, err = file.WriteString(fmt.Sprintf("Date,%s\n", user))
			if err != nil {
				return fmt.Errorf("unable to write to CSV file: %v", err)
			}
		} else {
			return fmt.Errorf("unable to open CSV file: %v", err)
		}
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close CSV file: %v\n", err)
		}
	}(file)

	date := weight.Time.Format("2006/01/02 15:04")
	_, err = file.WriteString(fmt.Sprintf("%s,%.1f\n", date, weight.GetWeight()))
	if err != nil {
		return fmt.Errorf("unable to write to CSV file: %v", err)
	}

	return nil
}
