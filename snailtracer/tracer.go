package snailtracer

type Vector struct {
	x, y, z int
}

func (v *Vector) add(u *Vector) *Vector {
	return &Vector{v.x + u.x, v.y + u.y, v.z + u.z}
}

func (v *Vector) sub(u *Vector) *Vector {
	return &Vector{v.x - u.x, v.y - u.y, v.z - u.z}
}

func (v *Vector) scaleMul(m int) *Vector {
	return &Vector{m * v.x, m * v.y, m * v.z}
}

func (v *Vector) mul(u *Vector) *Vector {
	return &Vector{v.x * u.x, v.y * u.y, v.z * u.z}
}

func (v *Vector) scaleDiv(d int) *Vector {
	return &Vector{v.x / d, v.y / d, v.z / d}
}

func (v *Vector) dot(u *Vector) int {
	return v.x*u.x + v.y*u.y + v.z*u.z
}

func (v *Vector) cross(u *Vector) *Vector {
	return &Vector{
		x: v.y*u.z - v.z*u.y,
		y: v.z*u.x - v.x*u.z,
		z: v.x*u.y - v.y*u.x,
	}
}

func (v *Vector) norm() *Vector {
	length := sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	return &Vector{
		x: v.x * 1000000 / length,
		y: v.y * 1000000 / length,
		z: v.z * 1000000 / length,
	}
}

func (v *Vector) clamp() *Vector {
	return &Vector{
		x: clamp(v.x),
		y: clamp(v.y),
		z: clamp(v.z),
	}
}

type Ray struct {
	origin, direction *Vector
	depth             int
	refract           bool
}

type Sphere struct {
	radius     int
	position   *Vector
	emission   *Vector
	color      *Vector
	reflection Material
}

func (s *Sphere) intersect(r *Ray) int {
	op := s.position.sub(r.origin)

	b := op.dot(r.direction) / 1000000
	det := b*b - op.dot(op) + s.radius*s.radius

	if det < 0 {
		return 0
	}

	det = sqrt(det)
	if b-det > 1000 {
		return b - det
	}
	if b+det > 1000 {
		return b + det
	}
	return 0
}

type Triangle struct {
	a, b, c    *Vector
	normal     *Vector
	emission   *Vector
	color      *Vector
	reflection Material
}

func (t *Triangle) intersect(r *Ray) int {
	e1 := t.b.sub(t.a)
	e2 := t.c.sub(t.a)

	p := e1.cross(e2)

	det := e1.dot(p) / 1000000
	if det > -1000 && det < 1000 {
		return 0
	}

	d := r.origin.sub(t.a)

	u := d.dot(p) / det
	if u < 0 || u > 1000000 {
		return 0
	}

	q := d.cross(e1)

	v := r.direction.dot(q) / det
	if v < 0 || u+v > 1000000 {
		return 0
	}

	dist := e2.dot(q) / det
	if dist < 1000 {
		return 0
	}
	return dist
}

type Material int

const (
	material_diffuse Material = iota
	material_specular
	material_refractive
)

type Primitive int

const (
	SpherePrimitive Primitive = iota
	primitive_triangle
)

type Scene struct {
	seed           int
	width, height  int
	camera         *Ray
	deltaX, deltaY *Vector
	spheres        []*Sphere
	triangles      []*Triangle
}

func newScene(w, h int) *Scene {
	s := &Scene{}
	s.width = w
	s.height = h
	s.camera = &Ray{
		origin:    &Vector{50000000, 52000000, 295600000},
		direction: (&Vector{0, -42612, -1000000}).norm(),
	}
	s.deltaX = &Vector{int(s.width * 513500 / s.height), 0, 0}
	s.deltaY = s.deltaX.cross(s.camera.direction).norm().scaleMul(513500).scaleDiv(1000000)
	return s
}

func (s *Scene) rand() int {
	s.seed = 1103515245*s.seed + 12345
	return s.seed
}

func (s *Scene) trace(x, y, spp int) *Vector {
	s.seed = int(uint32(y*s.width + x))

	color := &Vector{0, 0, 0}

	for k := 0; k < spp; k++ {
		pixel := s.deltaX.scaleMul((1000000*x+s.rand()%500000)/s.width - 500000).
			add(s.deltaY.scaleMul((1000000*y+s.rand()%500000)/s.height - 500000)).
			scaleDiv(1000000).
			add(s.camera.direction)
		ray := &Ray{
			origin:    s.camera.origin.add(pixel.scaleMul(140)),
			direction: pixel.norm(),
		}

		color = color.add(s.radiance(ray).scaleDiv(spp))
	}

	return color.clamp().scaleMul(255).scaleDiv(1000000)
}

func (s *Scene) radiance(ray *Ray) *Vector {
	if ray.depth > 10 {
		return &Vector{0, 0, 0}
	}
	dist, p, id := s.traceRay(ray)
	if dist == 0 {
		return &Vector{0, 0, 0}
	}

	var color, emission *Vector
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

	ref := 1
	if color.z > ref {
		ref = color.z
	}
	if color.y > ref {
		ref = color.y
	}
	if color.z > ref {
		ref = color.z
	}

	ray.depth++
	if ray.depth > 5 {
		if s.rand()%1000000 < ref {
			color = color.scaleMul(1000000).scaleDiv(ref)
		} else {
			return emission
		}
	}

	var result *Vector
	if p == SpherePrimitive {
		result = s.radianceSphere(ray, sphere, dist)
	} else {
		result = s.radianceTriangle(ray, triangle, dist)
	}

	return emission.add(color.mul(result).scaleDiv(1000000))
}

func (s *Scene) radianceSphere(ray *Ray, obj *Sphere, dist int) *Vector {
	intersect := ray.origin.add(ray.direction.scaleMul(dist).scaleDiv(1000000))
	normal := intersect.sub(obj.position).norm()

	if obj.reflection == material_diffuse {
		if normal.dot(ray.direction) >= 0 {
			normal = normal.scaleMul(-1)
		}
		return s.diffuse(ray, intersect, normal)
	}
	return s.specular(ray, intersect, normal)
}

func (s *Scene) radianceTriangle(ray *Ray, obj *Triangle, dist int) *Vector {
	intersect := ray.origin.add(ray.direction.scaleMul(dist).scaleDiv(1000000))

	nnt := 666666
	if ray.refract {
		nnt = 1500000
	}
	ddn := obj.normal.dot(ray.direction) / 1000000
	if ddn >= 0 {
		ddn = -ddn
	}
	cos2t := 1000000000000 - nnt*nnt*(1000000000000-ddn*ddn)/1000000000000
	if cos2t < 0 {
		return s.specular(ray, intersect, obj.normal)
	}
	return s.refractive(ray, intersect, obj.normal, nnt, ddn, cos2t)
}

func (s *Scene) diffuse(ray *Ray, intersect, normal *Vector) *Vector {
	r1 := 6283184 * (s.rand() % 1000000) / 1000000
	r2 := s.rand() % 1000000
	r2s := sqrt(r2) * 1000

	var u *Vector
	if abs(normal.x) > 100000 {
		u = &Vector{0, 1000000, 0}
	} else {
		u = &Vector{1000000, 0, 0}
	}
	u = u.cross(normal).norm()
	v := normal.cross(u).norm()

	u = u.scaleMul(cos(r1) * r2s / 1000000).
		add(v.scaleMul(sin(r1) * r2s / 1000000)).
		add(normal.scaleMul(sqrt(1000000-r2) * 1000)).
		norm()
	return s.radiance(&Ray{intersect, u, ray.depth, ray.refract})
}

func (s *Scene) specular(ray *Ray, intersect, normal *Vector) *Vector {
	reflection := ray.direction.sub(normal.scaleMul(2 * normal.dot(ray.direction) / 1000000)).norm()
	return s.radiance(&Ray{intersect, reflection, ray.depth, ray.refract})
}

func (s *Scene) refractive(ray *Ray, intersect, normal *Vector, nnt, ddn, cos2t int) *Vector {
	sign := -1
	if ray.refract {
		sign = 1
	}
	refraction := ray.direction.scaleMul(nnt).
		sub(normal.scaleMul(sign * (ddn*nnt/1000000 + sqrt(cos2t)))).
		scaleDiv(1000000).
		norm()

	c := 1000000 + ddn
	if !ray.refract {
		c = 1000000 - refraction.dot(normal)/1000000
	}
	re := 40000 + (1000000-40000)*c*c*c*c*c/1000000000000000/1000000000000000

	if ray.depth <= 2 {
		refraction = s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).scaleMul(1000000 - re)
		refraction = refraction.add(s.specular(ray, intersect, normal).scaleMul(re))
		return refraction.scaleDiv(1000000)
	}
	if s.rand()%1000000 < 250000+re/2 {
		return s.specular(ray, intersect, normal).scaleMul(re).scaleDiv(250000 + re/2)
	}
	return s.radiance(&Ray{intersect, refraction, ray.depth, !ray.refract}).scaleMul(1000000 - re).scaleDiv(750000 - re/2)
}

func (s *Scene) traceRay(ray *Ray) (int, Primitive, int) {
	var p Primitive
	var id int

	dist := 0

	for i := 0; i < len(s.spheres); i++ {
		d := s.spheres[i].intersect(ray)
		if d > 0 && (dist == 0 || d < dist) {
			dist = d
			p = SpherePrimitive
			id = i
		}
	}

	for i := 0; i < len(s.triangles); i++ {
		d := s.triangles[i].intersect(ray)
		if d > 0 && (dist == 0 || d < dist) {
			dist = d
			p = SpherePrimitive
			id = i
		}
	}

	return dist, p, id
}
