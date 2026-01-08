// +build js,wasm

package main

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"syscall/js"

	. "raytracer/common"
	. "raytracer/material"
	. "raytracer/objects"
	"raytracer/scenes"
)

var currentScene string = "random_spheres"

// Render state for progressive rendering
var renderState struct {
	cam           *Camera
	bvh           Hittable
	pixels        []byte     // RGBA output buffer
	accumulator   []float64  // RGB accumulator (3 floats per pixel)
	indices       []int      // Shuffled pixel indices
	width         int
	height        int
	samples       int
	depth         int
	totalPixels   int
	currentSample int  // For iterative refinement
	initialized   bool
}

// getScenes returns a list of available scene names
func getScenes() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return []interface{}{"random_spheres", "quads", "boxes", "planet", "crystals", "frog"}
	})
}

// setScene sets the current scene by name
func setScene() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return false
		}
		currentScene = args[0].String()
		return true
	})
}

// render renders the current scene and returns pixel data (non-progressive)
func render() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		width := 200
		height := 112
		samples := 10
		depth := 10

		if len(args) >= 1 {
			width = args[0].Int()
		}
		if len(args) >= 2 {
			height = args[1].Int()
		}
		if len(args) >= 3 {
			samples = args[2].Int()
		}
		if len(args) >= 4 {
			depth = args[3].Int()
		}

		var world Hittable_list
		var cam Camera

		switch currentScene {
		case "quads":
			world, cam = scenes.Quads()
		case "boxes":
			world, cam = scenes.Boxes()
		case "planet":
			world, cam = createPlanetScene()
		case "crystals":
			world, cam = scenes.CrystalCave()
		case "frog":
			world, cam = createFrogScene()
		default:
			world, cam = scenes.RandomSpheres()
		}

		// Override camera settings for web demo
		cam.Image_width = width
		cam.Aspect_ratio = float64(width) / float64(height)
		cam.Sample_per_pixel = samples
		cam.Max_depth = depth
		cam.Log_scanlines = false

		// Build BVH and render
		bvh := NewBvh(world.Objects)
		pixels := cam.RenderToBuffer(&bvh)

		// Create JS Uint8ClampedArray from pixel data
		jsArray := js.Global().Get("Uint8ClampedArray").New(len(pixels))
		js.CopyBytesToJS(jsArray, pixels)

		return jsArray
	})
}

// initProgressiveRender initializes the render state for iterative refinement
func initProgressiveRender() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		width := 200
		height := 112
		samples := 10
		depth := 10

		if len(args) >= 1 {
			width = args[0].Int()
		}
		if len(args) >= 2 {
			height = args[1].Int()
		}
		if len(args) >= 3 {
			samples = args[2].Int()
		}
		if len(args) >= 4 {
			depth = args[3].Int()
		}

		var world Hittable_list
		var cam Camera

		switch currentScene {
		case "quads":
			world, cam = scenes.Quads()
		case "boxes":
			world, cam = scenes.Boxes()
		case "planet":
			world, cam = createPlanetScene()
		case "crystals":
			world, cam = scenes.CrystalCave()
		case "frog":
			world, cam = createFrogScene()
		default:
			world, cam = scenes.RandomSpheres()
		}

		// Override camera settings
		cam.Image_width = width
		cam.Aspect_ratio = float64(width) / float64(height)
		cam.Sample_per_pixel = samples
		cam.Max_depth = depth
		cam.Log_scanlines = false
		cam.InitializeForWASM()

		// Build BVH
		bvh := NewBvh(world.Objects)

		// Create shuffled indices for random pixel order within each sample pass
		totalPixels := width * height
		indices := make([]int, totalPixels)
		for i := range indices {
			indices[i] = i
		}
		rand.Shuffle(len(indices), func(i, j int) {
			indices[i], indices[j] = indices[j], indices[i]
		})

		// Store render state
		renderState.cam = &cam
		renderState.bvh = &bvh
		renderState.pixels = make([]byte, totalPixels*4)
		renderState.accumulator = make([]float64, totalPixels*3) // RGB per pixel
		renderState.indices = indices
		renderState.width = width
		renderState.height = height
		renderState.samples = samples
		renderState.depth = depth
		renderState.totalPixels = totalPixels
		renderState.currentSample = 0
		renderState.initialized = true

		// Return info: totalPixels and totalSamples
		return map[string]interface{}{
			"totalPixels":  totalPixels,
			"totalSamples": samples,
		}
	})
}

// renderSamplePass renders ONE sample for all pixels and returns the accumulated result
// This enables iterative refinement - image progressively denoises
func renderSamplePass() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !renderState.initialized {
			return nil
		}

		// Increment current sample
		renderState.currentSample++
		sampleNum := renderState.currentSample

		// Render 1 sample for each pixel (in shuffled order for visual appeal)
		for _, pixelIdx := range renderState.indices {
			x := pixelIdx % renderState.width
			y := pixelIdx / renderState.width

			// Get one ray sample for this pixel
			r := renderState.cam.GetRay(x, y)
			color := renderState.cam.RayColor(&r, renderState.depth, renderState.bvh)

			// Accumulate to float buffer (RGB)
			accIdx := pixelIdx * 3
			renderState.accumulator[accIdx] += color.X()
			renderState.accumulator[accIdx+1] += color.Y()
			renderState.accumulator[accIdx+2] += color.Z()

			// Convert accumulated value to display RGB
			// Average = accumulated / sampleNum, then gamma correct
			scale := 1.0 / float64(sampleNum)
			rVal := math.Sqrt(renderState.accumulator[accIdx] * scale)
			gVal := math.Sqrt(renderState.accumulator[accIdx+1] * scale)
			bVal := math.Sqrt(renderState.accumulator[accIdx+2] * scale)

			// Clamp and convert to bytes
			pixIdx := pixelIdx * 4
			renderState.pixels[pixIdx] = byte(256 * clamp(rVal, 0, 0.999))
			renderState.pixels[pixIdx+1] = byte(256 * clamp(gVal, 0, 0.999))
			renderState.pixels[pixIdx+2] = byte(256 * clamp(bVal, 0, 0.999))
			renderState.pixels[pixIdx+3] = 255
		}

		// Return updated pixel buffer
		jsArray := js.Global().Get("Uint8ClampedArray").New(len(renderState.pixels))
		js.CopyBytesToJS(jsArray, renderState.pixels)

		return jsArray
	})
}

// renderSampleChunk renders 1 sample for a CHUNK of pixels (for hybrid progressive)
// Args: startIdx, endIdx, sampleNum (which sample this is for)
// This allows continuous updates within each sample pass
func renderSampleChunk() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !renderState.initialized {
			return nil
		}

		startIdx := 0
		endIdx := renderState.totalPixels
		sampleNum := renderState.currentSample

		if len(args) >= 1 {
			startIdx = args[0].Int()
		}
		if len(args) >= 2 {
			endIdx = args[1].Int()
		}
		if len(args) >= 3 {
			sampleNum = args[2].Int()
		}

		// Clamp indices
		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > renderState.totalPixels {
			endIdx = renderState.totalPixels
		}

		// Update current sample tracking
		renderState.currentSample = sampleNum

		// Render 1 sample for this chunk of pixels
		for i := startIdx; i < endIdx; i++ {
			pixelIdx := renderState.indices[i]
			x := pixelIdx % renderState.width
			y := pixelIdx / renderState.width

			// Get one ray sample for this pixel
			r := renderState.cam.GetRay(x, y)
			color := renderState.cam.RayColor(&r, renderState.depth, renderState.bvh)

			// Accumulate to float buffer (RGB)
			accIdx := pixelIdx * 3
			renderState.accumulator[accIdx] += color.X()
			renderState.accumulator[accIdx+1] += color.Y()
			renderState.accumulator[accIdx+2] += color.Z()

			// Convert accumulated value to display RGB
			scale := 1.0 / float64(sampleNum)
			rVal := math.Sqrt(renderState.accumulator[accIdx] * scale)
			gVal := math.Sqrt(renderState.accumulator[accIdx+1] * scale)
			bVal := math.Sqrt(renderState.accumulator[accIdx+2] * scale)

			// Clamp and convert to bytes
			pixIdx := pixelIdx * 4
			renderState.pixels[pixIdx] = byte(256 * clamp(rVal, 0, 0.999))
			renderState.pixels[pixIdx+1] = byte(256 * clamp(gVal, 0, 0.999))
			renderState.pixels[pixIdx+2] = byte(256 * clamp(bVal, 0, 0.999))
			renderState.pixels[pixIdx+3] = 255
		}

		// Return updated pixel buffer
		jsArray := js.Global().Get("Uint8ClampedArray").New(len(renderState.pixels))
		js.CopyBytesToJS(jsArray, renderState.pixels)

		return jsArray
	})
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// renderChunk renders a chunk of pixels (for per-pixel progressive mode)
func renderChunk() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if !renderState.initialized {
			return nil
		}

		startIdx := 0
		endIdx := renderState.totalPixels

		if len(args) >= 1 {
			startIdx = args[0].Int()
		}
		if len(args) >= 2 {
			endIdx = args[1].Int()
		}

		// Clamp indices
		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > renderState.totalPixels {
			endIdx = renderState.totalPixels
		}

		// Render this chunk of pixels (all samples per pixel)
		for i := startIdx; i < endIdx; i++ {
			pixelIdx := renderState.indices[i]
			x := pixelIdx % renderState.width
			y := pixelIdx / renderState.width

			// Render single pixel with all samples
			pixel_color := NewColor(0, 0, 0)
			for sample := 0; sample < renderState.samples; sample++ {
				r := renderState.cam.GetRay(x, y)
				pixel_color = pixel_color.Add(renderState.cam.RayColor(&r, renderState.depth, renderState.bvh))
			}

			// Write to buffer
			idx := pixelIdx * 4
			r, g, b := Color_to_rgb(pixel_color, renderState.samples)
			renderState.pixels[idx] = r
			renderState.pixels[idx+1] = g
			renderState.pixels[idx+2] = b
			renderState.pixels[idx+3] = 255
		}

		// Return updated pixel buffer
		jsArray := js.Global().Get("Uint8ClampedArray").New(len(renderState.pixels))
		js.CopyBytesToJS(jsArray, renderState.pixels)

		return jsArray
	})
}

// createPlanetScene creates the planet scene using embedded textures
func createPlanetScene() (Hittable_list, Camera) {
	var world Hittable_list

	cam := NewCamera()
	cam.Aspect_ratio = 16.0 / 9.0
	cam.Image_width = 400
	cam.Sample_per_pixel = 100
	cam.Max_depth = 30
	cam.Vfov = 68
	cam.Look_from = NewPoint3(14, 12, -5)
	cam.Look_at = NewPoint3(0, -2, 0)
	cam.Vup = NewVec3(0, 1, 0)
	cam.Defocus_angle = 0.0
	cam.Focus_dist = 10.0
	cam.Background = NewColor(0.0, 0.0, 0.0085)
	cam.Log_scanlines = false

	// Load embedded textures
	earthData, err1 := GetEmbeddedAsset("earthmap.jpg")
	moonData, err2 := GetEmbeddedAsset("moon.jpg")

	if err1 != nil || err2 != nil {
		// Fallback to colored spheres if textures fail
		earthMat := NewLambertian(NewColor(0.1, 0.3, 0.8))
		moonMat := NewLambertian(NewColor(0.5, 0.5, 0.5))
		
		s1 := NewSphere(NewPoint3(0, -10, 0), 8.0, &earthMat)
		s2 := NewSphere(NewPoint3(0, 6, -5), 2.0, &moonMat)
		world.Add(&s1)
		world.Add(&s2)
	} else {
		// Create textured spheres
		earthTex := NewImage_textureFromBytes(earthData)
		earthSurface := NewTexturedLambertian(&earthTex)
		
		moonTex := NewImage_textureFromBytes(moonData)
		moonSurface := NewTexturedLambertian(&moonTex)

		s1 := NewSphere(NewPoint3(0, -10, 0), 8.0, &earthSurface)
		s2 := NewSphere(NewPoint3(0, 6, -5), 2.0, &moonSurface)
		world.Add(&s1)
		world.Add(&s2)
	}

	// Add sun light
	sunSurface := NewDiffuse_light(NewColor(18, 18, 18))
	s3 := NewSphere(NewPoint3(1400, 1600, 3400), 2000, &sunSurface)
	world.Add(&s3)

	return world, cam
}

func main() {
	c := make(chan struct{})

	js.Global().Set("goGetScenes", getScenes())
	js.Global().Set("goSetScene", setScene())
	js.Global().Set("goRender", render())
	js.Global().Set("goInitProgressiveRender", initProgressiveRender())
	js.Global().Set("goRenderChunk", renderChunk())
	js.Global().Set("goRenderSamplePass", renderSamplePass())
	js.Global().Set("goRenderSampleChunk", renderSampleChunk())

	// Log that WASM is ready
	js.Global().Get("console").Call("log", "Go WASM raytracer initialized")

	// Keep the program running
	<-c
}

// createFrogScene creates the frog scene using embedded STL
func createFrogScene() (Hittable_list, Camera) {
	var world Hittable_list

	cam := NewCamera()
	cam.Aspect_ratio = 16.0 / 9.0
	cam.Image_width = 400
	cam.Sample_per_pixel = 100
	cam.Max_depth = 30
	cam.Vfov = 40
	cam.Look_from = NewPoint3(20, 20, 40)
	cam.Look_at = NewPoint3(0, 5, 0)
	cam.Vup = NewVec3(0, 1, 0)
	cam.Defocus_angle = 0.6
	cam.Focus_dist = 10.0
	cam.Background = NewColor(0.7, 0.8, 1.0) // Skyish background
	cam.Log_scanlines = false

	// Load embedded asset
	frogData, err := GetEmbeddedAsset("lowpolyfrog.stl")
	
	if err == nil {
		frogMat := NewLambertian(NewColor(0.2, 0.5, 0.2)) // Green frog
		
		// Create reader from bytes
		reader := bytes.NewReader(frogData)
		

		// Parse mesh
		frogMesh := NewMeshFromReader(reader, &frogMat, 0.5) // Scale down
		
		fmt.Printf("Loaded frog mesh with %d triangles\n", len(frogMesh.Objects))
		js.Global().Get("console").Call("log", fmt.Sprintf("Loaded frog mesh with %d triangles", len(frogMesh.Objects)))

		// Create BVH for mesh acceleration
		frogBvh := NewBvh(frogMesh.Objects)

		// Rotate to face camera (Camera at 20, 20, 40)
		rotatedFrogY := NewRotationZ(&frogBvh, 20.0)
		rotatedFrogX := NewRotationX(&rotatedFrogY, -270.0) // Rotate -90 on X so it sits flat if it was standing upright
		
		// Translate to ground (ensure it sits on y=0)
		translatedFrog := NewTranslation(rotatedFrogX, NewVec3(0, 0, 0))

		world.Add(translatedFrog)
	} else {
		// Fallback: Green sphere
		frogMat := NewLambertian(NewColor(0.2, 0.5, 0.2))
		sphere := NewSphere(NewPoint3(0, 5, 0), 5.0, &frogMat)
		world.Add(&sphere)
	}

	// Ground plane
	groundMat := NewLambertian(NewColor(0.3, 0.3, 0.6))
	ground := NewSphere(NewPoint3(0, -1000, 0), 1000, &groundMat)
	world.Add(&ground)

	// Sun light
	sunKey := NewDiffuse_light(NewColor(10, 10, 10))
	sun := NewSphere(NewPoint3(50, 100, 50), 10, &sunKey)
	world.Add(&sun)

	return world, cam
}
