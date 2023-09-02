package snailtracer

import (
	"github.com/holiman/uint256"
)

type Vector struct {
	X, Y, Z *uint256.Int
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func NewVector(x, y, z int64) Vector {
	ux := uint256.NewInt(uint64(abs(x)))
	uy := uint256.NewInt(uint64(abs(y)))
	uz := uint256.NewInt(uint64(abs(z)))
	if x < 0 {
		ux.Neg(ux)
	}
	if y < 0 {
		uy.Neg(uy)
	}
	if z < 0 {
		uz.Neg(uz)
	}
	return Vector{ux, uy, uz}
}

func (v Vector) Add(u Vector) Vector {
	return Vector{
		new(uint256.Int).Add(v.X, u.X),
		new(uint256.Int).Add(v.Y, u.Y),
		new(uint256.Int).Add(v.Z, u.Z),
	}
}

func (v Vector) Sub(u Vector) Vector {
	return Vector{
		new(uint256.Int).Sub(v.X, u.X),
		new(uint256.Int).Sub(v.Y, u.Y),
		new(uint256.Int).Sub(v.Z, u.Z),
	}
}

func (v Vector) ScaleMul(m *uint256.Int) Vector {
	return Vector{
		new(uint256.Int).Mul(m, v.X),
		new(uint256.Int).Mul(m, v.Y),
		new(uint256.Int).Mul(m, v.Z),
	}
}

func (v Vector) ScaleDiv(d *uint256.Int) Vector {
	return Vector{
		new(uint256.Int).SDiv(v.X, d),
		new(uint256.Int).SDiv(v.Y, d),
		new(uint256.Int).SDiv(v.Z, d),
	}
}

func (v Vector) Mul(u Vector) Vector {
	return Vector{
		new(uint256.Int).Mul(v.X, u.X),
		new(uint256.Int).Mul(v.Y, u.Y),
		new(uint256.Int).Mul(v.Z, u.Z),
	}
}

func (v Vector) Dot(u Vector) *uint256.Int {
	return new(uint256.Int).Add(
		new(uint256.Int).Mul(v.X, u.X),
		new(uint256.Int).Add(
			new(uint256.Int).Mul(v.Y, u.Y),
			new(uint256.Int).Mul(v.Z, u.Z),
		),
	)
}

func (v Vector) Cross(u Vector) Vector {
	return Vector{
		new(uint256.Int).Sub(
			new(uint256.Int).Mul(v.Y, u.Z),
			new(uint256.Int).Mul(v.Z, u.Y),
		),
		new(uint256.Int).Sub(
			new(uint256.Int).Mul(v.Z, u.X),
			new(uint256.Int).Mul(v.X, u.Z),
		),
		new(uint256.Int).Sub(
			new(uint256.Int).Mul(v.X, u.Y),
			new(uint256.Int).Mul(v.Y, u.X),
		),
	}
}

func (v Vector) Length() *uint256.Int {
	xSq := new(uint256.Int).Mul(v.X, v.X)
	ySq := new(uint256.Int).Mul(v.Y, v.Y)
	zSq := new(uint256.Int).Mul(v.Z, v.Z)
	return Sqrt(new(uint256.Int).Add(xSq, new(uint256.Int).Add(ySq, zSq)))
}

func (v Vector) Norm() Vector {
	length := v.Length()
	if Cmp(length, Big0) == 0 {
		return Vector{NewBig0(), NewBig0(), NewBig0()}
	}
	nx := new(uint256.Int).SDiv(new(uint256.Int).Mul(v.X, Big1e6), length)
	ny := new(uint256.Int).SDiv(new(uint256.Int).Mul(v.Y, Big1e6), length)
	nz := new(uint256.Int).SDiv(new(uint256.Int).Mul(v.Z, Big1e6), length)
	return Vector{nx, ny, nz}
}

func (v Vector) Clamp() Vector {
	return Vector{
		X: Clamp(v.X),
		Y: Clamp(v.Y),
		Z: Clamp(v.Z),
	}
}
