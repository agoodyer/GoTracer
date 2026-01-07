// Go Raytracer WASM Interface

const go = new Go();
let wasmReady = false;
let isRendering = false;

// DOM elements
const canvas = document.getElementById('canvas');
const ctx = canvas.getContext('2d');
const renderBtn = document.getElementById('renderBtn');
const sceneSelect = document.getElementById('scene');
const widthInput = document.getElementById('width');
const samplesInput = document.getElementById('samples');
const depthInput = document.getElementById('depth');
const progressiveCheckbox = document.getElementById('progressive');
const statusDiv = document.getElementById('status');

// Show status message
function showStatus(message, type = 'loading') {
    statusDiv.textContent = message;
    statusDiv.className = `status ${type}`;
    statusDiv.classList.remove('hidden');
}

// Hide status
function hideStatus() {
    statusDiv.classList.add('hidden');
}

// Initialize WASM
async function initWasm() {
    showStatus('Loading WebAssembly module...');

    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
        );

        go.run(result.instance);
        wasmReady = true;

        renderBtn.textContent = 'Render';
        renderBtn.disabled = false;
        showStatus('Ready! Click Render to start.', 'success');

        setTimeout(hideStatus, 2000);
    } catch (err) {
        console.error('Failed to load WASM:', err);
        showStatus('Failed to load WebAssembly module. See console for details.', 'error');
    }
}

// Render the scene (standard mode - blocking)
async function renderStandard(width, height, samples, depth) {
    const startTime = performance.now();

    const pixels = goRender(width, height, samples, depth);
    const imageData = new ImageData(pixels, width, height);
    ctx.putImageData(imageData, 0, 0);

    const elapsed = ((performance.now() - startTime) / 1000).toFixed(2);
    showStatus(`Rendered in ${elapsed}s`, 'success');
}

// Render the scene with progressive updates (chunked)
async function renderProgressiveChunked(width, height, samples, depth) {
    const startTime = performance.now();

    // Initialize the render (sets up scene, BVH, shuffled indices)
    const totalPixels = goInitProgressiveRender(width, height, samples, depth);
    console.log(`Progressive render initialized: ${totalPixels} pixels`);

    // Chunk size: render 1% of pixels per chunk for ~100 updates
    const chunkSize = Math.max(1, Math.ceil(totalPixels / 100));
    let renderedPixels = 0;

    // Create ImageData once for efficiency
    let imageData = null;

    // Render in chunks with yields to browser
    while (renderedPixels < totalPixels) {
        const endIdx = Math.min(renderedPixels + chunkSize, totalPixels);

        // Render this chunk (Go call)
        const pixels = goRenderChunk(renderedPixels, endIdx);

        // Create/update ImageData and paint to canvas
        if (!imageData) {
            imageData = new ImageData(pixels, width, height);
        } else {
            imageData.data.set(pixels);
        }
        ctx.putImageData(imageData, 0, 0);

        // Update progress
        renderedPixels = endIdx;
        const percent = Math.round((renderedPixels / totalPixels) * 100);
        const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
        showStatus(`Rendering... ${percent}% (${elapsed}s)`);

        // Yield to browser for repaint (crucial for live updates!)
        await new Promise(resolve => setTimeout(resolve, 0));
    }

    const elapsed = ((performance.now() - startTime) / 1000).toFixed(2);
    showStatus(`Rendered in ${elapsed}s`, 'success');
}

// Main render function
async function render() {
    if (!wasmReady || isRendering) {
        showStatus('WASM not ready or already rendering', 'error');
        return;
    }

    isRendering = true;

    const width = parseInt(widthInput.value) || 400;
    const height = Math.round(width / (16 / 9));
    const samples = parseInt(samplesInput.value) || 10;
    const depth = parseInt(depthInput.value) || 10;
    const scene = sceneSelect.value;
    const progressive = progressiveCheckbox?.checked ?? true;

    // Update canvas size
    canvas.width = width;
    canvas.height = height;

    // Clear canvas with dark background
    ctx.fillStyle = '#0a0a0a';
    ctx.fillRect(0, 0, width, height);

    // Set scene
    goSetScene(scene);

    renderBtn.disabled = true;
    renderBtn.textContent = 'Rendering...';
    showStatus(`Rendering ${width}x${height} with ${samples} samples...`);

    // Small delay to allow UI to update
    await new Promise(resolve => setTimeout(resolve, 16));

    try {
        if (progressive) {
            await renderProgressiveChunked(width, height, samples, depth);
        } else {
            await renderStandard(width, height, samples, depth);
        }
    } catch (err) {
        console.error('Render error:', err);
        showStatus('Render failed. See console for details.', 'error');
    }

    renderBtn.disabled = false;
    renderBtn.textContent = 'Render';
    isRendering = false;
}

// Event listeners
renderBtn.addEventListener('click', render);

// Start loading WASM
initWasm();
