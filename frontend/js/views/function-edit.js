import { Icons } from '../icons.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';
import { FunctionDocs } from '../components/function-docs.js';
import { CodeEditor } from '../components/code-editor.js';

export const FunctionEdit = {
  func: null,
  loading: true,
  formData: {
    name: "",
    description: "",
    code: "",
  },

  oninit: (vnode) => {
    FunctionEdit.loadFunction(vnode.attrs.id);
  },

  loadFunction: async (id) => {
    FunctionEdit.loading = true;
    try {
      const func = await API.functions.get(id);
      FunctionEdit.func = func;
      FunctionEdit.formData = {
        name: func.name,
        description: func.description || "",
        code: func.active_version.code,
      };
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionEdit.loading = false;
      m.redraw();
    }
  },

  saveFunction: async () => {
    try {
      await API.functions.update(FunctionEdit.func.id, FunctionEdit.formData);
      m.route.set(`/functions/${FunctionEdit.func.id}`);
    } catch (e) {
      alert("Failed to save function");
    }
  },

  view: (vnode) => {
    if (FunctionEdit.loading) {
      return m(".loading", "Loading...");
    }

    if (!FunctionEdit.func) {
      return m(".container", m(".card", "Function not found"));
    }

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", FunctionEdit.func.name),
            m(".page-subtitle", "Edit function code and details"),
          ]),
          m("a.btn", { href: `#!/functions/${FunctionEdit.func.id}` }, [
            Icons.arrowLeft(),
            "  Back",
          ]),
        ]),
      ]),

      m(FunctionDocs),

      m(".card.mb-24", [
        m(".card-header", m(".card-title", "Code (Lua)")),
        m("div", { style: "padding: 16px;" }, [
          m(CodeEditor, {
            id: "code-editor",
            value: FunctionEdit.formData.code,
            onChange: (value) => {
              FunctionEdit.formData.code = value;
            },
          }),
        ]),
      ]),

      m(".card.mb-24", [
        m(".card-header", m(".card-title", "Function Details")),
        m("div", { style: "padding: 24px;" }, [
          m(".form-group", [
            m("label.form-label", "Name"),
            m("input.form-input", {
              value: FunctionEdit.formData.name,
              oninput: (e) => (FunctionEdit.formData.name = e.target.value),
            }),
          ]),
          m(".form-group", [
            m("label.form-label", "Description"),
            m("textarea.form-textarea", {
              value: FunctionEdit.formData.description,
              oninput: (e) =>
                (FunctionEdit.formData.description = e.target.value),
              rows: 2,
            }),
          ]),
        ]),
      ]),

      m(".card", [
        m(
          "div",
          {
            style:
              "padding: 16px; display: flex; justify-content: space-between;",
          },
          [
            m(
              "a.btn",
              { href: `#!/functions/${FunctionEdit.func.id}` },
              "Cancel",
            ),
            m(
              "button.btn.btn-primary",
              {
                onclick: FunctionEdit.saveFunction,
              },
              "Save Changes",
            ),
          ],
        ),
      ]),
    ]);
  },
};
