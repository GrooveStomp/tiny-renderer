package geometry

import (
	"math"
	"fmt"
)

type Vertex3 struct {
	X float64
	Y float64
	Z float64
}

func (v *Vertex3) Length() float64 {
	p := (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z)
	return math.Sqrt(p)
}

func (v *Vertex3) Normalize() {
	length := v.Length()
	v.X = v.X / length
	v.Y = v.Y / length
	v.Z = v.Z / length
}

func Subtract(a, b Vertex3) Vertex3 {
	x := a.X - b.X
	y := a.Y - b.Y
	z := a.Z - b.Z
	return Vertex3{x, y, z}
}

func Add(a, b Vertex3) Vertex3 {
	x := a.X + b.X
	y := a.Y + b.Y
	z := a.Z + b.Z
	return Vertex3{x, y, z}
}

func Multiply(v Vertex3, f float64) Vertex3 {
	x := v.X * f
	y := v.Y * f
	z := v.Z * f
	return Vertex3{x, y, z}
}

func DotProduct(a, b Vertex3) float64 {
	x := a.X * b.X
	y := a.Y * b.Y
	z := a.Z * b.Z
	return x + y + z
}

func CrossProduct(a, b Vertex3) Vertex3 {
	x := a.Y * b.Z - a.Z * b.Y
	y := a.Z * b.X - a.X * b.Z
	z := a.X * b.Y - a.Y * b.X
	return Vertex3{x, y, z}
}

func (v *Vertex3) String() string {
	return fmt.Sprintf("%.2f,%.2f,%.2f", v.X, v.Y, v.Z)
}

type Triangle struct {
	V0 Vertex3
	V1 Vertex3
	V2 Vertex3
}

func (t *Triangle) Barycentric(p Vertex3) Vertex3 {
	v0 := Subtract(t.V1, t.V0)
	v1 := Subtract(t.V2, t.V0)
	v2 := Subtract(p, t.V0)

	d00 := DotProduct(v0, v0)
	d01 := DotProduct(v0, v1)
	d11 := DotProduct(v1, v1)
	d20 := DotProduct(v2, v0)
	d21 := DotProduct(v2, v1)

	denom := (d00 * d11) - (d01 * d01)

	v := (d11 * d20 - d01 * d21) / denom
	w := (d00 * d21 - d01 * d20) / denom
	u := float64(1) - v - w

	sum := u + v + w
	if sum < 0.99 || sum > 1.001 {
		return Vertex3{-1, -1, -1}
	}

	return Vertex3{u, v, w}
}
