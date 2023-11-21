package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	walkAnim    AnimationData
	idleAnim    AnimationData
	direction   string // (LEFT, RIGHT) need to find a better way to represent enums
	gravity     float64
	onGround    bool
	spritesheet *ebiten.Image
}

func DrawPlayer(spritesheet *ebiten.Image) Player {
	width, height := 16, 16
	playerWalkAnimationData := AnimationData{SpriteCell{0, 1, width, height}, 4, 8}
	playerIdleAnimationData := AnimationData{SpriteCell{0, 0, width, height}, 2, 64}

	// Will need to change screenHeight-height*2 when physics/jump is created
	p := Player{0, screenHeight - float64(height)*2, 0, 0, playerWalkAnimationData, playerIdleAnimationData, "RIGHT", 1, true, spritesheet}
	return p
}

func (p *Player) IdleAnimation(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	if p.direction == "LEFT" {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(p.idleAnim.sc.frameWidth), 0)
	} else if p.direction == "RIGHT" {
		op.GeoM.Scale(1, 1)
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
	op.GeoM.Scale(-1, 1)
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
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(p.posX, p.posY)

	cellX := p.walkAnim.sc.cellX
	cellY := p.walkAnim.sc.cellY

	i := (game.count / p.walkAnim.frameFrequency) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.sc.getRow(cellX)+i*p.walkAnim.sc.frameWidth, p.walkAnim.sc.getRow(cellY)
	screen.DrawImage(chickenSpritesheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.sc.frameWidth, sy+p.walkAnim.sc.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) JumpAnimation(screen *ebiten.Image) {
	p.posX += p.vX
	p.posY -= p.vY
	p.vY += p.gravity
}

type Game struct {
	keys     []ebiten.Key
	player   *Player
	count    int
	dbg      bool
	entities []*Player
}

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	g.count++
	return nil
}

func drawGround(screen *ebiten.Image) {
	sc := SpriteCell{2, 3, 16, 16}

	// Collission box
	vector.DrawFilledRect(screen, 0, screenHeight-float32(sc.frameHeight), screenWidth, float32(sc.frameHeight), color.RGBA{0, 100, 0, 0}, false)

	x0, y0 := sc.getCol(sc.cellX), sc.getRow(sc.cellY)
	x1, y1 := x0+sc.frameWidth, y0+sc.frameHeight

	numberOfTiles := screenWidth / sc.frameWidth
	fmt.Println(numberOfTiles)
	for i := 0; i <= numberOfTiles; i++ {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(1, 1)
		op.GeoM.Translate(float64(i)*float64(numberOfTiles), screenHeight-float64(sc.frameHeight))
		screen.DrawImage(groundSpritesheet.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image), op)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.dbgMode(screen)
	drawGround(screen)
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
			g.player.JumpAnimation(screen)
		default:
			g.player.vX = 0
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func (g *Game) dbgMode(screen *ebiten.Image) {
	if g.dbg {
		for _, entity := range g.entities {
			vector.DrawFilledRect(screen, float32(entity.posX), float32(entity.posY), 16, 16, color.RGBA{100, 0, 0, 0}, false)

			xCoord := strconv.FormatFloat(entity.posX, 'f', -1, 64)
			yCoord := strconv.FormatFloat(entity.posY, 'f', -1, 64)
			dbgXY := fmt.Sprintf("(%s, %s)", xCoord, yCoord)
			ebitenutil.DebugPrintAt(screen, dbgXY, int(entity.posX)+16, int(entity.posY)-16)
		}
	}
}

const (
	screenWidth  = 320
	screenHeight = 240
)

var (
	game               *Game
	chickenSpritesheet *ebiten.Image
	groundSpritesheet  *ebiten.Image
)

func LoadSpritesheets() {
	var err error
	chickenSpritesheet, _, err = ebitenutil.NewImageFromFile("./assets/Characters/chicken_sprites.png")
	groundSpritesheet, _, err = ebitenutil.NewImageFromFile("./assets/Tilesets/Hills.png")
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	var entities []*Player

	LoadSpritesheets()

	p := DrawPlayer(chickenSpritesheet)

	entities = append(entities, &p)
	game = &Game{player: &p, dbg: true, entities: entities}
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Platformer")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
