package main

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type SpriteData struct {
	frameOX     int // column of spritesheet Ex: 0 is first col 16 is 2nd col
	frameOY     int // row of spritesheet Ex: 16 is 2nd row
	frameWidth  int // Size of Sprite frame (most likely 16x16)
	frameHeight int
	frameCount  int // Total number of columns for specific row
}

type Player struct {
	posX      float64
	posY      float64
	vX        float64
	vY        float64
	walkAnim  SpriteData
	idleAnim  SpriteData
	direction string // (LEFT, RIGHT)
}

func (p *Player) IdleAnimation(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	if p.direction == "LEFT" {
		op.GeoM.Scale(-1, 1)
	} else if p.direction == "RIGHT" {
		op.GeoM.Scale(1, 1)
	}

	op.GeoM.Translate(p.posX, p.posY)

	i := (game.count / 8) % p.idleAnim.frameCount
	sx, sy := p.idleAnim.frameOX+i*p.idleAnim.frameWidth, p.idleAnim.frameOY
	screen.DrawImage(chickenSpriteSheet.SubImage(image.Rect(sx, sy, sx+p.idleAnim.frameWidth, sy+p.idleAnim.frameHeight)).(*ebiten.Image), op)

}

func (p *Player) LeftWalkAnimation(screen *ebiten.Image) {
	p.posX -= .5
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-1, 1) // Probably something wrong with this
	op.GeoM.Translate(p.posX, p.posY)

	i := (game.count / 8) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.frameOX+i*p.walkAnim.frameWidth, p.walkAnim.frameOY
	screen.DrawImage(chickenSpriteSheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.frameWidth, sy+p.walkAnim.frameHeight)).(*ebiten.Image), op)
}

func (p *Player) RightWalkAnimation(screen *ebiten.Image) {
	p.posX += .5
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(p.posX, p.posY)

	i := (game.count / 8) % p.walkAnim.frameCount
	sx, sy := p.walkAnim.frameOX+i*p.walkAnim.frameWidth, p.walkAnim.frameOY
	screen.DrawImage(chickenSpriteSheet.SubImage(image.Rect(sx, sy, sx+p.walkAnim.frameWidth, sy+p.walkAnim.frameHeight)).(*ebiten.Image), op)
}

const (
	screenWidth  = 320
	screenHeight = 240
)

var (
	chickenSpriteSheet *ebiten.Image
	game               *Game
)

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

func (g *Game) Draw(screen *ebiten.Image) {
	// g.player
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
		// fmt.Println(g.entities)
		for _, entity := range g.entities {
			fmt.Println(entity.posX)
			vector.DrawFilledRect(screen, float32(entity.posX), float32(entity.posY), 16, 16, color.RGBA{100, 0, 0, 0}, false)
		}
	}
}

// One time initilizations like assets in here
func init() {
	var err error
	chickenSpriteSheet, _, err = ebitenutil.NewImageFromFile("./assets/Characters/chicken_sprites.png")
	if err != nil {
		log.Fatal(err)
	}

	var entities []*Player
	playerWalkAnimationData := SpriteData{0, 16, 16, 16, 4}
	playerIdleAnimationData := SpriteData{0, 0, 16, 16, 2}
	p := Player{50, 50, 0, 0, playerWalkAnimationData, playerIdleAnimationData, "RIGHT"}
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
