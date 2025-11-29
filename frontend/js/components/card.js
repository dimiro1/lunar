/**
 * @fileoverview Card components for content containers.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * Available card color variants.
 * @enum {string}
 */
export const CardVariant = {
  DEFAULT: "default",
  DANGER: "danger",
  SUCCESS: "success",
  WARNING: "warning",
  INFO: "info",
};

/**
 * Card container component.
 * @type {Object}
 */
export const Card = {
  /**
   * Renders the card container.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.variant='default'] - Color variant from CardVariant
   * @param {boolean} [vnode.attrs.padded=false] - Add padding to card body
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      variant = CardVariant.DEFAULT,
      padded = false,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "card",
      variant !== CardVariant.DEFAULT && `card--${variant}`,
      padded && "card--padded",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m("div", { class: classes, ...attrs }, vnode.children);
  },
};

/**
 * Card Header component with title and optional subtitle.
 * @type {Object}
 */
export const CardHeader = {
  /**
   * Renders the card header.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.title - Header title text
   * @param {string} [vnode.attrs.subtitle] - Optional subtitle text
   * @param {IconName} [vnode.attrs.icon] - Optional icon name
   * @param {string} [vnode.attrs.variant='default'] - Color variant from CardVariant
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (e.g., action buttons)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      title,
      subtitle,
      icon,
      variant = CardVariant.DEFAULT,
      class: className = "",
      actions = [],
      ...attrs
    } = vnode.attrs;

    const classes = [
      "card__header",
      variant === CardVariant.DANGER && "card__header--danger",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    const titleClasses = [
      "card__title",
      variant === CardVariant.DANGER && "card__title--danger",
    ]
      .filter(Boolean)
      .join(" ");

    const iconClasses = [
      "card__header-icon",
      variant === CardVariant.DANGER && "card__header-icon--danger",
    ]
      .filter(Boolean)
      .join(" ");

    // Use actions prop if provided, otherwise fall back to children
    const actionsContent = actions.length > 0 ? actions : vnode.children;

    return m("div", { class: classes, ...attrs }, [
      m(".card__header-title-wrapper", [
        icon && m("span", { class: iconClasses }, m.trust(icons[icon]())),
        subtitle
          ? m(".card__header-title-group", [
            m("h3", { class: titleClasses }, title),
            m("p.card__subtitle", subtitle),
          ])
          : m("h3", { class: titleClasses }, title),
      ]),
      actionsContent &&
      m(".card__header-actions", actionsContent),
    ]);
  },
};

/**
 * Card Content component for the main body.
 * @type {Object}
 */
export const CardContent = {
  /**
   * Renders the card content.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} [vnode.attrs.dark=false] - Dark background
   * @param {boolean} [vnode.attrs.large=false] - Large padding
   * @param {boolean} [vnode.attrs.noPadding=false] - Remove padding
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      dark = false,
      large = false,
      noPadding = false,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "card__content",
      large && "card__content--large",
      dark && "card__content--dark",
      noPadding && "card__content--no-padding",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m("div", { class: classes, ...attrs }, vnode.children);
  },
};

/**
 * Card Footer component.
 * @type {Object}
 */
export const CardFooter = {
  /**
   * Renders the card footer.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "", ...attrs } = vnode.attrs;

    return m(
      "div",
      {
        class: `card__footer ${className}`.trim(),
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Card Divider component - horizontal line separator.
 * @type {Object}
 */
export const CardDivider = {
  /**
   * Renders the card divider.
   * @returns {Object} Mithril vnode
   */
  view() {
    return m("hr.card__divider");
  },
};

/**
 * Maximize button component for card headers.
 * @type {Object}
 */
export const CardMaximizeBtn = {
  /**
   * Renders the maximize button.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {() => void} vnode.attrs.onclick - Click handler to trigger maximize
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    return m(
      "button.card-maximize-btn",
      {
        onclick: vnode.attrs.onclick,
        title: t("card.maximize"),
      },
      m.trust(icons.arrowsPointingOut()),
    );
  },
};

/**
 * Maximizable Card component that can expand to fullscreen overlay.
 * @type {Object}
 */
export const MaximizableCard = {
  /**
   * Renders the maximizable card.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.title - Card header title
   * @param {IconName} [vnode.attrs.icon] - Optional header icon
   * @param {string} [vnode.attrs.variant='default'] - Color variant from CardVariant
   * @param {Array} [vnode.attrs.headerActions=[]] - Additional header action elements
   * @param {string} [vnode.attrs.class] - Additional CSS classes for the card
   * @param {*} vnode.children - Child elements to render in card content
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      title,
      icon,
      variant = CardVariant.DEFAULT,
      headerActions = [],
      class: className = "",
      ...attrs
    } = vnode.attrs;

    // Initialize state
    if (vnode.state.isMaximized === undefined) {
      vnode.state.isMaximized = false;
      vnode.state.escapeHandler = null;
    }

    /**
     * Removes the global Escape key listener.
     */
    const removeEscapeListener = () => {
      if (vnode.state.escapeHandler) {
        document.removeEventListener("keydown", vnode.state.escapeHandler);
        vnode.state.escapeHandler = null;
      }
    };

    /**
     * Closes the maximized view.
     */
    const closeMaximized = () => {
      removeEscapeListener();
      vnode.state.isMaximized = false;
      m.redraw();
    };

    /**
     * Sets up the global Escape key listener when maximized.
     */
    const setupEscapeListener = () => {
      removeEscapeListener();
      vnode.state.escapeHandler = (e) => {
        if (e.key === "Escape") {
          e.preventDefault();
          e.stopPropagation();
          closeMaximized();
        }
      };
      document.addEventListener("keydown", vnode.state.escapeHandler);
    };

    /**
     * Toggles the maximized state.
     */
    const toggleMaximize = () => {
      vnode.state.isMaximized = !vnode.state.isMaximized;
      if (vnode.state.isMaximized) {
        setupEscapeListener();
      }
      m.redraw();
    };

    /**
     * Handles click events on the overlay background.
     * @param {MouseEvent} e - Mouse event
     */
    const handleOverlayClick = (e) => {
      if (e.target.classList.contains("card-maximized-overlay")) {
        closeMaximized();
      }
    };

    // Build header actions with maximize button
    const allActions = [
      ...headerActions,
      m(CardMaximizeBtn, { onclick: toggleMaximize }),
    ];

    // Render maximized overlay if active
    if (vnode.state.isMaximized) {
      return m(
        ".card-maximized-overlay",
        {
          onclick: handleOverlayClick,
          onremove: removeEscapeListener,
        },
        m(".card-maximized", [
          m(".card__header", [
            m(".card__header-title-wrapper", [
              icon &&
              m(
                "span.card__header-icon",
                m.trust(icons[icon]()),
              ),
              m("h3.card__title", title),
            ]),
            m(
              "button.card-maximized__close",
              {
                onclick: closeMaximized,
                title: t("card.minimize"),
              },
              m.trust(icons.arrowsPointingIn()),
            ),
          ]),
          m(".card-maximized__content", vnode.children),
        ]),
      );
    }

    // Render normal card
    return m(Card, { variant, class: className, ...attrs }, [
      m(CardHeader, {
        title,
        icon,
        variant,
        actions: allActions,
      }),
      m(CardContent, { noPadding: true }, vnode.children),
    ]);
  },
};
