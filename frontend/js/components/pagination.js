export const Pagination = {
  view: function (vnode) {
    const { total, limit, offset, onPageChange, onLimitChange } = vnode.attrs;

    const currentPage = Math.floor(offset / limit) + 1;
    const totalPages = Math.ceil(total / limit);
    const start = offset + 1;
    const end = Math.min(offset + limit, total);

    if (total === 0) {
      return m(".pagination", m(".pagination-info", "No items to display"));
    }

    const limitOptions = [10, 20, 50, 100];

    return m(".pagination", [
      m(".pagination-info", `Showing ${start}-${end} of ${total}`),

      m(".pagination-controls", [
        m(
          "button.btn.pagination-button",
          {
            disabled: currentPage === 1,
            onclick: () => onPageChange(offset - limit),
          },
          "Previous",
        ),

        m(
          ".pagination-pages",
          Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
            let pageNum;
            if (totalPages <= 7) {
              pageNum = i + 1;
            } else if (currentPage <= 4) {
              pageNum = i + 1;
            } else if (currentPage >= totalPages - 3) {
              pageNum = totalPages - 6 + i;
            } else {
              pageNum = currentPage - 3 + i;
            }

            return m(
              "button.btn.pagination-page",
              {
                class: pageNum === currentPage ? "active" : "",
                onclick: () => onPageChange((pageNum - 1) * limit),
              },
              pageNum,
            );
          }),
        ),

        m(
          "button.btn.pagination-button",
          {
            disabled: currentPage === totalPages,
            onclick: () => onPageChange(offset + limit),
          },
          "Next",
        ),
      ]),

      m(".pagination-limit", [
        m("label", "Items per page:"),
        m(
          "select.form-select",
          {
            value: limit,
            onchange: (e) => onLimitChange(parseInt(e.target.value)),
          },
          limitOptions.map((opt) => m("option", { value: opt }, opt)),
        ),
      ]),
    ]);
  },
};
