package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"sync"
)

type Game struct {
	GPSS    []GPSElement
	Players []Player
}

type Player struct {
	Position Position
	Score    int
	Color    string
	Name     string
}

type Position struct {
	X float64
	Y float64
}

type GPSElement struct {
	Position Position
	Distance float64
}

type World struct {
	socket   *websocket.Conn
	encoder  *json.Encoder
	position Position
}

type Command struct {
	Contents string `json:"contents"`
	Tag      string `json:"tag"`
}

func (w *World) Encode(v interface{}) error {
	return w.encoder.Encode(v)
}

func (w *World) CommandTag(tag, content string) error {
	log.WithField("tag", tag).WithField("content", content)
	c := Command{
		Tag:      tag,
		Contents: content,
	}
	return w.Encode(&c)
}

func (w *World) Run() {
	w.encoder = json.NewEncoder(w.socket)
	err := w.CommandTag("SetName", "team-golang")
	if err != nil {
		log.Fatal("error setting name: ", err)
	}

	err = w.CommandTag("SetColor", "#00ffff")
	if err != nil {
		log.Fatal("error setting color: ", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		var game Game
		decoder := json.NewDecoder(w.socket)
		for {
			err := decoder.Decode(&game)
			if err != nil {
				log.Warn("error parsing message: ", err)
				continue
			}
			log.Debugf("game: %+v", game)
			for _, player := range game.Players {
				if player.Name == "team-golang" {
					w.position = player.Position
					break
				}
			}
			log.Infof("my position: %+v", w.position)
		}
	}()

	wg.Wait()
}

func NewWorld(url string) (*World, error) {
	origin := "http://localhost/"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}

	w := &World{socket: ws}

	return w, nil
}

func main() {
	w, err := NewWorld("ws://game.clearercode.com")
	if err != nil {
		log.Fatal(err)
	}
	w.Run()
}
