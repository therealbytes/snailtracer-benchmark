package snailtracer

import (
	"github.com/holiman/uint256"
)

type Ray struct {
	origin, direction Vector
	depth             int
	refract           bool
}

type Sphere struct {
	radius     *uint256.Int
	position   Vector
	emission   Vector
	color      Vector
	reflection Material
}

func (s *Sphere) Intersect(r *Ray) *uint256.Int {
	op := s.position.Sub(r.origin)
	b := new(uint256.Int).SDiv(op.Dot(r.direction), Big1e6)

	bSq := new(uint256.Int).Mul(b, b)
	rSq := new(uint256.Int).Mul(s.radius, s.radius)
	sSq := new(uint256.Int).Add(rSq, bSq)
	det := new(uint256.Int).Sub(sSq, op.Dot(op))

	if Cmp(det, Big0) <= 0 {
		return NewBig0()
	}

	detSqrt := Sqrt(det)

	bMinusDetSqrt := new(uint256.Int).Sub(b, detSqrt)
	bPlusDetSqrt := new(uint256.Int).Add(b, detSqrt)

	if Cmp(bMinusDetSqrt, Big1e3) > 0 {
		return bMinusDetSqrt
	}
	if Cmp(bPlusDetSqrt, Big1e3) > 0 {
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

func (t *Triangle) Intersect(r *Ray) *uint256.Int {
	e1 := t.b.Sub(t.a)
	e2 := t.c.Sub(t.a)
	p := r.direction.Cross(e2)

	det := new(uint256.Int).SDiv(e1.Dot(p), Big1e6)
	if Cmp(det, new(uint256.Int).Neg(Big1e3)) > 0 && Cmp(det, Big1e3) < 0 {
		return NewBig0()
	}

	d := r.origin.Sub(t.a)
	u := new(uint256.Int).SDiv(d.Dot(p), det)

	if Cmp(u, Big0) < 0 || Cmp(u, Big1e6) > 0 {
		return NewBig0()
	}

	q := d.Cross(e1)
	v := new(uint256.Int).SDiv(r.direction.Dot(q), det)

	if Cmp(v, Big0) < 0 || Cmp(new(uint256.Int).Add(u, v), Big1e6) > 0 {
		return NewBig0()
	}

	dist := new(uint256.Int).SDiv(e2.Dot(q), det)

	if Cmp(dist, Big1e3) < 0 {
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

func (s *Scene) rand() *uint256.Int {
	s.seed = s.seed*1103515245 + 12345
	return uint256.NewInt(uint64(s.seed))
}

func (s *Scene) trace(x, y, spp int) Vector {
	s.seed = uint32(s.id*s.width*s.height + y*s.width + x)
	color := NewVector(0, 0, 0)

	for k := 0; k < spp; k++ {
		rdX := s.deltaX.ScaleMul(
			new(uint256.Int).Sub(
				new(uint256.Int).SDiv(
					new(uint256.Int).Add(
						new(uint256.Int).Mul(Big1e6, uint256.NewInt(uint64(x))),
						new(uint256.Int).SMod(s.rand(), uint256.NewInt(500000)),
					),
					uint256.NewInt(uint64(s.width)),
				),
				uint256.NewInt(500000),
			),
		)
		rdY := s.deltaY.ScaleMul(
			new(uint256.Int).Sub(
				new(uint256.Int).SDiv(
					new(uint256.Int).Add(
						new(uint256.Int).Mul(Big1e6, uint256.NewInt(uint64(y))),
						new(uint256.Int).SMod(s.rand(), uint256.NewInt(500000)),
					),
					uint256.NewInt(uint64(s.height)),
				),
				uint256.NewInt(500000),
			),
		)
		pixel := rdX.Add(rdY).ScaleDiv(Big1e6).Add(s.camera.direction)
		ray := &Ray{
			origin:    s.camera.origin.Add(pixel.ScaleMul(uint256.NewInt(140))),
			direction: pixel.Norm(),
		}
		rad := s.radiance(ray)
		color = color.Add(rad.ScaleDiv(uint256.NewInt(uint64(spp))))
	}

	return color.Clamp().ScaleMul(uint256.NewInt(255)).ScaleDiv(Big1e6)
}

func (s *Scene) radiance(ray *Ray) Vector {
	if ray.depth > 10 {
		return NewVector(0, 0, 0)
	}

	dist, p, id := s.traceRay(ray)
	if Cmp(dist, Big0) == 0 {
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
	if Cmp(color.X, ref) > 0 {
		ref = color.X
	}
	if Cmp(color.Y, ref) > 0 {
		ref = color.Y
	}
	if Cmp(color.Z, ref) > 0 {
		ref = color.Z
	}

	ray.depth++
	if ray.depth > 5 {
		if Cmp(new(uint256.Int).SMod(s.rand(), Big1e6), ref) < 0 {
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

func (s *Scene) radianceSphere(ray *Ray, obj *Sphere, dist *uint256.Int) Vector {
	intersect := ray.origin.Add(ray.direction.ScaleMul(dist).ScaleDiv(Big1e6))
	normal := intersect.Sub(obj.position).Norm()

	if obj.reflection == DiffuseMaterial {
		if Cmp(normal.Dot(ray.direction), Big0) >= 0 {
			normal = normal.ScaleMul(BigNeg1)
		}
		return s.diffuse(ray, intersect, normal)
	}
	return s.specular(ray, intersect, normal)
}

func (s *Scene) radianceTriangle(ray *Ray, obj *Triangle, dist *uint256.Int) Vector {
	intersect := ray.origin.Add(ray.direction.ScaleMul(dist).ScaleDiv(Big1e6))

	nnt := uint256.NewInt(666666)
	if ray.refract {
		nnt = uint256.NewInt(1500000)
	}
	ddn := new(uint256.Int).SDiv(obj.normal.Dot(ray.direction), Big1e6)
	if Cmp(ddn, Big0) >= 0 {
		ddn = new(uint256.Int).Neg(ddn)
	}
	cos2t := new(uint256.Int).Sub(
		Big1e12,
		new(uint256.Int).SDiv(
			new(uint256.Int).Mul(
				new(uint256.Int).Mul(nnt, nnt),
				new(uint256.Int).Sub(
					Big1e12,
					new(uint256.Int).Mul(ddn, ddn),
				),
			),
			Big1e12,
		),
	)
	if Cmp(cos2t, Big0) < 0 {
		return s.specular(ray, intersect, obj.normal)
	}
	return s.refractive(ray, intersect, obj.normal, nnt, ddn, cos2t)
}

func (s *Scene) diffuse(ray *Ray, intersect, normal Vector) Vector {
	r1 := uint256.NewInt(6283184)
	r1.Mul(r1, new(uint256.Int).SMod(s.rand(), Big1e6))
	r1.SDiv(r1, Big1e6)

	r2 := new(uint256.Int).SMod(s.rand(), Big1e6)
	r2s := new(uint256.Int).Mul(Sqrt(r2), Big1e3)

	var u Vector
	if Cmp(Abs(normal.X), Big1e5) > 0 {
		u = NewVector(0, 1000000, 0)
	} else {
		u = NewVector(1000000, 0, 0)
	}
	u = u.Cross(normal).Norm()

	v := normal.Cross(u).Norm()

	u1 := u.ScaleMul(new(uint256.Int).SDiv(new(uint256.Int).Mul(Cos(r1), r2s), Big1e6))
	v1 := v.ScaleMul(new(uint256.Int).SDiv(new(uint256.Int).Mul(Sin(r1), r2s), Big1e6))
	n1 := normal.ScaleMul(new(uint256.Int).Mul(Sqrt(new(uint256.Int).Sub(Big1e6, r2)), Big1e3))
	u = u1.Add(v1).Add(n1).Norm()

	return s.radiance(&Ray{intersect, u, ray.depth, ray.refract})
}

func (s *Scene) specular(ray *Ray, intersect, normal Vector) Vector {
	d2 := new(uint256.Int).Mul(Big2, normal.Dot(ray.direction))
	reflection := ray.direction.Sub(normal.ScaleMul(new(uint256.Int).SDiv(d2, Big1e6))).Norm()
	return s.radiance(&Ray{intersect, reflection, ray.depth, ray.refract})
}

func (s *Scene) refractive(ray *Ray, intersect, normal Vector, nnt, ddn, cos2t *uint256.Int) Vector {
	sign := BigNeg1
	if ray.refract {
		sign = Big1
	}

	temp := new(uint256.Int).Mul(ddn, nnt)
	temp.SDiv(temp, Big1e6)
	temp.Add(temp, Sqrt(cos2t))
	temp.Mul(temp, sign)

	refraction := ray.direction.ScaleMul(nnt).
		Sub(normal.ScaleMul(temp)).
		ScaleDiv(Big1e6).
		Norm()

	c := new(uint256.Int).Add(Big1e6, ddn)
	if !ray.refract {
		c = new(uint256.Int).SDiv(refraction.Dot(normal), Big1e6)
		c.Sub(Big1e6, c)
	}

	temp = new(uint256.Int).Sub(Big1e6, Big4e4)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.Mul(temp, c)
	temp.SDiv(temp, Big1e30)
	re := new(uint256.Int).Add(Big4e4, temp)

	if ray.depth <= 2 {
		refraction = s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).ScaleMul(new(uint256.Int).Sub(Big1e6, re))
		refraction = refraction.Add(s.specular(ray, intersect, normal).ScaleMul(re))
		return refraction.ScaleDiv(Big1e6)
	}

	reDiv2 := new(uint256.Int).SDiv(re, Big2)
	threshold := uint256.NewInt(250000)
	threshold.Add(threshold, reDiv2)

	if Cmp(new(uint256.Int).SMod(s.rand(), Big1e6), threshold) < 0 {
		return s.specular(ray, intersect, normal).ScaleMul(re).ScaleDiv(threshold)
	}

	return s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).
		ScaleMul(new(uint256.Int).Sub(Big1e6, re)).
		ScaleDiv(new(uint256.Int).Sub(uint256.NewInt(750000), reDiv2))
}

func (s *Scene) traceRay(ray *Ray) (*uint256.Int, Primitive, int) {
	var p Primitive
	var id int

	dist := NewBig0()

	for i := 0; i < len(s.spheres); i++ {
		d := s.spheres[i].Intersect(ray)
		if Cmp(d, Big0) > 0 && (Cmp(dist, Big0) == 0 || Cmp(d, dist) < 0) {
			dist.Set(d)
			p = SpherePrimitive
			id = i
		}
	}

	for i := 0; i < len(s.triangles); i++ {
		d := s.triangles[i].Intersect(ray)
		if Cmp(d, Big0) > 0 && (Cmp(dist, Big0) == 0 || Cmp(d, dist) < 0) {
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
