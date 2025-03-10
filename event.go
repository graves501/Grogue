package main

import (
	"fmt"

	ebi "github.com/hajimehoshi/ebiten/v2"
)

const (
	LEFT  = -1
	UP    = -1
	DOWN  = 1
	RIGHT = 1
)

const (
	KEY_LEFT  = ebi.KeyH
	KEY_DOWN  = ebi.KeyJ
	KEY_UP    = ebi.KeyK
	KEY_RIGHT = ebi.KeyL

	KEY_TOP_LEFT     = ebi.KeyY
	KEY_TOP_RIGHT    = ebi.KeyU
	KEY_BOTTOM_LEFT  = ebi.KeyB
	KEY_BOTTOM_RIGHT = ebi.KeyN

	KEY_QUIT   = ebi.KeyQ
	KEY_ESCAPE = ebi.KeyEscape

	KEY_YES = ebi.KeyY
	KEY_NO  = ebi.KeyN
)

// Handle user input, including moving the player.
func HandleInput(g *Game) {
	dx := 0
	dy := 0

	if IsKeyPressed(KEY_QUIT, KEY_ESCAPE) {
		g.PromptText = "Do you really want to quit? [y/n]"
		g.GameState = PROMPT
		return
	}

	if g.GameState == PROMPT {
		if IsKeyPressed(KEY_YES) {
			g.GameState = STOP
		} else if IsKeyPressed(KEY_NO, KEY_ESCAPE) {
			g.GameState = RUNNING
			g.PromptText = ""
		}
		return
	}

	// Up / Down
	if IsKeyPressed(KEY_UP) {
		dy = UP
	} else if IsKeyPressed(KEY_DOWN) {
		dy = DOWN
	}

	// Left / Right
	if IsKeyPressed(KEY_LEFT) {
		dx = LEFT
	} else if IsKeyPressed(KEY_RIGHT) {
		dx = RIGHT
	}

	// Diagonals
	if IsKeyPressed(KEY_TOP_LEFT) {
		dx = LEFT
		dy = UP
	} else if IsKeyPressed(KEY_TOP_RIGHT) {
		dx = RIGHT
		dy = UP
	} else if IsKeyPressed(KEY_BOTTOM_LEFT) {
		dx = LEFT
		dy = DOWN
	} else if IsKeyPressed(KEY_BOTTOM_RIGHT) {
		dx = RIGHT
		dy = DOWN
	}

	newPos := GetIndexFromCoords(g.Player.X+dx, g.Player.Y+dy)
	tile := g.CurrentLevel.Tiles[newPos]
	if !tile.Blocked {
		g.Player.X += dx
		g.Player.Y += dy
		g.CurrentLevel.PlayerView.Compute(g.CurrentLevel, g.Player.X, g.Player.Y, 8)
	}
}

func IsKeyPressed(keys ...ebi.Key) bool {
	for _, key := range keys {
		if ebi.IsKeyPressed(key) {
			return true
		}
	}
	return false
}
