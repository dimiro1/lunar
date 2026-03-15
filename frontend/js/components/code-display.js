/**
 * @fileoverview Code display component for prominent verification codes.
 */

/**
 * CodeDisplay component - renders a large, styled code value.
 * Designed for displaying verification codes, tokens, or similar short values.
 * @type {Object}
 * @example
 * m(CodeDisplay, { value: "ABCD-1234" })
 * m(CodeDisplay, { value: "ABCD-1234", label: "Verification Code" })
 */
export const CodeDisplay = {
  view(vnode) {
    const { value, label } = vnode.attrs;

    return m(".code-display", [
      label && m(".code-display__label", label),
      m(".code-display__value", value),
    ]);
  },
};
