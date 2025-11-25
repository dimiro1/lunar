import { icons } from "../icons.js";
import { API } from "../api.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp } from "../utils.js";
import { routes } from "../routes.js";
import { BackButton } from "../components/button.js";
import {
  Card,
  CardHeader,
  CardContent,
  CardVariant,
} from "../components/card.js";
import {
  Badge,
  BadgeVariant,
  BadgeSize,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { LogViewer } from "../components/log-viewer.js";
import { CodeViewer } from "../components/code-viewer.js";

export const ExecutionDetail = {
  func: null,
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

      // Load function details
      ExecutionDetail.func = await API.functions.get(execution.function_id);
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
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", "Loading execution..."),
      ]);
    }

    if (!ExecutionDetail.execution) {
      return m(".fade-in", m(Card, m(CardContent, "Execution not found")));
    }

    const exec = ExecutionDetail.execution;
    const func = ExecutionDetail.func;

    return m(".fade-in", [
      // Header
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functionExecutions(exec.function_id) }),
          m(".function-details-divider"),
          m(".function-details-info", [
            m("h1.function-details-title", [
              func ? func.name : "Function",
              func && m(IDBadge, { id: func.id }),
              m(
                Badge,
                {
                  variant: BadgeVariant.SECONDARY,
                  size: BadgeSize.SM,
                  mono: true,
                },
                `exec: ${exec.id.substring(0, 8)}`,
              ),
              m(
                Badge,
                {
                  variant:
                    exec.status === "success"
                      ? BadgeVariant.SUCCESS
                      : BadgeVariant.DESTRUCTIVE,
                  size: BadgeSize.SM,
                },
                exec.status.toUpperCase(),
              ),
              exec.duration_ms &&
                m(
                  Badge,
                  {
                    variant: BadgeVariant.OUTLINE,
                    size: BadgeSize.SM,
                    mono: true,
                  },
                  `${exec.duration_ms}ms`,
                ),
            ]),
            m(
              "p.function-details-description",
              formatUnixTimestamp(exec.created_at),
            ),
          ]),
        ]),
        m(".function-details-actions", [
          func && m(StatusBadge, { enabled: !func.disabled, glow: true }),
        ]),
      ]),

      m(".execution-details-panels", [
        // Error Details
        exec.status === "error" &&
          exec.error_message &&
          m(Card, { variant: CardVariant.DANGER, style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, {
              title: "Execution Error",
              icon: "exclamationTriangle",
              variant: CardVariant.DANGER,
            }),
            m(CardContent, [
              m("pre.error-message", {
                style: "max-height: 500px; overflow-y: auto; white-space: pre-wrap; font-family: monospace; margin: 0; padding: 1rem; background: var(--color-background); border-radius: 6px;"
              }, exec.error_message),
            ]),
          ]),

        // Event Data
        exec.event_json &&
          m(Card, { style: "margin-bottom: 1.5rem" }, [
            m(CardHeader, { title: "Input Event (JSON)" }),
            m(CardContent, { noPadding: true }, [
              m(CodeViewer, {
                code: JSON.stringify(JSON.parse(exec.event_json), null, 2),
                language: "json",
                maxHeight: "200px",
                noBorder: true,
                padded: true,
              }),
            ]),
          ]),

        // Execution Logs
        m(Card, [
          m(CardHeader, {
            title: "Execution Logs",
            subtitle: `${ExecutionDetail.logsTotal} log entries`,
          }),
          m(CardContent, { noPadding: true }, [
            m(LogViewer, {
              logs: ExecutionDetail.logs.map((log) => ({
                ...log,
                timestamp: formatUnixTimestamp(log.created_at, "time"),
              })),
              maxHeight: "300px",
              noBorder: true,
            }),
          ]),
          ExecutionDetail.logsTotal > ExecutionDetail.logsLimit &&
            m(Pagination, {
              total: ExecutionDetail.logsTotal,
              limit: ExecutionDetail.logsLimit,
              offset: ExecutionDetail.logsOffset,
              onPageChange: ExecutionDetail.handleLogsPageChange,
              onLimitChange: ExecutionDetail.handleLogsLimitChange,
            }),
        ]),
      ]),
    ]);
  },
};
