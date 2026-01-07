// +build js,wasm

package main

import (
	"math/rand"
	"syscall/js"

	. "raytracer/common"
	. "raytracer/material"
	. "raytracer/objects"
	"raytracer/scenes"
)

var currentScene string = "random_spheres"

// Render state for chunked progressive rendering
var renderState struct {
	cam         *Camera
	bvh         Hittable
	pixels      []byte
	indices     []int
	width       int
	height      int
	samples     int
	depth       int
	totalPixels int
	initialized bool
}

// getScenes returns a list of available scene names
func getScenes() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return []interface{}{"random_spheres", "quads"}
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

// initProgressiveRender initializes the render state and returns shuffled indices
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

		// Create shuffled indices
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
		renderState.indices = indices
		renderState.width = width
		renderState.height = height
		renderState.samples = samples
		renderState.depth = depth
		renderState.totalPixels = totalPixels
		renderState.initialized = true

		return totalPixels
	})
}

// renderChunk renders a chunk of pixels and returns the updated buffer
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

		// Render this chunk of pixels
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

func main() {
	c := make(chan struct{})

	js.Global().Set("goGetScenes", getScenes())
	js.Global().Set("goSetScene", setScene())
	js.Global().Set("goRender", render())
	js.Global().Set("goInitProgressiveRender", initProgressiveRender())
	js.Global().Set("goRenderChunk", renderChunk())

	// Log that WASM is ready
	js.Global().Get("console").Call("log", "Go WASM raytracer initialized")

	// Keep the program running
	<-c
}
