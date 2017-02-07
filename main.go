package main

import (
	"fmt"
	goimg "image"
	gocolor "image/color"
	"image/png"
	"math"
	"os"
)

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

type image struct {
	Pixels []color
	Width  uint
	Height uint
}

func (img image) Get(x, y uint) color {
	return img.Pixels[y*img.Width+x]
}

func (img *image) Set(x, y uint, c color) {
	img.Pixels[y*img.Width+x] = c
}

func (img *image) FlipVertical() {
	for y := uint(0); y < img.Height; y++ {
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

func MakeImage(width, height uint) *image {
	var img image
	img.Pixels = make([]color, width*height)
	img.Width = width
	img.Height = height
	return &img
}

func (img *image) Line(x0, y0, x1, y1 uint, c color) {
	steep := false
	var tint uint
	var t float64

	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		tint = x0
		x0 = y0
		y0 = tint
		tint = x1
		x1 = y1
		y1 = tint
		steep = true
	}

	if x0 > x1 {
		tint = x0
		x0 = x1
		x1 = tint
		tint = y0
		y0 = y1
		y1 = tint
	}

	for x := x0; x <= x1; x++ {
		t = float64(x-x0) / float64(x1-x0)
		y := float64(y0)*(1.0-t) + float64(y1)*t

		if steep {
			img.Set(uint(y), x, c)
		} else {
			img.Set(x, uint(y), c)
		}
	}
}

func main() {
	img := MakeImage(100, 100)
	img.Fill(0x00000000)
	img.Line(13, 20, 80, 40, 0xFFFFFFFF)
	img.Line(20, 13, 40, 80, 0xFF0000FF)
	img.Line(80, 40, 13, 20, 0xFF0000FF)
	img.FlipVertical()

	goImg := goimg.NewRGBA(goimg.Rectangle{goimg.Point{0, 0}, goimg.Point{100, 100}})
	for y := uint(0); y < img.Height; y++ {
		for x := uint(0); x < img.Width; x++ {
			r, g, b, a := img.Get(x, y).Rgba()
			goImg.Set(int(x), int(y), gocolor.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	file, err := os.OpenFile("out.png", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	png.Encode(file, goImg)
	file.Close()
	fmt.Println("hi")
}
