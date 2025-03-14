package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	Levels       []Level
	CurrentLevel *Level
	Entities     []Entity
	Player       *Entity
	TickCount    int
	PromptText   string
	GameState    GameState
}

// Creates a new Game object and initializes the data.
func NewGame() *Game {
	g := &Game{}
	g.Levels = append(g.Levels, NewLevel())
	g.CurrentLevel = &g.Levels[0]
	g.TickCount = 0

	startX, startY := g.CurrentLevel.Rooms[0].Center()
	player, err := NewEntity(startX, startY, "player")
	if err != nil {
		log.Fatal(err)
	}
	g.Entities = append(g.Entities, player)
	g.Player = &g.Entities[0]
	return g
}

func (g *Game) Update() error {
	if g.GameState == STOP {
		return ebiten.Termination
	}

	g.TickCount++

	if g.TickCount > 5 {
		HandleInput(g)
		g.TickCount = 0
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.GameState == PROMPT {
		if g.PromptText != "" {
			ebitenutil.DebugPrint(screen, g.PromptText)
		}
	}

	level := g.Levels[0]
	level.Draw(screen)
	RenderEntities(g, level, screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	gd := NewGameData()
	return gd.TileWidth * gd.ScreenWidth, gd.TileHeight * gd.ScreenHeight
}

func main() {
	g := NewGame()
	ebiten.SetWindowTitle("Grogue")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
