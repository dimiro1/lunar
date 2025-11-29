/**
 * @fileoverview Execution detail view with logs and error information.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { Pagination } from "../components/pagination.js";
import { formatUnixTimestamp } from "../utils.js";
import { routes } from "../routes.js";
import { BackButton } from "../components/button.js";
import {
  Card,
  CardContent,
  CardHeader,
  CardVariant,
} from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { LogViewer } from "../components/log-viewer.js";
import { CodeViewer } from "../components/code-viewer.js";
import { AIRequestViewer } from "../components/ai-request-viewer.js";
import { EmailRequestViewer } from "../components/email-request-viewer.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 * @typedef {import('../types.js').Execution} Execution
 * @typedef {import('../types.js').ExecutionLog} ExecutionLog
 * @typedef {import('../types.js').AIRequest} AIRequest
 * @typedef {import('../types.js').EmailRequest} EmailRequest
 */

/**
 * Execution detail view component.
 * Displays execution information, logs, errors, and input event data.
 * @type {Object}
 */
export const ExecutionDetail = {
  /**
   * Parent function of the execution.
   * @type {LunarFunction|null}
   */
  func: null,

  /**
   * Currently loaded execution.
   * @type {Execution|null}
   */
  execution: null,

  /**
   * Execution logs.
   * @type {ExecutionLog[]}
   */
  logs: [],

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Number of logs per page.
   * @type {number}
   */
  logsLimit: 20,

  /**
   * Current logs pagination offset.
   * @type {number}
   */
  logsOffset: 0,

  /**
   * Total number of log entries.
   * @type {number}
   */
  logsTotal: 0,

  /**
   * AI requests for this execution.
   * @type {AIRequest[]}
   */
  aiRequests: [],

  /**
   * Number of AI requests per page.
   * @type {number}
   */
  aiRequestsLimit: 20,

  /**
   * Current AI requests pagination offset.
   * @type {number}
   */
  aiRequestsOffset: 0,

  /**
   * Total number of AI requests.
   * @type {number}
   */
  aiRequestsTotal: 0,

  /**
   * Email requests for this execution.
   * @type {EmailRequest[]}
   */
  emailRequests: [],

  /**
   * Number of email requests per page.
   * @type {number}
   */
  emailRequestsLimit: 20,

  /**
   * Current email requests pagination offset.
   * @type {number}
   */
  emailRequestsOffset: 0,

  /**
   * Total number of email requests.
   * @type {number}
   */
  emailRequestsTotal: 0,

  /**
   * Initializes the view and loads execution data.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    ExecutionDetail.loadExecution(vnode.attrs.id);
  },

  /**
   * Loads execution data and associated function.
   * @param {string} id - Execution ID
   * @returns {Promise<void>}
   */
  loadExecution: async (id) => {
    ExecutionDetail.loading = true;
    try {
      const [execution, logsData, aiRequestsData, emailRequestsData] =
        await Promise.all([
          API.executions.get(id),
          API.executions.getLogs(
            id,
            ExecutionDetail.logsLimit,
            ExecutionDetail.logsOffset,
          ),
          API.executions.getAIRequests(
            id,
            ExecutionDetail.aiRequestsLimit,
            ExecutionDetail.aiRequestsOffset,
          ),
          API.executions.getEmailRequests(
            id,
            ExecutionDetail.emailRequestsLimit,
            ExecutionDetail.emailRequestsOffset,
          ),
        ]);
      ExecutionDetail.execution = execution;
      ExecutionDetail.logs = logsData.logs || [];
      ExecutionDetail.logsTotal = logsData.pagination?.total || 0;
      ExecutionDetail.aiRequests = aiRequestsData.ai_requests || [];
      ExecutionDetail.aiRequestsTotal = aiRequestsData.pagination?.total || 0;
      ExecutionDetail.emailRequests = emailRequestsData.email_requests || [];
      ExecutionDetail.emailRequestsTotal =
        emailRequestsData.pagination?.total || 0;

      // Load function details
      ExecutionDetail.func = await API.functions.get(execution.function_id);
    } catch (e) {
      console.error("Failed to load execution:", e);
    } finally {
      ExecutionDetail.loading = false;
      m.redraw();
    }
  },

  /**
   * Reloads logs with current pagination.
   * @returns {Promise<void>}
   */
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

  /**
   * Handles page change from logs pagination.
   * @param {number} newOffset - New pagination offset
   */
  handleLogsPageChange: (newOffset) => {
    ExecutionDetail.logsOffset = newOffset;
    ExecutionDetail.loadLogs();
  },

  /**
   * Handles limit change from logs pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleLogsLimitChange: (newLimit) => {
    ExecutionDetail.logsLimit = newLimit;
    ExecutionDetail.logsOffset = 0;
    ExecutionDetail.loadLogs();
  },

  /**
   * Reloads AI requests with current pagination.
   * @returns {Promise<void>}
   */
  loadAIRequests: async () => {
    try {
      const data = await API.executions.getAIRequests(
        ExecutionDetail.execution.id,
        ExecutionDetail.aiRequestsLimit,
        ExecutionDetail.aiRequestsOffset,
      );
      ExecutionDetail.aiRequests = data.ai_requests || [];
      ExecutionDetail.aiRequestsTotal = data.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load AI requests:", e);
    }
  },

  /**
   * Handles page change from AI requests pagination.
   * @param {number} newOffset - New pagination offset
   */
  handleAIRequestsPageChange: (newOffset) => {
    ExecutionDetail.aiRequestsOffset = newOffset;
    ExecutionDetail.loadAIRequests();
  },

  /**
   * Handles limit change from AI requests pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleAIRequestsLimitChange: (newLimit) => {
    ExecutionDetail.aiRequestsLimit = newLimit;
    ExecutionDetail.aiRequestsOffset = 0;
    ExecutionDetail.loadAIRequests();
  },

  /**
   * Reloads email requests with current pagination.
   * @returns {Promise<void>}
   */
  loadEmailRequests: async () => {
    try {
      const data = await API.executions.getEmailRequests(
        ExecutionDetail.execution.id,
        ExecutionDetail.emailRequestsLimit,
        ExecutionDetail.emailRequestsOffset,
      );
      ExecutionDetail.emailRequests = data.email_requests || [];
      ExecutionDetail.emailRequestsTotal = data.pagination?.total || 0;
      m.redraw();
    } catch (e) {
      console.error("Failed to load email requests:", e);
    }
  },

  /**
   * Handles page change from email requests pagination.
   * @param {number} newOffset - New pagination offset
   */
  handleEmailRequestsPageChange: (newOffset) => {
    ExecutionDetail.emailRequestsOffset = newOffset;
    ExecutionDetail.loadEmailRequests();
  },

  /**
   * Handles limit change from email requests pagination.
   * @param {number} newLimit - New items per page limit
   */
  handleEmailRequestsLimitChange: (newLimit) => {
    ExecutionDetail.emailRequestsLimit = newLimit;
    ExecutionDetail.emailRequestsOffset = 0;
    ExecutionDetail.loadEmailRequests();
  },

  /**
   * Renders the execution detail view.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    if (ExecutionDetail.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("execution.loadingExecution")),
      ]);
    }

    if (!ExecutionDetail.execution) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("execution.executionNotFound"))),
      );
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
                  variant: exec.status === "success"
                    ? BadgeVariant.SUCCESS
                    : BadgeVariant.DESTRUCTIVE,
                  size: BadgeSize.SM,
                },
                t(`common.status.${exec.status}`),
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
        (() => {
          // Parse error message sections
          const parts = exec.error_message.split(/\[CODE]|\[\/CODE]/);
          const errorDescription = parts[0] || "";
          const codeSnippet = parts[1] || "";
          const tipSection = parts[2] || "";

          // Only trim trailing whitespace to preserve line number alignment
          const trimmedCode = codeSnippet
            .replace(/^\n+/, "")
            .replace(/\n+$/, "");

          return m(
            Card,
            { variant: CardVariant.DANGER, style: "margin-bottom: 1.5rem" },
            [
              m(CardHeader, {
                title: t("execution.executionError"),
                icon: "exclamationTriangle",
                variant: CardVariant.DANGER,
              }),
              m(CardContent, [
                // Error description
                errorDescription &&
                m(
                  "pre.error-description",
                  {
                    style:
                      "white-space: pre-wrap; font-family: monospace; margin: 0 0 1rem 0;",
                  },
                  errorDescription.trim(),
                ),

                // Code snippet with line numbers and syntax highlighting
                trimmedCode &&
                m("div", { style: "margin: 1rem 0;" }, [
                  m(CodeViewer, {
                    code: trimmedCode,
                    language: "lua",
                    maxHeight: "300px",
                    noBorder: false,
                    padded: true,
                  }),
                ]),

                // Tip section
                tipSection &&
                m(
                  "pre.error-tip",
                  {
                    style:
                      "white-space: pre-wrap; font-family: monospace; margin: 1rem 0 0 0; padding: 1rem; background: var(--color-background); border-radius: 6px;",
                  },
                  tipSection.trim(),
                ),
              ]),
            ],
          );
        })(),

        // Event Data
        exec.event_json &&
        m(Card, { style: "margin-bottom: 1.5rem" }, [
          m(CardHeader, { title: t("execution.inputEvent") }),
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

        // AI Requests
        ExecutionDetail.aiRequestsTotal > 0 &&
        m(Card, { style: "margin-bottom: 1.5rem" }, [
          m(CardHeader, {
            title: t("execution.aiRequests"),
            subtitle: t("execution.aiRequestsCount", {
              count: ExecutionDetail.aiRequestsTotal,
            }),
            icon: "network",
          }),
          m(CardContent, { noPadding: true }, [
            m(AIRequestViewer, {
              requests: ExecutionDetail.aiRequests,
              maxHeight: "400px",
              noBorder: true,
            }),
          ]),
          ExecutionDetail.aiRequestsTotal > ExecutionDetail.aiRequestsLimit &&
          m(Pagination, {
            total: ExecutionDetail.aiRequestsTotal,
            limit: ExecutionDetail.aiRequestsLimit,
            offset: ExecutionDetail.aiRequestsOffset,
            onPageChange: ExecutionDetail.handleAIRequestsPageChange,
            onLimitChange: ExecutionDetail.handleAIRequestsLimitChange,
          }),
        ]),

        // Email Requests
        ExecutionDetail.emailRequestsTotal > 0 &&
        m(Card, { style: "margin-bottom: 1.5rem" }, [
          m(CardHeader, {
            title: t("execution.emailRequests"),
            subtitle: t("execution.emailsSent", {
              count: ExecutionDetail.emailRequestsTotal,
            }),
            icon: "mail",
          }),
          m(CardContent, { noPadding: true }, [
            m(EmailRequestViewer, {
              requests: ExecutionDetail.emailRequests,
              maxHeight: "400px",
              noBorder: true,
            }),
          ]),
          ExecutionDetail.emailRequestsTotal >
            ExecutionDetail.emailRequestsLimit &&
          m(Pagination, {
            total: ExecutionDetail.emailRequestsTotal,
            limit: ExecutionDetail.emailRequestsLimit,
            offset: ExecutionDetail.emailRequestsOffset,
            onPageChange: ExecutionDetail.handleEmailRequestsPageChange,
            onLimitChange: ExecutionDetail.handleEmailRequestsLimitChange,
          }),
        ]),

        // Execution Logs
        m(Card, [
          m(CardHeader, {
            title: t("execution.executionLogs"),
            subtitle: t("execution.logEntries", {
              count: ExecutionDetail.logsTotal,
            }),
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
