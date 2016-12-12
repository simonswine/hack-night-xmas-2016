package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
	"math"
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
	throttle      <-chan time.Time
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
	<-w.throttle
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
	var err error

	w.encoder = json.NewEncoder(w.socket)

	rate := time.Second / 11
	w.throttle = time.Tick(rate)

	err = w.CommandTag("SetName", "team-golang")
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
			x, y := calc(
				game.GPSS[0].Position.X,
				game.GPSS[0].Position.Y,
				game.GPSS[0].Distance,
				game.GPSS[1].Position.X,
				game.GPSS[1].Position.Y,
				game.GPSS[1].Distance,
				game.GPSS[2].Position.X,
				game.GPSS[2].Position.Y,
				game.GPSS[2].Distance,
			)
			w.prizePosition = Position{
				X: x,
				Y: y,
			}

			for _, player := range game.Players {
				if player.Name == "team-golang" {
					w.myPosition = player.Position
					w.myPosition.X++
					w.myPosition.Y++
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

func calc(x0, y0, r0, x1, y1, r1, x2, y2, r2 float64) (x float64, y float64) {
	EPSILON := 0.00001
	var a, dx, dy, d, h, rx, ry float64
	var point2_x, point2_y float64

	/* dx and dy are the vertical and horizontal distances between
	 * the circle centers.
	 */
	dx = x1 - x0
	dy = y1 - y0

	/* Determine the straight-line distance between the centers. */
	d = math.Sqrt((dy * dy) + (dx * dx))

	/* Check for solvability. */
	if d > (r0 + r1) {
		/* no solution. circles do not intersect. */
		return
	}
	if d < math.Abs(r0-r1) {
		/* no solution. one circle is contained in the other */
		return
	}

	/* 'point 2' is the point where the line through the circle
	 * intersection points crosses the line between the circle
	 * centers.
	 */

	/* Determine the distance from point 0 to point 2. */
	a = ((r0 * r0) - (r1 * r1) + (d * d)) / (2.0 * d)

	/* Determine the coordinates of point 2. */
	point2_x = x0 + (dx * a / d)
	point2_y = y0 + (dy * a / d)

	/* Determine the distance from point 2 to either of the
	 * intersection points.
	 */
	h = math.Sqrt((r0 * r0) - (a * a))

	/* Now determine the offsets of the intersection points from
	 * point 2.
	 */
	rx = -dy * (h / d)
	ry = dx * (h / d)

	/* Determine the absolute intersection points. */
	intersectionPoint1_x := point2_x + rx
	intersectionPoint2_x := point2_x - rx
	intersectionPoint1_y := point2_y + ry
	intersectionPoint2_y := point2_y - ry

	/* Lets determine if circle 3 intersects at either of the above intersection points. */
	dx = intersectionPoint1_x - x2
	dy = intersectionPoint1_y - y2
	d1 := math.Sqrt((dy * dy) + (dx * dx))

	dx = intersectionPoint2_x - x2
	dy = intersectionPoint2_y - y2
	d2 := math.Sqrt((dy * dy) + (dx * dx))

	if math.Abs(d1-r2) < EPSILON {
		return intersectionPoint1_x, intersectionPoint1_y
	} else if math.Abs(d2-r2) < EPSILON {
		return intersectionPoint2_x, intersectionPoint2_y
	} else {
		return
	}
	return
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
