package snailtracer

import (
	"math/big"
)

type Ray struct {
	origin, direction Vector
	depth             int
	refract           bool
}

type Sphere struct {
	radius     *big.Int
	position   Vector
	emission   Vector
	color      Vector
	reflection Material
}

func (s *Sphere) Intersect(r *Ray) *big.Int {
	op := s.position.Sub(r.origin)
	b := new(big.Int).Quo(op.Dot(r.direction), Big1e6)

	bSq := new(big.Int).Mul(b, b)
	rSq := new(big.Int).Mul(s.radius, s.radius)
	sSq := new(big.Int).Add(rSq, bSq)
	det := new(big.Int).Sub(sSq, op.Dot(op))

	if det.Cmp(Big0) <= 0 {
		return NewBig0()
	}

	detSqrt := Sqrt(det)

	bMinusDetSqrt := new(big.Int).Sub(b, detSqrt)
	bPlusDetSqrt := new(big.Int).Add(b, detSqrt)

	if bMinusDetSqrt.Cmp(Big1e3) > 0 {
		return bMinusDetSqrt
	}
	if bPlusDetSqrt.Cmp(Big1e3) > 0 {
		return bPlusDetSqrt
	}
	return NewBig0()
}

type Triangle struct {
	a, b, c    Vector
	normal     Vector
	emission   Vector
	color      Vector
	reflection Material
}

func (t *Triangle) Intersect(r *Ray) *big.Int {
	e1 := t.b.Sub(t.a)
	e2 := t.c.Sub(t.a)
	p := r.direction.Cross(e2)

	det := new(big.Int).Quo(e1.Dot(p), Big1e6)
	if det.Cmp(new(big.Int).Neg(Big1e3)) > 0 && det.Cmp(Big1e3) < 0 {
		return NewBig0()
	}

	d := r.origin.Sub(t.a)
	u := new(big.Int).Quo(d.Dot(p), det)

	if u.Cmp(Big0) < 0 || u.Cmp(Big1e6) > 0 {
		return NewBig0()
	}

	q := d.Cross(e1)
	v := new(big.Int).Quo(r.direction.Dot(q), det)

	if v.Cmp(Big0) < 0 || new(big.Int).Add(u, v).Cmp(Big1e6) > 0 {
		return NewBig0()
	}

	dist := new(big.Int).Quo(e2.Dot(q), det)

	if dist.Cmp(Big1e3) < 0 {
		return NewBig0()
	}
	return dist
}

type Material int

const (
	DiffuseMaterial Material = iota
	SpecularMaterial
	RefractiveMaterial
)

type Primitive int

const (
	SpherePrimitive Primitive = iota
	TrianglePrimitive
)

type Scene struct {
	id             int
	seed           uint32
	width, height  int
	camera         *Ray
	deltaX, deltaY Vector
	spheres        []*Sphere
	triangles      []*Triangle
}

func newScene(w, h int) *Scene {
	s := &Scene{}
	s.width = w
	s.height = h
	return s
}

func (s *Scene) rand() *big.Int {
	s.seed = s.seed*1103515245 + 12345
	return big.NewInt(int64(s.seed))
}

func (s *Scene) trace(x, y, spp int) Vector {
	s.seed = uint32(s.id*s.width*s.height + y*s.width + x)
	color := NewVector(0, 0, 0)

	for k := 0; k < spp; k++ {
		rdX := s.deltaX.ScaleMul(
			new(big.Int).Sub(
				new(big.Int).Quo(
					new(big.Int).Add(
						new(big.Int).Mul(Big1e6, big.NewInt(int64(x))),
						new(big.Int).Rem(s.rand(), big.NewInt(500000)),
					),
					big.NewInt(int64(s.width)),
				),
				big.NewInt(500000),
			),
		)
		rdY := s.deltaY.ScaleMul(
			new(big.Int).Sub(
				new(big.Int).Quo(
					new(big.Int).Add(
						new(big.Int).Mul(Big1e6, big.NewInt(int64(y))),
						new(big.Int).Rem(s.rand(), big.NewInt(500000)),
					),
					big.NewInt(int64(s.height)),
				),
				big.NewInt(500000),
			),
		)
		pixel := rdX.Add(rdY).ScaleDiv(Big1e6).Add(s.camera.direction)
		ray := &Ray{
			origin:    s.camera.origin.Add(pixel.ScaleMul(big.NewInt(140))),
			direction: pixel.Norm(),
		}
		rad := s.radiance(ray)
		color = color.Add(rad.ScaleDiv(big.NewInt(int64(spp))))
	}

	return color.Clamp().ScaleMul(big.NewInt(255)).ScaleDiv(Big1e6)
}

func (s *Scene) radiance(ray *Ray) Vector {
	if ray.depth > 10 {
		return NewVector(0, 0, 0)
	}

	dist, p, id := s.traceRay(ray)
	if dist.Cmp(Big0) == 0 {
		return NewVector(0, 0, 0)
	}

	var color, emission Vector
	var sphere *Sphere
	var triangle *Triangle

	if p == SpherePrimitive {
		sphere = s.spheres[id]
		color = sphere.color
		emission = sphere.emission
	} else {
		triangle = s.triangles[id]
		color = triangle.color
		emission = triangle.emission
	}

	ref := Big1
	if color.X.Cmp(ref) > 0 {
		ref = color.X
	}
	if color.Y.Cmp(ref) > 0 {
		ref = color.Y
	}
	if color.Z.Cmp(ref) > 0 {
		ref = color.Z
	}

	ray.depth++
	if ray.depth > 5 {
		if new(big.Int).Rem(s.rand(), Big1e6).Cmp(ref) < 0 {
			color = color.ScaleMul(Big1e6).ScaleDiv(ref)
		} else {
			return emission
		}
	}

	var result Vector
	if p == SpherePrimitive {
		result = s.radianceSphere(ray, sphere, dist)
	} else {
		result = s.radianceTriangle(ray, triangle, dist)
	}
	return emission.Add(color.Mul(result).ScaleDiv(Big1e6))
}

func (s *Scene) radianceSphere(ray *Ray, obj *Sphere, dist *big.Int) Vector {
	intersect := ray.origin.Add(ray.direction.ScaleMul(dist).ScaleDiv(Big1e6))
	normal := intersect.Sub(obj.position).Norm()

	if obj.reflection == DiffuseMaterial {
		if normal.Dot(ray.direction).Cmp(Big0) >= 0 {
			normal = normal.ScaleMul(BigNeg1)
		}
		return s.diffuse(ray, intersect, normal)
	}
	return s.specular(ray, intersect, normal)
}

func (s *Scene) radianceTriangle(ray *Ray, obj *Triangle, dist *big.Int) Vector {
	intersect := ray.origin.Add(ray.direction.ScaleMul(dist).ScaleDiv(Big1e6))

	nnt := big.NewInt(666666)
	if ray.refract {
		nnt = big.NewInt(1500000)
	}
	ddn := new(big.Int).Quo(obj.normal.Dot(ray.direction), Big1e6)
	if ddn.Cmp(Big0) >= 0 {
		ddn = new(big.Int).Neg(ddn)
	}
	cos2t := new(big.Int).Sub(
		Big1e12,
		new(big.Int).Quo(
			new(big.Int).Mul(
				new(big.Int).Mul(nnt, nnt),
				new(big.Int).Sub(
					Big1e12,
					new(big.Int).Mul(ddn, ddn),
				),
			),
			Big1e12,
		),
	)
	if cos2t.Cmp(Big0) < 0 {
		return s.specular(ray, intersect, obj.normal)
	}
	return s.refractive(ray, intersect, obj.normal, nnt, ddn, cos2t)
}

func (s *Scene) diffuse(ray *Ray, intersect, normal Vector) Vector {
	r1 := big.NewInt(6283184)
	r1.Mul(r1, new(big.Int).Rem(s.rand(), Big1e6))
	r1.Quo(r1, Big1e6)

	r2 := new(big.Int).Rem(s.rand(), Big1e6)
	r2s := new(big.Int).Mul(Sqrt(r2), Big1e3)

	var u Vector
	if Abs(normal.X).Cmp(Big1e5) > 0 {
		u = NewVector(0, 1000000, 0)
	} else {
		u = NewVector(1000000, 0, 0)
	}
	u = u.Cross(normal).Norm()

	v := normal.Cross(u).Norm()

	u1 := u.ScaleMul(new(big.Int).Quo(new(big.Int).Mul(Cos(r1), r2s), Big1e6))
	v1 := v.ScaleMul(new(big.Int).Quo(new(big.Int).Mul(Sin(r1), r2s), Big1e6))
	n1 := normal.ScaleMul(new(big.Int).Mul(Sqrt(new(big.Int).Sub(Big1e6, r2)), Big1e3))
	u = u1.Add(v1).Add(n1).Norm()

	return s.radiance(&Ray{intersect, u, ray.depth, ray.refract})
}

func (s *Scene) specular(ray *Ray, intersect, normal Vector) Vector {
	d2 := new(big.Int).Mul(Big2, normal.Dot(ray.direction))
	reflection := ray.direction.Sub(normal.ScaleMul(new(big.Int).Quo(d2, Big1e6))).Norm()
	return s.radiance(&Ray{intersect, reflection, ray.depth, ray.refract})
}

func (s *Scene) refractive(ray *Ray, intersect, normal Vector, nnt, ddn, cos2t *big.Int) Vector {
	sign := BigNeg1
	if ray.refract {
		sign = Big1
	}

	temp := new(big.Int).Mul(ddn, nnt)
	temp.Quo(temp, Big1e6)
	temp.Add(temp, Sqrt(cos2t))
	temp.Mul(temp, sign)

	refraction := ray.direction.ScaleMul(nnt).
		Sub(normal.ScaleMul(temp)).
		ScaleDiv(Big1e6).
		Norm()

	c := new(big.Int).Add(Big1e6, ddn)
	if !ray.refract {
		c = new(big.Int).Quo(refraction.Dot(normal), Big1e6)
		c.Sub(Big1e6, c)
	}

	temp = new(big.Int).Sub(Big1e6, Big4e4)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Quo(temp, Big1e30)
	re := new(big.Int).Add(Big4e4, temp)

	if ray.depth <= 2 {
		refraction = s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).ScaleMul(new(big.Int).Sub(Big1e6, re))
		refraction = refraction.Add(s.specular(ray, intersect, normal).ScaleMul(re))
		return refraction.ScaleDiv(Big1e6)
	}

	reDiv2 := new(big.Int).Quo(re, Big2)
	threshold := big.NewInt(250000)
	threshold.Add(threshold, reDiv2)

	if new(big.Int).Rem(s.rand(), Big1e6).Cmp(threshold) < 0 {
		return s.specular(ray, intersect, normal).ScaleMul(re).ScaleDiv(threshold)
	}

	return s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).
		ScaleMul(new(big.Int).Sub(Big1e6, re)).
		ScaleDiv(new(big.Int).Sub(big.NewInt(750000), reDiv2))
}

func (s *Scene) traceRay(ray *Ray) (*big.Int, Primitive, int) {
	var p Primitive
	var id int

	dist := NewBig0()

	for i := 0; i < len(s.spheres); i++ {
		d := s.spheres[i].Intersect(ray)
		if d.Cmp(Big0) > 0 && (dist.Cmp(Big0) == 0 || d.Cmp(dist) < 0) {
			dist.Set(d)
			p = SpherePrimitive
			id = i
		}
	}

	for i := 0; i < len(s.triangles); i++ {
		d := s.triangles[i].Intersect(ray)
		if d.Cmp(Big0) > 0 && (dist.Cmp(Big0) == 0 || d.Cmp(dist) < 0) {
			dist.Set(d)
			p = TrianglePrimitive
			id = i
		}
	}

	return dist, p, id
}

func (s *Scene) Trace(x, y, spp int) Vector {
	return s.trace(x, y, spp)
}
