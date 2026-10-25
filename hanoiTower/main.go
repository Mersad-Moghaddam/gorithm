package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 500
	baseY        = 400.0
	clearY       = 120.0
)

var (
	pegX = [3]float64{180, 400, 620}

	diskColors = []color.RGBA{
		{R: 235, G: 75, B: 75, A: 255},  // Red
		{R: 245, G: 140, B: 50, A: 255}, // Orange
		{R: 240, G: 200, B: 50, A: 255}, // Yellow
		{R: 60, G: 195, B: 110, A: 255}, // Green
		{R: 50, G: 190, B: 220, A: 255}, // Cyan
		{R: 65, G: 125, B: 245, A: 255}, // Blue
		{R: 160, G: 85, B: 235, A: 255}, // Purple
		{R: 235, G: 80, B: 175, A: 255}, // Magenta
	}
)

type Move struct {
	disk, from, to int
}

func genMoves(n, from, to, aux int, out *[]Move) {
	if n == 0 {
		return
	}
	genMoves(n-1, from, aux, to, out)
	*out = append(*out, Move{disk: n, from: from, to: to})
	genMoves(n-1, aux, to, from, out)
}

type Game struct {
	n         int
	pegs      [3][]int
	moves     []Move
	moveIdx   int
	progress  float64
	speed     float64
	paused    bool
	prevSpace bool
	done      bool
}

func NewGame(n int) *Game {
	g := &Game{speed: 3.5}
	g.reset(n)
	return g
}

func (g *Game) reset(n int) {
	g.n = n
	g.pegs = [3][]int{}
	for i := n; i >= 1; i-- {
		g.pegs[0] = append(g.pegs[0], i)
	}
	g.moves = nil
	genMoves(n, 0, 2, 1, &g.moves)
	g.moveIdx = 0
	g.progress = 0
	g.done = false
}

func (g *Game) diskH() float64 {
	h := 180.0 / float64(g.n+1)
	if h > 24.0 {
		return 24.0
	}
	return h
}

func (g *Game) diskW(d int) float64 {
	if g.n <= 1 {
		return 120.0
	}
	minW, maxW := 40.0, 180.0
	return minW + (maxW-minW)*float64(d-1)/float64(g.n-1)
}

func ease(t float64) float64 {
	return (1.0 - math.Cos(t*math.Pi)) / 2.0
}

func (g *Game) Update() error {
	// Keys 1-8 reset with new disk count
	keys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
		ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8,
	}
	for i, k := range keys {
		if ebiten.IsKeyPressed(k) && g.n != i+1 {
			g.reset(i + 1)
			return nil
		}
	}

	// Space pauses
	spaceNow := ebiten.IsKeyPressed(ebiten.KeySpace)
	if spaceNow && !g.prevSpace {
		g.paused = !g.paused
	}
	g.prevSpace = spaceNow

	// R restarts
	if ebiten.IsKeyPressed(ebiten.KeyR) {
		g.reset(g.n)
		return nil
	}

	// Up / Down adjusts speed
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) && g.speed < 12.0 {
		g.speed += 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && g.speed > 0.5 {
		g.speed -= 0.05
	}

	if g.paused || g.done {
		return nil
	}

	if g.moveIdx >= len(g.moves) {
		g.done = true
		return nil
	}

	g.progress += g.speed * (1.0 / 60.0)
	if g.progress >= 3.0 {
		m := g.moves[g.moveIdx]
		g.pegs[m.from] = g.pegs[m.from][:len(g.pegs[m.from])-1]
		g.pegs[m.to] = append(g.pegs[m.to], m.disk)
		g.moveIdx++
		g.progress = 0
		if g.moveIdx >= len(g.moves) {
			g.done = true
		}
	}

	return nil
}

func (g *Game) drawDisk(screen *ebiten.Image, d int, cx, cy float64) {
	w := g.diskW(d)
	h := g.diskH()
	x := cx - w/2
	y := cy - h/2
	clr := diskColors[(d-1)%len(diskColors)]

	ebitenutil.DrawRect(screen, x, y, w, h, clr)
	ebitenutil.DrawRect(screen, x+2, y+2, w-4, 2, color.RGBA{255, 255, 255, 110}) // shine highlight
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{22, 25, 32, 255})

	// Base plate
	ebitenutil.DrawRect(screen, 60, baseY, 680, 18, color.RGBA{70, 76, 92, 255})

	// Peg shafts and labels
	h := g.diskH()
	for p := range 3 {
		ebitenutil.DrawRect(screen, pegX[p]-4, 150, 8, baseY-150, color.RGBA{105, 115, 138, 255})
		tag := fmt.Sprintf("PEG %c", 'A'+p)
		ebitenutil.DebugPrintAt(screen, tag, int(pegX[p])-18, int(baseY)+24)
	}

	// Stationary disks
	inFlightDisk := 0
	if !g.done && g.moveIdx < len(g.moves) {
		inFlightDisk = g.moves[g.moveIdx].disk
	}

	for p := range 3 {
		for slot, d := range g.pegs[p] {
			if d == inFlightDisk && slot == len(g.pegs[p])-1 {
				continue // flying, drawn below
			}
			cy := baseY - float64(slot)*h - h/2
			g.drawDisk(screen, d, pegX[p], cy)
		}
	}

	// Flying disk
	if !g.done && g.moveIdx < len(g.moves) {
		m := g.moves[g.moveIdx]
		startSlot := len(g.pegs[m.from]) - 1
		endSlot := len(g.pegs[m.to])

		startY := baseY - float64(startSlot)*h - h/2
		endY := baseY - float64(endSlot)*h - h/2

		var cx, cy float64
		switch {
		case g.progress < 1.0:
			t := ease(g.progress)
			cx = pegX[m.from]
			cy = startY + (clearY-startY)*t
		case g.progress < 2.0:
			t := ease(g.progress - 1.0)
			cx = pegX[m.from] + (pegX[m.to]-pegX[m.from])*t
			cy = clearY
		default:
			t := ease(g.progress - 2.0)
			cx = pegX[m.to]
			cy = clearY + (endY-clearY)*t
		}

		g.drawDisk(screen, m.disk, cx, cy)
	}

	// HUD text
	status := fmt.Sprintf("N: %d | Move: %d / %d | Speed: %.1fx", g.n, g.moveIdx, len(g.moves), g.speed)
	if g.paused {
		status += " [PAUSED]"
	} else if g.done {
		status += " [SOLVED!]"
	}
	controls := "[1-8] Disks  [Space] Pause  [Up/Down] Speed  [R] Reset"
	ebitenutil.DebugPrintAt(screen, status, 20, 16)
	ebitenutil.DebugPrintAt(screen, controls, 20, 36)
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	// ponytail: Ebitengine 2D ceiling 60 FPS; upgrade to Raylib/Vulkan if 3D mesh geometry needed.
	n := 4
	fmt.Print("Enter disk count N (1-8) [default 4]: ")
	var input int
	if _, err := fmt.Scan(&input); err == nil && input >= 1 && input <= 8 {
		n = input
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Tower of Hanoi - 60 FPS")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(NewGame(n)); err != nil {
		panic(err)
	}
}
