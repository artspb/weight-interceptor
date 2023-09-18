package storage

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

const usersName = "data/users.txt"

var UnknownUser = errors.New("unknown user")

var users []User

func init() {
	file, err := os.Open(usersName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No users specified, the name 'default' will be used")
			users = append(users, User{
				Name:      "default",
				MinWeight: 0,
				MaxWeight: math.MaxFloat64,
			})
			return
		}
		log.Fatalf("Unable to open users file: %v\n", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Unable to close users file: %v\n", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		user := scanner.Text()
		triple := strings.Split(user, " ")
		if len(triple) != 3 {
			log.Fatalln("Each user should have a name and a pair of min and max weight")
		}
		minWeight, err := strconv.ParseFloat(triple[1], 64)
		if err != nil {
			log.Fatalf("Unable to parse min weight (%s): %v\n", triple[1], err)
		}
		maxWeight, err := strconv.ParseFloat(triple[2], 64)
		if err != nil {
			log.Fatalf("Unable to parse max weight (%s): %v\n", triple[2], err)
		}
		users = append(users, User{
			Name:      triple[0],
			MinWeight: minWeight,
			MaxWeight: maxWeight,
		})
	}

	if err = scanner.Err(); err != nil {
		log.Fatalf("Unable to read users file: %v\n", err)
	}
}

type User struct {
	Name      string
	MinWeight float64
	MaxWeight float64
}

func FindUser(weight float64) (string, error) {
	for _, user := range users {
		if weight >= user.MinWeight && weight <= user.MaxWeight {
			return user.Name, nil
		}
	}
	return "", UnknownUser
}
