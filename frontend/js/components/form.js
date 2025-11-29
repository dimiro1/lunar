/**
 * @fileoverview Form components for user input.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').IconName} IconName
 */

/**
 * Form Group component - wrapper for form fields.
 * @type {Object}
 */
export const FormGroup = {
  /**
   * Renders the form group wrapper.
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
        class: `form-group ${className}`.trim(),
        ...attrs,
      },
      vnode.children,
    );
  },
};

/**
 * Form Label component.
 * @type {Object}
 */
export const FormLabel = {
  /**
   * Renders the form label.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.for] - ID of the input this label is for
   * @param {string} [vnode.attrs.text] - Label text
   * @param {boolean} [vnode.attrs.required=false] - Show required asterisk
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled styling
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      for: htmlFor,
      text,
      required = false,
      disabled = false,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "form-label",
      disabled && "form-label--disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m(
      "label",
      {
        for: htmlFor,
        class: classes,
        ...attrs,
      },
      [
        text,
        required &&
        m("span.form-label__required", { "aria-hidden": "true" }, "*"),
      ],
    );
  },
};

/**
 * Form Input component.
 * @type {Object}
 */
export const FormInput = {
  /**
   * Renders the form input.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.type='text'] - Input type
   * @param {string} [vnode.attrs.placeholder] - Placeholder text
   * @param {string} [vnode.attrs.value] - Input value
   * @param {string} [vnode.attrs.name] - Input name
   * @param {string} [vnode.attrs.id] - Input ID
   * @param {boolean} [vnode.attrs.mono=false] - Monospace font
   * @param {boolean} [vnode.attrs.error=false] - Error styling
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.readonly=false] - Read-only state
   * @param {boolean} [vnode.attrs.required=false] - Required field
   * @param {IconName} [vnode.attrs.icon] - Icon name to show on left
   * @param {(e: Event) => void} [vnode.attrs.oninput] - Input handler
   * @param {(e: Event) => void} [vnode.attrs.onchange] - Change handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      type = "text",
      placeholder,
      value,
      name,
      id,
      mono = false,
      error = false,
      disabled = false,
      readonly = false,
      required = false,
      icon,
      oninput,
      onchange,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const inputClasses = [
      "form-input",
      mono && "form-input--mono",
      error && "form-input--error",
      disabled && "form-input--disabled",
      icon && "form-input--with-icon",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    const inputElement = m("input", {
      type,
      id,
      name,
      placeholder,
      value,
      class: inputClasses,
      disabled,
      readonly,
      required,
      oninput,
      onchange,
      ...attrs,
    });

    if (icon) {
      return m(".form-input-wrapper", [
        m(
          "span.form-input__icon",
          { "aria-hidden": "true" },
          m.trust(icons[icon]()),
        ),
        inputElement,
      ]);
    }

    return inputElement;
  },
};

/**
 * Password Input with visibility toggle.
 * @type {Object}
 */
export const PasswordInput = {
  /**
   * Initializes component state.
   * @param {Object} vnode - Mithril vnode
   */
  oninit(vnode) {
    vnode.state.visible = false;
  },

  /**
   * Renders the password input.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.placeholder] - Placeholder text
   * @param {string} [vnode.attrs.value] - Input value
   * @param {string} [vnode.attrs.name] - Input name
   * @param {string} [vnode.attrs.id] - Input ID
   * @param {boolean} [vnode.attrs.mono=false] - Monospace font
   * @param {boolean} [vnode.attrs.error=false] - Error styling
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.required=false] - Required field
   * @param {(e: Event) => void} [vnode.attrs.oninput] - Input handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {state} vnode.state - Component state { visible: boolean }
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      placeholder,
      value,
      name,
      id,
      mono = false,
      error = false,
      disabled = false,
      required = false,
      oninput,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const inputClasses = [
      "form-input",
      "form-input--password",
      mono && "form-input--mono",
      error && "form-input--error",
      disabled && "form-input--disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m(".form-password-wrapper", [
      m("input", {
        type: vnode.state.visible ? "text" : "password",
        id,
        name,
        placeholder,
        value,
        class: inputClasses,
        disabled,
        required,
        oninput,
        ...attrs,
      }),
      m(
        "button.form-password-toggle",
        {
          type: "button",
          title: vnode.state.visible
            ? t("form.hidePassword")
            : t("form.showPassword"),
          onclick: () => {
            vnode.state.visible = !vnode.state.visible;
          },
        },
        [
          vnode.state.visible
            ? m.trust(icons.eyeSlash())
            : m.trust(icons.eye()),
        ],
      ),
    ]);
  },
};

/**
 * Copy Input with the copy-to-clipboard button.
 * @type {Object}
 */
export const CopyInput = {
  /**
   * Initializes component state.
   * @param {Object} vnode - Mithril vnode
   */
  oninit(vnode) {
    vnode.state.copied = false;
  },

  /**
   * Renders the copy input.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.value - Value to copy
   * @param {string} [vnode.attrs.name] - Input name
   * @param {string} [vnode.attrs.id] - Input ID
   * @param {boolean} [vnode.attrs.mono=true] - Monospace font
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.readonly=true] - Read-only state
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {state} vnode.state - Component state { copied: boolean }
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      value,
      name,
      id,
      mono = true,
      disabled = false,
      readonly = true,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const inputClasses = [
      "form-input",
      "form-input--copy",
      mono && "form-input--mono",
      disabled && "form-input--disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    /**
     * Copies the value to clipboard.
     * @returns {Promise<void>}
     */
    const handleCopy = async () => {
      try {
        await navigator.clipboard.writeText(value);
        vnode.state.copied = true;
        setTimeout(() => {
          vnode.state.copied = false;
          m.redraw();
        }, 2000);
      } catch (err) {
        console.error("Failed to copy:", err);
      }
    };

    return m(".form-copy-wrapper", [
      m("input", {
        type: "text",
        id,
        name,
        value,
        class: inputClasses,
        disabled,
        readonly,
        ...attrs,
      }),
      m(
        "button.form-copy-button",
        {
          type: "button",
          title: vnode.state.copied
            ? t("form.copied")
            : t("form.copyToClipboard"),
          "aria-label": t("form.copyToClipboard"),
          onclick: handleCopy,
        },
        [
          vnode.state.copied
            ? m(
              "span",
              { style: "color: var(--color-success)" },
              m.trust(icons.check()),
            )
            : m.trust(icons.copy()),
        ],
      ),
    ]);
  },
};

/**
 * Form Textarea component.
 * @type {Object}
 */
export const FormTextarea = {
  /**
   * Renders the form textarea.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.placeholder] - Placeholder text
   * @param {string} [vnode.attrs.value] - Textarea value
   * @param {string} [vnode.attrs.name] - Textarea name
   * @param {string} [vnode.attrs.id] - Textarea ID
   * @param {number} [vnode.attrs.rows] - Number of visible rows
   * @param {boolean} [vnode.attrs.error=false] - Error styling
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.readonly=false] - Read-only state
   * @param {boolean} [vnode.attrs.required=false] - Required field
   * @param {(e: Event) => void} [vnode.attrs.oninput] - Input handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      placeholder,
      value,
      name,
      id,
      rows,
      error = false,
      disabled = false,
      readonly = false,
      required = false,
      oninput,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = ["form-textarea", error && "form-input--error", className]
      .filter(Boolean)
      .join(" ");

    return m(
      "textarea",
      {
        id,
        name,
        placeholder,
        rows,
        class: classes,
        disabled,
        readonly,
        required,
        oninput,
        ...attrs,
      },
      value,
    );
  },
};

/**
 * @typedef {Object} SelectOption
 * @property {string} value - Option value
 * @property {string} label - Option display label
 */

/**
 * Form Select component.
 * @type {Object}
 */
export const FormSelect = {
  /**
   * Renders the form select.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {(string|SelectOption)[]} [vnode.attrs.options=[]] - Select options
   * @param {string} [vnode.attrs.selected] - Currently selected value
   * @param {string} [vnode.attrs.name] - Select name
   * @param {string} [vnode.attrs.id] - Select ID
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.required=false] - Required field
   * @param {(e: Event) => void} [vnode.attrs.onchange] - Change handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      options = [],
      selected,
      name,
      id,
      disabled = false,
      required = false,
      onchange,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    return m(
      "select",
      {
        id,
        name,
        class: `form-select ${className}`.trim(),
        disabled,
        required,
        onchange,
        ...attrs,
      },
      options.map((opt) => {
        const value = typeof opt === "object" ? opt.value : opt;
        const label = typeof opt === "object" ? opt.label : opt;
        return m(
          "option",
          {
            value,
            selected: value === selected,
          },
          label,
        );
      }),
    );
  },
};

/**
 * Checkbox component.
 * @type {Object}
 */
export const FormCheckbox = {
  /**
   * Renders the form checkbox.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.label] - Checkbox label text
   * @param {string} [vnode.attrs.description] - Additional description text
   * @param {boolean} [vnode.attrs.checked=false] - Checked state
   * @param {string} [vnode.attrs.name] - Checkbox name
   * @param {string} [vnode.attrs.id] - Checkbox ID
   * @param {string} [vnode.attrs.value] - Checkbox value
   * @param {boolean} [vnode.attrs.disabled=false] - Disabled state
   * @param {boolean} [vnode.attrs.required=false] - Required field
   * @param {(e: Event) => void} [vnode.attrs.onchange] - Change handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      label,
      description,
      checked = false,
      name,
      id,
      value,
      disabled = false,
      required = false,
      onchange,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "form-checkbox",
      disabled && "form-checkbox--disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m("label", { class: classes }, [
      m("input.form-checkbox__input", {
        type: "checkbox",
        id,
        name,
        value,
        checked,
        disabled,
        required,
        onchange,
        ...attrs,
      }),
      label && m("span.form-checkbox__label", label),
      description && m("span.form-checkbox__description", description),
    ]);
  },
};

/**
 * Help Text component.
 * @type {Object}
 */
export const FormHelp = {
  /**
   * Renders the form help text.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.text] - Help text content
   * @param {boolean} [vnode.attrs.error=false] - Error styling
   * @param {boolean} [vnode.attrs.success=false] - Success styling
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (alternative to text)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      text,
      error = false,
      success = false,
      class: className = "",
      ...attrs
    } = vnode.attrs;

    const classes = [
      "form-help",
      error && "form-help--error",
      success && "form-help--success",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m("p", { class: classes, ...attrs }, text || vnode.children);
  },
};
