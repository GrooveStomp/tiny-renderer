package main

import (
	goimg "image"
	gocolor "image/color"
	"image/png"
	"math"
	"os"
	//"github.com/pkg/profile"
)

//------------------------------------------------------------------------------

type color uint32

func (c color) Rgba() (uint32, uint32, uint32, uint32) {
	r, g, b, a := uint32(c), uint32(c), uint32(c), uint32(c)
	r >>= 24
	g >>= 16
	b >>= 8

	return r, g, b, a
}

func (c *color) Set(r, g, b, a uint32) {
	t := r<<24 |
		g<<16 |
		b<<8 |
		a

	*c = color(t)
}

//------------------------------------------------------------------------------

type image struct {
	Pixels []color
	Width  uint
	Height uint
}

func MakeImage(width, height uint) *image {
	var img image
	img.Pixels = make([]color, width*height)
	img.Width = width
	img.Height = height
	return &img
}

func (img image) Get(x, y uint) color {
	return img.Pixels[y*img.Width+x]
}

func (img *image) Set(x, y uint, c color) {
	img.Pixels[y*img.Width+x] = c
}

func (img *image) FlipVertical() {
	for y := uint(0); y < (img.Height/2)+1; y++ {
		y2 := (img.Height - 1) - y
		for x := uint(0); x < img.Width; x++ {
			c1 := img.Get(x, y)
			c2 := img.Get(x, y2)
			img.Set(x, y, c2)
			img.Set(x, y, c1)
		}
	}
}

func (img *image) Fill(c color) {
	for y := uint(0); y < img.Height; y++ {
		for x := uint(0); x < img.Width; x++ {
			img.Set(x, y, c)
		}
	}
}

func (src *image) WritePng(filename string) error {
	img := goimg.NewRGBA(goimg.Rectangle{goimg.Point{0, 0}, goimg.Point{int(src.Width), int(src.Height)}})

	for y := uint(0); y < src.Height; y++ {
		y2 := (src.Height - 1) - y
		for x := uint(0); x < src.Width; x++ {
			r, g, b, a := src.Get(x, y).Rgba()
			img.Set(int(x), int(y2), gocolor.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	png.Encode(file, img)
	file.Close()

	return nil
}

func (img *image) Line(x0, y0, x1, y1 int, c color) {
	steep := false
	var t int

	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		t = x0
		x0 = y0
		y0 = t
		t = x1
		x1 = y1
		y1 = t
		steep = true
	}

	if x0 > x1 {
		t = x0
		x0 = x1
		x1 = t
		t = y0
		y0 = y1
		y1 = t
	}

	dx := x1 - x0
	dy := y1 - y0
	derr2 := int(math.Abs(float64(dy)) * 2)
	err2 := 0

	y := y0

	for x := x0; x <= x1; x++ {
		if steep {
			img.Set(uint(y), uint(x), c)
		} else {
			img.Set(uint(x), uint(y), c)
		}

		err2 += derr2
		if err2 > dx {
			if y1 > y0 {
				y += 1
			} else {
				y -= 1
			}
			err2 -= (dx * 2)
		}
	}
}

//------------------------------------------------------------------------------

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	img := MakeImage(100, 100)
	img.Fill(0x000000FF)

	//for i := 0; i < 100000; i++ {
	img.Line(13, 20, 80, 40, 0xFFFFFFFF)
	img.Line(20, 13, 40, 80, 0xFF0000FF)
	img.Line(80, 40, 13, 20, 0xFF0000FF)
	//}

	img.WritePng("out.png")
}
