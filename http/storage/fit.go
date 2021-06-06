package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"log"
	"os"
)

func ReadTokenFromFile(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read token from file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Unable to close token file: %v\n", err)
		}
	}(file)

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	return token, err
}

func SaveTokenToFile(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close token file: %v\n", err)
		}
	}(file)

	err = json.NewEncoder(file).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode oauth token: %w", err)
	}
	return nil
}

func ReadDataSourceIdFromFile(path string) (id string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("unable to read data source id from file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close data source id file: %v\n", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id = scanner.Text()
	}
	err = scanner.Err()
	if err != nil {
		return "", fmt.Errorf("unable to read data source id from file: %w", err)
	}
	return id, nil
}

func SaveDataSourceIdToFile(path string, id string) error {
	fmt.Printf("Saving data source id file to: %s\n", path)
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache data source id: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close data source id file: %v\n", err)
		}
	}(file)

	_, err = file.Write([]byte(id))
	if err != nil {
		return fmt.Errorf("unable to read data source id from file: %w", err)
	}
	return nil
}
