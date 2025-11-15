import { Icons } from "../icons.js";
import { API } from "../api.js";
import { Pagination } from "../components/pagination.js";

export const FunctionsList = {
  functions: [],
  loading: true,
  limit: 20,
  offset: 0,
  total: 0,

  oninit: () => {
    FunctionsList.loadFunctions();
  },

  loadFunctions: async () => {
    FunctionsList.loading = true;
    try {
      const response = await API.functions.list(
        FunctionsList.limit,
        FunctionsList.offset,
      );
      FunctionsList.functions = response.functions || [];
      FunctionsList.total = response.pagination?.total || 0;
    } catch (e) {
      console.error("Failed to load functions:", e);
    } finally {
      FunctionsList.loading = false;
      m.redraw();
    }
  },

  handlePageChange: (newOffset) => {
    FunctionsList.offset = newOffset;
    FunctionsList.loadFunctions();
  },

  handleLimitChange: (newLimit) => {
    FunctionsList.limit = newLimit;
    FunctionsList.offset = 0;
    FunctionsList.loadFunctions();
  },

  view: () => {
    if (FunctionsList.loading) {
      return m(".loading", "Loading functions...");
    }

    return m(".container", [
      m(".page-header", [
        m(".page-title", [
          m("div", [
            m("h1", "Functions"),
            m(".page-subtitle", "Manage your serverless functions"),
          ]),
          m(
            "a.btn.btn-primary",
            {
              href: "#!/functions/new",
            },
            [Icons.plus(), "  New Function"],
          ),
        ]),
      ]),

      m(".card", [
        m(".card-header", [
          m(".card-title", "All Functions"),
          m(".card-subtitle", `${FunctionsList.total} functions total`),
        ]),

        FunctionsList.functions.length === 0
          ? m(
              ".text-center.mt-24.mb-24",
              "No functions yet. Create your first function to get started.",
            )
          : [
              m("table", [
                m(
                  "thead",
                  m("tr", [
                    m("th", "Name"),
                    m("th", "Description"),
                    m("th", "Active Version"),
                    m("th.th-actions", "Actions"),
                  ]),
                ),
                m(
                  "tbody",
                  FunctionsList.functions.map((func) =>
                    m("tr", { key: func.id }, [
                      m("td", func.name),
                      m("td", func.description || "No description"),
                      m(
                        "td",
                        m(
                          ".badge.badge-success",
                          `v${func.active_version.version}`,
                        ),
                      ),
                      m(
                        "td.td-actions",
                        m(".actions", [
                          m(
                            "a.btn.btn-icon",
                            {
                              href: `#!/functions/${func.id}`,
                              title: "View",
                            },
                            Icons.eye(),
                          ),
                          m(
                            "a.btn.btn-icon",
                            {
                              href: `#!/functions/${func.id}/edit`,
                              title: "Edit",
                            },
                            Icons.pencil(),
                          ),
                          m(
                            "a.btn.btn-icon",
                            {
                              href: `#!/functions/${func.id}?tab=test`,
                              title: "Test",
                            },
                            Icons.play(),
                          ),
                        ]),
                      ),
                    ]),
                  ),
                ),
              ]),
              m(Pagination, {
                total: FunctionsList.total,
                limit: FunctionsList.limit,
                offset: FunctionsList.offset,
                onPageChange: FunctionsList.handlePageChange,
                onLimitChange: FunctionsList.handleLimitChange,
              }),
            ],
      ]),
    ]);
  },
};
