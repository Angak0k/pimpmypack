//go:build ignore
// +build ignore

package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
)

func main() {
	// Create a 6000x6000 image to ensure > 5MB
	width := 6000
	height := 6000

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create complex pattern for larger file size
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			// Add noise to prevent compression
			r = uint8((int(r) + (x*y)%50) % 256)
			g = uint8((int(g) + (x+y)%50) % 256)
			b = uint8((int(b) + (x-y)%50) % 256)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	f, err := os.Create("too_large.jpg")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	// Use max quality to maximize file size
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatalf("Failed to encode: %v", err)
	}

	stat, _ := f.Stat()
	log.Printf("Created too_large.jpg (%.2f MB)", float64(stat.Size())/1024/1024)
}
