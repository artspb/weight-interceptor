package storage

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

const retryName = "data/retry.txt"

func AddRetry(request []byte) error {
	return addRetry(retryName, request)
}

func RetryAll(retry func(request string) error) (success int, failure int, err error) {
	file, err := os.Open(retryName)
	if err != nil {
		if os.IsNotExist(err) {
			return success, failure, nil
		}
		return success, failure, fmt.Errorf("unable to open retry file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close retry file: %v\n", err)
		}
	}(file)

	const retryTempName = "data/retry.temp.txt"
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		request := scanner.Text()
		err := retry(request)
		if err != nil {
			failure++
			if err := addRetry(retryTempName, []byte(request)); err != nil {
				return success, failure, err
			}
		} else {
			success++
		}
	}

	if err = scanner.Err(); err != nil {
		return success, failure, err
	}

	err = os.Remove(retryName)
	if err != nil {
		return success, failure, err
	}

	if _, err := os.Stat(retryTempName); err == nil {
		err := os.Rename(retryTempName, retryName)
		if err != nil {
			return success, failure, err
		}
	}

	return success, failure, nil
}

func addRetry(path string, request []byte) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("unable to open retry file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Unable to close retry file: %v\n", err)
		}
	}(file)
	return writeRequest(file, request)
}
