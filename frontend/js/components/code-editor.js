export const CodeEditor = {
  view: (vnode) => {
    const {
      id = "code-editor",
      value = "",
      onChange = null,
      readOnly = false,
      language = "lua",
      theme = "vs-dark",
      lineNumbers = true,
      minimap = false,
      height = "400px",
    } = vnode.attrs;

    return m("div", {
      id: id,
      style: `height: ${height}; border: 1px solid #444;`,
      oncreate: (divVnode) => {
        const container = divVnode.dom;
        if (container && window.monaco) {
          const editor = monaco.editor.create(container, {
            value: value || "",
            language: language,
            theme: theme,
            readOnly: readOnly,
            automaticLayout: true,
            minimap: {
              enabled: minimap,
            },
            lineNumbers: lineNumbers ? "on" : "off",
            scrollBeyondLastLine: false,
            fontSize: 14,
            tabSize: 2,
          });

          if (onChange) {
            editor.onDidChangeModelContent(() => {
              onChange(editor.getValue());
            });
          }

          vnode.state.editor = editor;
        }
      },
      onupdate: (divVnode) => {
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
    });
  },
};
