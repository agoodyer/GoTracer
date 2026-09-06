# GoTracer

GoTracer is a small path tracer written in Go. It renders scenes natively to PNG and also runs in the browser through WebAssembly.

**[Try the browser demo](https://agoodyer.github.io/GoTracer/)**

![A glass sphere, textured spheres, and diffuse objects rendered by GoTracer](sample_renders/sphere_scene.png)

## Why I built it

I wanted a project that would force me to learn Go through a problem with plenty of room for iteration. Ray tracing was a good fit: the first image only needs a little vector math, but realistic materials, meshes, acceleration structures, and parallel rendering each introduce a new problem to work through.

The project began as a native renderer based loosely on [_Ray Tracing in One Weekend_](https://raytracing.github.io/books/RayTracingInOneWeekend.html). I later added STL loading and a browser interface so the renderer could be explored without installing anything.

## What it can render

- Spheres, quads, boxes, and triangle meshes loaded from ASCII or binary STL files
- Diffuse, metallic, dielectric, and emissive materials
- Solid-colour, checkerboard, and image textures
- Reflections, refractions, soft shadows, anti-aliasing, and depth of field
- Translated and rotated objects
- Several ready-made scenes, including a Cornell box, textured planets, and an STL frog

The renderer uses only the Go standard library. The web build uses the same rendering packages as the native program, compiled to WebAssembly.

## How it works

Each pixel is estimated by tracing jittered camera rays into the scene. A ray can miss, hit an emissive surface, or scatter according to the material it reaches. The renderer follows those scattered rays recursively until they leave the scene or reach the configured bounce limit, then averages multiple samples to reduce noise.

Every renderable object implements the same intersection interface. GoTracer groups those objects in a bounding volume hierarchy (BVH), recursively splitting them along the longest axis of their combined bounding box. During a render, a ray can reject whole branches of the hierarchy before running the more expensive shape or triangle intersection tests.

The native renderer divides the image into 24 horizontal chunks and processes them with goroutines. The browser build takes a different approach: JavaScript requests small batches of work from the WebAssembly module, updates the canvas between batches, and yields to the browser so progress and cancellation remain responsive.

## Sample renders

| STL mesh | Earth and Moon |
| :---: | :---: |
| ![An STL mesh rendered with a reflective material](sample_renders/mesh_scene.png) | ![Textured Earth and Moon spheres](sample_renders/earth_scene.png) |
| Glass and metal | Cornell box |
| ![Glass, metal, and diffuse spheres](sample_renders/sphere_scene.png) | ![A Cornell box scene with coloured walls and boxes](sample_renders/cornell_box_scene.png) |

## Profiling the renderer

Ray tracing spends most of its time answering the same question: what does this ray hit next? I used Go's `pprof` tooling to inspect that path while rendering a high-resolution STL scene. The checked-in profile shows the cost concentrated in recursive ray traversal, axis-aligned bounding-box checks, and triangle intersections.

![CPU profile from a high-resolution STL render](sample_renders/cpu_profiling.svg)

The repository does not currently include a repeatable benchmark harness, so performance numbers are deliberately not presented as a portable comparison. Render time depends heavily on the selected scene, image size, samples per pixel, bounce depth, and machine.

## Run it locally

GoTracer targets Go 1.21.4 or newer and has no third-party dependencies.

Clone the repository and render the default scene:

```sh
git clone https://github.com/agoodyer/GoTracer.git
cd GoTracer
go run .
```

The native entry point selects a scene in [`main.go`](main.go) and writes the finished image to `output.png`.

To build and serve the browser version:

```sh
make serve
```

Then open <http://localhost:8080>. The demo exposes six scenes and controls for image width, samples per pixel, and maximum bounce depth.

Useful checks:

```sh
go test ./...
go vet ./...
make wasm
```

There is not yet an automated test suite; the first command currently acts as a compile check across the packages.

## Repository map

- `common/` — vectors, rays, colours, intervals, and shared math
- `material/` — materials, textures, hittable collections, bounding boxes, and the BVH
- `objects/` — cameras, geometric primitives, transforms, and STL parsing
- `scenes/` — native scene definitions
- `wasm/` — browser-facing renderer and embedded demo assets
- `web/` — the static interface deployed to GitHub Pages
- `sample_renders/` — example output and the saved CPU profile

## Status

This is a personal learning project rather than a general-purpose rendering library. Native scene selection and render settings are configured in code, and the worker count is currently fixed rather than selected from the host at runtime. The browser demo is the easiest way to explore the renderer as it stands.

## License

[MIT](LICENSE)
