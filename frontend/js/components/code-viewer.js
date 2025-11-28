/**
 * @fileoverview Simple read-only code viewer component with syntax highlighting.
 */

/**
 * Highlights code using highlight.js programmatic API.
 * This avoids the security warning from highlightElement() which uses innerHTML.
 * @param {string} code - Code to highlight
 * @param {string} language - Language for highlighting
 * @returns {{html: string, highlighted: boolean}} Highlighted HTML string and whether highlighting was applied
 */
function highlightCode(code, language) {
  if (!code || !window.hljs) {
    return { html: "", highlighted: false };
  }

  try {
    if (language && hljs.getLanguage(language)) {
      return {
        html: hljs.highlight(code, { language }).value,
        highlighted: true,
      };
    }
    return { html: hljs.highlightAuto(code).value, highlighted: true };
  } catch (e) {
    console.warn("highlight.js error:", e);
    return { html: "", highlighted: false };
  }
}

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
   * @param {string} [vnode.attrs.title=''] - Custom header title (overrides language display)
   * @param {boolean} [vnode.attrs.wrap=false] - Enable word wrapping
   * @param {*} vnode.children - Child elements to render in header (e.g., action buttons)
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
      title = "",
      wrap = false,
    } = vnode.attrs;

    const hasHeader = showHeader || title;
    const headerTitle = title || (language ? language.toUpperCase() : "");
    const { html, highlighted } = highlightCode(code, language);

    return m(
      ".code-viewer",
      {
        class: noBorder ? "code-viewer--no-border" : "",
      },
      [
        hasHeader &&
        headerTitle &&
        m(".code-viewer__header", [
          m("span.code-viewer__language", headerTitle),
          vnode.children,
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
                    class: [
                      language ? `language-${language}` : "",
                      highlighted ? "hljs" : "",
                    ].filter(Boolean).join(" "),
                  },
                  highlighted ? m.trust(html) : code,
                ),
              ],
            ),
          ],
        ),
      ],
    );
  },
};
