import { Icons } from "../icons.js";
import { API } from "../api.js";
import { IdPill } from "../components/id-pill.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp } from "../utils.js";

export const ExecutionDetail = {
  execution: null,
  logs: [],
  loading: true,
  logsLimit: 20,
  logsOffset: 0,
  logsTotal: 0,

  oninit: (vnode) => {
    ExecutionDetail.loadExecution(vnode.attrs.id);
  },

  loadExecution: async (id) => {
    ExecutionDetail.loading = true;
    try {
      const [execution, logsData] = await Promise.all([
        API.executions.get(id),
        API.executions.getLogs(
          id,
          ExecutionDetail.logsLimit,
          ExecutionDetail.logsOffset,
        ),
      ]);
      ExecutionDetail.execution = execution;
      ExecutionDetail.logs = logsData.logs || [];
      ExecutionDetail.logsTotal = logsData.pagination?.total || 0;
    } catch (e) {
      console.error("Failed to load execution:", e);
    } finally {
      ExecutionDetail.loading = false;
      m.redraw();
    }
  },

  loadLogs: async () => {
    try {
      const logsData = await API.executions.getLogs(
        ExecutionDetail.execution.id,
        ExecutionDetail.logsLimit,
        ExecutionDetail.logsOffset,
      );
      ExecutionDetail.logs = logsData.logs || [];
      ExecutionDetail.logsTotal = logsData.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load logs:", e);
    }
  },

  handleLogsPageChange: (newOffset) => {
    ExecutionDetail.logsOffset = newOffset;
    ExecutionDetail.loadLogs();
  },

  handleLogsLimitChange: (newLimit) => {
    ExecutionDetail.logsLimit = newLimit;
    ExecutionDetail.logsOffset = 0;
    ExecutionDetail.loadLogs();
  },

  view: () => {
    if (ExecutionDetail.loading) {
      return m(".loading", "Loading...");
    }

    if (!ExecutionDetail.execution) {
      return m(".container", m(".card", "Execution not found"));
    }

    const exec = ExecutionDetail.execution;

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", "Execution Details"),
            m(".page-subtitle", m(IdPill, { id: exec.id })),
          ]),
          m("a.btn", { href: `#!/functions/${exec.function_id}` }, [
            Icons.arrowLeft(),
            "  Back",
          ]),
        ]),
      ]),

      m(".card.mb-24", [
        m(".card-header", m(".card-title", "Overview")),
        m("table", [
          m("tbody", [
            m("tr", [
              m("td", m("strong", "Execution ID")),
              m("td", m(IdPill, { id: exec.id })),
            ]),
            m("tr", [
              m("td", m("strong", "Function ID")),
              m("td", m(IdPill, { id: exec.function_id })),
            ]),
            m("tr", [
              m("td", m("strong", "Function Version ID")),
              m("td", m(IdPill, { id: exec.function_version_id })),
            ]),
            m("tr", [
              m("td", m("strong", "Status")),
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
            ]),
            m("tr", [
              m("td", m("strong", "Duration")),
              m("td", exec.duration_ms ? `${exec.duration_ms}ms` : "N/A"),
            ]),
            m("tr", [
              m("td", m("strong", "Created")),
              m("td", formatUnixTimestamp(exec.created_at)),
            ]),
            exec.error_message &&
              m("tr", [
                m("td", m("strong", "Error")),
                m("td", { style: "color: #ef4444;" }, exec.error_message),
              ]),
          ]),
        ]),
      ]),

      exec.event_json &&
        m(".card.mb-24", [
          m(".card-header", m(".card-title", "Event Data")),
          m(
            "pre",
            m(
              "code.language-json",
              {
                oncreate: (vnode) => {
                  hljs.highlightElement(vnode.dom);
                },
              },
              JSON.stringify(JSON.parse(exec.event_json), null, 2),
            ),
          ),
        ]),

      ExecutionDetail.logsTotal > 0 &&
        m(".card.mb-24", [
          m(
            ".card-header",
            m(".card-title", `Logs (${ExecutionDetail.logsTotal})`),
          ),
          m("table", [
            m("thead", [
              m("tr", [
                m("th", { style: "width: 80px;" }, "Level"),
                m("th", { style: "width: 180px;" }, "Timestamp"),
                m("th", "Message"),
              ]),
            ]),
            m(
              "tbody",
              ExecutionDetail.logs.map((log) =>
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
                    formatUnixTimestamp(log.created_at, "time"),
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
          m(Pagination, {
            total: ExecutionDetail.logsTotal,
            limit: ExecutionDetail.logsLimit,
            offset: ExecutionDetail.logsOffset,
            onPageChange: ExecutionDetail.handleLogsPageChange,
            onLimitChange: ExecutionDetail.handleLogsLimitChange,
          }),
        ]),
    ]);
  },
};
