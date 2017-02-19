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
	"github.com/GrooveStomp/tiny-renderer/texture"
)

var LIGHT_DIR = geometry.Vertex3{0, 0, -1}

//------------------------------------------------------------------------------

type zbuffer struct {
	Buffer []float64
	Width  int
	Height int
}

func MakeZBuffer(width, height int) *zbuffer {
	var buf zbuffer
	buf.Buffer = make([]float64, width*height)
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

func FromTexture(t *texture.Texture) *image {
	var i image
	i.Pixels = t.Texels
	i.Width = t.Width
	i.Height = t.Height
	return &i
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

	swap := func(a, b *int) {
		t := *a
		*a = *b
		*b = t
	}

	steep := false
	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		swap(&x0, &y0)
		swap(&x1, &y1)
		steep = true
	}

	if x0 > x1 {
		swap(&x0, &x1)
		swap(&y0, &y1)
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

func RasterPolygon(p *Polygon, img *image, zbuf *zbuffer, step float64) {
	v0, v1, v2 := p.Vertices[0], p.Vertices[1], p.Vertices[2]
	if v0.Y == v1.Y && v0.Y == v2.Y { return }

	t := geometry.Triangle{p.Vertices[0], p.Vertices[1], p.Vertices[2]}

	swap := func(v0, v1 *geometry.Vertex3) {
		t := geometry.Vertex3{v0.X, v0.Y, v0.Z}
		*v0 = geometry.Vertex3{v1.X, v1.Y, v1.Z}
		*v1 = geometry.Vertex3{t.X, t.Y, t.Z}
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

	intensity := geometry.DotProduct(p.Normals[0], LIGHT_DIR)
	totalHeight := v2.Y - v0.Y

	for y := 0.0; y < totalHeight; y += step {
		isSecondHalf := y > (v1.Y - v0.Y) || v1.Y == v0.Y

		firstSegmentHeight := v1.Y - v0.Y
		secondSegmentHeight := v2.Y - v1.Y
		segmentHeight := firstSegmentHeight
		if isSecondHalf { segmentHeight = secondSegmentHeight }

		alpha := y / totalHeight

		beta := y / segmentHeight
		if isSecondHalf { beta = (y - firstSegmentHeight) / segmentHeight }

		a := geometry.Add(v0, geometry.Multiply(geometry.Subtract(v2, v0), alpha))
		b := geometry.Add(v0, geometry.Multiply(geometry.Subtract(v1, v0), beta))
		if isSecondHalf {
			b = geometry.Add(v1, geometry.Multiply(geometry.Subtract(v2, v1), beta))
		}

		if a.X > b.X {
			swap(&a, &b)
		}

		barycentricA := t.Barycentric(a)
		barycentricB := t.Barycentric(b)

		if barycentricA.X == -1 || barycentricB.X == -1 {
			;
		} else {
			az := (v0.Z * barycentricA.X) + (v1.Z * barycentricA.Y) + (v2.Z * barycentricA.Z)
			bz := (v0.Z * barycentricB.X) + (v1.Z * barycentricB.Y) + (v2.Z * barycentricB.Z)
			 au := (p.TexCoords[0].X * barycentricA.X) + (p.TexCoords[1].X * barycentricA.Y) + (p.TexCoords[2].X * barycentricA.Z)
			 av := (p.TexCoords[0].Y * barycentricA.X) + (p.TexCoords[1].Y * barycentricA.Y) + (p.TexCoords[2].Y * barycentricA.Z)
			 bu := (p.TexCoords[0].X * barycentricB.X) + (p.TexCoords[1].X * barycentricB.Y) + (p.TexCoords[2].X * barycentricB.Z)
			 bv := (p.TexCoords[0].Y * barycentricB.X) + (p.TexCoords[1].Y * barycentricB.Y) + (p.TexCoords[2].Y * barycentricB.Z)
			// ac := color.Multiply(p.Colors[0], barycentricA.X) + color.Multiply(p.Colors[1], barycentricA.Y) + color.Multiply(p.Colors[2], barycentricA.Z)
			// bc := color.Multiply(p.Colors[0], barycentricB.X) + color.Multiply(p.Colors[1], barycentricB.Y) + color.Multiply(p.Colors[2], barycentricB.Z)

			for x := a.X; x <= b.X; x += step {
				t := (x - a.X)
				if t > 0 {
					t = t / (b.X - a.X)
				}

				z := (az * (float64(1) - t)) + (bz * t)
				u := (au * (float64(1) - t)) + (bu * t)
				v := (av * (float64(1) - t)) + (bv * t)
				// c := color.Multiply(ac, float64(1)-t) + color.Multiply(bc, t)
				// c.SetAlpha(0xFF)

				if zbuf.Get(int(x), int(v0.Y + y)) < z {
					zbuf.Set(int(x), int(v0.Y + y), z)

					xTex := math.Max(u * float64(p.Texture.Width), 0.0)
					yTex := math.Max(v * float64(p.Texture.Height), 0.0)
					tex := p.Texture.Get(int(xTex), int(yTex))
					f := color.Multiply(tex, intensity)
					f.SetAlpha(byte(0xFF))

					//col := color.Multiply(color.White, intensity)
					//col.SetAlpha(0xFF)

					img.Set(int(x), int(v0.Y + y), f)
				}
			}
		}
	}
}

//------------------------------------------------------------------------------

type Polygon struct { // Assumed to be a triangle!
	Vertices  [3]geometry.Vertex3
	Normals   [3]geometry.Vertex3
	TexCoords [3]geometry.Vertex3
	Colors    [3]color.Color
	Texture   *image
}

func NewPolygon(v0, v1, v2, n0, n1, n2, uv0, uv1, uv2 geometry.Vertex3, c0, c1, c2 color.Color, tex *image) *Polygon {
	var p Polygon
	p.Vertices = [3]geometry.Vertex3{v0, v1, v2}
	p.Normals = [3]geometry.Vertex3{n0, n1, n2}
	p.TexCoords = [3]geometry.Vertex3{uv0, uv1, uv2}
	p.Colors = [3]color.Color{c0, c1, c2}
	p.Texture = tex
	return &p
}

//------------------------------------------------------------------------------

func usage(progName string) {
	fmt.Printf("Usage: %s file.obj texture.png output.png\n", progName)
	os.Exit(0)
}

func main() {
	//defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()

	if len(os.Args) < 4 {
		usage(os.Args[0])
	}

	objFile := os.Args[1]
	texFile := os.Args[2]
	outFile := os.Args[3]

	width := 512
	height := 512

	img := MakeImage(width, height)
	img.Fill(0x000000FF)

	zbuf := MakeZBuffer(width, height)

	model := obj.NewModel()
	model.ReadFromFile(objFile)

	tex, err := texture.FromFile(texFile)
	if err != nil {
		panic(err)
	}
	texImage := FromTexture(tex)

	for i := 0; i < len(model.FaceVertices); i++ {
		face := model.FaceVertices[i]
		tex := model.FaceTexCoords[i]

		var screenCoords [3]geometry.Vertex3
		var worldCoords [3]geometry.Vertex3
		var uvs [3]geometry.Vertex3

		for j := 0; j < 3; j++ {
			v := model.Vertices[face[j]-1]
			x := (v.X + 1) * (float64(width-1) / 2)
			y := (v.Y + 1) * (float64(height-1) / 2)
			z := (v.Z + 1) * (float64(height-1) / 2) // TODO(AARON): Need proper viewing volume.
			screenCoords[j] = geometry.Vertex3{x, y, z}
			worldCoords[j] = v
			uvs[j] = model.TexCoords[tex[j]-1]
		}

		n := geometry.CrossProduct(geometry.Subtract(worldCoords[2], worldCoords[0]), geometry.Subtract(worldCoords[1], worldCoords[0]))
		n.Normalize()
		intensity := geometry.DotProduct(n, LIGHT_DIR)

		if intensity > 0 {
			var c color.Color
			c = color.Multiply(color.White, intensity)
			c.SetAlpha(byte(0xFF))

			p := NewPolygon(
				screenCoords[0], screenCoords[1], screenCoords[2],
				n, n, n,
				uvs[0], uvs[1], uvs[2],
				c, c, c,
				texImage,
			)

			RasterPolygon(p, img, zbuf, 0.5)
		}
	}

	img.WritePng(outFile)
}
