package scenes

import (
	. "raytracer/common"
	. "raytracer/material"
	. "raytracer/objects"
)

func Quads() (Hittable_list, Camera) {

	cam := NewCamera()

	cam.Aspect_ratio = 1.0
	cam.Image_width = 400
	cam.Sample_per_pixel = 100
	cam.Max_depth = 50

	cam.Vfov = 80
	cam.Look_from = NewPoint3(0, 0, 9)
	cam.Look_at = NewPoint3(0, 0, 0)
	cam.Vup = NewVec3(0, 1, 0)
	cam.Log_scanlines = true

	cam.Defocus_angle = 0

	cam.Background = NewColor(0.6, 0.6, 0.6)

	var world Hittable_list

	red := NewLambertian(NewColor(1.0, 0.2, 0.2))
	green := NewLambertian(NewColor(0.2, 1.0, 0.2))
	blue := NewLambertian(NewColor(0.2, 0.2, 1.0))
	orange := NewLambertian(NewColor(1.0, 0.5, 0.0))
	teal := NewLambertian(NewColor(0.2, 0.8, 0.8))

	q1 := NewQuad(NewPoint3(-3, -2, 5), NewVec3(0, 0, -4), NewVec3(0, 4, 0), &red)
	q2 := NewQuad(NewPoint3(-2, -2, 1), NewVec3(4, 0, 0), NewVec3(0, 4, 0), &green) // Back wall
	q3 := NewQuad(NewPoint3(3, -2, 1), NewVec3(0, 0, 4), NewVec3(0, 4, 0), &blue)
	q4 := NewQuad(NewPoint3(-2, 3, 1), NewVec3(4, 0, 0), NewVec3(0, 0, 4), &orange)
	q5 := NewQuad(NewPoint3(-2, -3, 5), NewVec3(4, 0, 0), NewVec3(0, 0, -4), &teal)

	world.Add(&q1)
	world.Add(&q2)
	world.Add(&q3)
	world.Add(&q4)
	world.Add(&q5)

	return world, cam

}
