package snailtracer

import (
	"math/big"
)

var (
	BigNeg1    = big.NewInt(-1)
	Big0       = big.NewInt(0)
	Big1       = big.NewInt(1)
	Big2       = big.NewInt(2)
	Big1e3     = big.NewInt(1e3)
	Big4e5     = big.NewInt(40000)
	Big1e6     = big.NewInt(1e6)
	Big1e12    = big.NewInt(1e12)
	Big1e30, _ = new(big.Int).SetString("1000000000000000000000000000000", 10)
)

func NewBig0() *big.Int {
	return big.NewInt(0)
}

func NewBig1() *big.Int {
	return big.NewInt(1)
}

func NewBig2() *big.Int {
	return big.NewInt(2)
}

func NewBig1e3() *big.Int {
	return big.NewInt(1000)
}

func NewBig1e6() *big.Int {
	return big.NewInt(1000000)
}

func NewBig1e12() *big.Int {
	return big.NewInt(1000000000000)
}

func Abs(x *big.Int) *big.Int {
	if x.Sign() > 0 {
		return new(big.Int).Set(x)
	}
	return new(big.Int).Neg(x)
}

func Clamp(x *big.Int) *big.Int {
	if x.Cmp(Big0) < 0 {
		return NewBig0()
	}
	if x.Cmp(Big1e6) > 0 {
		return NewBig1e6()
	}
	return new(big.Int).Set(x)
}

func Sqrt(x *big.Int) *big.Int {
	z := new(big.Int).Add(x, Big1)
	z.Quo(z, Big2)
	y := new(big.Int).Set(x)
	for z.Cmp(y) < 0 {
		y.Set(z)
		z.Quo(x, y)
		z.Add(z, y)
		z.Quo(z, Big2)
	}
	return y
}

func Sin(x *big.Int) *big.Int {
	constant := big.NewInt(6283184)
	for x.Sign() < 0 {
		x.Add(x, constant)
	}
	for x.Cmp(constant) >= 0 {
		x.Sub(x, constant)
	}

	n := new(big.Int).Set(x)
	y := NewBig0()
	s := NewBig1()
	d := NewBig1()
	f := NewBig2()
	for n.Cmp(d) > 0 {
		t := new(big.Int).Mul(s, n)
		t.Quo(t, d)
		y.Add(y, t)

		n.Mul(n, x)
		n.Mul(n, x)
		n.Quo(n, Big1e6)
		n.Quo(n, Big1e6)

		d.Mul(d, f)
		d.Mul(d, new(big.Int).Add(f, Big1))

		s.Neg(s)

		f.Add(f, Big2)
	}
	return y
}

func Cos(x *big.Int) *big.Int {
	s := Sin(x)
	sSquared := new(big.Int).Mul(s, s)
	difference := new(big.Int).Sub(Big1e12, sSquared)
	return Sqrt(difference)
}
