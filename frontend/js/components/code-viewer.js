/**
 * @fileoverview Simple read-only code viewer component with syntax highlighting.
 */

/**
 * Code viewer component for displaying code with syntax highlighting.
 * Uses highlight.js for syntax highlighting.
 * @type {Object}
 */
export const CodeViewer = {
  /**
   * Renders the code viewer component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.code=''] - Code to display
   * @param {string} [vnode.attrs.language=''] - Language for syntax highlighting
   * @param {string} [vnode.attrs.maxHeight=''] - Maximum height with overflow scroll
   * @param {boolean} [vnode.attrs.noBorder=false] - Remove border styling
   * @param {boolean} [vnode.attrs.padded=false] - Add padding to code block
   * @param {boolean} [vnode.attrs.showHeader=false] - Show language header
   * @param {boolean} [vnode.attrs.wrap=false] - Enable word wrapping
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      code = "",
      language = "",
      maxHeight = "",
      noBorder = false,
      padded = false,
      showHeader = false,
      wrap = false,
    } = vnode.attrs;

    return m(
      ".code-viewer",
      {
        class: noBorder ? "code-viewer--no-border" : "",
      },
      [
        showHeader &&
        language &&
        m(".code-viewer__header", [
          m("span.code-viewer__language", language.toUpperCase()),
        ]),
        m(
          ".code-viewer__content",
          {
            style: maxHeight ? `max-height: ${maxHeight}` : "",
          },
          [
            m(
              "pre.code-viewer__pre",
              {
                class: [
                  padded ? "code-viewer__pre--padded" : "",
                  wrap ? "code-viewer__pre--wrap" : "",
                ].filter(Boolean).join(" "),
              },
              [
                m(
                  "code",
                  {
                    class: language ? `language-${language}` : "",
                    oncreate: (vnode) => {
                      if (window.hljs) {
                        hljs.highlightElement(vnode.dom);
                      }
                    },
                    onupdate: (vnode) => {
                      if (window.hljs) {
                        vnode.dom.removeAttribute("data-highlighted");
                        hljs.highlightElement(vnode.dom);
                      }
                    },
                  },
                  code,
                ),
              ],
            ),
          ],
        ),
      ],
    );
  },
};
