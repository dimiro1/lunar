import { Icons } from '../icons.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';

export const FunctionEnv = {
  func: null,
  envVars: [],
  loading: true,
  saving: false,

  oninit: (vnode) => {
    FunctionEnv.loadData(vnode.attrs.id);
  },

  loadData: async (id) => {
    FunctionEnv.loading = true;
    try {
      const func = await API.functions.get(id);
      FunctionEnv.func = func;
      FunctionEnv.envVars = Object.entries(func.env_vars || {}).map(
        ([key, value]) => ({ key, value })
      );
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionEnv.loading = false;
      m.redraw();
    }
  },

  saveEnvVars: async () => {
    FunctionEnv.saving = true;
    try {
      const env_vars = {};
      FunctionEnv.envVars.forEach((envVar) => {
        if (envVar.key && envVar.value) {
          env_vars[envVar.key] = envVar.value;
        }
      });

      await API.functions.updateEnv(FunctionEnv.func.id, env_vars);
      Toast.show("Environment variables updated", "success");
      m.route.set(`/functions/${FunctionEnv.func.id}`);
    } catch (e) {
      Toast.show("Failed to update environment variables", "error");
    } finally {
      FunctionEnv.saving = false;
    }
  },

  view: () => {
    if (FunctionEnv.loading) {
      return m(".loading", "Loading...");
    }

    if (!FunctionEnv.func) {
      return m(".container", m(".card", "Function not found"));
    }

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", "Environment Variables"),
            m(".page-subtitle", FunctionEnv.func.name),
          ]),
          m("a.btn", { href: `#!/functions/${FunctionEnv.func.id}` }, [
            Icons.arrowLeft(),
            "  Back",
          ]),
        ]),
      ]),

      m(".card.mb-24", [
        m(".card-header", [
          m("div", [
            m(".card-title", "Environment Variables"),
            m(".card-subtitle", `${FunctionEnv.envVars.length} variables`),
          ]),
          m(
            "button.btn.btn-icon",
            {
              onclick: () => FunctionEnv.envVars.push({ key: "", value: "" }),
            },
            Icons.plus()
          ),
        ]),
        m("div", { style: "padding: 24px;" }, [
          FunctionEnv.envVars.length === 0
            ? m(
                ".text-center",
                { style: "color: #a3a3a3; padding: 48px 0;" },
                "No environment variables. Click + to add one."
              )
            : FunctionEnv.envVars.map((envVar, i) =>
                m(
                  ".form-group",
                  {
                    key: i,
                    style: "display: flex; gap: 8px; align-items: center;",
                  },
                  [
                    m("input.form-input", {
                      value: envVar.key,
                      oninput: (e) => (envVar.key = e.target.value),
                      placeholder: "KEY",
                      style: "flex: 1;",
                    }),
                    m("input.form-input", {
                      value: envVar.value,
                      oninput: (e) => (envVar.value = e.target.value),
                      placeholder: "value",
                      style: "flex: 1;",
                    }),
                    m(
                      "button.btn.btn-icon",
                      {
                        onclick: () => FunctionEnv.envVars.splice(i, 1),
                      },
                      Icons.xMark()
                    ),
                  ]
                )
              ),
        ]),
      ]),

      m(
        "div",
        { style: "display: flex; justify-content: flex-end; gap: 12px;" },
        [
          m(
            "a.btn",
            { href: `#!/functions/${FunctionEnv.func.id}` },
            "Cancel"
          ),
          m(
            "button.btn.btn-primary",
            {
              onclick: FunctionEnv.saveEnvVars,
              disabled: FunctionEnv.saving,
            },
            FunctionEnv.saving ? "Saving..." : "Save Changes"
          ),
        ]
      ),
    ]);
  },
};
