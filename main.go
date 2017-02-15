package main

import (
	"fmt"
	obj "github.com/GrooveStomp/tiny-renderer/obj-loader"
	goimg "image"
	gocolor "image/color"
	"image/png"
	"math"
	"os"
	// "github.com/pkg/profile"
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
	if x >= img.Width {
		panic(fmt.Sprintf("x(%v) is out of range!", x))
	} else if y >= img.Height {
		panic(fmt.Sprintf("y(%v) is out of range!", y))
	}
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
	if x0 < 0 || x0 >= int(img.Width) {
		panic(fmt.Sprintf("x0(%v) is out of range!", x0))
	} else if x1 < 0 || x1 >= int(img.Width) {
		panic(fmt.Sprintf("x1(%v) is out of range!", x1))
	} else if y0 < 0 || y0 >= int(img.Height) {
		panic(fmt.Sprintf("y0(%v) is out of range!", y0))
	} else if y1 < 0 || y1 >= int(img.Height) {
		panic(fmt.Sprintf("y1(%v) is out of range!", y1))
	}

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

func (img *image) Triangle(t Triangle, c color) {
	v0, v1, v2 := t.v0, t.v1, t.v2

	swap := func(v0, v1 *obj.Vertex) {
		t := obj.Vertex{v0.X, v0.Y, v0.Z}
		v0.X = v1.X
		v0.Y = v1.Y
		v0.Z = v1.Z
		v1.X = t.X
		v1.Y = t.Y
		v1.Z = t.Z
	}

	if v0.Y > v1.Y { swap(&v0, &v1) }
	if v0.Y > v2.Y { swap(&v0, &v2) }
	if v1.Y > v2.Y { swap(&v1, &v2) }

	// img.Line(int(v0.X), int(v0.Y), int(v1.X), int(v1.Y), 0x00FF00FF)
	// img.Line(int(v1.X), int(v1.Y), int(v2.X), int(v2.Y), 0x00FF00FF)
	// img.Line(int(v2.X), int(v2.Y), int(v0.X), int(v0.Y), 0xFF0000FF)

	totalHeight := v2.Y - v0.Y

	for y := v0.Y; y < v1.Y; y++ {
		segmentHeight := v1.Y - v0.Y + 1
		alpha := (y - v0.Y) / totalHeight
		beta := (y - v0.Y) / segmentHeight

		a := obj.Add(v0, (obj.DotProduct(obj.Subtract(v2, v0), alpha)))
		b := obj.Add(v0, (obj.DotProduct(obj.Subtract(v1, v0), beta)))

		img.Line(int(a.X), int(y), int(b.X), int(y), c)
	}

	for y := v1.Y; y < v2.Y; y++ {
		segmentHeight := v2.Y - v1.Y + 1
		alpha := (y - v0.Y) / totalHeight
		beta := (y - v1.Y) / segmentHeight

		a := obj.Add(v0, (obj.DotProduct(obj.Subtract(v2, v0), alpha)))
		b := obj.Add(v1, (obj.DotProduct(obj.Subtract(v2, v1), beta)))

		if a.X > b.X { swap(&a, &b) }

		img.Line(int(a.X), int(y), int(b.X), int(y), c)
	}
}

//------------------------------------------------------------------------------

type Triangle struct {
	v0 obj.Vertex
	v1 obj.Vertex
	v2 obj.Vertex
}

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	width := uint(200)
	height := uint(200)

	img := MakeImage(width, height)
	img.Fill(0x000000FF)

	t0 := Triangle{obj.Vertex{10, 70, 0}, obj.Vertex{50, 160, 0}, obj.Vertex{70, 80, 0}}
	t1 := Triangle{obj.Vertex{180, 50, 0}, obj.Vertex{150, 1, 0}, obj.Vertex{70, 180, 0}}
	t2 := Triangle{obj.Vertex{180, 150, 0}, obj.Vertex{120, 160, 0}, obj.Vertex{130, 180, 0}}

	img.Triangle(t0, 0xFF0000FF)
	img.Triangle(t1, 0xFFFFFFFF)
	img.Triangle(t2, 0x00FF00FF)

	// model := obj.NewModel()
	// model.ReadFromFile("african_head.obj")

	// for i := 0; i < len(model.Faces); i++ {
	// 	face := model.Faces[i]
	// 	for j := 0; j < 3; j++ {
	// 		v0 := model.Vertices[face[j]-1]
	// 		v1 := model.Vertices[face[(j+1)%3]-1]

	// 		x0 := (v0.X + 1) * (float64(width-1) / 2)
	// 		y0 := (v0.Y + 1) * (float64(height-1) / 2)
	// 		x1 := (v1.X + 1) * (float64(width-1) / 2)
	// 		y1 := (v1.Y + 1) * (float64(height-1) / 2)
	// 		img.Line(int(x0), int(y0), int(x1), int(y1), 0xFFFFFFFF)
	// 	}
	// }

	img.WritePng("out.png")
}
