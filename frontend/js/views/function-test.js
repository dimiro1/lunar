/**
 * @fileoverview Function test view with request builder and response viewer.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { BackButton } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { TabContent, Tabs } from "../components/tabs.js";
import { getFunctionTabs } from "../utils.js";
import { routes } from "../routes.js";
import { CodeExamples } from "../components/code-examples.js";
import { RequestBuilder } from "../components/request-builder.js";
import { LogViewer } from "../components/log-viewer.js";
import { CodeViewer } from "../components/code-viewer.js";
import { FormGroup, FormLabel } from "../components/form.js";

/**
 * @typedef {import('../types.js').FaaSFunction} FaaSFunction
 * @typedef {import('../types.js').ExecuteResponse} ExecuteResponse
 * @typedef {import('../types.js').ExecutionLog} ExecutionLog
 */

/**
 * @typedef {Object} TestRequest
 * @property {string} method - HTTP method
 * @property {string} query - Query string
 * @property {string} body - Request body
 */

/**
 * Function test view component.
 * Provides a request builder to test function execution with code examples.
 * @type {Object}
 */
export const FunctionTest = {
  /**
   * Currently loaded function.
   * @type {FaaSFunction|null}
   */
  func: null,

  /**
   * Whether the view is loading.
   * @type {boolean}
   */
  loading: true,

  /**
   * Test request configuration.
   * @type {TestRequest}
   */
  testRequest: {
    method: "GET",
    query: "",
    body: "",
  },

  /**
   * Last test response (null if no test run yet).
   * @type {ExecuteResponse|null}
   */
  testResponse: null,

  /**
   * Logs from last test execution.
   * @type {ExecutionLog[]}
   */
  testLogs: [],

  /**
   * Whether test execution is in progress.
   * @type {boolean}
   */
  executing: false,

  /**
   * Initializes the view and loads the function.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    FunctionTest.testResponse = null;
    FunctionTest.testLogs = [];
    FunctionTest.executing = false;
    FunctionTest.loadFunction(vnode.attrs.id);
  },

  /**
   * Loads a function by ID.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadFunction: async (id) => {
    FunctionTest.loading = true;
    try {
      FunctionTest.func = await API.functions.get(id);
    } catch (e) {
      console.error("Failed to load function:", e);
    } finally {
      FunctionTest.loading = false;
      m.redraw();
    }
  },

  /**
   * Executes a test request against the function.
   * @returns {Promise<void>}
   */
  executeTest: async () => {
    FunctionTest.executing = true;
    m.redraw();
    try {
      const response = await API.execute(
        FunctionTest.func.id,
        FunctionTest.testRequest,
      );
      FunctionTest.testResponse = response;
      FunctionTest.testLogs = [];
      m.redraw();

      const executionId = response.headers &&
        response.headers["X-Execution-Id"];
      if (executionId) {
        try {
          const logsData = await API.executions.getLogs(executionId);
          FunctionTest.testLogs = logsData.logs || [];
          m.redraw();
        } catch (e) {
          console.error("Failed to load logs:", e);
          FunctionTest.testLogs = [];
        }
      }
    } catch (e) {
      Toast.show("Execution failed: " + e.message, "error");
    } finally {
      FunctionTest.executing = false;
      m.redraw();
    }
  },

  /**
   * Renders the function test view.
   * @param {Object} _vnode - Mithril vnode
   * @returns {Object} Mithril vnode
   */
  view: (_vnode) => {
    if (FunctionTest.loading) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", "Loading function..."),
      ]);
    }

    if (!FunctionTest.func) {
      return m(".fade-in", m(Card, m(CardContent, "Function not found")));
    }

    const func = FunctionTest.func;

    return m(".fade-in", [
      // Header
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functions() }),
          m(".function-details-divider"),
          m(".function-details-info", [
            m("h1.function-details-title", [
              func.name,
              m(IDBadge, { id: func.id }),
              m(
                Badge,
                {
                  variant: BadgeVariant.OUTLINE,
                  size: BadgeSize.SM,
                  mono: true,
                },
                `v${func.active_version.version}`,
              ),
            ]),
            m(
              "p.function-details-description",
              func.description || "No description",
            ),
          ]),
        ]),
        m(".function-details-actions", [
          m(StatusBadge, { enabled: !func.disabled, glow: true }),
        ]),
      ]),

      // Tabs
      m(Tabs, {
        tabs: getFunctionTabs(func.id),
        activeTab: "test",
      }),

      // Content
      m(TabContent, [
        m(".test-tab-container", [
          m(CodeExamples, {
            functionId: func.id,
            method: FunctionTest.testRequest.method,
            query: FunctionTest.testRequest.query,
            body: FunctionTest.testRequest.body,
          }),

          m(".test-panels", [
            // Request Builder
            m(RequestBuilder, {
              url: `${window.location.origin}/fn/${func.id}`,
              method: FunctionTest.testRequest.method,
              query: FunctionTest.testRequest.query,
              body: FunctionTest.testRequest.body,
              onMethodChange: (
                value,
              ) => (FunctionTest.testRequest.method = value),
              onQueryChange: (
                value,
              ) => (FunctionTest.testRequest.query = value),
              onBodyChange: (value) => (FunctionTest.testRequest.body = value),
              onExecute: FunctionTest.executeTest,
              loading: FunctionTest.executing,
            }),

            // Response Viewer
            m(Card, { class: "response-viewer" }, [
              m(CardHeader, {
                title: "Response",
                subtitle: FunctionTest.testResponse
                  ? `Status: ${FunctionTest.testResponse.status}`
                  : null,
              }),
              m(CardContent, [
                FunctionTest.testResponse
                  ? m("div", [
                    m(FormGroup, [
                      m(FormLabel, { text: "Status" }),
                      m(
                        "div",
                        {
                          style:
                            "display: flex; align-items: center; gap: 0.5rem;",
                        },
                        [
                          m(
                            Badge,
                            {
                              variant: FunctionTest.testResponse.status === 200
                                ? BadgeVariant.SUCCESS
                                : BadgeVariant.DESTRUCTIVE,
                              size: BadgeSize.SM,
                            },
                            FunctionTest.testResponse.status,
                          ),
                          FunctionTest.testResponse.headers &&
                          FunctionTest.testResponse.headers[
                            "X-Execution-Id"
                          ] &&
                          m(
                            "a",
                            {
                              href: routes.execution(
                                FunctionTest.testResponse.headers[
                                  "X-Execution-Id"
                                ],
                              ),
                              style: "text-decoration: none;",
                            },
                            m(
                              Badge,
                              {
                                variant: BadgeVariant.OUTLINE,
                                size: BadgeSize.SM,
                              },
                              "View Execution",
                            ),
                          ),
                        ],
                      ),
                    ]),
                    m(FormGroup, [
                      m(FormLabel, { text: "Body" }),
                      m(CodeViewer, {
                        code: FunctionTest.testResponse.body || "",
                        language: "json",
                        maxHeight: "300px",
                        wrap: true,
                      }),
                    ]),
                    FunctionTest.testLogs.length > 0 &&
                    m(FormGroup, [
                      m(FormLabel, { text: "Logs" }),
                      m(LogViewer, {
                        logs: FunctionTest.testLogs,
                        maxHeight: "200px",
                      }),
                    ]),
                  ])
                  : m(".no-response", [
                    m("p", "No response yet"),
                    m(
                      "p.text-muted",
                      "Execute a request to see the response",
                    ),
                  ]),
              ]),
            ]),
          ]),
        ]),
      ]),
    ]);
  },
};
