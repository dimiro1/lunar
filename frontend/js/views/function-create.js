import { Icons } from '../icons.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';
import { FunctionDocs } from '../components/function-docs.js';
import { CodeEditor } from '../components/code-editor.js';

export const FunctionCreate = {
  formData: {
    name: "",
    description: "",
    code: `function handler(ctx, event)
  log.info("Function started")

  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = '{"message": "Hello"}'
  }
end`,
  },

  oninit: () => {
    FunctionCreate.formData = {
      name: "",
      description: "",
      code: `function handler(ctx, event)
  log.info("Function started")

  return {
    statusCode = 200,
    headers = { ["Content-Type"] = "application/json" },
    body = '{"message": "Hello"}'
  }
end`,
    };
  },

  createFunction: async () => {
    try {
      const payload = {
        name: FunctionCreate.formData.name,
        description: FunctionCreate.formData.description,
        code: FunctionCreate.formData.code,
      };

      await API.functions.create(payload);
      m.route.set("/functions");
    } catch (e) {
      alert("Failed to create function: " + e.message);
    }
  },

  view: () => {
    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", "Create New Function"),
            m(".page-subtitle", "Define your serverless function"),
          ]),
          m("a.btn", { href: "#!/functions" }, [Icons.arrowLeft(), "  Back"]),
        ]),
      ]),

      m(".card.mb-24", [
        m(".card-header", m(".card-title", "Function Details")),
        m("div", { style: "padding: 24px;" }, [
          m(".form-group", [
            m("label.form-label", "Name"),
            m("input.form-input", {
              value: FunctionCreate.formData.name,
              oninput: (e) => (FunctionCreate.formData.name = e.target.value),
              placeholder: "my-function",
              required: true,
            }),
          ]),
          m(".form-group", [
            m("label.form-label", "Description"),
            m("textarea.form-textarea", {
              value: FunctionCreate.formData.description,
              oninput: (e) =>
                (FunctionCreate.formData.description = e.target.value),
              placeholder: "What does this function do?",
              rows: 2,
            }),
          ]),
        ]),
      ]),

      m(FunctionDocs),

      m(".card.mb-24", [
        m(".card-header", m(".card-title", "Function Code (Lua)")),
        m("div", { style: "padding: 24px;" }, [
          m(CodeEditor, {
            id: "code-editor",
            value: FunctionCreate.formData.code,
            onChange: (value) => {
              FunctionCreate.formData.code = value;
            },
          }),
        ]),
      ]),

      m(
        "div",
        { style: "display: flex; justify-content: flex-end; gap: 12px;" },
        [
          m("a.btn", { href: "#!/functions" }, "Cancel"),
          m(
            "button.btn.btn-primary",
            {
              onclick: FunctionCreate.createFunction,
              disabled: !FunctionCreate.formData.name,
            },
            "Create Function",
          ),
        ],
      ),
    ]);
  },
};
