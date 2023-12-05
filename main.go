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
	screenWidth int  = 480
	screenHeight int  = 320
	tileSize     = 16
)

var (
	game               *Game
	chickenSpritesheet *ebiten.Image
	groundSpritesheet  *ebiten.Image
	waterSpritesheet   *ebiten.Image
	spriteScale        = float64(1) // Default 1
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

func CreatePlayer(spritesheet *ebiten.Image) Player {
	width, height := tileSize, tileSize
	playerWalkAnimationData := AnimationData{SpriteCell{0, 1, width, height}, 4, 8}
	playerIdleAnimationData := AnimationData{SpriteCell{0, 0, width, height}, 2, 64}

	// Will need to change screenHeight-height*2 when physics/jump is created
	p := Player{0, float64(screenHeight) - float64(height)*2, 0, 0, tileSize, tileSize, playerWalkAnimationData, playerIdleAnimationData, "RIGHT", 1, true, spritesheet}
	return p
}

func (p *Player) IdleAnimation(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if p.direction == "LEFT" {
		op.GeoM.Scale(spriteScale*-1, spriteScale)
		op.GeoM.Translate(float64(p.idleAnim.sc.frameWidth), 0)
	} else if p.direction == "RIGHT" {
		op.GeoM.Scale(spriteScale, spriteScale)
	}

	op.GeoM.Translate(p.posX, p.posY)

	cellX := p.idleAnim.sc.cellX
	cellY := p.idleAnim.sc.cellY

	i := (game.count / p.idleAnim.frameFrequency) % p.idleAnim.frameCount
	sx, sy := p.idleAnim.sc.getCol(cellX)+i*p.idleAnim.sc.frameWidth, p.idleAnim.sc.getRow(cellY)
	screen.DrawImage(p.spritesheet.SubImage(image.Rect(sx, sy, sx+p.idleAnim.sc.frameWidth, sy+p.idleAnim.sc.frameHeight)).(*ebiten.Image), op)

}

func (p *Player) LeftWalkAnimation(screen *ebiten.Image) {
	p.posX -= .5
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(spriteScale*-1, spriteScale)
	op.GeoM.Translate(p.posX+float64(p.walkAnim.sc.frameWidth), p.posY)

	cellX := p.walkAnim.sc.cellX
	cellY := p.walkAnim.sc.cellY

	i := (game.count / p.walkAnim.frameFrequency) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.sc.getRow(cellX)+i*p.walkAnim.sc.frameWidth, p.walkAnim.sc.getRow(cellY)
	screen.DrawImage(chickenSpritesheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.sc.frameWidth, sy+p.walkAnim.sc.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) RightWalkAnimation(screen *ebiten.Image) {
	p.posX += .5
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(spriteScale, spriteScale)
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

type Game struct {
	keys      []ebiten.Key
	player    *Player
	count     int
	dbg       bool
	entities  []*Player
	mapLayers []Layer
}

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	g.count++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	drawMap(g.mapLayers, screen)
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
		case ebiten.KeySpace:
			g.player.JumpAnimation()
		default:
			g.player.vX = 0
		}
	}
}


func (g *Game) dbgMode(screen *ebiten.Image) {
	if g.dbg {
		for _, entity := range g.entities {
			w := float32(entity.idleAnim.sc.frameWidth)
			h := float32(entity.idleAnim.sc.frameHeight)
			vector.StrokeRect(screen, float32(entity.posX), float32(entity.posY), w, h, 1, color.White, false)

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
func drawMap(layers []Layer, screen *ebiten.Image) {
	gPngWidth := groundSpritesheet.Bounds().Dx()
	wPngWidth := waterSpritesheet.Bounds().Dx()

	gTileCount := gPngWidth / tileSize
	wTileCount := wPngWidth / tileSize

	xCount := screenWidth / tileSize
	for _, layer := range layers {
		for i, globalTileID := range layer.Data {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((i%xCount)*tileSize), float64((i/xCount)*tileSize))
			if layer.Name == "water" {
				tile := globalTileID - 37
				sx := (tile % wTileCount) * tileSize
				sy := (tile / wTileCount) * tileSize
				screen.DrawImage(waterSpritesheet.SubImage(image.Rect(sx, sy, sx+tileSize, sy+tileSize)).(*ebiten.Image), op)
			} else if layer.Name == "floor" {
				tile := globalTileID - 1
				sx := (tile % gTileCount) * tileSize
				sy := (tile / gTileCount) * tileSize
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
	game = &Game{player: &p, dbg: true, entities: entities, mapLayers: layers}
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
