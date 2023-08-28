package snailtracer

import (
	"math/big"
)

type Vector struct {
	X, Y, Z *big.Int
}

func NewVector(x, y, z int) Vector {
	return Vector{big.NewInt(int64(x)), big.NewInt(int64(y)), big.NewInt(int64(z))}
}

func (v Vector) Add(u Vector) Vector {
	return Vector{
		new(big.Int).Add(v.X, u.X),
		new(big.Int).Add(v.Y, u.Y),
		new(big.Int).Add(v.Z, u.Z),
	}
}

func (v Vector) Sub(u Vector) Vector {
	return Vector{
		new(big.Int).Sub(v.X, u.X),
		new(big.Int).Sub(v.Y, u.Y),
		new(big.Int).Sub(v.Z, u.Z),
	}
}

func (v Vector) ScaleMul(m *big.Int) Vector {
	return Vector{
		new(big.Int).Mul(m, v.X),
		new(big.Int).Mul(m, v.Y),
		new(big.Int).Mul(m, v.Z),
	}
}

func (v Vector) ScaleDiv(d *big.Int) Vector {
	return Vector{
		new(big.Int).Quo(v.X, d),
		new(big.Int).Quo(v.Y, d),
		new(big.Int).Quo(v.Z, d),
	}
}

func (v Vector) Mul(u Vector) Vector {
	return Vector{
		new(big.Int).Mul(v.X, u.X),
		new(big.Int).Mul(v.Y, u.Y),
		new(big.Int).Mul(v.Z, u.Z),
	}
}

func (v Vector) Dot(u Vector) *big.Int {
	return new(big.Int).Add(
		new(big.Int).Mul(v.X, u.X),
		new(big.Int).Add(
			new(big.Int).Mul(v.Y, u.Y),
			new(big.Int).Mul(v.Z, u.Z),
		),
	)
}

func (v Vector) Cross(u Vector) Vector {
	return Vector{
		new(big.Int).Sub(
			new(big.Int).Mul(v.Y, u.Z),
			new(big.Int).Mul(v.Z, u.Y),
		),
		new(big.Int).Sub(
			new(big.Int).Mul(v.Z, u.X),
			new(big.Int).Mul(v.X, u.Z),
		),
		new(big.Int).Sub(
			new(big.Int).Mul(v.X, u.Y),
			new(big.Int).Mul(v.Y, u.X),
		),
	}
}

func (v Vector) Length() *big.Int {
	xSq := new(big.Int).Mul(v.X, v.X)
	ySq := new(big.Int).Mul(v.Y, v.Y)
	zSq := new(big.Int).Mul(v.Z, v.Z)
	return Sqrt(new(big.Int).Add(xSq, new(big.Int).Add(ySq, zSq)))
}

func (v Vector) Norm() Vector {
	length := v.Length()
	if length.Cmp(Big0) == 0 {
		return Vector{NewBig0(), NewBig0(), NewBig0()}
	}
	nx := new(big.Int).Quo(new(big.Int).Mul(v.X, Big1e6), length)
	ny := new(big.Int).Quo(new(big.Int).Mul(v.Y, Big1e6), length)
	nz := new(big.Int).Quo(new(big.Int).Mul(v.Z, Big1e6), length)
	return Vector{nx, ny, nz}
}

func (v Vector) Clamp() Vector {
	return Vector{
		X: Clamp(v.X),
		Y: Clamp(v.Y),
		Z: Clamp(v.Z),
	}
}
