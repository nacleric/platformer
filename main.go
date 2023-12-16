package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  int = 480
	screenHeight int = 320
	// Hardcoded values for now but will be dynamic depending on map
	mapWidth  int = 30
	mapHeight int = 20
	tileSize  int = 16
	// These values will change
	cameraWidth  int     = 30
	cameraHeight int     = 20
	cameraScale  float64 = float64(screenWidth) / float64(tileSize) / float64(cameraWidth)

	startPosX = 0
	startPosY = 2
)

var (
	game               *Game
	chickenSpritesheet *ebiten.Image
	groundSpritesheet  *ebiten.Image
	waterSpritesheet   *ebiten.Image
)

type SpriteCell struct {
	cellX       int // column of spritesheet Ex: 0 is first col 16 is 2nd col
	cellY       int // row of spritesheet Ex: 16 is 2nd row
	frameWidth  int // Size of Sprite frame (most likely 16x16)
	frameHeight int
}

func (sc *SpriteCell) getRow(cellY int) int {
	return cellY * sc.frameHeight
}

func (sc *SpriteCell) getCol(cellX int) int {
	return cellX * sc.frameWidth
}

type AnimationData struct {
	sc             SpriteCell
	frameCount     int // Total number of columns for specific row
	frameFrequency int // How often frames transition
}

type Player struct {
	posX        float64
	posY        float64
	vX          float64
	vY          float64
	width       int
	height      int
	walkAnim    AnimationData
	idleAnim    AnimationData
	direction   string // (LEFT, RIGHT) need to find a better way to represent enums
	gravity     float64
	onGround    bool
	spritesheet *ebiten.Image
}

func TilePos(position float64) float64 {
	return position * float64(tileSize)
}

func CreatePlayer(spritesheet *ebiten.Image) Player {
	playerWalkAnimationData := AnimationData{SpriteCell{0, 1, tileSize, tileSize}, 4, 8}
	playerIdleAnimationData := AnimationData{SpriteCell{0, 0, tileSize, tileSize}, 2, 64}

	// Will need to change screenHeight-height*2 when physics/jump is created
	p := Player{TilePos(startPosX), TilePos(startPosY), 0, 0, tileSize, tileSize, playerWalkAnimationData, playerIdleAnimationData, "RIGHT", 1, true, spritesheet}
	return p
}

func (p *Player) IdleAnimation(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if p.direction == "LEFT" {
		op.GeoM.Scale(cameraScale*-1, cameraScale)
		op.GeoM.Translate(float64(p.idleAnim.sc.frameWidth)*cameraScale, 0)
	} else if p.direction == "RIGHT" {
		op.GeoM.Scale(cameraScale, cameraScale)
	}

	op.GeoM.Translate(p.posX, p.posY)

	cellX := p.idleAnim.sc.cellX
	cellY := p.idleAnim.sc.cellY

	i := (game.count / p.idleAnim.frameFrequency) % p.idleAnim.frameCount
	sx, sy := p.idleAnim.sc.getCol(cellX)+i*p.idleAnim.sc.frameWidth, p.idleAnim.sc.getRow(cellY)
	screen.DrawImage(p.spritesheet.SubImage(image.Rect(sx, sy, sx+p.idleAnim.sc.frameWidth, sy+p.idleAnim.sc.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) LeftWalkAnimation(screen *ebiten.Image) {
	p.posX -= .5 * cameraScale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cameraScale*-1, cameraScale)
	foo := float64(p.walkAnim.sc.frameWidth) * cameraScale

	op.GeoM.Translate(p.posX+foo, p.posY)

	cellX := p.walkAnim.sc.cellX
	cellY := p.walkAnim.sc.cellY

	i := (game.count / p.walkAnim.frameFrequency) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.sc.getRow(cellX)+i*p.walkAnim.sc.frameWidth, p.walkAnim.sc.getRow(cellY)
	screen.DrawImage(chickenSpritesheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.sc.frameWidth, sy+p.walkAnim.sc.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) RightWalkAnimation(screen *ebiten.Image) {
	p.posX += .5 * cameraScale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cameraScale, cameraScale)
	op.GeoM.Translate(p.posX, p.posY)

	cellX := p.walkAnim.sc.cellX
	cellY := p.walkAnim.sc.cellY

	i := (game.count / p.walkAnim.frameFrequency) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.sc.getRow(cellX)+i*p.walkAnim.sc.frameWidth, p.walkAnim.sc.getRow(cellY)
	screen.DrawImage(chickenSpritesheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.sc.frameWidth, sy+p.walkAnim.sc.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) JumpAnimation() {
	p.posX += p.vX
	p.posY -= p.vY
	p.vY += p.gravity
}

type Camera struct {
	width  int
	height int
	posX   int
	posY   int
}

type Game struct {
	keys      []ebiten.Key
	player    *Player
	count     int
	dbg       bool
	entities  []*Player
	mapLayers []Layer
	camera    Camera
}

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	g.count++
	// g.camera.posX = int(g.player.posX)
	// g.camera.posY = int(g.player.posY)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawCamera(g.mapLayers, screen)
	g.dbgMode(screen)
	g.player.IdleAnimation(screen)

	for _, keyPress := range g.keys {
		switch keyPress {
		case ebiten.KeyLeft:
			g.player.LeftWalkAnimation(screen)
			g.player.direction = "LEFT"
		case ebiten.KeyRight:
			g.player.RightWalkAnimation(screen)
			g.player.direction = "RIGHT"
		case ebiten.KeyJ:
			g.camera.posX -= 1
		case ebiten.KeyL:
			g.camera.posX += 1
		case ebiten.KeyI:
			g.camera.posY -= 1
		case ebiten.KeyK:
			g.camera.posY += 1
		case ebiten.KeySpace:
			g.player.JumpAnimation()
		default:
			g.player.vX = 0
		}
	}
}

func (g *Game) dbgMode(screen *ebiten.Image) {
	if g.dbg {
		dbgCamera := fmt.Sprintf("Camera Position: (%d, %d)", g.camera.posX, g.camera.posY)
		ebitenutil.DebugPrintAt(screen, dbgCamera, 0, 0)
		for _, entity := range g.entities {
			w := float32(entity.idleAnim.sc.frameWidth) * float32(cameraScale)
			h := float32(entity.idleAnim.sc.frameHeight) * float32(cameraScale)
			vector.StrokeRect(screen, float32(entity.posX), float32(entity.posY), w, h, float32(cameraScale), color.White, false)

			xCoord := strconv.FormatFloat(entity.posX, 'f', -1, 64)
			yCoord := strconv.FormatFloat(entity.posY, 'f', -1, 64)
			dbgXY := fmt.Sprintf("(%s, %s)", xCoord, yCoord)
			ebitenutil.DebugPrintAt(screen, dbgXY, int(entity.posX)+16, int(entity.posY)-16)
		}
	}
}

func LoadSpritesheets() {
	var err error
	chickenSpritesheet, _, err = ebitenutil.NewImageFromFile("./assets/Characters/chicken_sprites.png")
	groundSpritesheet, _, err = ebitenutil.NewImageFromFile("./assets/Tilesets/Hills.png")
	waterSpritesheet, _, err = ebitenutil.NewImageFromFile("./assets/Tilesets/Water.png")
	if err != nil {
		log.Fatal(err)
	}
}

type Layer struct {
	Data    []int  `json:"data"`
	Height  int    `json:"height"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Opacity int    `json:"opacity"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
	Width   int    `json:"width"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

type TileSet struct {
	Columns     int    `json:"columns"`
	FirstGid    int    `json:"firstgid"`
	Image       string `json:"image"`
	ImageHeight int    `json:"imageheight"`
	ImageWidth  int    `json:"imagewidth"`
	Margin      int    `json:"margin"`
	Name        string `json:"name"`
	Spacing     int    `json:"spacing"`
	TileCount   int    `json:"tilecount"`
	TileHeight  int    `json:"tileheight"`
	TileWidth   int    `json:"tilewidth"`
}

type TileMap struct {
	CompressionLevel int       `json:"compressionlevel"`
	Height           int       `json:"height"`
	Infinite         bool      `json:"infinite"`
	Layers           []Layer   `json:"layers"`
	NextLayerId      int       `json:"nextlayerid"`
	NextObjectId     int       `json:"nextobjectid"`
	Orientation      string    `json:"orientation"`
	RenderOrder      string    `json:"renderorder"`
	TiledVersion     string    `json:"tiledversion"`
	TiledHeight      int       `json:"tileheight"`
	TileSets         []TileSet `json:"tileSet"`
	TileWidth        int       `json:"tilewidth"`
	Type             string    `json:"type"`
	Version          string    `json:"version"`
	Width            int       `json:"width"`
}

func loadMap(file string) []Layer {
	mapFile, err := os.Open(file)
	if err != nil {
		log.Printf("%s,%s", "Can't find map", err)
	}
	defer mapFile.Close()

	byteArr, err := io.ReadAll(mapFile)

	var tm TileMap
	err = json.Unmarshal(byteArr, &tm)
	if err != nil {
		log.Println(err)
	}

	return tm.Layers
}

// Note: https://discourse.mapeditor.org/t/array-files-are-one-number-off-from-tile-set/1884/2
func (g *Game) drawCamera(layers []Layer, screen *ebiten.Image) {
	gPngWidth := groundSpritesheet.Bounds().Dx()
	wPngWidth := waterSpritesheet.Bounds().Dx()

	gTileCountX := gPngWidth / tileSize
	wTileCountX := wPngWidth / tileSize

	var cameraOffsetX float64 = float64(g.camera.posX) * float64(tileSize) * -1
	var cameraOffsetY float64 = float64(g.camera.posY) * float64(tileSize) * -1

	xCount := screenWidth / tileSize
	for _, layer := range layers {
		for i, globalTileID := range layer.Data {
			// can the XY of a 1d array can be its own function
			col := float64(i % xCount)
			row := float64(i / xCount)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(TilePos(col)+cameraOffsetX, TilePos(row)+cameraOffsetY)
			op.GeoM.Scale(float64(cameraScale), float64(cameraScale))
			if layer.Name == "water" {
				localTile := globalTileID - 37
				sx := (localTile % wTileCountX) * tileSize
				sy := (localTile / wTileCountX) * tileSize
				screen.DrawImage(waterSpritesheet.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
			} else if layer.Name == "floor" {
				localTile := globalTileID - 1
				sx := (localTile % gTileCountX) * tileSize
				sy := (localTile / gTileCountX) * tileSize
				screen.DrawImage(groundSpritesheet.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
			}
		}
	}

}

func init() {
	var entities []*Player

	LoadSpritesheets()

	p := CreatePlayer(chickenSpritesheet)

	entities = append(entities, &p)
	layers := loadMap("./maps/map1.tmj")

	c := Camera{width: cameraWidth, height: cameraHeight, posX: 0, posY: 0}
	game = &Game{player: &p, dbg: true, entities: entities, mapLayers: layers, camera: c}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Platformer")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
