package util

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

var Redis = &Redis{}

type Redis struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

var cfg *ini.File

func ConfigInit() {
	var err error
	cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("ConfigInit fail to parse 'app.ini': %v", err)
	}

	mapTo("redis", Redis)

	Redis.IdleTimeout = Redis.IdleTimeout * time.Second
}

func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo Redis err: %v", err)
	}
}
