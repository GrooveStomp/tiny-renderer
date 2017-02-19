package mesh

import (
	"bufio"
	"fmt"
	"github.com/GrooveStomp/tiny-renderer/geometry"
	"io"
	"os"
	"regexp"
	"strconv"
)

type Face struct { // Assumed to be a triangle!
	Vertices  [3]geometry.Vector3
	Normals   [3]geometry.Vector3
	TexCoords [3]geometry.Vector3
}

func NewFace(v0, v1, v2, n0, n1, n2, uv0, uv1, uv2 geometry.Vector3) *Face {
	var p Face
	p.Vertices = [3]geometry.Vector3{v0, v1, v2}
	p.Normals = [3]geometry.Vector3{n0, n1, n2}
	p.TexCoords = [3]geometry.Vector3{uv0, uv1, uv2}
	return &p
}

func NewFaceFromMesh(mesh *Mesh, i, width, height int) *Face {
	face := mesh.FaceVertices[i]
	tex := mesh.FaceTexCoords[i]

	var screenCoords [3]geometry.Vector3
	var worldCoords [3]geometry.Vector3
	var uvs [3]geometry.Vector3

	for j := 0; j < 3; j++ {
		v := mesh.Vertices[face[j]-1]
		x := (v.X + 1) * (float64(width-1) / 2)
		y := (v.Y + 1) * (float64(height-1) / 2)
		z := (v.Z + 1) * (float64(height-1) / 2) // TODO(AARON): Need proper viewing volume.
		screenCoords[j] = geometry.Vector3{x, y, z}
		worldCoords[j] = v
		uvs[j] = mesh.TexCoords[tex[j]-1]
	}

	n := geometry.CrossProduct(geometry.Subtract(worldCoords[2], worldCoords[0]), geometry.Subtract(worldCoords[1], worldCoords[0]))
	n.Normalize()

	p := NewFace(
		screenCoords[0], screenCoords[1], screenCoords[2],
		n, n, n,
		uvs[0], uvs[1], uvs[2],
	)

	return p
}

// Swaps elements at index a with those at index b.
func (f *Face) Swap(a, b int) {
	swp := func(a, b *geometry.Vector3) {
		t := geometry.Vector3{a.X, a.Y, a.Z}
		*a = geometry.Vector3{b.X, b.Y, b.Z}
		*b = geometry.Vector3{t.X, t.Y, t.Z}
	}

	swp(&f.Vertices[a], &f.Vertices[b])
	swp(&f.Normals[a], &f.Normals[b])
	swp(&f.TexCoords[a], &f.TexCoords[b])
}

//------------------------------------------------------------------------------

type Mesh struct {
	Vertices      []geometry.Vector3
	FaceVertices  [][3]uint
	FaceTexCoords [][3]uint
	TexCoords     []geometry.Vector3
}

func NewMesh() *Mesh {
	var mesh Mesh
	return &mesh
}

func (self *Mesh) ReadFromFile(filename string) {
	var (
		line          []byte
		isPrefix      bool
		err           error
		numVertices   uint
		vertexIndex   uint
		numFaces      uint
		faceIndex     uint
		numTexCoords  uint
		texCoordIndex uint
		matches       []string
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

	self.Vertices = make([]geometry.Vector3, numVertices)
	self.FaceVertices = make([][3]uint, numFaces)
	self.TexCoords = make([]geometry.Vector3, numTexCoords)
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
			self.Vertices[vertexIndex] = geometry.Vector3{x, y, z}
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
			self.TexCoords[texCoordIndex] = geometry.Vector3{x, y, z}
			texCoordIndex++
		}
	}
	osFile.Close()
}
