package scenes

import (
	"math"
	"math/rand"

	. "raytracer/common"
	. "raytracer/material"
	. "raytracer/objects"
)

// CrystalCave creates a dark cave with glowing and refractive crystals
func CrystalCave() (Hittable_list, Camera) {
	var world Hittable_list

	cam := NewCamera()
	cam.Aspect_ratio = 16.0 / 9.0
	cam.Image_width = 400
	cam.Sample_per_pixel = 100
	cam.Max_depth = 50
	cam.Vfov = 60
	cam.Look_from = NewPoint3(0, 3, 12)
	cam.Look_at = NewPoint3(0, 2, 0)
	cam.Vup = NewVec3(0, 1, 0)
	cam.Defocus_angle = 0.0
	cam.Focus_dist = 10.0
	cam.Background = NewColor(0.04, 0.04, 0.06) // Dark but with slight ambient visibility
	cam.Log_scanlines = true

	// Cave floor - dark rocky surface
	floorMat := NewLambertian(NewColor(0.05, 0.05, 0.06))
	floor := NewQuad(NewPoint3(-20, 0, -20), NewVec3(40, 0, 0), NewVec3(0, 0, 40), &floorMat)
	world.Add(&floor)

	// Cave ceiling
	ceilingMat := NewLambertian(NewColor(0.03, 0.03, 0.04))
	ceiling := NewQuad(NewPoint3(-20, 12, -20), NewVec3(40, 0, 0), NewVec3(0, 0, 40), &ceilingMat)
	world.Add(&ceiling)

	// Back wall
	backMat := NewLambertian(NewColor(0.04, 0.04, 0.05))
	back := NewQuad(NewPoint3(-20, 0, -10), NewVec3(40, 0, 0), NewVec3(0, 12, 0), &backMat)
	world.Add(&back)

	// Glowing crystal colors (emissive)
	glowColors := []Color{
		NewColor(0.8, 0.2, 1.0),  // Purple
		NewColor(0.2, 0.8, 1.0),  // Cyan
		NewColor(1.0, 0.3, 0.5),  // Pink
		NewColor(0.3, 1.0, 0.5),  // Green
		NewColor(0.4, 0.4, 1.0),  // Blue
	}

	// Create glowing crystals (emissive pointed shapes)
	crystalPositions := []Point3{
		NewPoint3(-4, 0, 2),
		NewPoint3(3, 0, 0),
		NewPoint3(-1, 0, -3),
		NewPoint3(5, 0, 4),
		NewPoint3(-6, 0, -2),
		NewPoint3(0.5, 0, 5),
		NewPoint3(-3, 0, 6),
	}

	for i, pos := range crystalPositions {
		// Vary crystal properties
		height := 1.5 + rand.Float64()*2.5
		baseSize := 0.3 + rand.Float64()*0.4
		glowIntensity := 2.0 + rand.Float64()*3.0
		
		glowColor := glowColors[i%len(glowColors)]
		glowMat := NewDiffuse_light(glowColor.Mult(glowIntensity))
		
		// Create crystal as a set of triangles (pyramid shape)
		addCrystal(&world, pos, height, baseSize, &glowMat)
	}

	// Add some glass/diamond crystals (refractive)
	glassMat := NewDielectric(2.4) // Diamond-like refraction
	
	glassCrystalPositions := []Point3{
		NewPoint3(1, 0, 2),
		NewPoint3(-2, 0, 4),
		NewPoint3(4, 0, -1),
	}

	for _, pos := range glassCrystalPositions {
		height := 2.0 + rand.Float64()*1.5
		baseSize := 0.4 + rand.Float64()*0.3
		addCrystal(&world, pos, height, baseSize, &glassMat)
	}

	// Add ceiling stalactite crystals (hanging down, glowing)
	stalactitePositions := []Point3{
		NewPoint3(-2, 12, 1),
		NewPoint3(2, 12, -2),
		NewPoint3(0, 12, 3),
		NewPoint3(-4, 12, -1),
		NewPoint3(4, 12, 2),
	}

	for i, pos := range stalactitePositions {
		height := 1.0 + rand.Float64()*2.0
		baseSize := 0.2 + rand.Float64()*0.25
		glowIntensity := 1.5 + rand.Float64()*2.0
		
		glowColor := glowColors[(i+2)%len(glowColors)]
		glowMat := NewDiffuse_light(glowColor.Mult(glowIntensity))
		
		// Hanging crystal (inverted)
		addInvertedCrystal(&world, pos, height, baseSize, &glowMat)
	}

	// Add some reflective metal crystal clusters
	metalMat := NewMetal(NewColor(0.7, 0.7, 0.8), 0.1)
	addCrystal(&world, NewPoint3(6, 0, 1), 1.8, 0.35, &metalMat)
	addCrystal(&world, NewPoint3(-5, 0, 3), 1.5, 0.3, &metalMat)

	// Small accent spheres (like crystal orbs)
	orbMat := NewDielectric(1.5)
	orb1 := NewSphere(NewPoint3(-1, 0.5, 7), 0.5, &orbMat)
	orb2 := NewSphere(NewPoint3(2, 0.4, 6), 0.4, &orbMat)
	world.Add(&orb1)
	world.Add(&orb2)

	return world, cam
}

// addCrystal adds a crystal (pyramid) shape at the given position
func addCrystal(world *Hittable_list, base Point3, height float64, baseSize float64, mat Material) {
	// Tip of crystal
	tip := NewPoint3(base.X(), base.Y()+height, base.Z())
	
	// Base corners (hexagonal-ish for more crystal-like appearance)
	corners := make([]Point3, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		corners[i] = NewPoint3(
			base.X()+baseSize*math.Cos(angle),
			base.Y(),
			base.Z()+baseSize*math.Sin(angle),
		)
	}

	// Create triangular faces from base to tip
	for i := 0; i < 6; i++ {
		next := (i + 1) % 6
		tri := Triangle(corners[i], corners[next], tip, mat)
		world.Add(&tri)
	}

	// Optional: add base triangles for closed shape
	center := base
	for i := 0; i < 6; i++ {
		next := (i + 1) % 6
		tri := Triangle(corners[i], center, corners[next], mat)
		world.Add(&tri)
	}
}

// addInvertedCrystal adds an upside-down crystal (stalactite)
func addInvertedCrystal(world *Hittable_list, ceiling Point3, height float64, baseSize float64, mat Material) {
	// Tip points down
	tip := NewPoint3(ceiling.X(), ceiling.Y()-height, ceiling.Z())
	
	// Base corners at ceiling level
	corners := make([]Point3, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		corners[i] = NewPoint3(
			ceiling.X()+baseSize*math.Cos(angle),
			ceiling.Y(),
			ceiling.Z()+baseSize*math.Sin(angle),
		)
	}

	// Create triangular faces from ceiling to tip
	for i := 0; i < 6; i++ {
		next := (i + 1) % 6
		tri := Triangle(corners[i], tip, corners[next], mat)
		world.Add(&tri)
	}
}
