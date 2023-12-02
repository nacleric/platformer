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


https://gamedev.stackexchange.com/a/29618
http://higherorderfun.com/blog/2012/05/20/the-guide-to-implementing-2d-platformers/

Ground detection
https://www.youtube.com/shorts/706pUVt3xwg

gravity
https://love2d.org/forums/viewtopic.php?p=175824#p175824
