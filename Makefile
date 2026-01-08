.PHONY: wasm serve clean

# Build WASM module
wasm:
	GOOS=js GOARCH=wasm go build -o web/main.wasm ./wasm

# Start local development server
serve: wasm
	@echo "Starting server at http://localhost:8080"
	cd web && python3 -m http.server 8080

# Clean build artifacts
clean:
	rm -f web/main.wasm

# Build native binary (original behavior)
build:
	go build -o raytracer main.go
