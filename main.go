package main

import (
	"flag"
	"go-websocket/server"
	"go-websocket/util"
	"net/http"
)

func init() {
	util.ConfigInit()
	util.RedisInit()
}

func main() {
	p := flag.String("p", "7777", "http listen port")
	flag.Parse()
	hub := server.NewHub()
	hub.Run()
	http.HandleFunc("/ws", server.ServeWS(hub))
	http.ListenAndServe("0.0.0.0:"+*p, nil)
}
