/**
 * @fileoverview Monaco-based code editor component. Language-agnostic: hover and
 * completion providers (and the Lunar API docs that drive them) live in
 * editor-completions.js + editor-<lang>-api.js, so this file only owns the
 * editor lifecycle, theme, and Mithril wiring.
 */

import {
  monacoLanguageFor,
  registerEditorProviders,
} from "./editor-completions.js";

/**
 * Registers the GitHub Dark theme for Monaco editor.
 * Only registers once using a global flag.
 */
const registerGitHubDarkTheme = () => {
  if (!window.monaco || window.__githubDarkThemeRegistered) return;
  window.__githubDarkThemeRegistered = true;

  monaco.editor.defineTheme("github-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "comment", foreground: "8b949e" },
      { token: "keyword", foreground: "ff7b72" },
      { token: "string", foreground: "a5d6ff" },
      { token: "number", foreground: "79c0ff" },
      { token: "type", foreground: "ffa657" },
      { token: "function", foreground: "d2a8ff" },
      { token: "variable", foreground: "ffa657" },
      { token: "constant", foreground: "79c0ff" },
      { token: "operator", foreground: "ff7b72" },
    ],
    colors: {
      "editor.background": "#0d1117",
      "editor.foreground": "#c9d1d9",
      "editor.lineHighlightBackground": "#161b22",
      "editor.selectionBackground": "#264f78",
      "editorCursor.foreground": "#c9d1d9",
      "editorLineNumber.foreground": "#7d8590",
      "editorLineNumber.activeForeground": "#c9d1d9",
      "editorGutter.background": "#0d1117",
    },
  });
};

/**
 * Monaco-based code editor component.
 * @type {Object}
 */
export const CodeEditor = {
  /**
   * Renders the code editor.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.id='code-editor'] - DOM element ID
   * @param {string} [vnode.attrs.value=''] - Initial code value
   * @param {(value: string) => void} [vnode.attrs.onChange] - Change callback
   * @param {boolean} [vnode.attrs.readOnly=false] - Read-only mode
   * @param {string} [vnode.attrs.language='lua'] - Lunar function language ("lua" or "starlark")
   * @param {string} [vnode.attrs.theme='github-dark'] - Editor theme
   * @param {boolean} [vnode.attrs.lineNumbers=true] - Show line numbers
   * @param {boolean} [vnode.attrs.minimap=false] - Show minimap
   * @param {string} [vnode.attrs.height='500px'] - Editor height
   * @returns {Object} Mithril vnode
   */
  view: (vnode) => {
    const {
      id = "code-editor",
      value = "",
      onChange = null,
      readOnly = false,
      language = "lua",
      theme = "github-dark",
      lineNumbers = true,
      minimap = false,
      height = "500px",
    } = vnode.attrs;

    // Map the Lunar function language to the Monaco grammar (Starlark → python).
    const monacoLanguage = monacoLanguageFor(language);

    /**
     * Creates the Monaco editor in the given container.
     * @param {HTMLElement} container - Container element
     */
    const createEditor = (container) => {
      require(["vs/editor/editor.main"], function () {
        if (!window.monaco) return;

        registerGitHubDarkTheme();
        registerEditorProviders(language);

        const editor = monaco.editor.create(container, {
          value: value || "",
          language: monacoLanguage,
          theme: theme,
          readOnly: readOnly,
          automaticLayout: true,
          minimap: {
            enabled: minimap,
          },
          lineNumbers: lineNumbers ? "on" : "off",
          scrollBeyondLastLine: false,
          fontFamily:
            '"Berkeley Mono", "JetBrains Mono", "SF Mono", Menlo, Monaco, "Cascadia Mono", monospace',
          fontSize: 14,
          tabSize: 2,
          suggestOnTriggerCharacters: true,
          quickSuggestions: true,
          padding: {
            top: 8,
            bottom: 8,
          },
        });

        if (onChange) {
          editor.onDidChangeModelContent(() => {
            onChange(editor.getValue());
          });
        }

        vnode.state.editor = editor;
      });
    };

    // Render the editor container
    return m(".code-editor-container", { style: `height: ${height};` }, [
      m("div", {
        id: id,
        style: `height: ${height};`,
        oncreate: (divVnode) => {
          const container = divVnode.dom;
          if (container) {
            createEditor(container);
          }
        },
        onupdate: () => {
          if (vnode.state.editor && value !== vnode.state.editor.getValue()) {
            const position = vnode.state.editor.getPosition();
            vnode.state.editor.setValue(value || "");
            if (position) {
              vnode.state.editor.setPosition(position);
            }
          }
        },
        onremove: () => {
          if (vnode.state.editor) {
            vnode.state.editor.dispose();
            vnode.state.editor = null;
          }
        },
      }),
    ]);
  },
};
