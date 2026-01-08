package objects

import (
	"math"
	. "raytracer/common"
	. "raytracer/material"
)

type Translation struct {
	object Hittable
	offset Vec3
	bbox   Aabb
}

func NewTranslation(object Hittable, offset Vec3) Translation {

	bbox := object.Bounding_box()
	bbox = bbox.Add(offset)
	return Translation{object: object, offset: offset, bbox: bbox}

}

func (tr Translation) Hit(r *Ray, ray_t Interval, rec *Hit_record) bool {

	offset_r := NewRay(r.Origin.Sub(tr.offset), r.Direction)

	if !tr.object.Hit(&offset_r, ray_t, rec) {
		return false
	}

	rec.P = rec.P.Add(tr.offset)

	return true

}

func (tr Translation) Bounding_box() Aabb {

	return NewAabb(Universe, Universe, Universe)
	// return tr.bbox
}

type RotationY struct {
	object    Hittable
	cos_theta float64
	sin_theta float64
	bbox      Aabb
}

func (rt RotationY) Hit(r *Ray, ray_t Interval, rec *Hit_record) bool {

	x := rt.cos_theta*r.Origin.X() - rt.sin_theta*r.Origin.Z()
	z := rt.sin_theta*r.Origin.X() + rt.cos_theta*r.Origin.Z()

	dir_x := rt.cos_theta*r.Direction.X() - rt.sin_theta*r.Direction.Z()
	dir_z := rt.sin_theta*r.Direction.X() + rt.cos_theta*r.Direction.Z()

	rotated_r := NewRay(NewPoint3(x, r.Origin.Y(), z), NewVec3(dir_x, r.Direction.Y(), dir_z))

	if !rt.object.Hit(&rotated_r, ray_t, rec) {
		return false
	}

	p_x := rt.cos_theta*rec.P.X() + rt.sin_theta*rec.P.Z()
	p_z := -rt.sin_theta*rec.P.X() + rt.cos_theta*rec.P.Z()
	rec.P = NewVec3(p_x, rec.P.Y(), p_z)

	normal_x := rt.cos_theta*rec.Normal.X() + rt.sin_theta*rec.Normal.Z()
	normal_z := -rt.sin_theta*rec.Normal.X() + rt.cos_theta*rec.Normal.Z()

	rec.Normal = NewVec3(normal_x, rec.Normal.Y(), normal_z)
	return true

}

func NewRotationY(object Hittable, angle float64) RotationY {
	radians := Degrees_to_radians(angle)
	cos_theta := math.Cos(radians)
	sin_theta := math.Sin(radians)

	bbox := object.Bounding_box()

	min_x, min_y, min_z := Infinity, Infinity, Infinity
	max_x, max_y, max_z := -Infinity, -Infinity, -Infinity

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float64(i)*bbox.X().Max + (1-float64(i))*bbox.X().Min
				y := float64(j)*bbox.Y().Max + (1-float64(j))*bbox.Y().Min
				z := float64(k)*bbox.Z().Max + (1-float64(k))*bbox.Z().Min

				new_x := cos_theta*x + sin_theta*z
				new_z := -sin_theta*x + cos_theta*z

				tester := NewVec3(new_x, y, new_z)

				min_x = math.Min(min_x, tester.X())
				max_x = math.Max(max_x, tester.X())

				min_y = math.Min(min_y, tester.Y())
				max_y = math.Max(max_y, tester.Y())

				min_z = math.Min(min_z, tester.Z())
				max_z = math.Max(max_z, tester.Z())

			}
		}

		bbox = NewAabbFromPoints(NewPoint3(min_x, min_y, min_z), NewPoint3(max_x, max_y, max_z))

	}

	return RotationY{object: object, cos_theta: cos_theta, sin_theta: sin_theta, bbox: bbox}

}

func (rt RotationY) Bounding_box() Aabb {
	return NewAabb(Universe, Universe, Universe)
	// return rt.bbox
}

type RotationX struct {
	object    Hittable
	cos_theta float64
	sin_theta float64
	bbox      Aabb
}

func (rt RotationX) Hit(r *Ray, ray_t Interval, rec *Hit_record) bool {

	y := rt.cos_theta*r.Origin.Y() - rt.sin_theta*r.Origin.Z()
	z := rt.sin_theta*r.Origin.Y() + rt.cos_theta*r.Origin.Z()

	dir_y := rt.cos_theta*r.Direction.Y() - rt.sin_theta*r.Direction.Z()
	dir_z := rt.sin_theta*r.Direction.Y() + rt.cos_theta*r.Direction.Z()

	rotated_r := NewRay(NewPoint3(r.Origin.X(), y, z), NewVec3(r.Direction.X(), dir_y, dir_z))

	if !rt.object.Hit(&rotated_r, ray_t, rec) {
		return false
	}

	p_y := rt.cos_theta*rec.P.Y() + rt.sin_theta*rec.P.Z()
	p_z := -rt.sin_theta*rec.P.Y() + rt.cos_theta*rec.P.Z()
	rec.P = NewVec3(rec.P.X(), p_y, p_z)

	normal_y := rt.cos_theta*rec.Normal.Y() + rt.sin_theta*rec.Normal.Z()
	normal_z := -rt.sin_theta*rec.Normal.Y() + rt.cos_theta*rec.Normal.Z()

	rec.Normal = NewVec3(rec.Normal.X(), normal_y, normal_z)
	return true

}

func NewRotationX(object Hittable, angle float64) RotationX {
	radians := Degrees_to_radians(angle)
	cos_theta := math.Cos(radians)
	sin_theta := math.Sin(radians)

	bbox := object.Bounding_box()

	min_x, min_y, min_z := Infinity, Infinity, Infinity
	max_x, max_y, max_z := -Infinity, -Infinity, -Infinity

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float64(i)*bbox.X().Max + (1-float64(i))*bbox.X().Min
				y := float64(j)*bbox.Y().Max + (1-float64(j))*bbox.Y().Min
				z := float64(k)*bbox.Z().Max + (1-float64(k))*bbox.Z().Min

				new_y := cos_theta*y - sin_theta*z
				new_z := sin_theta*y + cos_theta*z

				tester := NewVec3(x, new_y, new_z)

				min_x = math.Min(min_x, tester.X())
				max_x = math.Max(max_x, tester.X())

				min_y = math.Min(min_y, tester.Y())
				max_y = math.Max(max_y, tester.Y())

				min_z = math.Min(min_z, tester.Z())
				max_z = math.Max(max_z, tester.Z())

			}
		}

		bbox = NewAabbFromPoints(NewPoint3(min_x, min_y, min_z), NewPoint3(max_x, max_y, max_z))

	}

	return RotationX{object: object, cos_theta: cos_theta, sin_theta: sin_theta, bbox: bbox}

}

func (rt RotationX) Bounding_box() Aabb {
	return NewAabb(Universe, Universe, Universe)
}

type RotationZ struct {
	object    Hittable
	cos_theta float64
	sin_theta float64
	bbox      Aabb
}

func (rt RotationZ) Hit(r *Ray, ray_t Interval, rec *Hit_record) bool {

	x := rt.cos_theta*r.Origin.X() - rt.sin_theta*r.Origin.Y()
	y := rt.sin_theta*r.Origin.X() + rt.cos_theta*r.Origin.Y()

	dir_x := rt.cos_theta*r.Direction.X() - rt.sin_theta*r.Direction.Y()
	dir_y := rt.sin_theta*r.Direction.X() + rt.cos_theta*r.Direction.Y()

	rotated_r := NewRay(NewPoint3(x, y, r.Origin.Z()), NewVec3(dir_x, dir_y, r.Direction.Z()))

	if !rt.object.Hit(&rotated_r, ray_t, rec) {
		return false
	}

	p_x := rt.cos_theta*rec.P.X() + rt.sin_theta*rec.P.Y()
	p_y := -rt.sin_theta*rec.P.X() + rt.cos_theta*rec.P.Y()
	rec.P = NewVec3(p_x, p_y, rec.P.Z())

	normal_x := rt.cos_theta*rec.Normal.X() + rt.sin_theta*rec.Normal.Y()
	normal_y := -rt.sin_theta*rec.Normal.X() + rt.cos_theta*rec.Normal.Y()

	rec.Normal = NewVec3(normal_x, normal_y, rec.Normal.Z())
	return true

}

func NewRotationZ(object Hittable, angle float64) RotationZ {
	radians := Degrees_to_radians(angle)
	cos_theta := math.Cos(radians)
	sin_theta := math.Sin(radians)

	bbox := object.Bounding_box()

	min_x, min_y, min_z := Infinity, Infinity, Infinity
	max_x, max_y, max_z := -Infinity, -Infinity, -Infinity

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float64(i)*bbox.X().Max + (1-float64(i))*bbox.X().Min
				y := float64(j)*bbox.Y().Max + (1-float64(j))*bbox.Y().Min
				z := float64(k)*bbox.Z().Max + (1-float64(k))*bbox.Z().Min

				new_x := cos_theta*x - sin_theta*y
				new_y := sin_theta*x + cos_theta*y

				tester := NewVec3(new_x, new_y, z)

				min_x = math.Min(min_x, tester.X())
				max_x = math.Max(max_x, tester.X())

				min_y = math.Min(min_y, tester.Y())
				max_y = math.Max(max_y, tester.Y())

				min_z = math.Min(min_z, tester.Z())
				max_z = math.Max(max_z, tester.Z())

			}
		}

		bbox = NewAabbFromPoints(NewPoint3(min_x, min_y, min_z), NewPoint3(max_x, max_y, max_z))

	}

	return RotationZ{object: object, cos_theta: cos_theta, sin_theta: sin_theta, bbox: bbox}

}

func (rt RotationZ) Bounding_box() Aabb {
	return NewAabb(Universe, Universe, Universe)
}
