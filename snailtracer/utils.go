package snailtracer

import (
	"github.com/holiman/uint256"
)

var (
	BigNeg1    = new(uint256.Int).Neg(uint256.NewInt(1))
	Big0       = uint256.NewInt(0)
	Big1       = uint256.NewInt(1)
	Big2       = uint256.NewInt(2)
	Big1e3     = uint256.NewInt(1e3)
	Big4e4     = uint256.NewInt(40000)
	Big1e5     = uint256.NewInt(100000)
	Big1e6     = uint256.NewInt(1e6)
	Big1e12    = uint256.NewInt(1e12)
	Big1e30, _ = uint256.FromHex("0xC9F2C9CD04674EDEA40000000")
)

func NewBig0() *uint256.Int {
	return uint256.NewInt(0)
}

func NewBig1() *uint256.Int {
	return uint256.NewInt(1)
}

func NewBig2() *uint256.Int {
	return uint256.NewInt(2)
}

func NewBig1e3() *uint256.Int {
	return uint256.NewInt(1000)
}

func NewBig1e6() *uint256.Int {
	return uint256.NewInt(1000000)
}

func NewBig1e12() *uint256.Int {
	return uint256.NewInt(1000000000000)
}

func Abs(x *uint256.Int) *uint256.Int {
	if x.Sign() > 0 {
		return new(uint256.Int).Set(x)
	}
	return new(uint256.Int).Neg(x)
}

func Clamp(x *uint256.Int) *uint256.Int {
	if Cmp(x, Big0) < 0 {
		return NewBig0()
	}
	if Cmp(x, Big1e6) > 0 {
		return NewBig1e6()
	}
	return new(uint256.Int).Set(x)
}

func Sqrt(x *uint256.Int) *uint256.Int {
	z := new(uint256.Int).Add(x, Big1)
	z.SDiv(z, Big2)
	y := new(uint256.Int).Set(x)
	for Cmp(z, y) < 0 {
		y.Set(z)
		z.SDiv(x, y)
		z.Add(z, y)
		z.SDiv(z, Big2)
	}
	return y
}

func Sin(x *uint256.Int) *uint256.Int {
	constant := uint256.NewInt(6283184)
	for x.Sign() < 0 {
		x.Add(x, constant)
	}
	for Cmp(x, constant) >= 0 {
		x.Sub(x, constant)
	}

	n := new(uint256.Int).Set(x)
	y := NewBig0()
	s := NewBig1()
	d := NewBig1()
	f := NewBig2()
	for Cmp(n, d) > 0 {
		t := new(uint256.Int).Mul(s, n)
		t.SDiv(t, d)
		y.Add(y, t)

		n.Mul(n, x)
		n.Mul(n, x)
		n.SDiv(n, Big1e6)
		n.SDiv(n, Big1e6)

		d.Mul(d, f)
		d.Mul(d, new(uint256.Int).Add(f, Big1))

		s.Neg(s)

		f.Add(f, Big2)
	}
	return y
}

func Cos(x *uint256.Int) *uint256.Int {
	s := Sin(x)
	sSquared := new(uint256.Int).Mul(s, s)
	difference := new(uint256.Int).Sub(Big1e12, sSquared)
	return Sqrt(difference)
}

func Cmp(x, y *uint256.Int) int {
	if x.Sign() < y.Sign() {
		return -1
	}
	if x.Sign() > y.Sign() {
		return 1
	}
	return x.Cmp(y)
}
