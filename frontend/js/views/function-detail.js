import { Icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { IdPill } from "../components/id-pill.js";
import { CodeEditor } from "../components/code-editor.js";
import { CodeExamples } from "../components/code-examples.js";
import { Pagination } from "../components/pagination.js";

export const FunctionDetail = {
  func: null,
  versions: [],
  executions: [],
  loading: true,
  activeTab: "overview",
  testRequest: {
    method: "GET",
    query: "",
    body: "",
  },
  testResponse: null,
  testLogs: [],
  selectedVersions: [],
  envVars: [],
  savingEnv: false,
  versionsLimit: 20,
  versionsOffset: 0,
  versionsTotal: 0,
  executionsLimit: 20,
  executionsOffset: 0,
  executionsTotal: 0,

  oninit: (vnode) => {
    const tab = m.route.param("tab");
    if (tab) FunctionDetail.activeTab = tab;

    // Clear test results and logs when visiting the page
    FunctionDetail.testResponse = null;
    FunctionDetail.testLogs = [];

    FunctionDetail.loadData(vnode.attrs.id);
  },

  loadData: async (id) => {
    FunctionDetail.loading = true;
    try {
      const [func, versions, executions] = await Promise.all([
        API.functions.get(id),
        API.versions.list(
          id,
          FunctionDetail.versionsLimit,
          FunctionDetail.versionsOffset,
        ),
        API.executions.list(
          id,
          FunctionDetail.executionsLimit,
          FunctionDetail.executionsOffset,
        ),
      ]);
      FunctionDetail.func = func;
      FunctionDetail.versions = versions.versions || [];
      FunctionDetail.versionsTotal = versions.pagination?.total || 0;
      FunctionDetail.executions = executions.executions || [];
      FunctionDetail.executionsTotal = executions.pagination?.total || 0;
      FunctionDetail.envVars = Object.entries(func.env_vars || {}).map(
        ([key, value]) => ({ key, value }),
      );
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionDetail.loading = false;
      m.redraw();
    }
  },

  loadVersions: async () => {
    try {
      const versions = await API.versions.list(
        FunctionDetail.func.id,
        FunctionDetail.versionsLimit,
        FunctionDetail.versionsOffset,
      );
      FunctionDetail.versions = versions.versions || [];
      FunctionDetail.versionsTotal = versions.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load versions:", e);
    }
  },

  loadExecutions: async () => {
    try {
      const executions = await API.executions.list(
        FunctionDetail.func.id,
        FunctionDetail.executionsLimit,
        FunctionDetail.executionsOffset,
      );
      FunctionDetail.executions = executions.executions || [];
      FunctionDetail.executionsTotal = executions.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load executions:", e);
    }
  },

  handleVersionsPageChange: (newOffset) => {
    FunctionDetail.versionsOffset = newOffset;
    FunctionDetail.loadVersions();
  },

  handleVersionsLimitChange: (newLimit) => {
    FunctionDetail.versionsLimit = newLimit;
    FunctionDetail.versionsOffset = 0;
    FunctionDetail.loadVersions();
  },

  handleExecutionsPageChange: (newOffset) => {
    FunctionDetail.executionsOffset = newOffset;
    FunctionDetail.loadExecutions();
  },

  handleExecutionsLimitChange: (newLimit) => {
    FunctionDetail.executionsLimit = newLimit;
    FunctionDetail.executionsOffset = 0;
    FunctionDetail.loadExecutions();
  },

  setTab: (tab) => {
    FunctionDetail.activeTab = tab;
  },

  executeTest: async () => {
    try {
      const response = await API.execute(
        FunctionDetail.func.id,
        FunctionDetail.testRequest,
      );
      FunctionDetail.testResponse = response;
      FunctionDetail.testLogs = [];
      m.redraw();

      // Fetch logs for this execution
      const executionId = response.headers["X-Execution-Id"];
      if (executionId) {
        try {
          const logsData = await API.executions.getLogs(executionId);
          FunctionDetail.testLogs = logsData.logs || [];
          m.redraw();
        } catch (e) {
          console.error("Failed to load logs:", e);
          FunctionDetail.testLogs = [];
        }
      }

      FunctionDetail.loadData(FunctionDetail.func.id);
    } catch (e) {
      alert("Execution failed");
    }
  },

  activateVersion: async (version) => {
    if (!confirm(`Activate version ${version}?`)) return;
    try {
      await API.versions.activate(FunctionDetail.func.id, version);
      FunctionDetail.loadData(FunctionDetail.func.id);
    } catch (e) {
      alert("Failed to activate version");
    }
  },

  saveEnvVars: async () => {
    FunctionDetail.savingEnv = true;
    try {
      const env_vars = {};
      FunctionDetail.envVars.forEach((envVar) => {
        if (envVar.key && envVar.value) {
          env_vars[envVar.key] = envVar.value;
        }
      });

      await API.functions.updateEnv(FunctionDetail.func.id, env_vars);
      Toast.show("Environment variables updated", "success");
      FunctionDetail.loadData(FunctionDetail.func.id);
    } catch (e) {
      Toast.show("Failed to update environment variables", "error");
    } finally {
      FunctionDetail.savingEnv = false;
    }
  },

  view: (vnode) => {
    if (FunctionDetail.loading) {
      return m(".loading", "Loading...");
    }

    if (!FunctionDetail.func) {
      return m(".container", m(".card", "Function not found"));
    }

    const func = FunctionDetail.func;

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", func.name),
            m(".page-subtitle", func.description || "No description"),
          ]),
          m(".actions", [
            m("a.btn", { href: "#!/functions" }, [Icons.arrowLeft(), "  Back"]),
            m("a.btn.btn-primary", { href: `#!/functions/${func.id}/edit` }, [
              Icons.pencil(),
              "  Edit",
            ]),
          ]),
        ]),
      ]),

      m(".tabs", [
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "overview" ? "active" : "",
            onclick: () => FunctionDetail.setTab("overview"),
          },
          "Overview",
        ),
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "code" ? "active" : "",
            onclick: () => FunctionDetail.setTab("code"),
          },
          "Code",
        ),
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "env" ? "active" : "",
            onclick: () => FunctionDetail.setTab("env"),
          },
          `Environment (${FunctionDetail.envVars.length})`,
        ),
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "versions" ? "active" : "",
            onclick: () => FunctionDetail.setTab("versions"),
          },
          `Versions (${FunctionDetail.versionsTotal})`,
        ),
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "executions" ? "active" : "",
            onclick: () => FunctionDetail.setTab("executions"),
          },
          `Executions (${FunctionDetail.executionsTotal})`,
        ),
        m(
          "a.tab",
          {
            class: FunctionDetail.activeTab === "test" ? "active" : "",
            onclick: () => FunctionDetail.setTab("test"),
          },
          "Test",
        ),
      ]),

      // Overview Tab
      FunctionDetail.activeTab === "overview" &&
        m(".card", [
          m(".card-header", m(".card-title", "Function Details")),
          m("table", [
            m("tbody", [
              m("tr", [
                m("td", m("strong", "ID")),
                m("td", m(IdPill, { id: func.id })),
              ]),
              m("tr", [m("td", m("strong", "Name")), m("td", func.name)]),
              m("tr", [
                m("td", m("strong", "Active Version")),
                m(
                  "td",
                  m(".badge.badge-success", `v${func.active_version.version}`),
                ),
              ]),
              m("tr", [
                m("td", m("strong", "Created")),
                m("td", new Date(func.created_at).toLocaleString()),
              ]),
              m("tr", [
                m("td", m("strong", "Updated")),
                m("td", new Date(func.updated_at).toLocaleString()),
              ]),
            ]),
          ]),
        ]),

      // Code Tab
      FunctionDetail.activeTab === "code" &&
        m(".card", [
          m(
            ".card-header",
            m(".card-title", `Code (Version ${func.active_version.version})`),
          ),
          m("div", { style: "padding: 16px;" }, [
            m(CodeEditor, {
              id: "code-viewer",
              value: func.active_version.code,
              readOnly: true,
            }),
          ]),
        ]),

      // Versions Tab
      FunctionDetail.activeTab === "versions" &&
        m(".card", [
          m(".card-header", [
            m(".card-title", "Version History"),
            m(
              "button.btn",
              {
                class:
                  FunctionDetail.selectedVersions.length === 2
                    ? "btn-primary"
                    : "",
                disabled: FunctionDetail.selectedVersions.length !== 2,
                style:
                  FunctionDetail.selectedVersions.length !== 2
                    ? "cursor: not-allowed; background: #404040; color: #666; border-color: #404040;"
                    : "",
                onclick:
                  FunctionDetail.selectedVersions.length === 2
                    ? () => {
                        const sorted = [
                          ...FunctionDetail.selectedVersions,
                        ].sort((a, b) => a - b);
                        m.route.set(
                          `/functions/${func.id}/diff/${sorted[0]}/${sorted[1]}`,
                        );
                      }
                    : undefined,
              },
              "Compare Selected",
            ),
          ]),
          FunctionDetail.versions.length === 0
            ? m(".text-center.mt-24.mb-24", "No versions yet")
            : [
                m("table", [
                  m(
                    "thead",
                    m("tr", [
                      m("th", { style: "width: 50px;" }, ""),
                      m("th", { style: "width: 100px;" }, "Version"),
                      m("th", "Created"),
                      m("th", { style: "width: 100px;" }, "Status"),
                      m("th.th-actions", { style: "width: 120px;" }, "Actions"),
                    ]),
                  ),
                  m(
                    "tbody",
                    FunctionDetail.versions.map((v) =>
                      m("tr", [
                        m(
                          "td",
                          m("input[type=checkbox]", {
                            checked: FunctionDetail.selectedVersions.includes(
                              v.version,
                            ),
                            disabled:
                              FunctionDetail.selectedVersions.length === 2 &&
                              !FunctionDetail.selectedVersions.includes(
                                v.version,
                              ),
                            onchange: (e) => {
                              if (e.target.checked) {
                                if (
                                  FunctionDetail.selectedVersions.length < 2
                                ) {
                                  FunctionDetail.selectedVersions.push(
                                    v.version,
                                  );
                                }
                              } else {
                                const index =
                                  FunctionDetail.selectedVersions.indexOf(
                                    v.version,
                                  );
                                if (index > -1) {
                                  FunctionDetail.selectedVersions.splice(
                                    index,
                                    1,
                                  );
                                }
                              }
                            },
                          }),
                        ),
                        m("td", m(".badge", `v${v.version}`)),
                        m("td", new Date(v.created_at).toLocaleString()),
                        m(
                          "td",
                          v.is_active
                            ? m(".badge.badge-success", "Active")
                            : m(".badge", "Inactive"),
                        ),
                        m(
                          "td.td-actions",
                          !v.is_active &&
                            m(
                              "button.btn.btn-sm",
                              {
                                onclick: () =>
                                  FunctionDetail.activateVersion(v.version),
                              },
                              [Icons.arrowsRightLeft(), " Activate"],
                            ),
                        ),
                      ]),
                    ),
                  ),
                ]),
                m(Pagination, {
                  total: FunctionDetail.versionsTotal,
                  limit: FunctionDetail.versionsLimit,
                  offset: FunctionDetail.versionsOffset,
                  onPageChange: FunctionDetail.handleVersionsPageChange,
                  onLimitChange: FunctionDetail.handleVersionsLimitChange,
                }),
              ],
        ]),

      // Executions Tab
      FunctionDetail.activeTab === "executions" &&
        m(".card", [
          m(".card-header", m(".card-title", "Recent Executions")),
          FunctionDetail.executions.length === 0
            ? m(".text-center.mt-24.mb-24", "No executions yet")
            : [
                m("table", [
                  m(
                    "thead",
                    m("tr", [
                      m("th", "Execution ID"),
                      m("th", "Status"),
                      m("th", "Duration"),
                      m("th", "Created"),
                      m("th.th-actions", "Actions"),
                    ]),
                  ),
                  m(
                    "tbody",
                    FunctionDetail.executions.map((exec) =>
                      m("tr", [
                        m("td", m(IdPill, { id: exec.id })),
                        m(
                          "td",
                          m(
                            ".badge",
                            {
                              class:
                                exec.status === "success"
                                  ? "badge-success"
                                  : exec.status === "error"
                                    ? "badge-error"
                                    : "",
                            },
                            exec.status,
                          ),
                        ),
                        m(
                          "td",
                          exec.duration_ms ? `${exec.duration_ms}ms` : "N/A",
                        ),
                        m(
                          "td",
                          new Date(exec.created_at * 1000).toLocaleString(),
                        ),
                        m(
                          "td.td-actions",
                          m(
                            ".actions",
                            m(
                              "a.btn.btn-icon",
                              {
                                href: `#!/executions/${exec.id}`,
                              },
                              Icons.eye(),
                            ),
                          ),
                        ),
                      ]),
                    ),
                  ),
                ]),
                m(Pagination, {
                  total: FunctionDetail.executionsTotal,
                  limit: FunctionDetail.executionsLimit,
                  offset: FunctionDetail.executionsOffset,
                  onPageChange: FunctionDetail.handleExecutionsPageChange,
                  onLimitChange: FunctionDetail.handleExecutionsLimitChange,
                }),
              ],
        ]),

      // Test Tab
      FunctionDetail.activeTab === "test" &&
        m("div", [
          m(CodeExamples, {
            functionId: func.id,
            method: FunctionDetail.testRequest.method,
            query: FunctionDetail.testRequest.query,
            body: FunctionDetail.testRequest.body,
          }),

          m(".card", [
            m(".card-header", m(".card-title", "Test Function")),
            m("div", { style: "padding: 24px;" }, [
              m(".form-group", [
                m("label.form-label", "Method"),
                m(
                  "select.form-select",
                  {
                    value: FunctionDetail.testRequest.method,
                    onchange: (e) =>
                      (FunctionDetail.testRequest.method = e.target.value),
                  },
                  [
                    m("option", { value: "GET" }, "GET"),
                    m("option", { value: "POST" }, "POST"),
                    m("option", { value: "PUT" }, "PUT"),
                    m("option", { value: "DELETE" }, "DELETE"),
                  ],
                ),
              ]),
              m(".form-group", [
                m("label.form-label", "Query Parameters"),
                m("input.form-input", {
                  value: FunctionDetail.testRequest.query,
                  oninput: (e) =>
                    (FunctionDetail.testRequest.query = e.target.value),
                  placeholder: "key1=value1&key2=value2",
                  style: "font-family: monospace;",
                }),
              ]),
              m(".form-group", [
                m("label.form-label", "Request Body"),
                m("textarea.form-textarea", {
                  value: FunctionDetail.testRequest.body,
                  oninput: (e) =>
                    (FunctionDetail.testRequest.body = e.target.value),
                  rows: 8,
                  style: "font-family: monospace;",
                }),
              ]),
              m(
                "button.btn.btn-primary",
                {
                  onclick: FunctionDetail.executeTest,
                },
                [Icons.play(), "  Execute"],
              ),

              FunctionDetail.testResponse &&
                m("div", { style: "margin-top: 24px;" }, [
                  m(".form-group", [
                    m("label.form-label", "Response"),
                    m(
                      "div",
                      m(
                        ".badge",
                        {
                          class:
                            FunctionDetail.testResponse.status === 200
                              ? "badge-success"
                              : "",
                        },
                        `Status: ${FunctionDetail.testResponse.status}`,
                      ),
                    ),
                  ]),
                  m(".form-group", [
                    m("label.form-label", "Body"),
                    m(
                      "pre",
                      {
                        style:
                          "background: #0a0a0a; padding: 12px; border-radius: 4px; margin: 0; border: 1px solid #262626; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace; font-size: 13px; line-height: 1.5; overflow-x: auto; white-space: pre-wrap; word-wrap: break-word; min-height: 100px; max-height: 400px; overflow-y: auto;",
                      },
                      FunctionDetail.testResponse.body,
                    ),
                  ]),
                  FunctionDetail.testLogs.length > 0 &&
                    m(".form-group", { style: "margin-top: 24px;" }, [
                      m("label.form-label", "Logs"),
                      m(
                        "div",
                        {
                          style:
                            "max-height: 300px; overflow-y: auto; border: 1px solid #262626; border-radius: 4px;",
                        },
                        m("table", { style: "margin: 0;" }, [
                          m("thead", [
                            m("tr", [
                              m("th", { style: "width: 80px;" }, "Level"),
                              m("th", { style: "width: 180px;" }, "Timestamp"),
                              m("th", "Message"),
                            ]),
                          ]),
                          m(
                            "tbody",
                            FunctionDetail.testLogs.map((log) =>
                              m("tr", [
                                m(
                                  "td",
                                  m(
                                    "span.badge",
                                    {
                                      class:
                                        log.level.toLowerCase() === "error"
                                          ? "badge-error"
                                          : log.level.toLowerCase() === "warn"
                                            ? "badge-warn"
                                            : log.level.toLowerCase() === "info"
                                              ? "badge-info"
                                              : "badge-debug",
                                    },
                                    log.level.toUpperCase(),
                                  ),
                                ),
                                m(
                                  "td",
                                  {
                                    style:
                                      "font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace; font-size: 12px; color: #a3a3a3;",
                                  },
                                  new Date(log.timestamp).toLocaleTimeString(),
                                ),
                                m(
                                  "td",
                                  {
                                    style:
                                      "font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace; font-size: 13px;",
                                  },
                                  log.message,
                                ),
                              ]),
                            ),
                          ),
                        ]),
                      ),
                    ]),
                ]),
            ]),
          ]),
        ]),

      // Environment Tab
      FunctionDetail.activeTab === "env" &&
        m(".card", [
          m(".card-header", [
            m(".card-title", "Environment Variables"),
            m(
              "button.btn.btn-icon",
              {
                onclick: () =>
                  FunctionDetail.envVars.push({ key: "", value: "" }),
              },
              Icons.plus(),
            ),
          ]),
          m("div", { style: "padding: 24px;" }, [
            FunctionDetail.envVars.length === 0
              ? m(
                  ".text-center",
                  { style: "color: #a3a3a3; padding: 48px 0;" },
                  "No environment variables. Click + to add one.",
                )
              : FunctionDetail.envVars.map((envVar, i) =>
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
                          onclick: () => FunctionDetail.envVars.splice(i, 1),
                        },
                        Icons.xMark(),
                      ),
                    ],
                  ),
                ),
            m(
              "div",
              {
                style:
                  "display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px;",
              },
              [
                m(
                  "button.btn.btn-primary",
                  {
                    onclick: FunctionDetail.saveEnvVars,
                    disabled: FunctionDetail.savingEnv,
                  },
                  FunctionDetail.savingEnv ? "Saving..." : "Save Changes",
                ),
              ],
            ),
          ]),
        ]),
    ]);
  },
};
