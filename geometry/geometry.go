package geometry // import "code.groovestomp.com/tiny-renderer/geometry"

import (
	"fmt"
	"math"
)

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

func (v *Vector3) Length() float64 {
	p := (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z)
	return math.Sqrt(p)
}

func (v *Vector3) Normalize() {
	length := v.Length()
	v.X = v.X / length
	v.Y = v.Y / length
	v.Z = v.Z / length
}

func Subtract(a, b Vector3) Vector3 {
	x := a.X - b.X
	y := a.Y - b.Y
	z := a.Z - b.Z
	return Vector3{x, y, z}
}

func Add(a, b Vector3) Vector3 {
	x := a.X + b.X
	y := a.Y + b.Y
	z := a.Z + b.Z
	return Vector3{x, y, z}
}

func Multiply(v Vector3, f float64) Vector3 {
	x := v.X * f
	y := v.Y * f
	z := v.Z * f
	return Vector3{x, y, z}
}

func DotProduct(a, b Vector3) float64 {
	x := a.X * b.X
	y := a.Y * b.Y
	z := a.Z * b.Z
	return x + y + z
}

func CrossProduct(a, b Vector3) Vector3 {
	x := a.Y*b.Z - a.Z*b.Y
	y := a.Z*b.X - a.X*b.Z
	z := a.X*b.Y - a.Y*b.X
	return Vector3{x, y, z}
}

func (v *Vector3) String() string {
	return fmt.Sprintf("%.2f,%.2f,%.2f", v.X, v.Y, v.Z)
}

type Triangle struct {
	V0 Vector3
	V1 Vector3
	V2 Vector3
}

func (t *Triangle) Barycentric(p Vector3) Vector3 {
	v0 := Subtract(t.V1, t.V0)
	v1 := Subtract(t.V2, t.V0)
	v2 := Subtract(p, t.V0)

	d00 := DotProduct(v0, v0)
	d01 := DotProduct(v0, v1)
	d11 := DotProduct(v1, v1)
	d20 := DotProduct(v2, v0)
	d21 := DotProduct(v2, v1)

	denom := (d00 * d11) - (d01 * d01)

	v := (d11*d20 - d01*d21) / denom
	w := (d00*d21 - d01*d20) / denom
	u := float64(1) - v - w

	sum := u + v + w
	if sum < 0.99 || sum > 1.001 {
		return Vector3{-1, -1, -1}
	}

	return Vector3{u, v, w}
}
