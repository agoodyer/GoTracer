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

// Hybrid progressive: chunked updates within each sample pass
// This provides continuous visual feedback - no stasis periods
async function renderHybridProgressive(width, height, samples, depth) {
    const startTime = performance.now();

    // Initialize the render (sets up scene, BVH, accumulator)
    const info = goInitProgressiveRender(width, height, samples, depth);
    const totalPixels = info.totalPixels;
    console.log(`Hybrid render initialized: ${totalPixels} pixels, ${samples} samples`);

    // Chunk size: update ~50 times per sample (2% of pixels per chunk)
    const chunksPerSample = 50;
    const chunkSize = Math.max(1, Math.ceil(totalPixels / chunksPerSample));

    let imageData = null;
    let totalChunks = samples * Math.ceil(totalPixels / chunkSize);
    let currentChunk = 0;

    // Outer loop: for each sample
    for (let sample = 1; sample <= samples; sample++) {
        // Inner loop: chunked rendering within this sample
        for (let startIdx = 0; startIdx < totalPixels; startIdx += chunkSize) {
            const endIdx = Math.min(startIdx + chunkSize, totalPixels);

            // Render 1 sample for this chunk of pixels
            const pixels = goRenderSampleChunk(startIdx, endIdx, sample);

            // Update canvas
            if (!imageData) {
                imageData = new ImageData(pixels, width, height);
            } else {
                imageData.data.set(pixels);
            }
            ctx.putImageData(imageData, 0, 0);

            currentChunk++;

            // Update status periodically (not every chunk to avoid spam)
            if (currentChunk % 10 === 0 || startIdx === 0) {
                const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
                const percent = Math.round((currentChunk / totalChunks) * 100);
                showStatus(`Sample ${sample}/${samples} - ${percent}% (${elapsed}s)`);
            }

            // Yield to browser for repaint
            await new Promise(resolve => setTimeout(resolve, 0));
        }
    }

    const elapsed = ((performance.now() - startTime) / 1000).toFixed(2);
    showStatus(`Rendered in ${elapsed}s (${samples} samples)`, 'success');
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
            // Use hybrid progressive for continuous updates
            await renderHybridProgressive(width, height, samples, depth);
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
