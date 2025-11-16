import { Icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { IdPill } from "../components/id-pill.js";
import { CodeEditor } from "../components/code-editor.js";
import { CodeExamples } from "../components/code-examples.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp } from "../utils.js";

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

  deleteFunction: async () => {
    if (
      !confirm(
        `Are you sure you want to delete "${FunctionDetail.func.name}"? This action cannot be undone.`,
      )
    ) {
      return;
    }

    try {
      await API.functions.delete(FunctionDetail.func.id);
      Toast.show("Function deleted successfully", "success");
      m.route.set("/functions");
    } catch (e) {
      Toast.show("Failed to delete function: " + e.message, "error");
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
            m(
              "button.btn",
              {
                onclick: FunctionDetail.deleteFunction,
                style: "color: #f48771;",
              },
              [Icons.trash(), "  Delete"],
            ),
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
            class: FunctionDetail.activeTab === "env" ? "active" : "",
            onclick: () => FunctionDetail.setTab("env"),
          },
          `Environment (${Object.keys(FunctionDetail.func.env_vars || {}).length})`,
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

      // Overview Tab (combined with Code)
      FunctionDetail.activeTab === "overview" &&
        m(".layout-with-sidebar", [
          m(".main-column", [
            m(".card", [
              m(
                ".card-header",
                m(".card-title", `Code v${func.active_version.version}`),
              ),
              m("div", { style: "padding: 16px;" }, [
                m(CodeEditor, {
                  id: "code-viewer",
                  value: func.active_version.code,
                  readOnly: true,
                }),
              ]),
            ]),
          ]),
          m(".docs-sidebar", [
            m(".docs-header", [
              m("h3", "Details"),
              m("p.docs-subtitle", "Function metadata"),
            ]),
            m(".docs-content", { style: "padding: 12px;" }, [
              m("table", { style: "width: 100%; font-size: 11px;" }, [
                m("tbody", [
                  m("tr", [
                    m(
                      "td",
                      { style: "padding: 6px 8px; color: #858585;" },
                      "ID",
                    ),
                    m(
                      "td",
                      { style: "padding: 6px 8px;" },
                      m(IdPill, { id: func.id }),
                    ),
                  ]),
                  m("tr", [
                    m(
                      "td",
                      { style: "padding: 6px 8px; color: #858585;" },
                      "Name",
                    ),
                    m(
                      "td",
                      {
                        style:
                          "padding: 6px 8px; font-family: var(--font-mono);",
                      },
                      func.name,
                    ),
                  ]),
                  m("tr", [
                    m(
                      "td",
                      { style: "padding: 6px 8px; color: #858585;" },
                      "Version",
                    ),
                    m(
                      "td",
                      { style: "padding: 6px 8px;" },
                      m(
                        ".badge.badge-success",
                        { style: "font-size: 10px;" },
                        `v${func.active_version.version}`,
                      ),
                    ),
                  ]),
                  m("tr", [
                    m(
                      "td",
                      { style: "padding: 6px 8px; color: #858585;" },
                      "Created",
                    ),
                    m(
                      "td",
                      {
                        style:
                          "padding: 6px 8px; font-family: var(--font-mono); font-size: 10px;",
                      },
                      formatUnixTimestamp(func.created_at),
                    ),
                  ]),
                  m("tr", [
                    m(
                      "td",
                      { style: "padding: 6px 8px; color: #858585;" },
                      "Updated",
                    ),
                    m(
                      "td",
                      {
                        style:
                          "padding: 6px 8px; font-family: var(--font-mono); font-size: 10px;",
                      },
                      formatUnixTimestamp(func.updated_at),
                    ),
                  ]),
                ]),
              ]),
            ]),
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
                        m("td", formatUnixTimestamp(v.created_at)),
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
                        m("td", formatUnixTimestamp(exec.created_at)),
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

          m(".layout-with-sidebar", [
            // Request (left)
            m(".main-column", [
              m(".card", [
                m(".card-header", m(".card-title", "Request")),
                m("div", { style: "padding: 16px;" }, [
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
                ]),
              ]),
            ]),

            // Response (right)
            m(".docs-sidebar", { style: "top: 0; position: relative;" }, [
              FunctionDetail.testResponse
                ? m("div", { style: "padding: 16px;" }, [
                    m(
                      ".docs-header",
                      {
                        style:
                          "padding: 0 0 12px 0; border-bottom: 1px solid #3e3e42;",
                      },
                      [
                        m("h3", "Response"),
                        m(
                          ".badge",
                          {
                            class:
                              FunctionDetail.testResponse.status === 200
                                ? "badge-success"
                                : FunctionDetail.testResponse.status >= 400
                                  ? "badge-error"
                                  : "",
                            style: "margin-top: 4px;",
                          },
                          `${FunctionDetail.testResponse.status}`,
                        ),
                      ],
                    ),
                    m("div", { style: "margin-top: 12px;" }, [
                      m(
                        "label.form-label",
                        { style: "font-size: 11px; margin-bottom: 6px;" },
                        "Body",
                      ),
                      m(
                        "pre",
                        {
                          style:
                            "background: #1e1e1e; padding: 10px; border-radius: 4px; margin: 0; border: 1px solid #3e3e42; font-family: var(--font-mono); font-size: 11px; line-height: 1.5; overflow-x: auto; white-space: pre-wrap; word-wrap: break-word; max-height: 300px; overflow-y: auto;",
                        },
                        FunctionDetail.testResponse.body,
                      ),
                    ]),
                    FunctionDetail.testLogs.length > 0 &&
                      m("div", { style: "margin-top: 16px;" }, [
                        m(
                          "label.form-label",
                          { style: "font-size: 11px; margin-bottom: 6px;" },
                          "Logs",
                        ),
                        m(
                          "div",
                          {
                            style:
                              "max-height: 300px; overflow-y: auto; border: 1px solid #3e3e42; border-radius: 4px;",
                          },
                          m("table", { style: "margin: 0; font-size: 11px;" }, [
                            m("thead", [
                              m("tr", [
                                m(
                                  "th",
                                  { style: "width: 60px; padding: 8px;" },
                                  "Level",
                                ),
                                m("th", { style: "padding: 8px;" }, "Message"),
                              ]),
                            ]),
                            m(
                              "tbody",
                              FunctionDetail.testLogs.map((log) =>
                                m("tr", [
                                  m(
                                    "td",
                                    { style: "padding: 6px 8px;" },
                                    m(
                                      "span.badge",
                                      {
                                        class:
                                          log.level.toLowerCase() === "error"
                                            ? "badge-error"
                                            : log.level.toLowerCase() === "warn"
                                              ? "badge-warn"
                                              : log.level.toLowerCase() ===
                                                  "info"
                                                ? "badge-info"
                                                : "badge-debug",
                                        style:
                                          "font-size: 9px; padding: 2px 6px;",
                                      },
                                      log.level.toUpperCase(),
                                    ),
                                  ),
                                  m(
                                    "td",
                                    {
                                      style:
                                        "font-family: var(--font-mono); font-size: 11px; padding: 6px 8px;",
                                    },
                                    log.message,
                                  ),
                                ]),
                              ),
                            ),
                          ]),
                        ),
                      ]),
                  ])
                : m(
                    "div",
                    {
                      style:
                        "padding: 16px; text-align: center; color: #858585;",
                    },
                    [
                      m("p", { style: "margin: 0;" }, "No response yet"),
                      m(
                        "p",
                        { style: "margin: 8px 0 0 0; font-size: 11px;" },
                        "Execute a request to see the response",
                      ),
                    ],
                  ),
            ]),
          ]),
        ]),

      // Environment Tab
      FunctionDetail.activeTab === "env" &&
        m(".layout-with-sidebar", [
          m(".main-column", [
            m(".card", [
              m(".card-header", [
                m("div", [
                  m(".card-title", "Environment Variables"),
                  m(
                    ".card-subtitle",
                    `${Object.keys(FunctionDetail.func.env_vars || {}).length} variables`,
                  ),
                ]),
                m(
                  "a.btn.btn-primary",
                  { href: `#!/functions/${FunctionDetail.func.id}/env` },
                  "Manage Variables",
                ),
              ]),
              Object.keys(FunctionDetail.func.env_vars || {}).length === 0
                ? m(
                    ".text-center",
                    { style: "color: #858585; padding: 48px 0;" },
                    "No environment variables configured.",
                  )
                : m("div", { style: "padding: 24px;" }, [
                    Object.entries(FunctionDetail.func.env_vars).map(
                      ([key, value]) =>
                        m(
                          "div",
                          {
                            key: key,
                            style:
                              "display: flex; gap: 12px; margin-bottom: 8px; font-family: var(--font-mono); font-size: 13px;",
                          },
                          [
                            m(
                              "div",
                              { style: "color: #858585; min-width: 150px;" },
                              key,
                            ),
                            m(
                              "div",
                              {
                                style: "color: #cccccc; word-break: break-all;",
                              },
                              value,
                            ),
                          ],
                        ),
                    ),
                  ]),
            ]),
          ]),
          m(".docs-sidebar", [
            m(".docs-header", [
              m("h3", "Environment Variables"),
              m("p.docs-subtitle", "Usage in functions"),
            ]),
            m(".docs-content", [
              m(".docs-section", [
                m(
                  "p",
                  {
                    style:
                      "font-size: 11px; color: #cccccc; line-height: 1.5; margin-bottom: 12px;",
                  },
                  "Environment variables are available in your functions through the env API.",
                ),
                m(".docs-methods", [
                  m(".docs-method", [
                    m("code.docs-signature", "env.get(key)"),
                    m("span.docs-desc", "Get environment variable value"),
                  ]),
                  m(".docs-method", [
                    m("code.docs-signature", "env.set(key, value)"),
                    m("span.docs-desc", "Set environment variable (persisted)"),
                  ]),
                  m(".docs-method", [
                    m("code.docs-signature", "env.delete(key)"),
                    m("span.docs-desc", "Delete environment variable"),
                  ]),
                ]),
              ]),
              m(
                ".docs-section",
                {
                  style:
                    "margin-top: 16px; padding-top: 16px; border-top: 1px solid #3e3e42;",
                },
                [
                  m("h4.docs-section-title", "Example"),
                  m(
                    "pre.docs-example",
                    {
                      oncreate: (vnode) => {
                        hljs.highlightElement(vnode.dom.querySelector("code"));
                      },
                    },
                    [
                      m(
                        "code.language-lua",
                        `-- Get API key
local apiKey = env.get("API_KEY")

-- Use in HTTP request
local res = http.get(url, {
  headers = {
    ["Authorization"] = apiKey
  }
})`,
                      ),
                    ],
                  ),
                ],
              ),
              m(
                ".docs-section",
                {
                  style:
                    "margin-top: 16px; padding-top: 16px; border-top: 1px solid #3e3e42;",
                },
                [
                  m("h4.docs-section-title", "Best Practices"),
                  m(
                    "ul",
                    {
                      style:
                        "font-size: 11px; color: #858585; line-height: 1.6; margin: 8px 0 0 0; padding-left: 20px;",
                    },
                    [
                      m("li", "Use UPPER_CASE for keys"),
                      m("li", "Store secrets and API keys"),
                      m("li", "Don't hardcode sensitive data"),
                      m("li", "Use descriptive key names"),
                    ],
                  ),
                ],
              ),
            ]),
          ]),
        ]),
    ]);
  },
};
