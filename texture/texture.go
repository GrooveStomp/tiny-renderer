package texture

import (
	gscolor "github.com/GrooveStomp/tiny-renderer/color"
	"image/png"
	"os"
)

type Texture struct {
	Texels []gscolor.Color
	Width  int
	Height int
}

func FromFile(filename string) (*Texture, error) {
	var t Texture

	file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return &t, err
	}

	img, err := png.Decode(file)
	file.Close()
	if err != nil {
		return &t, err
	}

	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y

	t.Texels = make([]gscolor.Color, width*height)
	t.Width = width
	t.Height = height

	ratio := float64(0xFF) / float64(0xFFFF)

	// Now copy image data into Texels.
	for y := 0; y < height; y++ {
		y2 := (height - 1) - y
		for x := 0; x < width; x++ {
			// Get RGBA components.
			r, g, b, a := img.At(x, y2).RGBA()
			rb := float64(r) * ratio
			gb := float64(g) * ratio
			bb := float64(b) * ratio
			ab := float64(a) * ratio
			t.Texels[y*width+x] = gscolor.NewColorRgba(byte(rb), byte(gb), byte(bb), byte(ab))
		}
	}

	return &t, nil
}
