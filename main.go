package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"sync"
	"time"
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

type PositionInt struct {
	X int `json:"x"`
	Y int `json:"y"`
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
	socket        *websocket.Conn
	encoder       *json.Encoder
	myPosition    Position
	prizePosition Position
}

type Command struct {
	Contents string `json:"contents"`
	Tag      string `json:"tag"`
}

type Move struct {
	Contents PositionInt `json:"contents"`
	Tag      string      `json:"tag"`
}

func (w *World) Encode(v interface{}) error {
	return w.encoder.Encode(v)
}

func q_func(e1 GPSElement, e2 GPSElement) float64 {
	x1 := e1.Position.X
	y1 := e1.Position.Y
	d1 := e1.Distance
	x2 := e2.Position.X
	y2 := e2.Position.Y
	d2 := e2.Distance
	return ((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1) + d2*d2 - d1*d1) / 2.0
}

func k_func(e1 GPSElement, e2 GPSElement) float64 {
	y1 := e1.Position.Y
	y2 := e2.Position.Y

	return q_func(e1, e2) / (y2 - y1)
}

func r_func(e1 GPSElement, e2 GPSElement) float64 {
	x1 := e1.Position.X
	y1 := e1.Position.Y
	x2 := e2.Position.X
	y2 := e2.Position.Y
	return (x2 - x1) / (y2 - y1)
}

func x_c_func(e1 GPSElement, e2 GPSElement, e3 GPSElement) float64 {
	return (k_func(e2, e3)-k_func(e1, e2))/(r_func(e1, e2)-r_func(e2, e3)) - e1.Position.X
}

func y_c_func(e1 GPSElement, e2 GPSElement, e3 GPSElement) float64 {
	return r_func(e1, e1)*((k_func(e2, e3)-k_func(e1, e2))/(r_func(e1, e2)-r_func(e2, e3))) - k_func(e1, e2) - e1.Position.Y
}

func (w *World) CommandTag(tag, content string) error {
	log.WithField("tag", tag).WithField("content", content).Info()
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
			w.prizePosition = Position{
				X: x_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2]),
				Y: y_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2]),
			}

			for _, player := range game.Players {
				if player.Name == "team-golang" {
					w.myPosition = player.Position
					break
				}
			}
			log.Infof("my position: %+v", w.myPosition)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			w.Move(int(w.prizePosition.X-w.myPosition.X), int(w.prizePosition.Y-w.myPosition.Y))
			time.Sleep(500 * time.Millisecond)
		}
	}()

	wg.Wait()
}

func (w *World) Move(x, y int) error {
	log.Infof("move X=%d, Y=%d", x, y)
	c := &Move{
		Tag:      "Move",
		Contents: PositionInt{X: x, Y: y},
	}
	return w.Encode(&c)
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

	/*
				// get our position from game
				// calculate gift position from game[0..2]
				// move  gift positino - our poisitin

				our_pos_x := 0
				our_pos_y := 0

				prize_x := x_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2])
				prize_y := y_c_func(game.GPSS[0], game.GPSS[1], game.GPSS[2])

				// command Move (prize_x - our_x, prize_y - our_y)
			}
		>>>>>>> fe26090f0ae216d1c5a34e36da1b5948b40c5566
	*/
}
