package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	tb "github.com/nsf/termbox-go"
)

// Entity represents any object that has an X/Y position and can be alive or dead.
type Entity struct {
	x, y  int
	alive bool
}

// Game holds the entire game state.
type Game struct {
	width, height int
	player        Entity
	aliens        []Entity
	bullets       []Entity
	score         int
	tick          int
}

func newGame() *Game {
	g := &Game{width: 40, height: 20}
	g.player = Entity{x: g.width / 2, y: g.height - 1, alive: true}
	for r := 1; r <= 4; r++ {
		for c := 5; c < g.width-5; c += 3 {
			g.aliens = append(g.aliens, Entity{x: c, y: r, alive: true})
		}
	}
	return g
}

func (g *Game) draw() {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	if g.player.alive {
		tb.SetCell(g.player.x, g.player.y, '^', tb.ColorGreen, tb.ColorDefault)
	}
	for _, a := range g.aliens {
		if a.alive {
			tb.SetCell(a.x, a.y, 'W', tb.ColorRed, tb.ColorDefault)
		}
	}
	for _, b := range g.bullets {
		tb.SetCell(b.x, b.y, '|', tb.ColorYellow, tb.ColorDefault)
	}
	fmtStr := fmt.Sprintf("Score: %d", g.score)
	for i, ch := range fmtStr {
		tb.SetCell(i, 0, ch, tb.ColorWhite, tb.ColorDefault)
	}
	tb.Flush()
}

func (g *Game) update() {
	var aliveBullets []Entity
	for _, b := range g.bullets {
		b.y--
		if b.y > 0 {
			aliveBullets = append(aliveBullets, b)
		}
	}
	g.bullets = aliveBullets

	// collisions
	for i, b := range g.bullets {
		for j, a := range g.aliens {
			if a.alive && b.x == a.x && b.y == a.y {
				a.alive = false
				g.score += 10
				g.aliens[j] = a
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				break
			}
		}
	}

	// move aliens every few ticks
	if g.tick%10 == 0 {
		dir := 1
		for _, a := range g.aliens {
			if !a.alive {
				continue
			}
			if a.x+dir >= g.width-1 || a.x+dir <= 0 {
				dir = -1
				break
			}
		}
		for i, a := range g.aliens {
			if a.alive {
				a.x += dir
				if dir == -1 && a.y < g.height-2 {
					a.y++
				}
				g.aliens[i] = a
			}
		}
	}
	g.tick++
}

func (g *Game) handleEvent(ev tb.Event) bool {
	switch ev.Type {
	case tb.EventKey:
		if !g.player.alive {
			return false
		}
		switch ev.Key {
		case tb.KeyArrowLeft:
			if g.player.x > 1 {
				g.player.x--
			}
		case tb.KeyArrowRight:
			if g.player.x < g.width-2 {
				g.player.x++
			}
		case tb.KeyEsc:
			return false
		case tb.KeySpace:
			g.bullets = append(g.bullets, Entity{x: g.player.x, y: g.player.y - 1, alive: true})
		}
		if ev.Ch == ' ' {
			g.bullets = append(g.bullets, Entity{x: g.player.x, y: g.player.y - 1, alive: true})
		}
	case tb.EventError:
		fmt.Fprintln(os.Stderr, ev.Err)
		return false
	}
	return true
}

func main() {
	if err := tb.Init(); err != nil {
		fmt.Println("termbox init failed", err)
		return
	}
	defer tb.Close()
	rand.Seed(time.Now().UnixNano())
	game := newGame()

	eventChan := make(chan tb.Event)
	go func() {
		for {
			eventChan <- tb.PollEvent()
		}
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	running := true
	for running {
		select {
		case ev := <-eventChan:
			if !game.handleEvent(ev) {
				running = false
			}
		case <-ticker.C:
			game.update()
			game.draw()
		}
	}
}
