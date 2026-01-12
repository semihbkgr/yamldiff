// yamldiff Web Application

(function() {
    'use strict';

    // DOM elements
    const leftTextarea = document.getElementById('left');
    const rightTextarea = document.getElementById('right');
    const compareBtn = document.getElementById('compare');
    const statBtn = document.getElementById('stat');
    const outputEl = document.getElementById('output');
    const statusEl = document.getElementById('status');
    const versionEl = document.getElementById('version');
    const ignoreOrderCheckbox = document.getElementById('ignoreOrder');
    const pathOnlyCheckbox = document.getElementById('pathOnly');
    const metadataCheckbox = document.getElementById('metadata');

    // Clear buttons
    document.querySelectorAll('.clear-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.target;
            document.getElementById(target).value = '';
        });
    });

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
            statBtn.disabled = false;

            // Show version
            if (typeof yamldiffVersion === 'function') {
                const ver = yamldiffVersion();
                versionEl.textContent = 'Version: ' + (ver || 'dev');
            }

            console.log('WASM loaded successfully');
        } catch (err) {
            console.error('Failed to load WASM:', err);
            compareBtn.textContent = 'WASM Error';
            outputEl.innerHTML = '<span class="error">Failed to load WASM module. Please refresh the page.</span>';
        }
    }

    // Format diff output with colors
    function formatDiffOutput(text) {
        if (!text || text.trim() === '') {
            return '<span class="no-diff">No differences found</span>';
        }

        const lines = text.split('\n');
        return lines.map(line => {
            const trimmed = line.trim();
            let className = '';

            if (trimmed.startsWith('+')) {
                className = 'diff-added';
            } else if (trimmed.startsWith('-')) {
                className = 'diff-deleted';
            } else if (trimmed.startsWith('~')) {
                className = 'diff-modified';
            }

            // Escape HTML
            const escaped = line
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');

            return className
                ? `<span class="diff-line ${className}">${escaped}</span>`
                : `<span class="diff-line">${escaped}</span>`;
        }).join('\n');
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

    // Stat handler
    function handleStat() {
        const left = leftTextarea.value;
        const right = rightTextarea.value;

        if (!left && !right) {
            outputEl.innerHTML = '<span class="no-diff">Enter YAML content in both panels to compare</span>';
            statusEl.textContent = '';
            return;
        }

        const options = {
            ignoreOrder: ignoreOrderCheckbox.checked
        };

        try {
            const result = yamldiffStat(left, right, options);

            if (result.error) {
                outputEl.innerHTML = `<span class="error">Error: ${escapeHtml(result.error)}</span>`;
                statusEl.textContent = 'Parse error';
            } else {
                const stats = result.result;
                const parts = [];
                if (stats.added > 0) parts.push(`<span class="diff-added">${stats.added} added</span>`);
                if (stats.deleted > 0) parts.push(`<span class="diff-deleted">${stats.deleted} deleted</span>`);
                if (stats.modified > 0) parts.push(`<span class="diff-modified">${stats.modified} modified</span>`);

                if (parts.length === 0) {
                    outputEl.innerHTML = '<span class="no-diff">No differences found</span>';
                } else {
                    outputEl.innerHTML = parts.join(', ');
                }
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
    statBtn.addEventListener('click', handleStat);

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
