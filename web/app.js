// yamldiff Web Application

(function() {
    'use strict';

    // DOM elements
    const leftTextarea = document.getElementById('left');
    const rightTextarea = document.getElementById('right');
    const leftHighlight = document.getElementById('leftHighlight');
    const rightHighlight = document.getElementById('rightHighlight');
    const compareBtn = document.getElementById('compare');
    const outputEl = document.getElementById('output');
    const statusEl = document.getElementById('status');
    const versionEl = document.getElementById('version');
    const ignoreOrderCheckbox = document.getElementById('ignoreOrder');
    const pathOnlyCheckbox = document.getElementById('pathOnly');

    // Debounce helper
    function debounce(fn, delay) {
        let timeoutId;
        return function(...args) {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => fn.apply(this, args), delay);
        };
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
    function syncScroll(textarea, highlightEl) {
        highlightEl.parentElement.scrollTop = textarea.scrollTop;
        highlightEl.parentElement.scrollLeft = textarea.scrollLeft;
    }

    // Setup highlighting for a textarea
    function setupHighlighting(textarea, highlightEl) {
        const debouncedUpdate = debounce(() => updateHighlight(textarea, highlightEl), 30);

        textarea.addEventListener('input', debouncedUpdate);
        textarea.addEventListener('scroll', () => syncScroll(textarea, highlightEl));
    }

    // Initialize highlighting for both textareas
    setupHighlighting(leftTextarea, leftHighlight);
    setupHighlighting(rightTextarea, rightHighlight);

    // DEV: Default values for testing (remove before release)
    leftTextarea.value = `apiVersion: v1
kind: Pod
metadata:
  name: app
  labels:
    app: app
    instance: "app-v1"
    version: "v1"
spec:
  containers:
    - name: app
      image: app:1.0
      ports:
        - name: http-port
          containerPort: 80
      resources:
        requests:
          memory: 256Mi
          cpu: "100m"
        limits:
          memory: 512Mi
          cpu: "300m"
      volumeMounts:
        - name: config-volume
          subPath: app-config.yaml
          mountPath: /etc/config
  volumes:
    - name: config-volume
      configMap:
        name: app-config`;

    rightTextarea.value = `apiVersion: v1
kind: Pod
metadata:
  name: app
  labels:
    app: app
    instance: "app-v2"
    version: "v2"
spec:
  containers:
    - name: app
      image: app:2.0
      ports:
        - name: http-port
          containerPort: 80
          protocol: TCP
      resources:
        requests:
          memory: 256Mi
          cpu: "100m"
        limits:
          memory: 1Gi
          cpu: "300m"
      livenessProbe:
        httpGet:
          path: /healthz
          port: http-port
  initContainers:
    - name: init-app
      image: init-app:1.0`;

    // Clear buttons
    document.querySelectorAll('.clear-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.target;
            const textarea = document.getElementById(target);
            const highlightEl = document.getElementById(target + 'Highlight');
            textarea.value = '';
            if (highlightEl) {
                highlightEl.innerHTML = '';
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
            pathOnly: pathOnlyCheckbox.checked
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
