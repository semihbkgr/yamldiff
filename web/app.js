// yamldiff Web Application

import { basicSetup, EditorView } from 'codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';

// DOM elements
const compareBtn = document.getElementById('compare');
const outputEl = document.getElementById('output');
const statusEl = document.getElementById('status');
const versionEl = document.getElementById('version');
const ignoreOrderCheckbox = document.getElementById('ignoreOrder');
const pathOnlyCheckbox = document.getElementById('pathOnly');
const metadataCheckbox = document.getElementById('metadata');

// Shared CodeMirror extensions
const editorExtensions = [
    basicSetup,
    yaml(),
    oneDark,
    EditorView.theme({
        '&': { backgroundColor: 'var(--bg-secondary)' },
        '.cm-gutters': { backgroundColor: 'var(--bg-secondary)', border: 'none' },
    }),
];

// Create CodeMirror editors
const leftEditor = new EditorView({
    extensions: editorExtensions,
    parent: document.getElementById('left-editor'),
});

const rightEditor = new EditorView({
    extensions: editorExtensions,
    parent: document.getElementById('right-editor'),
});

const editors = { left: leftEditor, right: rightEditor };

// File drag-and-drop on editor panels
document.querySelectorAll('.editor-panel').forEach(panel => {
    let dragCounter = 0;
    const editorId = panel.querySelector('[id$="-editor"]').id.replace('-editor', '');

    panel.addEventListener('dragenter', (e) => {
        e.preventDefault();
        dragCounter++;
        panel.classList.add('drag-over');
    });

    panel.addEventListener('dragover', (e) => {
        e.preventDefault();
    });

    panel.addEventListener('dragleave', () => {
        dragCounter--;
        if (dragCounter === 0) {
            panel.classList.remove('drag-over');
        }
    });

    panel.addEventListener('drop', (e) => {
        e.preventDefault();
        dragCounter = 0;
        panel.classList.remove('drag-over');
        const file = e.dataTransfer.files[0];
        if (!file) return;
        file.text().then(text => {
            const editor = editors[editorId];
            editor.dispatch({
                changes: { from: 0, to: editor.state.doc.length, insert: text },
            });
        });
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

// Open file buttons
document.querySelectorAll('.open-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = '.yaml,.yml';
        input.addEventListener('change', () => {
            const file = input.files[0];
            if (!file) return;
            file.text().then(text => {
                const editor = editors[btn.dataset.target];
                editor.dispatch({
                    changes: { from: 0, to: editor.state.doc.length, insert: text },
                });
            });
        });
        input.click();
    });
});

// Clear buttons
document.querySelectorAll('.clear-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        const editor = editors[btn.dataset.target];
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: '' },
        });
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
    const left = leftEditor.state.doc.toString();
    const right = rightEditor.state.doc.toString();

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
