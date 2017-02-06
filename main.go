package main

import (
	"image"
	"fmt"
	"image/color"
	"image/png"
	"os"
)

func Set(img *image.RGBA, x, y int, c color.Color) error {
	// Rectangle returns color.Opaque for in-bounds pixels, and color.Transparent
	// for out-of-bounds pixels.
	if img.Rect.At(x,y) == color.Transparent {
		return fmt.Errorf("Pixel at (%v, %v) is out-of-bounds", x, y)
	}

	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	index := (y - img.Rect.Min.Y) * img.Stride + (x - img.Rect.Min.X) * 4
	r,g,b,a := c.RGBA()
	img.Pix[index + 0] = uint8(r)
	img.Pix[index + 1] = uint8(g)
	img.Pix[index + 2] = uint8(b)
	img.Pix[index + 3] = uint8(a)

	return nil
}

func main() {
	topLeft := image.Point{0,0}
	bottomRight := image.Point{5,5}
	img := image.NewRGBA(image.Rectangle{topLeft, bottomRight})

	err := Set(img, 4,4, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile("out.png", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	png.Encode(file, img)
	file.Close()
}
