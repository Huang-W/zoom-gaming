package utils

import (
  "fmt"
  "log"
)

// Helper Functions
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func WarnOnError(err error, msg string) {
	if err != nil {
		log.Println("%s: %s", msg, err)
	}
}
