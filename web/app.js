// yamldiff Web Application

import { basicSetup, EditorView } from 'codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { oneDark } from '@codemirror/theme-one-dark';
import { Decoration } from '@codemirror/view';
import { StateEffect, StateField, RangeSetBuilder } from '@codemirror/state';

// DOM elements
const outputEl = document.getElementById('output');
const versionEl = document.getElementById('version');
const ignoreOrderCheckbox = document.getElementById('ignoreOrder');

// Debounce helper
let debounceTimer;
function debounceCompare() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(handleCompare, 300);
}

// Diff highlight decorations
const setHighlights = StateEffect.define();

const highlightField = StateField.define({
    create() { return Decoration.none; },
    update(decorations, tr) {
        decorations = decorations.map(tr.changes);
        for (const e of tr.effects) {
            if (e.is(setHighlights)) decorations = e.value;
        }
        return decorations;
    },
    provide: f => EditorView.decorations.from(f),
});

const diffMarkClasses = {
    added: 'cm-diff-added',
    deleted: 'cm-diff-deleted',
    modified: 'cm-diff-modified',
};

// Convert UTF-8 byte offset to UTF-16 code-unit position
function byteOffsetToCharPosition(text, byteOffset) {
    let charPos = 0;
    let bytePos = 0;
    const encoder = new TextEncoder();
    
    for (const char of text) {
        if (bytePos >= byteOffset) break;
        bytePos += encoder.encode(char).length;
        charPos += char.length; // Account for surrogate pairs in UTF-16
    }
    
    return charPos;
}

function buildDecorations(diffs, side, docLength) {
    const ranges = [];
    const docText = side === 'leftSource' ? leftEditor.state.doc.toString() : rightEditor.state.doc.toString();
    
    for (const docDiffs of diffs) {
        for (const d of docDiffs) {
            const source = d[side];
            if (source) {
                const start = byteOffsetToCharPosition(docText, source.start);
                const end = byteOffsetToCharPosition(docText, source.end);
                if (end <= docLength) {
                    ranges.push({ start, end, type: d.type });
                }
            }
        }
    }
    ranges.sort((a, b) => a.start - b.start);

    const builder = new RangeSetBuilder();
    for (const r of ranges) {
        const cls = diffMarkClasses[r.type];
        if (cls) builder.add(r.start, r.end, Decoration.mark({ class: cls }));
    }
    return builder.finish();
}

// Shared CodeMirror extensions
const editorExtensions = [
    basicSetup,
    yaml(),
    oneDark,
    highlightField,
    EditorView.theme({
        '&': { backgroundColor: 'var(--bg-secondary)' },
        '.cm-gutters': { backgroundColor: 'var(--bg-secondary)', border: 'none' },
    }),
    EditorView.updateListener.of(update => {
        if (update.docChanged && wasmReady) {
            debounceCompare();
        }
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
let wasmReady = false;

async function initWasm() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch('yamldiff.wasm'),
            go.importObject
        );
        go.run(result.instance);
        wasmReady = true;
        handleCompare();

        // Show version
        if (typeof yamldiffVersion === 'function') {
            const ver = yamldiffVersion();
            versionEl.textContent = 'Version: ' + (ver || 'dev');
        }

        console.log('WASM loaded successfully');
    } catch (err) {
        console.error('Failed to load WASM:', err);
        outputEl.innerHTML = '<span class="error">Failed to load WASM module. Please refresh the page.</span>';
    }
}

// Scroll editor to a source position, centering it in the viewport
function scrollToSource(editor, source) {
    if (!source) return;
    const docText = editor.state.doc.toString();
    const charPos = byteOffsetToCharPosition(docText, source.start);
    const pos = Math.min(charPos, editor.state.doc.length);
    editor.dispatch({
        selection: { anchor: pos },
        effects: EditorView.scrollIntoView(pos, { y: 'center' }),
    });
    editor.focus();
}

// Format diff output from structured diffs array
function formatDiffOutput(diffs) {
    if (!diffs || diffs.length === 0) {
        return '<span class="no-diff">No differences found</span>';
    }

    const docParts = diffs.map(docDiffs => {
        if (!docDiffs || docDiffs.length === 0) return '';
        return docDiffs.map((d, i) => {
            const dataAttrs = `data-left-start="${d.leftSource?.start ?? ''}" data-left-end="${d.leftSource?.end ?? ''}" data-right-start="${d.rightSource?.start ?? ''}" data-right-end="${d.rightSource?.end ?? ''}"`;
            return `<span class="diff-entry diff-${d.type}" ${dataAttrs}>${d.format}</span>`;
        }).join('\n');
    });

    const output = docParts.join('\n---\n');
    if (!output.trim()) {
        return '<span class="no-diff">No differences found</span>';
    }
    return output;
}

// Handle clicks on diff entries
outputEl.addEventListener('click', (e) => {
    const entry = e.target.closest('.diff-entry');
    if (!entry) return;

    const leftStart = entry.dataset.leftStart;
    const rightStart = entry.dataset.rightStart;

    if (leftStart !== '') {
        scrollToSource(leftEditor, { start: parseInt(leftStart) });
    }
    if (rightStart !== '') {
        scrollToSource(rightEditor, { start: parseInt(rightStart) });
    }
});

// Apply diff highlights to editors
function applyHighlights(diffs) {
    leftEditor.dispatch({
        effects: setHighlights.of(buildDecorations(diffs, 'leftSource', leftEditor.state.doc.length)),
    });
    rightEditor.dispatch({
        effects: setHighlights.of(buildDecorations(diffs, 'rightSource', rightEditor.state.doc.length)),
    });
}

function clearHighlights() {
    leftEditor.dispatch({ effects: setHighlights.of(Decoration.none) });
    rightEditor.dispatch({ effects: setHighlights.of(Decoration.none) });
}

// Compare handler
function handleCompare() {
    const left = leftEditor.state.doc.toString();
    const right = rightEditor.state.doc.toString();

    if (!left || !right) {
        outputEl.innerHTML = '';
        clearHighlights();
        return;
    }

    const options = {
        ignoreOrder: ignoreOrderCheckbox.checked,
    };

    try {
        const result = yamldiffCompare(left, right, options);

        if (result.error) {
            outputEl.innerHTML = `<span class="error">Error: ${escapeHtml(result.error)}</span>`;
            clearHighlights();
        } else {
            outputEl.innerHTML = formatDiffOutput(result.diffs);
            applyHighlights(result.diffs);
        }
    } catch (err) {
        outputEl.innerHTML = `<span class="error">Error: ${escapeHtml(err.message)}</span>`;
        clearHighlights();
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

// Re-compare when options change
ignoreOrderCheckbox.addEventListener('change', () => {
    if (wasmReady) handleCompare();
});

// Initialize
initWasm();
