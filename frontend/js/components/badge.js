/**
 * @fileoverview Badge components for displaying labels, statuses, and tags.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * Available badge color variants.
 * @enum {string}
 */
export const BadgeVariant = {
  PRIMARY: "primary",
  SECONDARY: "secondary",
  DESTRUCTIVE: "destructive",
  OUTLINE: "outline",
  SUCCESS: "success",
  WARNING: "warning",
};

/**
 * Available badge sizes.
 * @enum {string}
 */
export const BadgeSize = {
  SM: "sm",
  DEFAULT: "default",
  LG: "lg",
};

/**
 * Badge component for displaying labels and tags.
 * @type {Object}
 */
export const Badge = {
  /**
   * Renders the badge component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.variant='primary'] - Color variant from BadgeVariant
   * @param {string} [vnode.attrs.size='default'] - Size from BadgeSize
   * @param {boolean} [vnode.attrs.uppercase=false] - Uppercase text
   * @param {boolean} [vnode.attrs.mono=false] - Monospace font
   * @param {boolean} [vnode.attrs.dot=false] - Show status dot
   * @param {boolean} [vnode.attrs.dotGlow=false] - Add glow effect to dot
   * @param {IconName} [vnode.attrs.icon] - Icon name for left icon
   * @param {IconName} [vnode.attrs.iconRight] - Icon name for right icon
   * @param {string} [vnode.attrs.href] - If provided, renders as anchor
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      variant = BadgeVariant.PRIMARY,
      size = BadgeSize.DEFAULT,
      uppercase = false,
      mono = false,
      dot = false,
      dotGlow = false,
      icon,
      iconRight,
      href,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "badge",
      `badge--${size}`,
      `badge--${variant}`,
      uppercase && "badge--uppercase",
      mono && "badge--mono",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    const content = [
      dot &&
      m("span", {
        class: `badge__dot ${dotGlow ? "badge__dot--glow" : ""}`.trim(),
      }),
      icon && m("span.badge__icon", m.trust(icons[icon]())),
      vnode.children,
      iconRight && m("span.badge__icon", m.trust(icons[iconRight]())),
    ];

    if (href) {
      return m("a", { href, class: classes, ...attrs }, content);
    }

    return m("span", { class: classes, ...attrs }, content);
  },
};

/**
 * ID Badge component - displays an ID with a hashtag icon.
 * @type {Object}
 */
export const IDBadge = {
  /**
   * Renders the ID badge.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.id - The ID to display
   * @param {string} [vnode.attrs.href] - Optional link URL
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { id, href } = vnode.attrs;

    return m(
      Badge,
      {
        variant: href ? BadgeVariant.OUTLINE : BadgeVariant.SECONDARY,
        size: BadgeSize.SM,
        icon: "hashtag",
        mono: true,
        href,
      },
      id,
    );
  },
};

/**
 * Status Badge component - displays enabled/disabled status.
 * @type {Object}
 */
export const StatusBadge = {
  /**
   * Renders the status badge.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {boolean} vnode.attrs.enabled - Whether the status is enabled
   * @param {boolean} [vnode.attrs.glow=false] - Add glow effect when enabled
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { enabled, glow = false } = vnode.attrs;

    return m(
      Badge,
      {
        variant: enabled ? BadgeVariant.SUCCESS : BadgeVariant.WARNING,
        size: glow ? BadgeSize.DEFAULT : BadgeSize.SM,
        dot: true,
        dotGlow: glow && enabled,
        uppercase: true,
        mono: true,
      },
      enabled ? t("badge.enabled") : t("badge.disabled"),
    );
  },
};

/**
 * Method Badges component - displays a list of HTTP methods.
 * @type {Object}
 */
export const MethodBadges = {
  /**
   * Renders method badges.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string[]} [vnode.attrs.methods=[]] - Array of HTTP method names
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { methods = [] } = vnode.attrs;

    return m(
      ".badge__methods",
      methods.map((method) =>
        m(
          Badge,
          {
            variant: BadgeVariant.SECONDARY,
            size: BadgeSize.SM,
            mono: true,
          },
          method,
        )
      ),
    );
  },
};

/**
 * Log Level Badge - for execution logs.
 * @type {Object}
 */
export const LogLevelBadge = {
  /**
   * Renders the log level badge.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {('INFO'|'WARN'|'ERROR'|'DEBUG')} vnode.attrs.level - Log level
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { level } = vnode.attrs;

    /** @type {Object.<string, string>} */
    const variantMap = {
      INFO: BadgeVariant.SUCCESS,
      WARN: BadgeVariant.WARNING,
      ERROR: BadgeVariant.DESTRUCTIVE,
      DEBUG: BadgeVariant.SECONDARY,
    };

    return m(
      Badge,
      {
        variant: variantMap[level] || BadgeVariant.SECONDARY,
        size: BadgeSize.SM,
        uppercase: true,
        mono: true,
      },
      level,
    );
  },
};
