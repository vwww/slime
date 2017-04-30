package slime

import (
	"math"
)

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(w Vec2) Vec2    { return Vec2{v.X + w.X, v.Y + w.Y} }
func (v Vec2) Sub(w Vec2) Vec2    { return Vec2{v.X - w.X, v.Y - w.Y} }
func (v Vec2) Mul(f float64) Vec2 { return Vec2{v.X * f, v.Y * f} }
func (v Vec2) Div(f float64) Vec2 { return Vec2{v.X / f, v.Y / f} }
func (v Vec2) Dot(w Vec2) float64 { return v.X*w.X + v.Y*w.Y }

func (v Vec2) Length() float64        { return math.Hypot(v.X, v.Y) }
func (v Vec2) LengthSquared() float64 { return v.Dot(v) }
func (v Vec2) Normalize() Vec2        { return v.Div(v.Length()) }
