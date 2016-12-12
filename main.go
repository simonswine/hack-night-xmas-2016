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

type World struct {
	socket *websocket.Conn
}

type Command struct {
	Contents string `json:"contents"`
	Tag      string `json:"tag"`
}

func (w *World) CommandTag(tag, content string) error {
	c := &Command{
		Tag:      tag,
		Contents: content,
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	bytes = append(bytes, byte('\n'))

	_, err = w.socket.Write(bytes)
	return err
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

	err = w.CommandTag("SetName", "team-golang")
	if err != nil {
		log.Fatal("error setting name: ", err)
	}

	err = w.CommandTag("SetColor", "ffff00")
	if err != nil {
		log.Fatal("error setting color: ", err)
	}

	var game Game
	buffer := make([]byte, 65535)
	for {
		n, err := w.socket.Read(buffer)
		if err != nil {
			log.Warn("End of input: ", err)
			break
		}

		line := buffer[:n]
		log.Infof("line: %+v", string(line))
		err = json.Unmarshal(line, &game)
		if err != nil {
			log.Warn("error parsing message: ", err)
			continue
		}
		log.Infof("message: %+v", game)
	}
}
