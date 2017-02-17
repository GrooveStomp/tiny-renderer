package color

import (
	"fmt"
)

type Color uint32

func NewColor(c Color) Color {
	return c
}

func NewColorRgba(r, g, b, a byte) Color {
	t := uint32(r)<<24 |
		uint32(g)<<16 |
		uint32(b)<<8 |
		uint32(a)

	return Color(t)
}

func NewColorFloat64(r, g, b, a float64) Color {
	r1 := r * float64(255)
	g1 := g * float64(255)
	b1 := b * float64(255)
	a1 := a * float64(255)

	return NewColorRgba(byte(r1), byte(g1), byte(b1), byte(a1))
}

func (c Color) Rgba() (byte, byte, byte, byte) {
	r, g, b, a := uint32(c), uint32(c), uint32(c), uint32(c)
	r >>= 24
	g >>= 16
	b >>= 8

	return byte(r), byte(g), byte(b), byte(a)
}

func (c *Color) Set(r, g, b, a byte) {
	new := NewColorRgba(r, g, b, a)
	*c = new
}

func (c *Color) SetAlpha(a byte) {
	*c = Color(uint32(*c) | uint32(a))
}

func Multiply(c Color, k float64) Color {
	r, g, b, a := c.Rgba()
	r = byte(float64(r) * k)
	g = byte(float64(g) * k)
	b = byte(float64(b) * k)
	a = byte(float64(a) * k)

	return NewColorRgba(r, g, b, a)
}

func Add(c Color, other Color) Color {
	r, g, b, a := c.Rgba()
	r2, g2, b2, a2 := other.Rgba()
	r += r2
	g += g2
	b += b2
	a += a2

	return NewColorRgba(r, g, b, a)
}

func (c Color) String() string {
	r, g, b, a := c.Rgba()
	return fmt.Sprintf("%v,%v,%v,%v", r, g, b, a)
}

var Red Color = NewColorFloat64(1.0, 0.0, 0.0, 1.0)
var Green Color = NewColorFloat64(0.0, 1.0, 0.0, 1.0)
var Blue Color = NewColorFloat64(0.0, 0.0, 1.0, 1.0)
var White Color = NewColorFloat64(1.0, 1.0, 1.0, 1.0)
