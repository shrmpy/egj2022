package main

import (
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
)
import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/shrmpy/egj2022/polarity"
	"github.com/tinne26/etxt"
)

//go:embed DejaVuSansMono.ttf
var dejavuSansMonoTTF []byte

var emptyImage = ebiten.NewImage(3, 3)

func init() {
	// todo is fill for alpha lvl here?
	emptyImage.Fill(color.White)
	log.SetFlags(log.Lshortfile | log.Ltime)
}
func main() {
	var (
		err    error
		name   string
		wd, ht = 640, 480
		ch     = make(chan string, 100)
		fonts  = etxt.NewFontLibrary()
	)
	defer close(ch)
	if name, err = fonts.ParseFontBytes(dejavuSansMonoTTF); err != nil {
		log.Fatalf("FAIL Parse error DejaVu Sans Mono, %s", err.Error())
	}
	log.Printf("INFO font %s", name)
	var renderer = etxt.NewStdRenderer()
	renderer.SetCacheHandler(etxt.NewDefaultCache(2 * 1024 * 1024).NewHandler())
	renderer.SetFont(fonts.GetFont("DejaVu Sans Mono"))
	renderer.SetColor(color.White)
	renderer.SetSizePx(12)
	ebiten.SetWindowSize(wd, ht)
	ebiten.SetWindowTitle("egj2022")
	var game = &Game{
		info:    ch,
		Width:   wd,
		Height:  ht,
		txtre:   renderer,
		history: make([]string, 0, 25),
	}
	game.maze = polarity.NewMaze(20, game.AddHistory)

	if err = ebiten.RunGame(game); err != nil {
		log.Fatalf("FAIL shutdown, %v", err)
	}
}

// Update runs game logic steps
func (g *Game) Update() error {
	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return errors.New("game quit by player")
	}

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		var fs = ebiten.IsFullscreen()
		ebiten.SetFullscreen(!fs)
	}

	if g.trouble {
		// troubleshooting, don't exit
		return nil
	}
	if err := g.maze.Update(); err != nil {
		// TODO error mgt
		log.Printf("DEBUG maze end, %v", err.Error())
		g.trouble = true
		//return err
	}

	return nil
}

// Draw renders one frame
func (g *Game) Draw(screen *ebiten.Image) {
	g.printDebugLog(screen)
	var (
		rgb     color.RGBA
		v       []ebiten.Vertex
		i       []uint16
		x, y    float32
		cpx     = float32(12)
		mm      = g.maze.Mini()
		padLeft = float32(g.Width) - g.maze.Width()*cpx
		src     = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	)
	// draw mini map in ne corner
	for row, rslc := range mm {
		for col, cell := range rslc {
			switch {
			case cell.Has(polarity.Barrier):
				rgb = color.RGBA{0x4b, 0x00, 0x82, 0xff}
			case cell.Has(polarity.Jaeger):
				rgb = color.RGBA{0xad, 0xff, 0x2f, 0xff}
			default:
				rgb = color.RGBA{0x33, 0x33, 0x33, 0xff}
			}
			x = float32(col)*cpx + padLeft
			y = float32(row) * cpx
			v, i = rect(x, y, cpx, cpx, rgb)
			screen.DrawTriangles(v, i, src, nil)
		}
	}
}

func (g *Game) printDebugLog(screen *ebiten.Image) {
	// help us troubleshoot
	g.txtre.SetTarget(screen)
	max := len(g.history)
	select {
	case dl := <-g.info:
		// TODO scroll messages
		if max < 25 {
			g.history = append(g.history, fmt.Sprintf("DEBUG: %s", dl))
		}
	default:
		g.txtre.SetAlign(etxt.Bottom, etxt.Left)
		for i := max; i > 0; i-- {
			msg := g.history[i-1]
			sz := g.txtre.SelectionRect(msg)
			g.txtre.Draw(msg, 0, g.Height-sz.Height.Ceil()*i)
		}
		// print frame rate in se corner
		g.txtre.SetAlign(etxt.Bottom, etxt.Right)
		g.txtre.Draw(fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()), g.Width-1, g.Height)
	}
}

// Game represents the main game state
type Game struct {
	Width   int
	Height  int
	info    chan string
	maze    *polarity.Maze
	txtre   *etxt.Renderer
	history []string
	trouble bool
}

// Layout is static for now, can be dynamic
func (g *Game) Layout(outsideWidth int, outsideHeight int) (int, int) {
	return g.Width, g.Height
}

// allow maze to bubble-up debug msg
func (g *Game) AddHistory(msg string) {
	g.info <- msg
}
