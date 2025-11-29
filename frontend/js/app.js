/**
 * @fileoverview Main application entry point.
 * Defines the Layout component and application routes.
 */

import { i18n, t } from "./i18n/index.js";
import { Toast } from "./components/toast.js";
import { Header } from "./components/navbar.js";

// Initialize i18n before rendering
i18n.init();
import { CommandPalette } from "./components/command-palette.js";
import { Login } from "./views/login.js";
import { FunctionsList } from "./views/functions-list.js";
import { FunctionCreate } from "./views/function-create.js";
import { FunctionCode } from "./views/function-code.js";
import { FunctionVersions } from "./views/function-versions.js";
import { FunctionExecutions } from "./views/function-executions.js";
import { FunctionSettings } from "./views/function-settings.js";
import { FunctionTest } from "./views/function-test.js";
import { ExecutionDetail } from "./views/execution-detail.js";
import { VersionDiff } from "./views/version-diff.js";
import { Preview } from "./views/preview.js";
import { API } from "./api.js";

/**
 * Layout component that wraps all authenticated pages.
 * Provides the header, main content area, toast notifications, and command palette.
 * @type {Object}
 */
const Layout = {
  /**
   * Handles user logout.
   * Calls the logout API and redirects to the login page.
   * @returns {Promise<void>}
   */
  handleLogout: async () => {
    try {
      await API.auth.logout();
      m.route.set("/login");
    } catch (e) {
      console.error("Logout failed:", e);
    }
  },

  /**
   * Renders the layout component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.breadcrumb] - Breadcrumb text for the current page
   * @param {string} [vnode.attrs.breadcrumbKey] - Translation key for breadcrumb
   * @param {Object} vnode.children - Child components to render in the main area
   * @returns {Object[]} Array of Mithril vnodes
   */
  view: (vnode) => {
    const breadcrumb = vnode.attrs.breadcrumbKey
      ? t(vnode.attrs.breadcrumbKey)
      : vnode.attrs.breadcrumb;

    return [
      m(Header, {
        breadcrumb,
        onLogout: Layout.handleLogout,
      }),
      m("main", vnode.children),
      m(Toast),
      m(CommandPalette),
    ];
  },
};

// Routes
m.route(document.getElementById("app"), "/functions", {
  "/login": {
    render: () => m(Login),
  },
  "/functions": {
    render: () => m(Layout, m(FunctionsList)),
  },
  "/functions/new": {
    render: () =>
      m(Layout, { breadcrumbKey: "functions.newFunction" }, m(FunctionCreate)),
  },
  "/functions/:id": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "tabs.code" },
        m(FunctionCode, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/functions/:id/versions": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "tabs.versions" },
        m(FunctionVersions, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/functions/:id/executions": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "tabs.executions" },
        m(FunctionExecutions, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/functions/:id/settings": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "tabs.settings" },
        m(FunctionSettings, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/functions/:id/test": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "tabs.test" },
        m(FunctionTest, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/functions/:id/diff/:v1/:v2": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumbKey: "diff.title" },
        m(VersionDiff, {
          ...vnode.attrs,
          key: `${vnode.attrs.id}-${vnode.attrs.v1}-${vnode.attrs.v2}`,
        }),
      ),
  },
  "/executions/:id": {
    render: (vnode) =>
      m(
        Layout,
        { breadcrumb: "Execution Details" },
        m(ExecutionDetail, { ...vnode.attrs, key: vnode.attrs.id }),
      ),
  },
  "/preview": {
    render: () => m(Layout, { breadcrumb: "Component Preview" }, m(Preview)),
  },
  "/preview/:component": {
    render: (vnode) =>
      m(Layout, { breadcrumb: "Component Preview" }, m(Preview, vnode.attrs)),
  },
});
