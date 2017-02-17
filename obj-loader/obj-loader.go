package objloader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"math"
)

type Vertex struct {
	X float64
	Y float64
	Z float64
}

func (v *Vertex) Length() float64 {
	p := (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z)
	return math.Sqrt(p)
}

func (v *Vertex) Normalize() {
	length := v.Length()
	v.X = v.X / length
	v.Y = v.Y / length
	v.Z = v.Z / length
}

func Subtract(a, b Vertex) Vertex {
	x := a.X - b.X
	y := a.Y - b.Y
	z := a.Z - b.Z
	return Vertex{x, y, z}
}

func Add(a, b Vertex) Vertex {
	x := a.X + b.X
	y := a.Y + b.Y
	z := a.Z + b.Z
	return Vertex{x, y, z}
}

func Multiply(v Vertex, f float64) Vertex {
	x := v.X * f
	y := v.Y * f
	z := v.Z * f
	return Vertex{x, y, z}
}

func DotProduct(a, b Vertex) float64 {
	x := a.X * b.X
	y := a.Y * b.Y
	z := a.Z * b.Z
	return x + y + z
}

func CrossProduct(a, b Vertex) Vertex {
	x := a.Y * b.Z - a.Z * b.Y
	y := a.Z * b.X - a.X * b.Z
	z := a.X * b.Y - a.Y * b.X
	return Vertex{x, y, z}
}

func (v *Vertex) String() string {
	return fmt.Sprintf("%.2f,%.2f,%.2f", v.X, v.Y, v.Z)
}

type Model struct {
	Vertices []Vertex
	Faces    [][3]uint
}

func NewModel() *Model {
	var model Model
	return &model
}

func (self *Model) ReadFromFile(filename string) {
	var (
		line        []byte
		isPrefix    bool
		err         error
		numVertices uint
		numFaces    uint
		vertexIndex uint
		faceIndex   uint
		matches     []string
	)

	simpleVertRegexp := regexp.MustCompile(`v -?\d+`)
	vertexRegexp := regexp.MustCompile(`v (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?) (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?) (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?)`)
	simpleFaceRegexp := regexp.MustCompile(`f \d+`)
	faceRegexp := regexp.MustCompile(`f (\d+)/\d+/\d+ (\d+)/\d+/\d+ (\d+)/\d+/\d+$`)

	osFile, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	f := bufio.NewReader(osFile)

	// Get vertex and face counts.
	for {
		line, isPrefix, err = f.ReadLine()
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}
		if isPrefix {
			panic("Didn't read entire line!")
		}

		if simpleVertRegexp.Match(line) { // line starts with "v "
			numVertices++
		} else if simpleFaceRegexp.Match(line) { // line starts with "f "
			numFaces++
		}
	}

	if numVertices == 0 {
		panic("Didn't find any vertices!")
	}

	osFile.Close()
	osFile, err = os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	f = bufio.NewReader(osFile)

	self.Vertices = make([]Vertex, numVertices)
	self.Faces = make([][3]uint, numFaces)

	// Now Read data into faces or vertex slot.
	for {
		line, isPrefix, err = f.ReadLine()
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == io.EOF {
			break
		}
		if isPrefix {
			panic("Didn't read entire line!")
		}

		if simpleVertRegexp.Match(line) { // line starts with "v "
			matches = vertexRegexp.FindStringSubmatch(string(line))
			if matches == nil {
				panic(fmt.Sprintf("%s", line))
			}
			x, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				panic(err)
			}
			y, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				panic(err)
			}
			z, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				panic(err)
			}
			self.Vertices[vertexIndex] = Vertex{x, y, z}
			vertexIndex++
		} else if simpleFaceRegexp.Match(line) { // line starts with "f "
			matches = faceRegexp.FindStringSubmatch(string(line))
			x, err := strconv.ParseUint(matches[1], 10, 64)
			if err != nil {
				panic(err)
			}
			y, err := strconv.ParseUint(matches[2], 10, 64)
			if err != nil {
				panic(err)
			}
			z, err := strconv.ParseUint(matches[3], 10, 64)
			if err != nil {
				panic(err)
			}
			self.Faces[faceIndex] = [3]uint{uint(x), uint(y), uint(z)}
			faceIndex++
		}
	}
	osFile.Close()
}
