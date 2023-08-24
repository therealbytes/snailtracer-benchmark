package snailtracer

func abs(x int) int {
	if x > 0 {
		return x
	}
	return -x
}

func clamp(x int) int {
	if x < 0 {
		return 0
	}
	if x > 1000000 {
		return 1000000
	}
	return x
}

func sqrt(x int) int {
	if x == 0 {
		return 0
	}
	z := (x + 1) / 2
	y := x
	for z < y {
		y = z
		z = (x/z + z) / 2
	}
	return y
}

func sin(x int) int {
	for x < 0 {
		x += 6283184
	}
	for x >= 6283184 {
		x -= 6283184
	}

	y := 0
	s := 1
	n := x
	d := 1
	f := 2
	for n > d {
		y += s * n / d
		n = n * x * x / 1000000 / 1000000
		d *= f * (f + 1)
		s *= -1
		f += 2
	}
	return y
}

func cos(x int) int {
	s := sin(x)
	return sqrt(1000000000000 - s*s)
}
