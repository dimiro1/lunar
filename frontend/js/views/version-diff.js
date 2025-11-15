import { Icons } from '../icons.js';
import { API } from '../api.js';

export const VersionDiff = {
  func: null,
  diffData: null,
  loading: true,

  oninit: (vnode) => {
    VersionDiff.loadData(vnode.attrs.id, vnode.attrs.v1, vnode.attrs.v2);
  },

  loadData: async (functionId, v1, v2) => {
    VersionDiff.loading = true;
    try {
      const [func, diffData] = await Promise.all([
        API.functions.get(functionId),
        API.versions.diff(functionId, v1, v2),
      ]);
      VersionDiff.func = func;
      VersionDiff.diffData = diffData;
    } catch (e) {
      console.error("Failed to load diff:", e);
    } finally {
      VersionDiff.loading = false;
      m.redraw();
    }
  },

  view: () => {
    if (VersionDiff.loading) {
      return m(".loading", "Loading...");
    }

    if (!VersionDiff.func || !VersionDiff.diffData) {
      return m(".container", m(".card", "Diff not found"));
    }

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", "Version Comparison"),
            m(
              ".page-subtitle",
              `${VersionDiff.func.name} - v${VersionDiff.diffData.old_version} → v${VersionDiff.diffData.new_version}`,
            ),
          ]),
          m("a.btn", { href: `#!/functions/${VersionDiff.func.id}` }, [
            Icons.arrowLeft(),
            "  Back",
          ]),
        ]),
      ]),

      m(".card.mb-24", [
        m(".card-header", [
          m(".card-title", "Code Changes"),
          m("div", { style: "display: flex; gap: 12px; font-size: 13px;" }, [
            m(
              "span",
              { style: "color: #ef4444;" },
              `- Version ${VersionDiff.diffData.old_version}`,
            ),
            m(
              "span",
              { style: "color: #86efac;" },
              `+ Version ${VersionDiff.diffData.new_version}`,
            ),
          ]),
        ]),
        m(
          "div",
          {
            style:
              "max-height: 600px; overflow-y: auto; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace; font-size: 13px; background: #0a0a0a;",
          },
          m(
            "table",
            { style: "margin: 0; width: 100%; border-collapse: collapse;" },
            [
              m(
                "tbody",
                VersionDiff.diffData.diff.map((line, idx) =>
                  m(
                    "tr",
                    {
                      key: idx,
                      style: `background: ${
                        line.line_type === "added"
                          ? "#14532d40"
                          : line.line_type === "removed"
                            ? "#7f1d1d40"
                            : "transparent"
                      };`,
                    },
                    [
                      m(
                        "td",
                        {
                          style:
                            "width: 40px; padding: 2px 8px; text-align: right; color: #666; border-right: 1px solid #262626; user-select: none; background: #171717;",
                        },
                        line.old_line || "",
                      ),
                      m(
                        "td",
                        {
                          style:
                            "width: 40px; padding: 2px 8px; text-align: right; color: #666; border-right: 1px solid #262626; user-select: none; background: #171717;",
                        },
                        line.new_line || "",
                      ),
                      m(
                        "td",
                        {
                          style:
                            "width: 20px; padding: 2px 8px; text-align: center; border-right: 1px solid #262626; user-select: none; font-weight: bold; color: " +
                            (line.line_type === "added"
                              ? "#86efac"
                              : line.line_type === "removed"
                                ? "#ef4444"
                                : "#666") +
                            ";",
                        },
                        line.line_type === "added"
                          ? "+"
                          : line.line_type === "removed"
                            ? "-"
                            : " ",
                      ),
                      m(
                        "td",
                        {
                          style:
                            "padding: 2px 12px; white-space: pre-wrap; word-break: break-all; color: " +
                            (line.line_type === "added"
                              ? "#86efac"
                              : line.line_type === "removed"
                                ? "#fca5a5"
                                : "#e5e5e5") +
                            ";",
                        },
                        line.content || " ",
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ]),
    ]);
  },
};
