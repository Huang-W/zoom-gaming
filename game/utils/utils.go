package utils

import (
	"fmt"
	"log"
)

// Helper Functions
func FailOnError(err error, format string, a ...interface{}) {
	if err != nil {
		log.Fatalf(format, append(a, err)...)
		panic(fmt.Sprintf(format, append(a, err)...))
	}
}

func WarnOnError(err error, format string, a ...interface{}) {
	if err != nil {
		log.Printf(format, append(a, err)...)
	}
}
