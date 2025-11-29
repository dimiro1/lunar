/**
 * @fileoverview Utility functions for the lunar Dashboard.
 */

import { routes } from "./routes.js";
import { i18n, t } from "./i18n/index.js";

/**
 * @typedef {import('./types.js').TabItem} TabItem
 */

/**
 * Generates tab navigation items for function detail pages.
 * @param {string} funcId - The function ID
 * @returns {TabItem[]} Array of tab configuration objects
 * @example
 * const tabs = getFunctionTabs("abc123");
 * // Returns tabs for Code, Versions, Executions, Settings, Test
 */
export const getFunctionTabs = (funcId) => [
  { id: "code", label: t("tabs.code"), href: routes.functionCode(funcId) },
  {
    id: "versions",
    label: t("tabs.versions"),
    href: routes.functionVersions(funcId),
  },
  {
    id: "executions",
    label: t("tabs.executions"),
    href: routes.functionExecutions(funcId),
  },
  {
    id: "settings",
    label: t("tabs.settings"),
    href: routes.functionSettings(funcId),
  },
  { id: "test", label: t("tabs.test"), href: routes.functionTest(funcId) },
];

/**
 * Format options for timestamp display.
 * @typedef {'time'|'date'|'datetime'} TimestampFormat
 */

/**
 * Formats a Unix timestamp (seconds) into a human-readable string.
 * @param {number} timestamp - Unix timestamp in seconds
 * @param {TimestampFormat} [format='datetime'] - Output format
 * @returns {string} Formatted date/time string or 'N/A' / 'Invalid Date'
 * @example
 * formatUnixTimestamp(1700000000, 'datetime'); // "11/14/2023, 2:13:20 PM"
 * formatUnixTimestamp(1700000000, 'date');     // "11/14/2023"
 * formatUnixTimestamp(1700000000, 'time');     // "2:13:20 PM"
 * formatUnixTimestamp(0);                       // "N/A"
 */
export const formatUnixTimestamp = (timestamp, format = "datetime") => {
  if (!timestamp || timestamp === 0) {
    return t("common.na");
  }

  const date = new Date(timestamp * 1000);

  if (isNaN(date.getTime())) {
    return t("common.invalidDate");
  }

  const locale = i18n.getLocale();

  switch (format) {
    case "time":
      return date.toLocaleTimeString(locale);
    case "date":
      return date.toLocaleDateString(locale);
    case "datetime":
    default:
      return date.toLocaleString(locale);
  }
};
