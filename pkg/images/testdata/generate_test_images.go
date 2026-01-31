// +build ignore

package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"os"
)

func main() {
	// 1. Create valid JPEG (800x600)
	createValidJPEG("valid.jpg", 800, 600)

	// 2. Create valid PNG (640x480)
	createValidPNG("valid.png", 640, 480)

	// 3. Create valid WebP (500x500)
	createValidWebP("valid.webp", 500, 500)

	// 4. Create large JPEG (3000x3000) for resize testing
	createValidJPEG("large.jpg", 3000, 3000)

	// 5. Create too large JPEG (6MB+)
	createLargeJPEG("too_large.jpg", 4000, 4000)

	// 6. Create invalid text file
	createInvalidFile("invalid.txt")

	// 7. Create corrupted JPEG
	createCorruptedJPEG("corrupted.jpg")

	log.Println("Test images generated successfully!")
}

func createValidJPEG(filename string, width, height int) {
	img := createColoredImage(width, height, color.RGBA{255, 100, 100, 255})
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	if err != nil {
		log.Fatalf("Failed to encode JPEG %s: %v", filename, err)
	}
	log.Printf("Created %s (%dx%d)", filename, width, height)
}

func createValidPNG(filename string, width, height int) {
	img := createColoredImage(width, height, color.RGBA{100, 255, 100, 255})
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		log.Fatalf("Failed to encode PNG %s: %v", filename, err)
	}
	log.Printf("Created %s (%dx%d)", filename, width, height)
}

func createValidWebP(filename string, width, height int) {
	// Create PNG first, then user can convert to WebP manually if needed
	// For now, create a PNG with .webp extension marker
	img := createColoredImage(width, height, color.RGBA{100, 100, 255, 255})
	f, err := os.Create(filename + ".png")
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		log.Fatalf("Failed to encode WebP placeholder %s: %v", filename, err)
	}
	log.Printf("Created %s.png (placeholder for WebP) (%dx%d)", filename, width, height)
}

func createLargeJPEG(filename string, width, height int) {
	img := createGradientImage(width, height)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	// Use lower quality to make file larger
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatalf("Failed to encode large JPEG %s: %v", filename, err)
	}

	stat, _ := f.Stat()
	log.Printf("Created %s (%dx%d, %.2f MB)", filename, width, height, float64(stat.Size())/1024/1024)
}

func createInvalidFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	_, err = f.WriteString("This is not an image file, just plain text!")
	if err != nil {
		log.Fatalf("Failed to write to %s: %v", filename, err)
	}
	log.Printf("Created %s (invalid format)", filename)
}

func createCorruptedJPEG(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer f.Close()

	// Write JPEG magic bytes but corrupt data
	corruptData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00}
	corruptData = append(corruptData, []byte("CORRUPTED DATA HERE!!!")...)
	_, err = f.Write(corruptData)
	if err != nil {
		log.Fatalf("Failed to write corrupted JPEG %s: %v", filename, err)
	}
	log.Printf("Created %s (corrupted)", filename)
}

func createColoredImage(width, height int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func createGradientImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}
