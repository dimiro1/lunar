/**
 * @fileoverview Button components.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * Available button color variants.
 * @enum {string}
 */
export const ButtonVariant = {
  PRIMARY: "primary",
  DESTRUCTIVE: "destructive",
  OUTLINE: "outline",
  SECONDARY: "secondary",
  GHOST: "ghost",
  LINK: "link",
};

/**
 * Available button sizes.
 * @enum {string}
 */
export const ButtonSize = {
  DEFAULT: "default",
  SM: "sm",
  LG: "lg",
  ICON: "icon",
};

/**
 * Button component - renders as button or anchor depending on props.
 * @type {Object}
 */
export const Button = {
  /**
   * Renders the button component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.variant='primary'] - Color variant from ButtonVariant
   * @param {string} [vnode.attrs.size='default'] - Size from ButtonSize
   * @param {boolean} [vnode.attrs.fullWidth=false] - Full width button
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.loading=false] - Loading state (shows spinner)
   * @param {string} [vnode.attrs.href] - If provided, renders as anchor
   * @param {string} [vnode.attrs.target] - Link target (e.g., "_blank")
   * @param {() => void} [vnode.attrs.onclick] - Click handler
   * @param {string} [vnode.attrs.type='button'] - Button type attribute
   * @param {IconName} [vnode.attrs.icon] - Icon name for left icon
   * @param {IconName} [vnode.attrs.iconRight] - Icon name for right icon
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {string} [vnode.attrs.ariaLabel] - Aria label for accessibility
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      variant = ButtonVariant.PRIMARY,
      size = ButtonSize.DEFAULT,
      fullWidth = false,
      disabled = false,
      loading = false,
      href,
      target,
      onclick,
      type = "button",
      icon,
      iconRight,
      class: className = "",
      ariaLabel,
      ...attrs
    } = vnode.attrs;

    const classes = [
      "btn",
      `btn--${size}`,
      `btn--${variant}`,
      fullWidth && "btn--full-width",
      (disabled || loading) && "btn--disabled",
      loading && "btn--loading",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    const content = [
      loading && m("span.btn__spinner", m.trust(icons.spinner())),
      !loading && icon && m.trust(icons[icon]()),
      vnode.children,
      !loading && iconRight && m.trust(icons[iconRight]()),
    ];

    if (href) {
      return m(
        "a",
        {
          href,
          target,
          class: classes,
          "aria-disabled": disabled || loading || undefined,
          "aria-busy": loading || undefined,
          "aria-label": ariaLabel,
          rel: target === "_blank" ? "noopener noreferrer" : undefined,
          onclick: disabled || loading ? (e) => e.preventDefault() : onclick,
          ...attrs,
        },
        content,
      );
    }

    return m(
      "button",
      {
        type,
        class: classes,
        disabled: disabled || loading,
        "aria-busy": loading || undefined,
        "aria-label": ariaLabel,
        onclick,
        ...attrs,
      },
      content,
    );
  },
};

/**
 * Back button component - styled link with back arrow.
 * @type {Object}
 */
export const BackButton = {
  /**
   * Renders the back button.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.href='#!/'] - Backlink URL
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { href = "#!/" } = vnode.attrs;

    return m("a.back-btn", { href }, [
      m.trust(icons.chevronLeft()),
      t("common.back"),
    ]);
  },
};
