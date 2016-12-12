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


func q_func(e1 GPSElement, e2 GPSElement) float64 {
	x1 := e1.Position.X
	y1 := e1.Position.Y
	d1 := e1.Distance
	x2 := e2.Position.X
	y2 := e2.Position.Y
	d2 := e2.Distance
	return ((x2 - x1) * (x2 - x1) + (y2 - y1) * (y2 - y1) + d2 * d2 - d1 * d1) / 2.0
}

func k_func(e1 GPSElement, e2 GPSElement) float64 {
	y1 := e1.Position.Y
	y2 := e2.Position.Y

	return q_func(e1, e2) / (y2 - y1)
}

func q_func(e1 GPSElement, e2 GPSElement) float64 {
	x1 := e1.Position.X
	y1 := e1.Position.Y
	x2 := e2.Position.X
	y2 := e2.Position.Y
	return (x2 - x1) / (y2 - y1)
}


func x_c_func(e1 GPSElement, e2 GPSElement, e3 GPSElement) float64 {
	return (k_func(e2, e3) - k_func(e1, e2)) / (r_func(e1, e2) - r_func(e2, e3)) - e1.Position.X
}


func y_c_func(e1 GPSElement, e2 GPSElement, e3 GPSElement) float64 {
	return r_func(e1, e1) * ((k_func(e2, e3) - k_func(e1, e2)) / (r_func(e1, e2) - r_func(e2, e3))) - k_func(e1, e2) - e1.Position.Y
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

		// get our position from game
		// calculate gift position from game[0..2]
		// move  gift positino - our poisitin

		our_pos_x := 0
		our_pos_y := 0

		prize_x := x_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2])
		prize_y := y_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2])

		// command Move (prize_x - our_x, prize_y - our_y)
	}
}
