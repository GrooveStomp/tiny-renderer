package objloader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"github.com/GrooveStomp/tiny-renderer/geometry"
)

type Model struct {
	Vertices []geometry.Vertex3
	FaceVertices [][3]uint
	FaceTexCoords [][3]uint
	TexCoords []geometry.Vertex3
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
		vertexIndex uint
		numFaces    uint
		faceIndex    uint
		numTexCoords uint
		texCoordIndex uint
		matches     []string
	)

	simpleVertRegexp := regexp.MustCompile(`v -?\d+`)
	vertexRegexp := regexp.MustCompile(`v (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?) (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?) (-?\d+(?:\.\d+?\d+(?:e-?\d+)?)?)`)
	simpleFaceRegexp := regexp.MustCompile(`f \d+`)
	faceRegexp := regexp.MustCompile(`f (\d+)/(\d+)/(\d+) (\d+)/(\d+)/(\d+) (\d+)/(\d+)/(\d+)$`)
	simpleTexRegexp := regexp.MustCompile(`vt\s+-?\d+`)
	texRegexp := regexp.MustCompile(`vt\s+(-?\d+(?:\.\d+))\s+(-?\d+(?:\.\d+))\s+(-?\d+(?:\.\d+))`)

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
		} else if simpleTexRegexp.Match(line) { // line starts with "vt "
			numTexCoords++
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

	self.Vertices = make([]geometry.Vertex3, numVertices)
	self.FaceVertices = make([][3]uint, numFaces)
	self.TexCoords = make([]geometry.Vertex3, numTexCoords)
	self.FaceTexCoords = make([][3]uint, numFaces)

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
			self.Vertices[vertexIndex] = geometry.Vertex3{x, y, z}
			vertexIndex++
		} else if simpleFaceRegexp.Match(line) { // line starts with "f "
			matches = faceRegexp.FindStringSubmatch(string(line))
			v0, err := strconv.ParseUint(matches[1], 10, 64)
			if err != nil {
				panic(err)
			}
			vt0, err := strconv.ParseUint(matches[2], 10, 64)
			if err != nil {
				panic(err)
			}
			v1, err := strconv.ParseUint(matches[4], 10, 64)
			if err != nil {
				panic(err)
			}
			vt1, err := strconv.ParseUint(matches[5], 10, 64)
			if err != nil {
				panic(err)
			}
			v2, err := strconv.ParseUint(matches[7], 10, 64)
			if err != nil {
				panic(err)
			}
			vt2, err := strconv.ParseUint(matches[8], 10, 64)
			if err != nil {
				panic(err)
			}

			self.FaceVertices[faceIndex] = [3]uint{uint(v0), uint(v1), uint(v2)}
			self.FaceTexCoords[faceIndex] = [3]uint{uint(vt0), uint(vt1), uint(vt2)}
			faceIndex++
		} else if simpleTexRegexp.Match(line) { // line starts with "vt "
			matches = texRegexp.FindStringSubmatch(string(line))
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
			self.TexCoords[texCoordIndex] = geometry.Vertex3{x, y, z}
			texCoordIndex++
		}
	}
	osFile.Close()
}
