/**
 * @fileoverview Route utility functions for generating URLs.
 * Provides two sets of helpers: `routes` for href attributes (with #!/ prefix)
 * and `paths` for m.route.set() calls (without prefix).
 */

/**
 * Route generators for href attributes (hashbang format).
 * Use these when setting href on anchor elements.
 * @example
 * m("a", { href: routes.functionCode("abc123") }, "View Code")
 * @namespace
 */
export const routes = {
  /**
   * Functions list page.
   * @returns {string} Route URL
   */
  functions: () => "#!/functions",

  /**
   * Create new function page.
   * @returns {string} Route URL
   */
  functionCreate: () => "#!/functions/new",

  /**
   * Function code editor page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionCode: (id) => `#!/functions/${id}`,

  /**
   * Function versions list page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionVersions: (id) => `#!/functions/${id}/versions`,

  /**
   * Function executions list page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionExecutions: (id) => `#!/functions/${id}/executions`,

  /**
   * Function settings page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionSettings: (id) => `#!/functions/${id}/settings`,

  /**
   * Function test page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionTest: (id) => `#!/functions/${id}/test`,

  /**
   * Function data page.
   * @param {string} id - Function ID
   * @returns {string} Route URL
   */
  functionData: (id) => `#!/functions/${id}/kv`,
  
  /**
   * Version diff comparison page.
   * @param {string} id - Function ID
   * @param {number} v1 - First version number
   * @param {number} v2 - Second version number
   * @returns {string} Route URL
   */
  functionDiff: (id, v1, v2) => `#!/functions/${id}/diff/${v1}/${v2}`,

  /**
   * Execution detail page.
   * @param {string} id - Execution ID
   * @returns {string} Route URL
   */
  execution: (id) => `#!/executions/${id}`,

  /**
   * Login page.
   * @returns {string} Route URL
   */
  login: () => "#!/login",

  /**
   * Component preview index page.
   * @returns {string} Route URL
   */
  preview: () => "#!/preview",

  /**
   * Component preview page for a specific component.
   * @param {string} component - Component name
   * @returns {string} Route URL
   */
  previewComponent: (component) => `#!/preview/${component}`,
};

/**
 * Path generators for m.route.set() calls (without #!/ prefix).
 * Use these when programmatically navigating.
 * @example
 * m.route.set(paths.functionCode("abc123"))
 * @namespace
 */
export const paths = {
  /**
   * Functions list page.
   * @returns {string} Path
   */
  functions: () => "/functions",

  /**
   * Create new function page.
   * @returns {string} Path
   */
  functionCreate: () => "/functions/new",

  /**
   * Function code editor page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionCode: (id) => `/functions/${id}`,

  /**
   * Function versions list page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionVersions: (id) => `/functions/${id}/versions`,

  /**
   * Function executions list page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionExecutions: (id) => `/functions/${id}/executions`,

  /**
   * Function settings page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionSettings: (id) => `/functions/${id}/settings`,

  /**
   * Function test page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionTest: (id) => `/functions/${id}/test`,

  /**
   * Function data page.
   * @param {string} id - Function ID
   * @returns {string} Path
   */
  functionData: (id) => `/functions/${id}/kv`,
  
  /**
   * Version diff comparison page.
   * @param {string} id - Function ID
   * @param {number} v1 - First version number
   * @param {number} v2 - Second version number
   * @returns {string} Path
   */
  functionDiff: (id, v1, v2) => `/functions/${id}/diff/${v1}/${v2}`,

  /**
   * Execution detail page.
   * @param {string} id - Execution ID
   * @returns {string} Path
   */
  execution: (id) => `/executions/${id}`,

  /**
   * Login page.
   * @returns {string} Path
   */
  login: () => "/login",
};

export default routes;
