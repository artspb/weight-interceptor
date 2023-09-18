package main

import (
	"fmt"
	"log"
	"os"
	"weight-interceptor-http/dataprocessor/fit"
	"weight-interceptor-http/dataservice"
)

func main() {
	args := os.Args
	if len(args) == 3 && args[1] == "fit" {
		err := fit.Authenticate(args[2])
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Authenticated successfully")
		return
	}

	dataservice.StartService()
}
