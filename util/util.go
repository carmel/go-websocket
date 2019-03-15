package util

import (
	"log"

	uuid "github.com/satori/go.uuid"
)

func CheckErr(tip string, err error) {
	if err != nil {
		log.Println(tip, ":", err)
	}
}

func Log(v ...interface{}) {
	log.Println(v...)
}

func UUID() string {
	return uuid.Must(uuid.NewV1()).String()
}
