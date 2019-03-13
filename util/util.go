package util

import "log"

func CheckErr(tip string, err error) {
	if err != nil {
		log.Println(tip, ":", err)
	}
}

func Log(info string) {
	log.Println(info)
}
