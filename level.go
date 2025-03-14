package main

import (
	_ "image/png"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/norendren/go-fov/fov"
)

type GameData struct {
	ScreenWidth  int
	ScreenHeight int
	TileWidth    int
	TileHeight   int
	MaxRoomSize  int
	MinRoomSize  int
	MaxRooms     int
}

// Creates a new instance of the static game data.
func NewGameData() GameData {
	gd := GameData{
		ScreenWidth:  80,
		ScreenHeight: 50,
		TileWidth:    16,
		TileHeight:   16,
		MaxRoomSize:  10,
		MinRoomSize:  6,
		MaxRooms:     30,
	}
	return gd
}

type MapTile struct {
	PixelX  int
	PixelY  int
	Blocked bool
	Opaque  bool
	Seen    bool
	Image   *ebiten.Image
}

const (
	TileFloor string = "floor"
	TileWall  string = "wall"
)

// Creates a MapTile of a given type at pixels x, y.
//
//	Supported types include 'Floor' and 'Wall'
func NewTile(x int, y int, tileType string) (MapTile, error) {
	blocked := true
	opaque := true

	image, _, err := ebitenutil.NewImageFromFile("assets/" + tileType + ".png")
	if err != nil {
		return MapTile{}, err
	}

	if tileType == TileFloor {
		blocked = false
		opaque = false
	}

	tile := MapTile{
		PixelX:  x,
		PixelY:  y,
		Blocked: blocked,
		Opaque:  opaque,
		Seen:    false,
		Image:   image,
	}
	return tile, nil
}

// Returns the logical index of the map slice from given X and Y tile coordinates.
func GetIndexFromCoords(x int, y int) int {
	gd := NewGameData()
	return (y * gd.ScreenWidth) + x
}

type Level struct {
	Tiles      []MapTile
	Rooms      []RectangularRoom
	PlayerView *fov.View
}

// Creates a new Level object
func NewLevel() Level {
	l := Level{}
	l.createTiles()
	l.PlayerView = fov.New()
	return l
}

// Creates a map of tiles.
func (level *Level) createTiles() {
	gd := NewGameData()
	tiles := make([]MapTile, gd.ScreenHeight*gd.ScreenWidth)

	// Fill with wall tiles
	for x := 0; x < gd.ScreenWidth; x++ {
		for y := 0; y < gd.ScreenHeight; y++ {
			idx := GetIndexFromCoords(x, y)
			wall, err := NewTile(x*gd.TileWidth, y*gd.TileHeight, TileWall)
			if err != nil {
				log.Fatal(err)
			}
			tiles[idx] = wall
		}
	}
	level.Tiles = tiles

	for i := 0; i <= gd.MaxRooms; i++ {
		// generate width and height as random numbers between gd.MinRoomSize and gd.MaxRoomSize
		width := rand.Intn(gd.MaxRoomSize-gd.MinRoomSize+1) + gd.MinRoomSize
		height := rand.Intn(gd.MaxRoomSize-gd.MinRoomSize+1) + gd.MinRoomSize

		xPos := rand.Intn(gd.ScreenWidth - width)
		yPos := rand.Intn(gd.ScreenHeight - height)

		newRoom := NewRectangularRoom(xPos, yPos, width, height)

		isOkay := true
		for _, room := range level.Rooms {
			// check through all existing rooms to ensure newRoom doesn't intersect
			if newRoom.IntersectsWith(room) {
				isOkay = false
				break
			}
		}

		if isOkay {
			level.Rooms = append(level.Rooms, newRoom)
		}
	}

	// Carve out rooms
	for roomNum, room := range level.Rooms {
		x1, x2, y1, y2 := room.Interior()
		for x := x1; x <= x2; x++ {
			for y := y1; y <= y2; y++ {
				idx := GetIndexFromCoords(x, y)
				floor, err := NewTile(x*gd.TileWidth, y*gd.TileHeight, TileFloor)
				if err != nil {
					log.Fatal(err)
				}
				level.Tiles[idx] = floor
			}
		}
		if roomNum > 0 {
			level.tunnelBetween(&level.Rooms[roomNum-1], &level.Rooms[roomNum])
		}
	}
}

// Draw the current level to 'screen'.
func (level *Level) Draw(screen *ebiten.Image) {
	gd := NewGameData()
	for x := 0; x < gd.ScreenWidth; x++ {
		for y := 0; y < gd.ScreenHeight; y++ {
			idx := GetIndexFromCoords(x, y)
			tile := level.Tiles[idx]
			if level.PlayerView.IsVisible(x, y) {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(tile.PixelX), float64(tile.PixelY))
				screen.DrawImage(tile.Image, op)
				level.Tiles[idx].Seen = true
			} else if tile.Seen {
				op := &colorm.DrawImageOptions{}
				var colorM colorm.ColorM
				op.GeoM.Translate(float64(tile.PixelX), float64(tile.PixelY))
				colorM.Translate(0, 0, 0, 0.30)
				colorM.Invert() // Make tile darker
				colorm.DrawImage(screen, tile.Image, colorM, op)
			}
		}
	}
}

type RectangularRoom struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

// Create a new RectangularRoom structure.
func NewRectangularRoom(x int, y int, width int, height int) RectangularRoom {
	return RectangularRoom{
		X1: x,
		Y1: y,
		X2: x + width,
		Y2: y + height,
	}
}

// Returns the tile coordinates of the center of the RectangularRoom.
func (r *RectangularRoom) Center() (int, int) {
	centerX := (r.X1 + r.X2) / 2
	centerY := (r.Y1 + r.Y2) / 2
	return centerX, centerY
}

// Returns the tile coordinates of the interior of the RectangularRoom.
func (r *RectangularRoom) Interior() (int, int, int, int) {
	return r.X1 + 1, r.X2 - 1, r.Y1 + 1, r.Y2 - 1
}

// Determines if this room intersects with otherRoom.
func (r *RectangularRoom) IntersectsWith(otherRoom RectangularRoom) bool {
	return r.X1 <= otherRoom.X2 && r.X2 >= otherRoom.X1 && r.Y1 <= otherRoom.Y2 && r.Y2 >= otherRoom.Y1
}

// Create a vertical tunnel.
func (level *Level) createVerticalTunnel(y1 int, y2 int, x int) {
	gd := NewGameData()
	for y := min(y1, y2); y < max(y1, y2)+1; y++ {
		idx := GetIndexFromCoords(x, y)

		if idx > 0 && idx < gd.ScreenHeight*gd.ScreenWidth {
			floor, err := NewTile(x*gd.TileWidth, y*gd.TileHeight, TileFloor)
			if err != nil {
				log.Fatal(err)
			}
			level.Tiles[idx] = floor
		}
	}
}

// Create a horizontal tunnel.
func (level *Level) createHorizontalTunnel(x1 int, x2 int, y int) {
	gd := NewGameData()
	for x := min(x1, x2); x < max(x1, x2)+1; x++ {
		idx := GetIndexFromCoords(x, y)

		if idx > 0 && idx < gd.ScreenHeight*gd.ScreenWidth {
			floor, err := NewTile(x*gd.TileWidth, y*gd.TileHeight, TileFloor)
			if err != nil {
				log.Fatal(err)
			}
			level.Tiles[idx] = floor
		}
	}
}

// Tunnel from this first room to second room.
func (level *Level) tunnelBetween(first *RectangularRoom, second *RectangularRoom) {
	startX, startY := first.Center()
	endX, endY := second.Center()

	if rand.Intn(2) == 0 {
		// Tunnel horizontally, then vertically
		level.createHorizontalTunnel(startX, endX, startY)
		level.createVerticalTunnel(startY, endY, endX)
	} else {
		// Tunnel vertically, then horizontally
		level.createVerticalTunnel(startY, endY, startX)
		level.createHorizontalTunnel(startX, endX, endY)
	}
}

// Determine if tile coordinates are on the screen.
func (level Level) InBounds(x int, y int) bool {
	gd := NewGameData()
	if x < 0 || x > gd.ScreenWidth || y < 0 || y > gd.ScreenHeight {
		return false
	}
	return true
}

// Determines if a tile at a given coordinate can be seen through.
func (level Level) IsOpaque(x int, y int) bool {
	return level.Tiles[GetIndexFromCoords(x, y)].Opaque
}
