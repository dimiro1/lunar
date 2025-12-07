/**
 * @fileoverview English (en) translations for Lunar Dashboard.
 */

export default {
  // Common/shared strings
  common: {
    loading: "Loading...",
    save: "Save",
    cancel: "Cancel",
    delete: "Delete",
    create: "Create",
    search: "Search",
    close: "Close",
    enabled: "Enabled",
    disabled: "Disabled",
    noDescription: "No description",
    na: "N/A",
    invalidDate: "Invalid Date",
    back: "Back",
    saveChanges: "Save Changes",
    functionNotFound: "Function not found",
    status: {
      success: "SUCCESS",
      error: "ERROR",
      timeout: "TIMEOUT",
    },
  },

  // Navigation
  nav: {
    dashboard: "Dashboard",
    logout: "Logout",
    search: "Search",
  },

  // Login page
  login: {
    title: "Lunar",
    subtitle: "Enter your API key to continue",
    apiKeyLabel: "API Key",
    apiKeyPlaceholder: "Enter your API key",
    loginButton: "Login",
    loggingIn: "Logging in...",
    invalidKey: "Invalid API key",
    footer: "Check the server logs for your API key if this is the first run.",
  },

  // Functions list
  functions: {
    title: "Functions",
    subtitle: "Manage your serverless functions",
    newFunction: "New Function",
    allFunctions: "All Functions",
    totalCount: "{{count}} functions total",
    emptyState: "No functions yet. Create your first function to get started.",
    loadingFunctions: "Loading functions...",
    loadingFunction: "Loading function...",
    columns: {
      name: "Name",
      description: "Description",
      status: "Status",
      version: "Version",
    },
  },

  // Function tabs
  tabs: {
    code: "Code",
    versions: "Versions",
    executions: "Executions",
    settings: "Settings",
    test: "Test",
  },

  // Function creation
  create: {
    title: "Create New Function",
    subtitle: "Initialize a new serverless function using Lua.",
    functionName: "Function Name",
    functionNamePlaceholder: "e.g., payment-webhook",
    starterTemplate: "Starter Template",
    createButton: "Create Function",
    failedToCreate: "Failed to create function",
  },

  // Templates
  templates: {
    http: {
      name: "HTTP Template",
      description: "Handle HTTP requests with custom logic",
    },
    api: {
      name: "REST API",
      description: "Build RESTful API endpoints",
    },
    aiChat: {
      name: "AI Chatbot",
      description: "Chat completion with OpenAI or Anthropic",
    },
    email: {
      name: "Send Email",
      description: "Send emails via Resend API",
    },
    blank: {
      name: "Blank",
      description: "Start with empty template",
    },
  },

  // Settings page
  settings: {
    generalConfig: "General Configuration",
    functionName: "Function Name",
    description: "Description",
    logRetention: "Log Retention Period",
    retentionHelp: "Executions older than this will be automatically deleted",
    envVars: "Environment Variables",
    variablesCount: "{{count}} variables",
    network: "Network & Triggers",
    invocationUrl: "Invocation URL",
    functionDescriprion: "function descriprion",
    supportedMethods: "Supported Methods",
    functionStatus: "Function Status",
    enableFunction: "Enable Function",
    disableWarning:
      "Disabling will stop all incoming requests to this function.",
    dangerZone: "Danger Zone",
    deleteFunction: "Delete Function",
    deleteWarning: "Once deleted, this function cannot be recovered.",
    deleteConfirm:
      'Are you sure you want to delete "{{name}}"? This action cannot be undone.',
    retention: {
      days7: "7 days",
      days15: "15 days",
      days30: "30 days",
      year1: "1 year",
    },
  },

  // Executions
  executions: {
    title: "Execution History",
    totalCount: "{{count}} total executions",
    emptyState:
      "No executions yet. Test your function to see execution history.",
    columns: {
      id: "Execution ID",
      status: "Status",
      duration: "Duration",
      time: "Time",
    },
  },

  // Test page
  test: {
    response: "Response",
    status: "Status",
    body: "Body",
    logs: "Logs",
    noResponse: "No response yet",
    executeHint: "Execute a request to see the response",
    viewExecution: "View Execution",
  },

  // Command palette
  commandPalette: {
    searchPlaceholder: "Search functions...",
    loading: "Loading...",
    noResults: "No results found",
    toNavigate: "to navigate",
    toSelect: "to select",
    toClose: "to close",
    actions: {
      viewFunctions: "View all functions",
      createFunction: "Create a new function",
      goToCode: "Go to Code",
      viewVersions: "View version history",
      viewExecutions: "View execution logs",
      configureFunction: "Configure function",
      testFunction: "Test function",
      switchLanguage: "Switch language",
    },
    currentLanguage: "(current)",
  },

  // Toast notifications
  toast: {
    closeNotification: "Close notification",
    envVarsUpdated: "Environment variables updated",
    settingsSaved: "Settings saved successfully",
    functionDeleted: "Function deleted successfully",
    functionEnabled: "Function enabled successfully",
    functionDisabled: "Function disabled successfully",
    failedToSave: "Failed to save settings",
    failedToDelete: "Failed to delete function",
    failedToUpdate: "Failed to update status",
    executionFailed: "Execution failed",
  },

  // Pagination
  pagination: {
    showing: "Showing",
    to: "to",
    of: "of",
    results: "results",
    perPage: "{{count}} per page",
    previous: "Previous",
    next: "Next",
  },

  // Versions
  versions: {
    title: "Version History",
    totalCount: "{{count}} total versions",
    emptyState: "No versions yet.",
    current: "Current",
    columns: {
      version: "Version",
      createdAt: "Created At",
      actions: "Actions",
    },
    compare: "Compare",
    deploy: "Deploy",
  },

  // Diff viewer
  diff: {
    title: "Version Diff",
    comparing: "Comparing",
    with: "with",
    addition: "addition",
    additions: "additions",
    deletion: "deletion",
    deletions: "deletions",
    codeDiff: "Code diff",
    lineAdded: "Line added",
    lineRemoved: "Line removed",
    unchangedLine: "Unchanged line",
  },

  // Request builder
  requestBuilder: {
    request: "Request",
    method: "Method",
    url: "URL",
    requestUrl: "Request URL",
    queryParams: "Query Parameters",
    headers: "Headers (JSON)",
    requestBody: "Request Body",
    execute: "Send Request",
    executing: "Sending...",
  },

  // Badge
  badge: {
    enabled: "Enabled",
    disabled: "Disabled",
  },

  // Code editor
  code: {
    codeSaved: "Code saved successfully",
    failedToSave: "Failed to save code",
  },

  // Execution detail
  execution: {
    loadingExecution: "Loading execution...",
    executionNotFound: "Execution not found",
    executionError: "Execution Error",
    inputEvent: "Input Event (JSON)",
    aiRequests: "AI Requests",
    aiRequestsCount: "{{count}} API calls",
    emailRequests: "Email Requests",
    emailsSent: "{{count}} emails sent",
    executionLogs: "Execution Logs",
    logEntries: "{{count}} log entries",
  },

  // Version diff
  versionDiff: {
    loadingDiff: "Loading diff...",
    diffNotFound: "Diff not found",
    codeChanges: "Code Changes",
  },

  // Environment variables
  envVars: {
    noVariables: 'No environment variables. Click "Add Variable" to add one.',
    addVariable: "Add Variable",
    keyPlaceholder: "KEY",
    valuePlaceholder: "Value",
    restore: "Restore",
    remove: "Remove",
  },

  // Versions
  versionsPage: {
    activateConfirm: "Activate version {{version}}?",
    versionActivated: "Version {{version}} activated",
    failedToActivate: "Failed to activate version",
    active: "ACTIVE",
    activate: "Activate",
    selectToCompare: "Select 2 versions to compare",
    compareVersions: "Compare v{{v1}} and v{{v2}}",
    versionsCount: "{{count}} versions",
  },

  // AI Request viewer
  aiRequestViewer: {
    noRequests: "No AI requests recorded for this execution.",
    provider: "Provider",
    model: "Model",
    status: "Status",
    tokens: "Tokens",
    duration: "Duration",
    time: "Time",
    in: "in",
    out: "out",
    error: "Error",
    endpoint: "Endpoint",
    request: "Request",
    response: "Response",
    truncated: "... (truncated)",
  },

  // API Reference
  apiReference: {
    llmDocumentation: "LLM Documentation",
  },

  // Card
  card: {
    maximize: "Maximize",
    minimize: "Minimize",
  },

  // Code examples
  codeExamples: {
    title: "Code Examples",
    subtitle: "Call this function from your application",
    copied: "Copied!",
    copyToClipboard: "Copy to clipboard",
    selectLanguage: "the selected language is",
  },

  // Email request viewer
  emailRequestViewer: {
    noRequests: "No email requests recorded for this execution.",
    to: "To",
    subject: "Subject",
    status: "Status",
    type: "Type",
    duration: "Duration",
    time: "Time",
    error: "Error",
    from: "From",
    emailId: "Email ID",
    request: "Request",
    response: "Response",
    truncated: "... (truncated)",
  },

  // Form
  form: {
    showPassword: "Show password",
    hidePassword: "Hide password",
    copied: "Copied!",
    copyToClipboard: "Copy to clipboard",
    checkBox: "checkbox",
  },

  // Log viewer
  logViewer: {
    noLogs: "No logs available",
  },

  // Table
  table: {
    noData: "No data available",
  },

  // Lua API Reference
  luaApi: {
    types: {
      string: "string",
      number: "number",
      table: "table",
      function: "function",
      module: "module",
    },
    ai: {
      name: "AI",
      description: "AI provider integrations",
      groups: { chat: "Chat (ai)" },
      items: { chat: "Chat completion with OpenAI or Anthropic" },
    },
    email: {
      name: "Email",
      description: "Email sending via Resend",
      groups: { send: "Send (email)" },
      items: { send: "Send email via Resend API" },
    },
    handler: {
      name: "Handler",
      description: "Handler function inputs",
      groups: {
        context: "Context (ctx)",
        event: "Event (event)",
      },
      items: {
        executionId: "Unique execution identifier",
        functionId: "Function identifier",
        functionName: "Function name",
        version: "Function version",
        requestId: "HTTP request identifier",
        startedAt: "Start timestamp (Unix)",
        baseUrl: "Server base URL",
        method: "HTTP method (GET, POST, etc.)",
        path: "Request path",
        body: "Request body as string",
        headers: "Request headers table",
        query: "Query parameters table",
      },
    },
    io: {
      name: "IO",
      description: "Input/output operations",
      groups: {
        logging: "Logging (log)",
        kv: "Key-Value Store (kv)",
        env: "Environment (env)",
        http: "HTTP Client (http)",
      },
      items: {
        logInfo: "Log info message",
        logDebug: "Log debug message",
        logWarn: "Log warning message",
        logError: "Log error message",
        kvGet: "Get value from store",
        kvSet: "Set key-value pair",
        kvDelete: "Delete key from store",
        envGet: "Get environment variable",
        httpGet: "GET request",
        httpPost: "POST request",
        httpPut: "PUT request",
        httpDelete: "DELETE request",
      },
    },
    data: {
      name: "Data",
      description: "Data transformation",
      groups: {
        json: "JSON (json)",
        base64: "Base64 (base64)",
        crypto: "Crypto (crypto)",
      },
      items: {
        jsonEncode: "Encode table to JSON",
        jsonDecode: "Decode JSON to table",
        base64Encode: "Encode to base64",
        base64Decode: "Decode from base64",
        md5: "MD5 hash (hex)",
        sha256: "SHA256 hash (hex)",
        hmacSha256: "HMAC-SHA256 (hex)",
        uuid: "Generate UUID v4",
      },
    },
    utils: {
      name: "Utils",
      description: "Utility functions",
      groups: {
        time: "Time (time)",
        strings: "Strings (strings)",
        random: "Random (random)",
      },
      items: {
        timeNow: "Current Unix timestamp",
        timeFormat: "Format timestamp",
        timeParse: "Parse time string",
        timeSleep: "Sleep milliseconds",
        trim: "Trim whitespace",
        split: "Split by separator",
        join: "Join with separator",
        contains: "Contains substring",
        replace: "Replace in string",
        randomInt: "Random integer",
        randomFloat: "Random float 0.0-1.0",
        randomString: "Random alphanumeric",
        randomId: "Unique sortable ID",
      },
    },
  },
};
