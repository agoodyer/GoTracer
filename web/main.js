// Go Raytracer WASM Interface

const go = new Go();
let wasmReady = false;
let isRendering = false;
let cancelRequested = false;

// DOM elements
const canvas = document.getElementById('canvas');
const ctx = canvas.getContext('2d');
const renderBtn = document.getElementById('renderBtn');
const cancelBtn = document.getElementById('cancelBtn');
const sceneSelect = document.getElementById('scene');
const widthInput = document.getElementById('width');
const samplesInput = document.getElementById('samples');
const depthInput = document.getElementById('depth');
const progressSection = document.getElementById('progressSection');
const progressBar = document.getElementById('progressBar');
const progressStatus = document.getElementById('progressStatus');
const progressTime = document.getElementById('progressTime');

// Stepper button handlers
document.querySelectorAll('.stepper button').forEach(btn => {
    btn.addEventListener('click', () => {
        const target = document.getElementById(btn.dataset.target);
        const step = parseInt(target.step) || 1;
        const min = parseInt(target.min) || 0;
        const max = parseInt(target.max) || Infinity;
        let value = parseInt(target.value) || 0;

        if (btn.dataset.action === 'increment') {
            value = Math.min(max, value + step);
        } else {
            value = Math.max(min, value - step);
        }

        target.value = value;
    });
});

// Update progress bar and text
function updateProgress(current, total, elapsed, statusText) {
    const percent = Math.min(100, Math.round((current / total) * 100));
    progressBar.style.width = `${percent}%`;
    progressStatus.textContent = statusText;
    progressTime.textContent = `${elapsed}s`;
}

// Show progress section
function showProgress() {
    progressSection.classList.remove('hidden');
    progressBar.style.width = '0%';
    progressBar.classList.remove('cancelled');
}

// Hide progress section  
function hideProgress() {
    progressSection.classList.add('hidden');
}

// Draw a checkerboard grid pattern as placeholder
function drawGridPattern(width, height) {
    const gridSize = 20;
    for (let y = 0; y < height; y += gridSize) {
        for (let x = 0; x < width; x += gridSize) {
            const isEven = ((x / gridSize) + (y / gridSize)) % 2 === 0;
            ctx.fillStyle = isEven ? '#2a2a2a' : '#1f1f1f';
            ctx.fillRect(x, y, gridSize, gridSize);
        }
    }
}

// Draw initial grid on page load
function initCanvas() {
    const width = parseInt(widthInput.value) || 400;
    const height = Math.round(width / (16 / 9));
    canvas.width = width;
    canvas.height = height;
    drawGridPattern(width, height);
}

// Initialize WASM
async function initWasm() {
    // Draw initial grid pattern immediately
    initCanvas();

    progressSection.classList.remove('hidden');
    progressStatus.textContent = 'Loading WebAssembly...';
    progressBar.style.width = '50%';

    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
        );

        go.run(result.instance);
        wasmReady = true;

        renderBtn.textContent = 'Render';
        renderBtn.disabled = false;

        progressBar.style.width = '100%';
        progressStatus.textContent = 'Ready';

        setTimeout(hideProgress, 1500);
    } catch (err) {
        console.error('Failed to load WASM:', err);
        progressStatus.textContent = 'Failed to load';
        progressBar.style.background = '#ef4444';
    }
}

// Progressive render with progress bar and cancel support
async function renderProgressive(width, height, samples, depth) {
    const startTime = performance.now();

    // Initialize the render
    const info = goInitProgressiveRender(width, height, samples, depth);
    const totalPixels = info.totalPixels;
    console.log(`Render initialized: ${totalPixels} pixels, ${samples} samples`);

    // Chunk size: ~50 updates per sample
    const chunksPerSample = 50;
    const chunkSize = Math.max(1, Math.ceil(totalPixels / chunksPerSample));

    let imageData = null;
    const totalChunks = samples * Math.ceil(totalPixels / chunkSize);
    let currentChunk = 0;

    // Render loop
    for (let sample = 1; sample <= samples; sample++) {
        for (let startIdx = 0; startIdx < totalPixels; startIdx += chunkSize) {
            // Check for cancellation
            if (cancelRequested) {
                const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
                progressBar.classList.add('cancelled');
                updateProgress(currentChunk, totalChunks, elapsed, `Cancelled at sample ${sample}/${samples}`);
                return;
            }

            const endIdx = Math.min(startIdx + chunkSize, totalPixels);

            // Render chunk
            const pixels = goRenderSampleChunk(startIdx, endIdx, sample);

            // Update canvas
            if (!imageData) {
                imageData = new ImageData(pixels, width, height);
            } else {
                imageData.data.set(pixels);
            }
            ctx.putImageData(imageData, 0, 0);

            currentChunk++;

            // Update progress bar
            if (currentChunk % 5 === 0 || startIdx === 0) {
                const elapsed = ((performance.now() - startTime) / 1000).toFixed(1);
                updateProgress(
                    currentChunk,
                    totalChunks,
                    elapsed,
                    `Sample ${sample}/${samples}`
                );
            }

            // Yield to browser
            await new Promise(resolve => setTimeout(resolve, 0));
        }
    }

    const elapsed = ((performance.now() - startTime) / 1000).toFixed(2);
    updateProgress(totalChunks, totalChunks, elapsed, `Done (${samples} samples)`);
}

// Main render function
async function render() {
    if (!wasmReady || isRendering) return;

    isRendering = true;
    cancelRequested = false;

    const width = parseInt(widthInput.value) || 400;
    const height = Math.round(width / (16 / 9));
    const samples = parseInt(samplesInput.value) || 20;
    const depth = parseInt(depthInput.value) || 10;
    const scene = sceneSelect.value;

    // Update canvas size
    canvas.width = width;
    canvas.height = height;

    // Draw grid placeholder pattern
    drawGridPattern(width, height);

    // Set scene
    goSetScene(scene);

    renderBtn.classList.add('hidden');
    cancelBtn.classList.remove('hidden');
    showProgress();
    updateProgress(0, 1, '0.0', 'Starting...');

    // Small delay for UI update
    await new Promise(resolve => setTimeout(resolve, 16));

    try {
        await renderProgressive(width, height, samples, depth);
    } catch (err) {
        console.error('Render error:', err);
        progressStatus.textContent = 'Error';
        progressBar.style.background = '#ef4444';
    }

    cancelBtn.classList.add('hidden');
    renderBtn.classList.remove('hidden');
    renderBtn.disabled = false;
    isRendering = false;
}

// Cancel render
function cancelRender() {
    if (isRendering) {
        cancelRequested = true;
        cancelBtn.disabled = true;
        cancelBtn.textContent = 'Cancelling...';
    }
}

// Reset cancel button state
function resetCancelBtn() {
    cancelBtn.disabled = false;
    cancelBtn.textContent = 'Cancel';
}

// Event listeners
renderBtn.addEventListener('click', render);
cancelBtn.addEventListener('click', cancelRender);

// Reset cancel button when showing render button
const observer = new MutationObserver(() => {
    if (!cancelBtn.classList.contains('hidden')) {
        resetCancelBtn();
    }
});
observer.observe(cancelBtn, { attributes: true, attributeFilter: ['class'] });

// Start loading WASM
initWasm();
