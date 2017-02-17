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

func NewColor(c color) color {
	return c
}

func NewColorRgba(r, g, b, a byte) color {
	t := uint32(r)<<24 |
		uint32(g)<<16 |
		uint32(b)<<8 |
		uint32(a)

	return color(t)
}

func NewColorFloat64(r, g, b, a float64) color {
	r1 := r * float64(255)
	g1 := g * float64(255)
	b1 := b * float64(255)
	a1 := a * float64(255)

	return NewColorRgba(byte(r1), byte(g1), byte(b1), byte(a1))
}

func (c color) Rgba() (byte, byte, byte, byte) {
	r, g, b, a := uint32(c), uint32(c), uint32(c), uint32(c)
	r >>= 24
	g >>= 16
	b >>= 8

	return byte(r), byte(g), byte(b), byte(a)
}

func (c *color) Set(r, g, b, a byte) {
	new := NewColorRgba(r, g, b, a)
	*c = new
}

func (c *color) SetAlpha(a byte) {
	*c = color(uint32(*c) | uint32(a))
}

func (c color) Multiply(k float64) color {
	r, g, b, a := c.Rgba()
	r = byte(float64(r) * k)
	g = byte(float64(g) * k)
	b = byte(float64(b) * k)
	a = byte(float64(a) * k)

	return NewColorRgba(r, g, b, a)
}

func (c color) Add(other color) color {
	r, g, b, a := c.Rgba()
	r2, g2, b2, a2 := other.Rgba()
	r += r2
	g += g2
	b += b2
	a += a2

	return NewColorRgba(r, g, b, a)
}

func (c color) String() string {
	r, g, b, a := c.Rgba()
	return fmt.Sprintf("%v,%v,%v,%v", r, g, b, a)
}

var red color = NewColorFloat64(1.0, 0.0, 0.0, 1.0)
var green color = NewColorFloat64(0.0, 1.0, 0.0, 1.0)
var blue color = NewColorFloat64(0.0, 0.0, 1.0, 1.0)
var white color = NewColorFloat64(1.0, 1.0, 1.0, 1.0)

//------------------------------------------------------------------------------

type image struct {
	Pixels []color
	Width  int
	Height int
}

func MakeImage(width, height int) *image {
	var img image
	img.Pixels = make([]color, width*height)
	img.Width = width
	img.Height = height
	return &img
}

func (img image) Get(x, y int) color {
	return img.Pixels[y*img.Width+x]
}

func (img *image) Set(x, y int, c color) {
	if x >= img.Width {
		panic(fmt.Sprintf("x(%v) is out of range!", x))
	} else if y >= img.Height {
		panic(fmt.Sprintf("y(%v) is out of range!", y))
	}
	img.Pixels[y*img.Width+x] = c
}

func (img *image) FlipVertical() {
	for y := 0; y < (img.Height/2)+1; y++ {
		y2 := (img.Height - 1) - y
		for x := 0; x < img.Width; x++ {
			c1 := img.Get(x, y)
			c2 := img.Get(x, y2)
			img.Set(x, y, c2)
			img.Set(x, y, c1)
		}
	}
}

func (img *image) Fill(c color) {
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			img.Set(x, y, c)
		}
	}
}

func (src *image) WritePng(filename string) error {
	img := goimg.NewRGBA(goimg.Rectangle{goimg.Point{0, 0}, goimg.Point{int(src.Width), int(src.Height)}})

	for y := 0; y < src.Height; y++ {
		y2 := (src.Height - 1) - y
		for x := 0; x < src.Width; x++ {
			r, g, b, a := src.Get(x, y).Rgba()
			img.Set(x, y2, gocolor.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
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

func (img *image) HorizontalLine(x0, x1 int, y int, c color) {
	if x0 > x1 {
		t := x0
		x0 = x1
		x1 = t
	}

	for x := x0; x <= x1; x++ {
		img.Set(x, y, c)
	}
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
			img.Set(y, x, c)
		} else {
			img.Set(x, y, c)
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

func (img *image) Triangle(t Triangle, c0, c1, c2 color) {
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

	// Bubblesort. Sort vertices by Y-component, ascending order.
	if v0.Y > v1.Y {
		swap(&v0, &v1)
	}
	if v0.Y > v2.Y {
		swap(&v0, &v2)
	}
	if v1.Y > v2.Y {
		swap(&v1, &v2)
	}

	totalHeight := v2.Y - v0.Y

	for y := v0.Y; y <= v1.Y; y++ {
		segmentHeight := v1.Y - v0.Y + 1
		alpha := (y - v0.Y) / totalHeight
		beta := (y - v0.Y) / segmentHeight

		a := obj.Add(v0, obj.Multiply(obj.Subtract(v2, v0), alpha))
		b := obj.Add(v0, obj.Multiply(obj.Subtract(v1, v0), beta))

		if a.X > b.X {
			swap(&a, &b)
		}

		barycentricA := t.Barycentric(a)
		barycentricB := t.Barycentric(b)

		colorA := c0.Multiply(barycentricA.X) + c1.Multiply(barycentricA.Y) + c2.Multiply(barycentricA.Z)
		colorB := c0.Multiply(barycentricB.X) + c1.Multiply(barycentricB.Y) + c2.Multiply(barycentricB.Z)

		delta := math.Abs(b.X - a.X)
		for x := a.X; x <= b.X; x++ {
			tp := (x - a.X)
			t := tp
			if tp > 0 {
				t = tp / delta
			}

			ap := colorA.Multiply(float64(1)-t)
			bp := colorB.Multiply(t)
			lerp := ap.Add(bp)

			if barycentricA.X == -1 || barycentricB.X == -1 {
				img.Set(int(x), int(y), white)
			} else {
				img.Set(int(x), int(y), lerp)
			}
		}

		// img.HorizontalLine(int(a.X), int(b.X), int(y), c)
	}

	for y := v1.Y; y <= v2.Y; y++ {
		segmentHeight := v2.Y - v1.Y + 1
		alpha := (y - v0.Y) / totalHeight
		beta := (y - v1.Y) / segmentHeight

		a := obj.Add(v0, obj.Multiply(obj.Subtract(v2, v0), alpha))
		b := obj.Add(v1, obj.Multiply(obj.Subtract(v2, v1), beta))

		if a.X > b.X {
			swap(&a, &b)
		}

		barycentricA := t.Barycentric(a)
		barycentricB := t.Barycentric(b)

		colorA := c0.Multiply(barycentricA.X) + c1.Multiply(barycentricA.Y) + c2.Multiply(barycentricA.Z)
		colorB := c0.Multiply(barycentricB.X) + c1.Multiply(barycentricB.Y) + c2.Multiply(barycentricB.Z)

		delta := math.Abs(b.X - a.X)
		for x := a.X; x <= b.X; x++ {
			tp := (x - a.X)
			t := tp
			if tp > 0 {
				t = tp / delta
			}

			ap := colorA.Multiply(float64(1)-t)
			bp := colorB.Multiply(t)
			lerp := ap.Add(bp)

			if barycentricA.X == -1 || barycentricB.X == -1 {
				img.Set(int(x), int(y), white)
			} else {
				img.Set(int(x), int(y), lerp)
			}
		}
	}
}

//------------------------------------------------------------------------------

type Triangle struct {
	v0 obj.Vertex
	v1 obj.Vertex
	v2 obj.Vertex
}

func (t *Triangle) Barycentric(p obj.Vertex) obj.Vertex {
	v0 := obj.Subtract(t.v1, t.v0)
	v1 := obj.Subtract(t.v2, t.v0)
	v2 := obj.Subtract(p, t.v0)

	d00 := obj.DotProduct(v0, v0)
	d01 := obj.DotProduct(v0, v1)
	d11 := obj.DotProduct(v1, v1)
	d20 := obj.DotProduct(v2, v0)
	d21 := obj.DotProduct(v2, v1)

	denom := (d00 * d11) - (d01 * d01)

	v := (d11 * d20 - d01 * d21) / denom
	w := (d00 * d21 - d01 * d20) / denom
	u := float64(1) - v - w

	sum := u + v + w
	if sum < 0.99 || sum > 1.001 {
		return obj.Vertex{-1, -1, -1}
	}

	return obj.Vertex{u, v, w}
}

func usage(progName string) {
	fmt.Printf("Usage: %s obj_file output_img.png\n", progName)
	os.Exit(0)
}

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	if len(os.Args) < 3 {
		usage(os.Args[0])
	}

	objFile := os.Args[1]
	outFile := os.Args[2]

	lightDir := obj.Vertex{0, 0, -1}

	width := 1000
	height := 1000

	img := MakeImage(width, height)
	img.Fill(0x000000FF)

	model := obj.NewModel()
	model.ReadFromFile(objFile)

	for i := 0; i < len(model.Faces); i++ {
		face := model.Faces[i]

		var screenCoords [3]obj.Vertex
		var worldCoords [3]obj.Vertex

		for j := 0; j < 3; j++ {
			v := model.Vertices[face[j]-1]
			x := (v.X + 1) * (float64(width-1) / 2)
			y := (v.Y + 1) * (float64(height-1) / 2)
			screenCoords[j] = obj.Vertex{x, y, float64(0)}
			worldCoords[j] = v
		}

		n := obj.CrossProduct(obj.Subtract(worldCoords[2], worldCoords[0]), obj.Subtract(worldCoords[1], worldCoords[0]))
		n.Normalize()
		intensity := obj.DotProduct(n, lightDir)
		if intensity > 0 {
			var c color
			c = white.Multiply(intensity)
			c.SetAlpha(byte(0xFF))
			img.Triangle(Triangle{screenCoords[0], screenCoords[1], screenCoords[2]}, c, c, c)
		}
	}

	img.WritePng(outFile)
}
