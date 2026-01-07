package objects

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"math/rand"
	"os"
	. "raytracer/common"
	. "raytracer/material"
	"time"
)

type Camera struct {
	Aspect_ratio     float64
	Image_width      int
	Sample_per_pixel int
	image_height     int
	center           Point3
	pixel00_loc      Point3
	pixel_delta_u    Vec3
	pixel_delta_v    Vec3
	Max_depth        int
	Vfov             float64
	Look_from        Point3
	Look_at          Point3
	Vup              Point3
	u, v, w          Vec3
	Defocus_angle    float64
	Focus_dist       float64
	defocus_disk_u   Vec3
	defocus_disk_v   Vec3
	Background       Color
	Log_scanlines    bool
	// SkySphere        Sphere
}

func NewCamera() Camera {

	var c Camera = Camera{}

	c.Aspect_ratio = 16.0 / 9.0
	c.Image_width = 100

	c.Sample_per_pixel = 300 //250
	c.Max_depth = 30         //50

	c.Vfov = 60
	c.Look_from = NewPoint3(13, 11, -4)
	c.Look_at = NewPoint3(0, -2, 0)
	c.Vup = NewVec3(0, 1, 0)

	c.Defocus_angle = 0.0
	c.Focus_dist = 10.0

	c.Background = NewColor(0.0, 0.0, 0.025)

	c.Log_scanlines = false

	return c
}

func (c *Camera) initialize() {

	//Calculate image height
	c.image_height = int(float64(c.Image_width) / c.Aspect_ratio)
	if c.image_height < 1 {
		c.image_height = 1
	}

	c.center = c.Look_from

	//Camera
	// focal_length := (c.Look_from.Sub(c.Look_at).Length())
	theta := Degrees_to_radians(c.Vfov)
	h := math.Tan(theta / 2)

	viewport_height := 2.0 * h * c.Focus_dist
	viewport_width := viewport_height * float64(c.Image_width) / float64(c.image_height)

	c.w = Unit_vector(c.Look_from.Sub(c.Look_at))
	c.u = Unit_vector(Cross(c.Vup, c.w))
	c.v = Cross(c.w, c.u)

	viewport_u := c.u.Mult(viewport_width)
	viewport_v := c.v.Mult(-viewport_height)

	c.pixel_delta_u = viewport_u.Div(float64(c.Image_width))
	c.pixel_delta_v = viewport_v.Div(float64(c.image_height))

	viewport_upper_left := c.center.Sub(c.w.Mult(c.Focus_dist)).Sub(viewport_u.Div(2)).Sub(viewport_v.Div(2))

	c.pixel00_loc = viewport_upper_left.Add(c.pixel_delta_u.Add(c.pixel_delta_v).Mult(0.5))

	defocus_radius := c.Focus_dist * math.Tan(Degrees_to_radians(c.Defocus_angle/2))
	c.defocus_disk_u = c.u.Mult(defocus_radius)
	c.defocus_disk_u = c.v.Mult(defocus_radius)

}

func (c *Camera) Render(world Hittable) {

	start := time.Now() //time execution

	c.initialize()

	img := image.NewRGBA(image.Rect(0, 0, c.Image_width, c.image_height))

	for j := 0; j < c.image_height; j++ {

		if c.Log_scanlines {
			fmt.Printf("Scanlines remaining: %d\n", c.image_height-j)
		}

		for i := 0; i < c.Image_width; i++ {

			pixel_color := NewColor(0, 0, 0)

			for sample := 0; sample < c.Sample_per_pixel; sample++ {
				r := c.get_ray(i, j)
				pixel_color = pixel_color.Add(ray_color(c, &r, c.Max_depth, world))
			}

			Write_color(pixel_color, c.Sample_per_pixel, img, i, j)

		}

	}

	//THREAD TESTING
	// var wg sync.WaitGroup

	// numWorkers := 2

	// wg.Add(numWorkers)

	// go c.renderBlock(0, c.image_height/4, world, img, &wg)

	// go c.renderBlock(c.image_height/4, c.image_height/4, world, img, &wg)

	// wg.Wait()

	//END TEST

	// Create a file to save the image
	file, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Encode the image to PNG and save to the file
	if err := png.Encode(file, img); err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~\nElapsed Time: ", elapsed, "\n~~~~~~~~~~~~~~~~~~~~~~~~~~\n")

}

// RenderToBuffer renders the scene and returns RGBA pixel data as []byte
// This is used for WASM builds where we can't write to files
func (c *Camera) RenderToBuffer(world Hittable) []byte {
	c.initialize()

	// Create pixel buffer (RGBA format for HTML canvas)
	pixels := make([]byte, c.Image_width*c.image_height*4)

	for j := 0; j < c.image_height; j++ {
		for i := 0; i < c.Image_width; i++ {
			pixel_color := NewColor(0, 0, 0)

			for sample := 0; sample < c.Sample_per_pixel; sample++ {
				r := c.get_ray(i, j)
				pixel_color = pixel_color.Add(ray_color(c, &r, c.Max_depth, world))
			}

			// Convert to RGBA bytes
			idx := (j*c.Image_width + i) * 4
			r, g, b := Color_to_rgb(pixel_color, c.Sample_per_pixel)
			pixels[idx] = r
			pixels[idx+1] = g
			pixels[idx+2] = b
			pixels[idx+3] = 255 // Alpha
		}
	}

	return pixels
}

// RenderProgressiveWASM renders pixels in random order with periodic callbacks for live updates
// This creates a "reveal" effect where the image progressively fills in
func (c *Camera) RenderProgressiveWASM(world Hittable, pixels []byte, updateCallback func(), updatePercent int) {
	c.initialize()
	totalPixels := c.Image_width * c.image_height

	// Create shuffled pixel indices using Fisher-Yates algorithm
	indices := make([]int, totalPixels)
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	// Calculate update interval (default to every 1% if not specified)
	if updatePercent <= 0 {
		updatePercent = 1
	}
	updateInterval := totalPixels * updatePercent / 100
	if updateInterval < 1 {
		updateInterval = 1
	}

	// Render pixels in shuffled order
	for count, pixelIdx := range indices {
		i := pixelIdx % c.Image_width
		j := pixelIdx / c.Image_width

		// Render single pixel with all samples
		pixel_color := NewColor(0, 0, 0)
		for sample := 0; sample < c.Sample_per_pixel; sample++ {
			r := c.get_ray(i, j)
			pixel_color = pixel_color.Add(ray_color(c, &r, c.Max_depth, world))
		}

		// Write to shared buffer
		idx := pixelIdx * 4
		r, g, b := Color_to_rgb(pixel_color, c.Sample_per_pixel)
		pixels[idx] = r
		pixels[idx+1] = g
		pixels[idx+2] = b
		pixels[idx+3] = 255

		// Periodic update callback
		if (count+1)%updateInterval == 0 {
			updateCallback()
		}
	}

	// Final callback to ensure last pixels are shown
	updateCallback()
}

type Chunk struct {
	world *Hittable
	rect  image.Rectangle
}

func (c *Camera) renderChunk(chunk Chunk, img *image.RGBA) {

	for j := chunk.rect.Bounds().Min.Y; j < chunk.rect.Bounds().Max.Y; j++ {

		// cmd := exec.Command("cmd", "/c", "cls") // Windows system
		// cmd.Stdout = os.Stdout
		// cmd.Run()
		// fmt.Println(float64(j) / float64(chunk.rect.Bounds().Max.Y) * 100)
		for i := chunk.rect.Bounds().Min.X; i < chunk.rect.Bounds().Max.X; i++ {

			pixel_color := NewColor(0, 0, 0)

			for sample := 0; sample < c.Sample_per_pixel; sample++ {
				r := c.get_ray(i, j)
				pixel_color = pixel_color.Add(ray_color(c, &r, c.Max_depth, *chunk.world))
			}
			Write_color(pixel_color, c.Sample_per_pixel, img, i, j)

		}

	}

}

func (c *Camera) Chunk_image(n int, world Hittable) []Chunk {

	var chunks []Chunk

	chunk_size := c.image_height / n

	for i := 0; i < n; i++ {
		chunks = append(chunks, Chunk{rect: image.Rect(0, i*chunk_size, c.Image_width, i*(chunk_size+1)), world: &world})
	}

	return chunks
}

func worker(c *Camera, chunk *Chunk, img *image.RGBA, done chan bool) {
	c.renderChunk(*chunk, img)
	fmt.Print("done!")
	done <- true
}

func (c *Camera) RenderMultithreaded(world Hittable) {

	start := time.Now()

	c.initialize()
	img := image.NewRGBA(image.Rect(0, 0, c.Image_width, c.image_height))

	done := make(chan bool)

	n_processes := 24

	for i := 0; i < n_processes; i++ {
		chunk := Chunk{rect: image.Rect(0, i*c.image_height/n_processes, c.Image_width, (i+1)*c.image_height/n_processes), world: &world}
		go worker(c, &chunk, img, done)
	}

	for i := 0; i < n_processes; i++ {
		<-done
	}

	file, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Encode the image to PNG and save to the file
	if err := png.Encode(file, img); err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	fmt.Print("~~~~~~~~~~~~~~~~~~~~~~~~~~\nElapsed Time: ", elapsed, "\n~~~~~~~~~~~~~~~~~~~~~~~~~~\n")

}

func ray_color(c *Camera, r *Ray, depth int, world Hittable) Color {
	var rec Hit_record

	if depth <= 0 {
		return NewColor(0, 0, 0)
	}

	if !world.Hit(r, NewInterval(0.001, Infinity), &rec) {
		return c.Background

		// return c.SkySphere.Mat.

	}

	// direction := rec.Normal.Add(Random_unit_vector())

	// return ray_color(NewRay(rec.P, direction), depth-1, world).Mult(0.50)

	var scattered Ray
	var attenuation Color

	m := *rec.Mat

	color_from_emission := m.Emitted(rec.U, rec.V, &rec.P)

	if !m.Scatter(r, &rec, &attenuation, &scattered) {
		return color_from_emission
	}

	// if rec.Mat.Scatter(r, &rec, &attenuation, &scattered) {
	color_from_scatter := ComponentMultiply(attenuation, ray_color(c, &scattered, depth-1, world))
	// }

	return color_from_emission.Add(color_from_scatter)

	// unit_direction := Unit_vector(r.Direction)
	// a := 0.5 * (unit_direction.Y() + 1.0)
	// return NewColor(1.0, 1.0, 1.0).Mult(1.0 - a).Add(NewColor(0.5, 0.7, 1.0).Mult(a))
}

func (c *Camera) get_ray(i int, j int) Ray {

	pixel_center := c.pixel00_loc.Add(c.pixel_delta_u.Mult(float64(i)).Add(c.pixel_delta_v.Mult(float64(j))))
	pixel_sample := pixel_center.Add(c.pixel_sample_square())

	ray_origin := c.center

	if c.Defocus_angle > 0 {
		ray_origin = c.defocus_disk_sample()
	}

	ray_direction := pixel_sample.Sub(ray_origin)

	return NewRay(ray_origin, ray_direction)
}

func (c *Camera) defocus_disk_sample() Point3 {
	p := Random_in_unit_disk()
	return c.center.Add(c.defocus_disk_u.Mult(p.X())).Add(c.defocus_disk_v.Mult(p.Y()))
}

func (c *Camera) pixel_sample_square() Vec3 {
	px := -0.5 + rand.Float64()
	py := -0.5 + rand.Float64()
	return c.pixel_delta_u.Mult(px).Add(c.pixel_delta_v.Mult(py))
}

// InitializeForWASM is an exported wrapper for initialize() for use in WASM
func (c *Camera) InitializeForWASM() {
	c.initialize()
}

// GetRay is an exported wrapper for get_ray() for use in WASM
func (c *Camera) GetRay(i, j int) Ray {
	return c.get_ray(i, j)
}

// RayColor is an exported wrapper for ray_color() for use in WASM
func (c *Camera) RayColor(r *Ray, depth int, world Hittable) Color {
	return ray_color(c, r, depth, world)
}

