/**
 * @fileoverview Toast notification system for displaying temporary messages.
 */

import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').ToastMessage} ToastMessage
 */

/**
 * Toast notification component and manager.
 * Provides static methods for showing notifications and renders the toast container.
 *
 * @example
 * // Show a success notification
 * Toast.show("Changes saved!", "success");
 *
 * // Show an error notification
 * Toast.show("Failed to save", "error", 5000);
 */
export const Toast = {
  /**
   * Array of active toast messages.
   * @type {ToastMessage[]}
   */
  messages: [],

  /**
   * Counter for generating unique message IDs.
   * @type {number}
   */
  nextId: 0,

  /**
   * Shows a new toast notification.
   * @param {string} message - The message to display
   * @param {('success'|'error'|'warning'|'info')} [type='success'] - Toast type/color
   * @param {number} [duration=3000] - Duration in milliseconds before auto-dismiss
   */
  show: (message, type = "success", duration = 3000) => {
    const id = Toast.nextId++;
    Toast.messages.push({ id, message, type });
    m.redraw();

    setTimeout(() => {
      Toast.remove(id);
    }, duration);
  },

  /**
   * Removes a toast notification by ID.
   * @param {number} id - The toast ID to remove
   */
  remove: (id) => {
    const index = Toast.messages.findIndex((msg) => msg.id === id);
    if (index !== -1) {
      Toast.messages.splice(index, 1);
      m.redraw();
    }
  },

  /**
   * Renders the toast container with all active notifications.
   * @returns {Object|null} Mithril vnode or null if no messages
   */
  view: () => {
    if (Toast.messages.length === 0) return null;

    return m(
      ".toast-container",
      Toast.messages.map((msg) =>
        m(
          ".toast",
          {
            key: msg.id,
            class: `toast--${msg.type}`,
          },
          [
            m("span", msg.message),
            m(
              "button.toast__close",
              {
                onclick: () => Toast.remove(msg.id),
                "aria-label": t("toast.closeNotification"),
              },
              m.trust(icons.xMark()),
            ),
          ],
        )
      ),
    );
  },
};
