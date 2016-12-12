package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type Game struct {
	GPSS []GPSElement
}

type GPSElement struct {
	Position struct {
		X int
		Y int
	}
	Distance float64
}

func main() {
	origin := "http://localhost/"
	url := "ws://game.clearercode.com"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	//if _, err := ws.Write([]byte("hello, world!\n")); err != nil {
	//	log.Fatal(err)
	//}
	var msg = make([]byte, 1024)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Fatal(err)
	}
	var game Game
	err = json.Unmarshal(msg[:n], &game)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("%+v.\n", game)
}
