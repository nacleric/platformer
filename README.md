// func readAsset(assetPath string) []byte {
// 	file, err := os.Open(assetPath)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 	}
// 	defer file.Close()

// 	// Read the file content into a byte slice
// 	fileBytes, err := io.ReadAll(file)
// 	if err != nil {
// 		fmt.Println("Error reading file:", err)
// 	}
// 	return fileBytes
// }

/*
func loadAsset() *ebiten.Image {
	fileBytes := readAsset("./assets/Characters/chicken_sprites.png")
	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		log.Fatal(err)
	}

	origEbitenImage := ebiten.NewImageFromImage(img)
	s := origEbitenImage.Bounds().Size()
	ebitenImage = ebiten.NewImage(s.X, s.Y)

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(0.5)
	ebitenImage.DrawImage(origEbitenImage, op)

	return ebitenImage
}
*/

/*
func drawPlayer(screen *ebiten.Image, p Player) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(playerScaleWidth, playerScaleHeight)
	op.GeoM.Translate(p.posX, p.posY)

	playerImage := ebiten.NewImage(imageWidth, imageHeight)
	playerImage.Fill(p.color)
	screen.DrawImage(playerImage, op)
}
*/

https://gamedev.stackexchange.com/a/29618
http://higherorderfun.com/blog/2012/05/20/the-guide-to-implementing-2d-platformers/
