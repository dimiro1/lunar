import { icons } from "../icons.js";
import { API } from "../api.js";
import { Toast } from "../components/toast.js";
import { BackButton } from "../components/button.js";
import { Card, CardHeader, CardContent } from "../components/card.js";
import {
  Badge,
  BadgeVariant,
  BadgeSize,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { Tabs, TabContent } from "../components/tabs.js";
import { getFunctionTabs } from "../utils.js";
import { routes } from "../routes.js";
import { CodeExamples } from "../components/code-examples.js";
import { RequestBuilder } from "../components/request-builder.js";
import { LogViewer } from "../components/log-viewer.js";
import { CodeViewer } from "../components/code-viewer.js";
import { FormGroup, FormLabel } from "../components/form.js";

export const FunctionTest = {
  func: null,
  loading: true,
  testRequest: {
    method: "GET",
    query: "",
    body: "",
  },
  testResponse: null,
  testLogs: [],

  oninit: (vnode) => {
    FunctionTest.testResponse = null;
    FunctionTest.testLogs = [];
    FunctionTest.loadFunction(vnode.attrs.id);
  },

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

  executeTest: async () => {
    try {
      const response = await API.execute(
        FunctionTest.func.id,
        FunctionTest.testRequest,
      );
      FunctionTest.testResponse = response;
      FunctionTest.testLogs = [];
      m.redraw();

      const executionId =
        response.headers && response.headers["X-Execution-Id"];
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
    }
  },

  view: (vnode) => {
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
              onMethodChange: (value) =>
                (FunctionTest.testRequest.method = value),
              onQueryChange: (value) =>
                (FunctionTest.testRequest.query = value),
              onBodyChange: (value) => (FunctionTest.testRequest.body = value),
              onExecute: FunctionTest.executeTest,
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
                                variant:
                                  FunctionTest.testResponse.status === 200
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
