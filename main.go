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
	"github.com/GrooveStomp/tiny-renderer/color"
	"github.com/GrooveStomp/tiny-renderer/geometry"
)

//------------------------------------------------------------------------------

type zbuffer struct {
	Buffer []float64
	Width int
	Height int
}

func MakeZBuffer(width, height int) *zbuffer {
	var buf zbuffer
	buf.Buffer = make([]float64, width * height)
	buf.Width = width
	buf.Height = height

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			buf.Set(x, y, math.Inf(-1))
		}
	}

	return &buf
}

func (buf *zbuffer) Get(x, y int) float64 {
	return buf.Buffer[y*buf.Width+x]
}

func (buf *zbuffer) Set(x, y int, z float64) {
	if x >= buf.Width {
		panic(fmt.Sprintf("x(%v) is out of range!", x))
	} else if y >= buf.Height {
		panic(fmt.Sprintf("y(%v) is out of range!", y))
	}
	buf.Buffer[y*buf.Width+x] = z
}

//------------------------------------------------------------------------------

type image struct {
	Pixels []color.Color
	Width  int
	Height int
}

func MakeImage(width, height int) *image {
	var img image
	img.Pixels = make([]color.Color, width*height)
	img.Width = width
	img.Height = height
	return &img
}

func (img image) Get(x, y int) color.Color {
	return img.Pixels[y*img.Width+x]
}

func (img *image) Set(x, y int, c color.Color) {
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

func (img *image) Fill(c color.Color) {
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

func (img *image) HorizontalLine(x0, x1 int, y int, c color.Color) {
	if x0 > x1 {
		t := x0
		x0 = x1
		x1 = t
	}

	for x := x0; x <= x1; x++ {
		img.Set(x, y, c)
	}
}

func (img *image) Line(x0, y0, x1, y1 int, c color.Color) {
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

func RasterTriangle(t geometry.Triangle, c color.Color, img *image, zbuf *zbuffer) {
	v0, v1, v2 := t.V0, t.V1, t.V2

	swap := func(v0, v1 *geometry.Vertex3) {
		t := geometry.Vertex3{v0.X, v0.Y, v0.Z}
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

		a := geometry.Add(v0, geometry.Multiply(geometry.Subtract(v2, v0), alpha))
		b := geometry.Add(v0, geometry.Multiply(geometry.Subtract(v1, v0), beta))

		if a.X > b.X {
			swap(&a, &b)
		}

		barycentricA := t.Barycentric(a)
		barycentricB := t.Barycentric(b)

		if barycentricA.X == -1 || barycentricB.X == -1 {
			;
		} else {
			// Now set the Z-Buffer
			az := (v0.Z * barycentricA.X) + (v1.Z * barycentricA.Y) + (v2.Z * barycentricA.Z)
			bz := (v0.Z * barycentricB.X) + (v1.Z * barycentricB.Y) + (v2.Z * barycentricB.Z)

			delta := math.Abs(b.X - a.X)

			for x := a.X; x <= b.X; x++ {
				tp := (x - a.X)
				t := tp
				if tp > 0 {
					t = tp / delta
				}

				lerp := (az * (float64(1)-t)) + (bz * t)
				if zbuf.Get(int(x), int(y)) < lerp {
					zbuf.Set(int(x), int(y), lerp)
					img.Set(int(x), int(y), c)
				}
			}
		}

		// img.HorizontalLine(int(a.X), int(b.X), int(y), c)
	}

	for y := v1.Y; y <= v2.Y; y++ {
		segmentHeight := v2.Y - v1.Y + 1
		alpha := (y - v0.Y) / totalHeight
		beta := (y - v1.Y) / segmentHeight

		a := geometry.Add(v0, geometry.Multiply(geometry.Subtract(v2, v0), alpha))
		b := geometry.Add(v1, geometry.Multiply(geometry.Subtract(v2, v1), beta))

		if a.X > b.X {
			swap(&a, &b)
		}

		barycentricA := t.Barycentric(a)
		barycentricB := t.Barycentric(b)

		if barycentricA.X == -1 || barycentricB.X == -1 {
			;
		} else {
			// Now set the Z-Buffer
			az := (v0.Z * barycentricA.X) + (v1.Z * barycentricA.Y) + (v2.Z * barycentricA.Z)
			bz := (v0.Z * barycentricB.X) + (v1.Z * barycentricB.Y) + (v2.Z * barycentricB.Z)

			delta := math.Abs(b.X - a.X)

			for x := a.X; x <= b.X; x++ {
				tp := (x - a.X)
				t := tp
				if tp > 0 {
					t = tp / delta
				}

				lerp := (az * (float64(1)-t)) + (bz * t)
				if zbuf.Get(int(x), int(y)) < lerp {
					zbuf.Set(int(x), int(y), lerp)
					img.Set(int(x), int(y), c)
				}
			}
		}

		//img.HorizontalLine(int(a.X), int(b.X), int(y), c)
	}
}

//------------------------------------------------------------------------------

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

	lightDir := geometry.Vertex3{0, 0, -1}

	width := 1000
	height := 1000

	img := MakeImage(width, height)
	img.Fill(0x000000FF)

	zbuf := MakeZBuffer(width, height)

	model := obj.NewModel()
	model.ReadFromFile(objFile)

	for i := 0; i < len(model.Faces); i++ {
		face := model.Faces[i]

		var screenCoords [3]geometry.Vertex3
		var worldCoords [3]geometry.Vertex3

		for j := 0; j < 3; j++ {
			v := model.Vertices[face[j]-1]
			x := (v.X + 1) * (float64(width-1) / 2)
			y := (v.Y + 1) * (float64(height-1) / 2)
			z := (v.Z + 1) * (float64(height-1) / 2) // TODO(AARON): Need proper viewing volume.
			screenCoords[j] = geometry.Vertex3{x, y, z}
			worldCoords[j] = v
		}

		n := geometry.CrossProduct(geometry.Subtract(worldCoords[2], worldCoords[0]), geometry.Subtract(worldCoords[1], worldCoords[0]))
		n.Normalize()
		intensity := geometry.DotProduct(n, lightDir)
		if intensity > 0 {
			var c color.Color
			c = color.Multiply(color.White, intensity)
			c.SetAlpha(byte(0xFF))
			RasterTriangle(geometry.Triangle{screenCoords[0], screenCoords[1], screenCoords[2]}, c, img, zbuf)
		}
	}

	img.WritePng(outFile)
}
