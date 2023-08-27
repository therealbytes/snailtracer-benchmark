package snailtracer

import (
	imageColor "image/color"
	"math/big"
)

type Vector struct {
	x, y, z *big.Int
}

func NewVector(x, y, z int) Vector {
	return Vector{big.NewInt(int64(x)), big.NewInt(int64(y)), big.NewInt(int64(z))}
}

func (v Vector) Add(u Vector) Vector {
	return Vector{
		new(big.Int).Add(v.x, u.x),
		new(big.Int).Add(v.y, u.y),
		new(big.Int).Add(v.z, u.z),
	}
}

func (v Vector) Sub(u Vector) Vector {
	return Vector{
		new(big.Int).Sub(v.x, u.x),
		new(big.Int).Sub(v.y, u.y),
		new(big.Int).Sub(v.z, u.z),
	}
}

func (v Vector) ScaleMul(m *big.Int) Vector {
	return Vector{
		new(big.Int).Mul(m, v.x),
		new(big.Int).Mul(m, v.y),
		new(big.Int).Mul(m, v.z),
	}
}

func (v Vector) ScaleDiv(d *big.Int) Vector {
	return Vector{
		new(big.Int).Div(v.x, d),
		new(big.Int).Div(v.y, d),
		new(big.Int).Div(v.z, d),
	}
}

func (v Vector) Mul(u Vector) Vector {
	return Vector{
		new(big.Int).Mul(v.x, u.x),
		new(big.Int).Mul(v.y, u.y),
		new(big.Int).Mul(v.z, u.z),
	}
}

func (v Vector) Dot(u Vector) *big.Int {
	return new(big.Int).Add(
		new(big.Int).Mul(v.x, u.x),
		new(big.Int).Add(
			new(big.Int).Mul(v.y, u.y),
			new(big.Int).Mul(v.z, u.z),
		),
	)
}

func (v Vector) Cross(u Vector) Vector {
	return Vector{
		new(big.Int).Sub(
			new(big.Int).Mul(v.y, u.z),
			new(big.Int).Mul(v.z, u.y),
		),
		new(big.Int).Sub(
			new(big.Int).Mul(v.z, u.x),
			new(big.Int).Mul(v.x, u.z),
		),
		new(big.Int).Sub(
			new(big.Int).Mul(v.x, u.y),
			new(big.Int).Mul(v.y, u.x),
		),
	}
}

func (v Vector) Length() *big.Int {
	xSq := new(big.Int).Mul(v.x, v.x)
	ySq := new(big.Int).Mul(v.y, v.y)
	zSq := new(big.Int).Mul(v.z, v.z)
	return Sqrt(new(big.Int).Add(xSq, new(big.Int).Add(ySq, zSq)))
}

func (v Vector) Norm() Vector {
	length := v.Length()
	if length.Cmp(Big0) == 0 {
		return Vector{NewBig0(), NewBig0(), NewBig0()}
	}
	nx := new(big.Int).Div(new(big.Int).Mul(v.x, Big1e6), length)
	ny := new(big.Int).Div(new(big.Int).Mul(v.y, Big1e6), length)
	nz := new(big.Int).Div(new(big.Int).Mul(v.z, Big1e6), length)
	return Vector{nx, ny, nz}
}

func (v Vector) Clamp() Vector {
	return Vector{
		x: Clamp(v.x),
		y: Clamp(v.y),
		z: Clamp(v.z),
	}
}

func (v Vector) Color() imageColor.Color {
	return imageColor.RGBA{R: byte(v.x.Int64()), G: byte(v.y.Int64()), B: byte(v.z.Int64()), A: 255}
}
