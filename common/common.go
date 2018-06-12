package common

import (
	"fmt"
	"log"
	"os"
)

var Logfile, _ = os.Create(os.Getenv("LOG_FILE"))

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func WriteLogFile(param string) interface{} {

	logger := log.New(Logfile, param+": ", log.LstdFlags|log.Lshortfile)
	return logger

}
