// yamldiff Web Application

(function() {
    'use strict';

    // DOM elements
    const leftTextarea = document.getElementById('left');
    const rightTextarea = document.getElementById('right');
    const leftHighlight = document.getElementById('leftHighlight');
    const rightHighlight = document.getElementById('rightHighlight');
    const leftLineNumbers = document.getElementById('leftLineNumbers');
    const rightLineNumbers = document.getElementById('rightLineNumbers');
    const compareBtn = document.getElementById('compare');
    const outputEl = document.getElementById('output');
    const statusEl = document.getElementById('status');
    const versionEl = document.getElementById('version');
    const ignoreOrderCheckbox = document.getElementById('ignoreOrder');
    const pathOnlyCheckbox = document.getElementById('pathOnly');
    const metadataCheckbox = document.getElementById('metadata');

    // Mutual exclusivity for pathOnly and metadata
    pathOnlyCheckbox.addEventListener('change', () => {
        if (pathOnlyCheckbox.checked) {
            metadataCheckbox.checked = false;
        }
    });

    metadataCheckbox.addEventListener('change', () => {
        if (metadataCheckbox.checked) {
            pathOnlyCheckbox.checked = false;
        }
    });

    // Debounce helper
    function debounce(fn, delay) {
        let timeoutId;
        return function(...args) {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => fn.apply(this, args), delay);
        };
    }

    // Update line numbers for a textarea
    function updateLineNumbers(textarea, lineNumbersEl) {
        const lines = textarea.value.split('\n');
        const lineCount = lines.length;
        let html = '';
        for (let i = 1; i <= lineCount; i++) {
            html += `<span>${i}</span>`;
        }
        lineNumbersEl.innerHTML = html;
    }

    // Update syntax highlighting for a textarea
    function updateHighlight(textarea, highlightEl) {
        if (typeof yamldiffColorize === 'function') {
            const highlighted = yamldiffColorize(textarea.value);
            highlightEl.innerHTML = highlighted || '';
        } else {
            // Fallback: show plain escaped text if WASM not loaded
            highlightEl.textContent = textarea.value;
        }
    }

    // Sync scroll positions
    function syncScroll(textarea, highlightEl, lineNumbersEl) {
        const scrollTop = textarea.scrollTop;
        const scrollLeft = textarea.scrollLeft;
        // highlightEl is the <code>, its parent <pre> is the scrollable element
        const preEl = highlightEl.parentElement;
        preEl.scrollTop = scrollTop;
        preEl.scrollLeft = scrollLeft;
        lineNumbersEl.style.transform = `translateY(-${scrollTop}px)`;
    }

    // Setup highlighting for a textarea
    function setupEditor(textarea, highlightEl, lineNumbersEl) {
        const debouncedUpdate = debounce(() => {
            updateHighlight(textarea, highlightEl);
            updateLineNumbers(textarea, lineNumbersEl);
        }, 30);

        textarea.addEventListener('input', debouncedUpdate);
        textarea.addEventListener('scroll', () => syncScroll(textarea, highlightEl, lineNumbersEl));

        // Initial line numbers
        updateLineNumbers(textarea, lineNumbersEl);
    }

    // Initialize editors for both textareas
    setupEditor(leftTextarea, leftHighlight, leftLineNumbers);
    setupEditor(rightTextarea, rightHighlight, rightLineNumbers);

    // Clear buttons
    document.querySelectorAll('.clear-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.target;
            const textarea = document.getElementById(target);
            const highlightEl = document.getElementById(target + 'Highlight');
            const lineNumbersEl = document.getElementById(target + 'LineNumbers');
            textarea.value = '';
            if (highlightEl) {
                highlightEl.innerHTML = '';
            }
            if (lineNumbersEl) {
                lineNumbersEl.innerHTML = '<span>1</span>';
            }
        });
    });

    // Initialize WASM
    async function initWasm() {
        try {
            const go = new Go();
            const result = await WebAssembly.instantiateStreaming(
                fetch('yamldiff.wasm'),
                go.importObject
            );
            go.run(result.instance);

            // Enable buttons
            compareBtn.disabled = false;
            compareBtn.textContent = 'Compare';

            // Show version
            if (typeof yamldiffVersion === 'function') {
                const ver = yamldiffVersion();
                versionEl.textContent = 'Version: ' + (ver || 'dev');
            }

            // Trigger initial highlighting for default values
            updateHighlight(leftTextarea, leftHighlight);
            updateHighlight(rightTextarea, rightHighlight);

            console.log('WASM loaded successfully');
        } catch (err) {
            console.error('Failed to load WASM:', err);
            compareBtn.textContent = 'WASM Error';
            outputEl.innerHTML = '<span class="error">Failed to load WASM module. Please refresh the page.</span>';
        }
    }

    // Format diff output - WASM now returns HTML-formatted output
    function formatDiffOutput(html) {
        if (!html || html.trim() === '') {
            return '<span class="no-diff">No differences found</span>';
        }
        // WASM output is already HTML-formatted and escaped
        return html;
    }

    // Compare handler
    function handleCompare() {
        const left = leftTextarea.value;
        const right = rightTextarea.value;

        if (!left && !right) {
            outputEl.innerHTML = '<span class="no-diff">Enter YAML content in both panels to compare</span>';
            statusEl.textContent = '';
            return;
        }

        const options = {
            ignoreOrder: ignoreOrderCheckbox.checked,
            pathOnly: pathOnlyCheckbox.checked,
            metadata: metadataCheckbox.checked
        };

        try {
            const result = yamldiffCompare(left, right, options);

            if (result.error) {
                outputEl.innerHTML = `<span class="error">Error: ${escapeHtml(result.error)}</span>`;
                statusEl.textContent = 'Parse error';
            } else {
                outputEl.innerHTML = formatDiffOutput(result.result);
                statusEl.textContent = result.hasDiff ? 'Differences found' : 'No differences';
            }
        } catch (err) {
            outputEl.innerHTML = `<span class="error">Error: ${escapeHtml(err.message)}</span>`;
            statusEl.textContent = 'Error';
        }
    }


    // Escape HTML helper
    function escapeHtml(text) {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }

    // Event listeners
    compareBtn.addEventListener('click', handleCompare);

    // Keyboard shortcut: Ctrl/Cmd + Enter to compare
    document.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            if (!compareBtn.disabled) {
                handleCompare();
            }
        }
    });

    // Initialize
    initWasm();
})();
