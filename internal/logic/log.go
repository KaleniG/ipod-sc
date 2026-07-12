package logic

import (
	"log"
	"os"
)

var errorLog *log.Logger

func init() {
	f, err := os.OpenFile(
		"errors.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}

	errorLog = log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
}
