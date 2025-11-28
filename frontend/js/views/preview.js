/**
 * @fileoverview Preview view for showcasing all UI components.
 * Development tool for viewing and testing component variations.
 */

import {
  BackButton,
  Button,
  ButtonSize,
  ButtonVariant,
} from "../components/button.js";
import {
  Card,
  CardContent,
  CardDivider,
  CardFooter,
  CardHeader,
} from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  LogLevelBadge,
  MethodBadges,
  StatusBadge,
} from "../components/badge.js";
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/table.js";
import { TabContent, Tabs } from "../components/tabs.js";
import {
  CopyInput,
  FormCheckbox,
  FormGroup,
  FormHelp,
  FormInput,
  FormLabel,
  FormSelect,
  FormTextarea,
  PasswordInput,
} from "../components/form.js";
import { Pagination } from "../components/pagination.js";
import { Toast } from "../components/toast.js";
import { Kbd, Separator } from "../components/kbd.js";
import {
  FunctionTemplates,
  TemplateCard,
  TemplateCards,
} from "../components/template-card.js";
import { APIReference, LuaAPISections } from "../components/api-reference.js";
import { LogViewer } from "../components/log-viewer.js";
import { CodeViewer } from "../components/code-viewer.js";
import { EnvEditor } from "../components/env-editor.js";
import { RequestBuilder } from "../components/request-builder.js";
import {
  DiffViewer,
  LineType,
  VersionLabels,
} from "../components/diff-viewer.js";
import { AIRequestViewer } from "../components/ai-request-viewer.js";

/**
 * @typedef {import('../components/env-editor.js').EnvVar} EnvVar
 */

/**
 * @typedef {Object} PreviewDemoState
 * @property {string} selectedTemplate - Currently selected template ID
 * @property {string} activeTab - Currently active tab ID
 * @property {string} apiSection - Currently selected API section
 * @property {boolean} checkboxChecked - Checkbox demo state
 * @property {string} selectValue - Select demo value
 * @property {number} paginationOffset - Pagination offset
 * @property {number} paginationLimit - Pagination limit
 * @property {EnvVar[]} envVars - Environment variables demo data
 * @property {string} requestMethod - Request builder method
 * @property {string} requestQuery - Request builder query
 * @property {string} requestBody - Request builder body
 */

/**
 * Preview view component for showcasing all UI components.
 * Provides an interactive gallery for development and testing.
 * @type {Object}
 */
export const Preview = {
  /**
   * Currently active component name in the sidebar.
   * @type {string}
   */
  activeComponent: "button",

  /**
   * State for interactive component demos.
   * @type {PreviewDemoState}
   */
  demoState: {
    selectedTemplate: "http",
    activeTab: "tab1",
    apiSection: "http",
    checkboxChecked: false,
    selectValue: "option1",
    paginationOffset: 0,
    paginationLimit: 10,
    envVars: [
      {
        key: "API_KEY",
        value: "secret123",
        state: "original",
        originalKey: "API_KEY",
      },
      { key: "DEBUG", value: "true", state: "original", originalKey: "DEBUG" },
    ],
    requestMethod: "GET",
    requestQuery: "",
    requestBody: "",
  },

  /**
   * Initializes the preview view with the component from route params.
   * @param {Object} vnode - Mithril vnode
   */
  oninit: (vnode) => {
    const component = m.route.param("component");
    if (component) {
      Preview.activeComponent = component;
    }
  },

  /**
   * Updates the active component when route changes.
   * @param {Object} vnode - Mithril vnode
   */
  onbeforeupdate: (vnode) => {
    const component = m.route.param("component");
    if (component && component !== Preview.activeComponent) {
      Preview.activeComponent = component;
    }
  },

  /**
   * List of available components to preview.
   * @type {string[]}
   */
  components: [
    "button",
    "card",
    "badge",
    "table",
    "tabs",
    "form",
    "pagination",
    "toast",
    "kbd",
    "template-card",
    "api-reference",
    "log-viewer",
    "code-viewer",
    "env-editor",
    "request-builder",
    "diff-viewer",
    "ai-request-viewer",
  ],

  /**
   * Renders the preview page with sidebar navigation and component display.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    return m(".preview-page", [
      m(".preview-sidebar", [
        m("h2.preview-sidebar__title", "Components"),
        m(
          "nav.preview-sidebar__nav",
          Preview.components.map((comp) =>
            m(
              "a.preview-sidebar__link",
              {
                href: `#!/preview/${comp}`,
                class: Preview.activeComponent === comp
                  ? "preview-sidebar__link--active"
                  : "",
              },
              comp.replace("-", " "),
            )
          ),
        ),
      ]),
      m(".preview-content", [
        m(
          "h1.preview-content__title",
          Preview.activeComponent.replace("-", " "),
        ),
        m(".preview-content__component", [
          Preview.renderComponent(Preview.activeComponent),
        ]),
      ]),
    ]);
  },

  /**
   * Renders the appropriate component preview based on name.
   * @param {string} name - Component name to render
   * @returns {Object} Mithril vnode
   */
  renderComponent: (name) => {
    switch (name) {
      case "button":
        return Preview.renderButtons();
      case "card":
        return Preview.renderCards();
      case "badge":
        return Preview.renderBadges();
      case "table":
        return Preview.renderTable();
      case "tabs":
        return Preview.renderTabs();
      case "form":
        return Preview.renderForms();
      case "pagination":
        return Preview.renderPagination();
      case "toast":
        return Preview.renderToast();
      case "kbd":
        return Preview.renderKbd();
      case "template-card":
        return Preview.renderTemplateCards();
      case "api-reference":
        return Preview.renderAPIReference();
      case "log-viewer":
        return Preview.renderLogViewer();
      case "code-viewer":
        return Preview.renderCodeViewer();
      case "env-editor":
        return Preview.renderEnvEditor();
      case "request-builder":
        return Preview.renderRequestBuilder();
      case "diff-viewer":
        return Preview.renderDiffViewer();
      case "ai-request-viewer":
        return Preview.renderAIRequestViewer();
      default:
        return m("p", "Component not found");
    }
  },

  /**
   * Renders button component previews.
   * @returns {Object} Mithril vnode
   */
  renderButtons: () => {
    return m(".preview-section", [
      m("h3", "Variants"),
      m(".preview-row", [
        m(Button, { variant: ButtonVariant.PRIMARY }, "Primary"),
        m(Button, { variant: ButtonVariant.SECONDARY }, "Secondary"),
        m(Button, { variant: ButtonVariant.OUTLINE }, "Outline"),
        m(Button, { variant: ButtonVariant.GHOST }, "Ghost"),
        m(Button, { variant: ButtonVariant.DESTRUCTIVE }, "Destructive"),
        m(Button, { variant: ButtonVariant.LINK }, "Link"),
      ]),

      m("h3", "Sizes"),
      m(".preview-row", [
        m(Button, { size: ButtonSize.SM }, "Small"),
        m(Button, { size: ButtonSize.DEFAULT }, "Default"),
        m(Button, { size: ButtonSize.LG }, "Large"),
        m(Button, { size: ButtonSize.ICON, icon: "plus" }),
      ]),

      m("h3", "With Icons"),
      m(".preview-row", [
        m(Button, { icon: "plus" }, "Add Item"),
        m(
          Button,
          { variant: ButtonVariant.DESTRUCTIVE, icon: "trash" },
          "Delete",
        ),
        m(Button, { variant: ButtonVariant.SECONDARY, icon: "copy" }, "Copy"),
      ]),

      m("h3", "States"),
      m(".preview-row", [
        m(Button, { disabled: true }, "Disabled"),
        m(Button, { loading: true }, "Loading"),
      ]),

      m("h3", "Back Button"),
      m(".preview-row", [m(BackButton, { href: "#!/preview" })]),
    ]);
  },

  /**
   * Renders card component previews.
   * @returns {Object} Mithril vnode
   */
  renderCards: () => {
    return m(".preview-section", [
      m("h3", "Basic Card"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardHeader, { title: "Card Title", subtitle: "Card subtitle text" }),
        m(CardContent, [
          m("p", "This is the card content. It can contain any elements."),
        ]),
        m(CardFooter, [
          m(Button, { variant: ButtonVariant.PRIMARY }, "Action"),
        ]),
      ]),

      m("h3", "Card Variants"),
      m(".preview-grid", [
        m(Card, { variant: "danger", style: "margin-bottom: 1rem;" }, [
          m(CardHeader, { title: "Danger Card" }),
          m(CardContent, "This is a danger variant card."),
        ]),
        m(Card, { variant: "warning", style: "margin-bottom: 1rem;" }, [
          m(CardHeader, { title: "Warning Card" }),
          m(CardContent, "This is a warning variant card."),
        ]),
      ]),

      m("h3", "With Divider"),
      m(Card, { style: "max-width: 400px;" }, [
        m(CardHeader, { title: "Section 1" }),
        m(CardContent, "First section content"),
        m(CardDivider),
        m(CardContent, "Second section content"),
      ]),
    ]);
  },

  /**
   * Renders badge component previews.
   * @returns {Object} Mithril vnode
   */
  renderBadges: () => {
    return m(".preview-section", [
      m("h3", "Variants"),
      m(".preview-row", [
        m(Badge, { variant: BadgeVariant.DEFAULT }, "Default"),
        m(Badge, { variant: BadgeVariant.PRIMARY }, "Primary"),
        m(Badge, { variant: BadgeVariant.SECONDARY }, "Secondary"),
        m(Badge, { variant: BadgeVariant.SUCCESS }, "Success"),
        m(Badge, { variant: BadgeVariant.DESTRUCTIVE }, "Destructive"),
        m(Badge, { variant: BadgeVariant.WARNING }, "Warning"),
        m(Badge, { variant: BadgeVariant.INFO }, "Info"),
      ]),

      m("h3", "Sizes"),
      m(".preview-row", [
        m(Badge, { size: BadgeSize.SM }, "Small"),
        m(Badge, { size: BadgeSize.DEFAULT }, "Default"),
        m(Badge, { size: BadgeSize.LG }, "Large"),
      ]),

      m("h3", "ID Badge"),
      m(".preview-row", [
        m(IDBadge, { id: "abc123def456" }),
        m(IDBadge, { id: "xyz789" }),
      ]),

      m("h3", "Status Badge"),
      m(".preview-row", [
        m(StatusBadge, { enabled: true }),
        m(StatusBadge, { enabled: false }),
        m(StatusBadge, { enabled: true, glow: true }),
      ]),

      m("h3", "Method Badges"),
      m(".preview-row", [m(MethodBadges)]),

      m("h3", "Log Level Badges"),
      m(".preview-row", [
        m(LogLevelBadge, { level: "INFO" }),
        m(LogLevelBadge, { level: "WARN" }),
        m(LogLevelBadge, { level: "ERROR" }),
        m(LogLevelBadge, { level: "DEBUG" }),
      ]),
    ]);
  },

  /**
   * Renders table component previews.
   * @returns {Object} Mithril vnode
   */
  renderTable: () => {
    const data = [
      { id: "func-1", name: "get-users", status: "active", version: "1.0.0" },
      {
        id: "func-2",
        name: "create-order",
        status: "active",
        version: "2.1.0",
      },
      {
        id: "func-3",
        name: "send-email",
        status: "disabled",
        version: "1.2.3",
      },
    ];

    return m(".preview-section", [
      m("h3", "Basic Table"),
      m(Card, [
        m(Table, [
          m(TableHeader, [
            m(TableRow, [
              m(TableHead, "Name"),
              m(TableHead, "Status"),
              m(TableHead, "Version"),
            ]),
          ]),
          m(
            TableBody,
            data.map((row) =>
              m(TableRow, { key: row.id }, [
                m(TableCell, { mono: true }, row.name),
                m(
                  TableCell,
                  m(StatusBadge, { enabled: row.status === "active" }),
                ),
                m(TableCell, row.version),
              ])
            ),
          ),
        ]),
      ]),

      m("h3", "Empty Table"),
      m(Card, [
        m(Table, [
          m(TableBody, [
            m(TableEmpty, {
              colspan: 3,
              icon: "inbox",
              message: "No items found. Create your first item to get started.",
            }),
          ]),
        ]),
      ]),
    ]);
  },

  /**
   * Renders tabs component previews.
   * @returns {Object} Mithril vnode
   */
  renderTabs: () => {
    const tabs = [
      { id: "tab1", label: "Overview" },
      { id: "tab2", label: "Settings" },
      { id: "tab3", label: "Logs (42)" },
    ];

    return m(".preview-section", [
      m("h3", "Tabs"),
      m(Tabs, {
        tabs,
        activeTab: Preview.demoState.activeTab,
        onTabChange: (id) => (Preview.demoState.activeTab = id),
      }),
      m(TabContent, [
        Preview.demoState.activeTab === "tab1" &&
        m(Card, [m(CardContent, "Overview content goes here")]),
        Preview.demoState.activeTab === "tab2" &&
        m(Card, [m(CardContent, "Settings content goes here")]),
        Preview.demoState.activeTab === "tab3" &&
        m(Card, [m(CardContent, "Logs content goes here")]),
      ]),
    ]);
  },

  /**
   * Renders form component previews.
   * @returns {Object} Mithril vnode
   */
  renderForms: () => {
    return m(".preview-section", [
      m("h3", "Input"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardContent, [
          m(FormGroup, [
            m(FormLabel, { text: "Name", required: true }),
            m(FormInput, { placeholder: "Enter name" }),
          ]),
          m(FormGroup, [
            m(FormLabel, { text: "Code" }),
            m(FormInput, { placeholder: "monospace", mono: true }),
          ]),
          m(FormGroup, [
            m(FormLabel, { text: "Error State" }),
            m(FormInput, { error: true, value: "Invalid value" }),
            m(FormHelp, { error: true, text: "This field has an error" }),
          ]),
        ]),
      ]),

      m("h3", "Password Input"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardContent, [
          m(FormGroup, [
            m(FormLabel, { text: "Password" }),
            m(PasswordInput, { placeholder: "Enter password" }),
          ]),
        ]),
      ]),

      m("h3", "Copy Input"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardContent, [
          m(FormGroup, [
            m(FormLabel, { text: "API URL" }),
            m(CopyInput, {
              value: "https://api.example.com/v1/functions",
              mono: true,
            }),
          ]),
        ]),
      ]),

      m("h3", "Textarea"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardContent, [
          m(FormGroup, [
            m(FormLabel, { text: "Description" }),
            m(FormTextarea, { placeholder: "Enter description...", rows: 3 }),
          ]),
        ]),
      ]),

      m("h3", "Select"),
      m(Card, { style: "max-width: 400px; margin-bottom: 1rem;" }, [
        m(CardContent, [
          m(FormGroup, [
            m(FormLabel, { text: "Option" }),
            m(FormSelect, {
              options: [
                { value: "option1", label: "Option 1" },
                { value: "option2", label: "Option 2" },
                { value: "option3", label: "Option 3" },
              ],
              selected: Preview.demoState.selectValue,
              onchange: (e) => (Preview.demoState.selectValue = e.target.value),
            }),
          ]),
        ]),
      ]),

      m("h3", "Checkbox"),
      m(Card, { style: "max-width: 400px;" }, [
        m(CardContent, [
          m(FormCheckbox, {
            id: "demo-checkbox",
            label: "Enable feature",
            description: "This will enable the experimental feature.",
            checked: Preview.demoState.checkboxChecked,
            onchange:
              () => (Preview.demoState.checkboxChecked = !Preview.demoState
                .checkboxChecked),
          }),
        ]),
      ]),
    ]);
  },

  /**
   * Renders pagination component previews.
   * @returns {Object} Mithril vnode
   */
  renderPagination: () => {
    return m(".preview-section", [
      m("h3", "Pagination"),
      m(Card, [
        m(CardContent, [
          m(Pagination, {
            total: 100,
            limit: Preview.demoState.paginationLimit,
            offset: Preview.demoState.paginationOffset,
            onPageChange: (
              offset,
            ) => (Preview.demoState.paginationOffset = offset),
            onLimitChange: (limit) => {
              Preview.demoState.paginationLimit = limit;
              Preview.demoState.paginationOffset = 0;
            },
          }),
        ]),
      ]),
    ]);
  },

  /**
   * Renders toast component previews.
   * @returns {Object} Mithril vnode
   */
  renderToast: () => {
    return m(".preview-section", [
      m("h3", "Toast Notifications"),
      m(".preview-row", [
        m(
          Button,
          {
            onclick: () =>
              Toast.show("Operation completed successfully", "success"),
          },
          "Success Toast",
        ),
        m(
          Button,
          {
            variant: ButtonVariant.DESTRUCTIVE,
            onclick: () => Toast.show("An error occurred", "error"),
          },
          "Error Toast",
        ),
        m(
          Button,
          {
            variant: ButtonVariant.SECONDARY,
            onclick: () => Toast.show("Please note this information", "info"),
          },
          "Info Toast",
        ),
      ]),
    ]);
  },

  /**
   * Renders keyboard shortcut component previews.
   * @returns {Object} Mithril vnode
   */
  renderKbd: () => {
    return m(".preview-section", [
      m("h3", "Keyboard Shortcuts"),
      m(".preview-row", [m(Kbd, "Ctrl"), m(Separator), m(Kbd, "C")]),
      m(".preview-row", { style: "margin-top: 1rem;" }, [
        m(Kbd, "⌘"),
        m(Separator),
        m(Kbd, "K"),
      ]),
    ]);
  },

  /**
   * Renders template card component previews.
   * @returns {Object} Mithril vnode
   */
  renderTemplateCards: () => {
    return m(".preview-section", [
      m("h3", "Template Cards"),
      m(
        TemplateCards,
        FunctionTemplates.map((template) =>
          m(TemplateCard, {
            key: template.id,
            name: template.name,
            description: template.description,
            icon: template.icon,
            selected: Preview.demoState.selectedTemplate === template.id,
            onclick: () => (Preview.demoState.selectedTemplate = template.id),
          })
        ),
      ),
    ]);
  },

  /**
   * Renders API reference component previews.
   * @returns {Object} Mithril vnode
   */
  renderAPIReference: () => {
    return m(".preview-section", [
      m("h3", "API Reference"),
      m(".preview-api-reference", { style: "max-width: 400px;" }, [
        m(APIReference, {
          sections: LuaAPISections,
          activeSection: Preview.demoState.apiSection,
          onSectionChange: (id) => {
            Preview.demoState.apiSection = id;
          },
        }),
      ]),
    ]);
  },

  /**
   * Renders log viewer component previews.
   * @returns {Object} Mithril vnode
   */
  renderLogViewer: () => {
    const logs = [
      {
        level: "INFO",
        message: "Function started",
        timestamp: "2024-01-15 10:30:00",
      },
      {
        level: "DEBUG",
        message: "Processing request...",
        timestamp: "2024-01-15 10:30:01",
      },
      {
        level: "WARN",
        message: "Rate limit approaching",
        timestamp: "2024-01-15 10:30:02",
      },
      {
        level: "ERROR",
        message: "Connection timeout",
        timestamp: "2024-01-15 10:30:03",
      },
      {
        level: "INFO",
        message: "Retrying connection...",
        timestamp: "2024-01-15 10:30:04",
      },
    ];

    return m(".preview-section", [
      m("h3", "Log Viewer"),
      m(Card, { style: "max-width: 600px;" }, [
        m(CardHeader, { title: "Execution Logs" }),
        m(CardContent, { noPadding: true }, [
          m(LogViewer, {
            logs,
            maxHeight: "250px",
            noBorder: true,
          }),
        ]),
      ]),

      m("h3", "Empty State"),
      m(Card, { style: "max-width: 600px;" }, [
        m(CardContent, { noPadding: true }, [
          m(LogViewer, { logs: [], noBorder: true }),
        ]),
      ]),
    ]);
  },

  /**
   * Renders code viewer component previews.
   * @returns {Object} Mithril vnode
   */
  renderCodeViewer: () => {
    const luaCode = `function handle(ctx)
    local method = ctx.request.method
    log.info("Received " .. method .. " request")

    return {
        status = 200,
        body = json.encode({ message = "Hello!" })
    }
end`;

    const jsonCode = `{
    "name": "my-function",
    "version": "1.0.0",
    "env": {
        "API_KEY": "secret"
    }
}`;

    const longLineCode =
      `{"id":"abc123","name":"my-function","description":"This is a very long description that demonstrates how word wrapping works in the code viewer component when the wrap attribute is enabled","status":"active","created_at":"2024-01-15T10:30:00Z"}`;

    return m(".preview-section", [
      m("h3", "Lua Code"),
      m(Card, { style: "max-width: 600px; margin-bottom: 1rem;" }, [
        m(CardContent, { noPadding: true }, [
          m(CodeViewer, {
            code: luaCode,
            language: "lua",
            showHeader: true,
            noBorder: true,
            padded: true,
          }),
        ]),
      ]),

      m("h3", "JSON"),
      m(Card, { style: "max-width: 600px; margin-bottom: 1rem;" }, [
        m(CardContent, { noPadding: true }, [
          m(CodeViewer, {
            code: jsonCode,
            language: "json",
            noBorder: true,
            padded: true,
          }),
        ]),
      ]),

      m("h3", "With Word Wrap"),
      m(Card, { style: "max-width: 600px; margin-bottom: 1rem;" }, [
        m(CardContent, { noPadding: true }, [
          m(CodeViewer, {
            code: longLineCode,
            language: "json",
            noBorder: true,
            padded: true,
            wrap: true,
          }),
        ]),
      ]),

      m("h3", "Without Word Wrap (horizontal scroll)"),
      m(Card, { style: "max-width: 600px;" }, [
        m(CardContent, { noPadding: true }, [
          m(CodeViewer, {
            code: longLineCode,
            language: "json",
            noBorder: true,
            padded: true,
            wrap: false,
          }),
        ]),
      ]),
    ]);
  },

  /**
   * Renders environment editor component previews.
   * @returns {Object} Mithril vnode
   */
  renderEnvEditor: () => {
    return m(".preview-section", [
      m("h3", "Environment Editor"),
      m(Card, { style: "max-width: 600px;" }, [
        m(CardHeader, { title: "Environment Variables" }),
        m(CardContent, [
          m(EnvEditor, {
            envVars: Preview.demoState.envVars,
            onAdd: () =>
              Preview.demoState.envVars.push({
                key: "",
                value: "",
                state: "added",
              }),
            onToggleRemove: (i) => {
              const envVar = Preview.demoState.envVars[i];
              if (envVar.state === "removed") {
                envVar.state = "original";
              } else if (envVar.state === "added") {
                Preview.demoState.envVars.splice(i, 1);
              } else {
                envVar.state = "removed";
              }
            },
            onChange: (i, key, value) => {
              Preview.demoState.envVars[i].key = key;
              Preview.demoState.envVars[i].value = value;
            },
          }),
        ]),
      ]),
    ]);
  },

  /**
   * Renders request builder component previews.
   * @returns {Object} Mithril vnode
   */
  renderRequestBuilder: () => {
    return m(".preview-section", [
      m("h3", "Request Builder"),
      m(".preview-request-builder", { style: "max-width: 500px;" }, [
        m(RequestBuilder, {
          url: "https://api.example.com/functions/abc123/invoke",
          method: Preview.demoState.requestMethod,
          query: Preview.demoState.requestQuery,
          body: Preview.demoState.requestBody,
          onMethodChange: (v) => (Preview.demoState.requestMethod = v),
          onQueryChange: (v) => (Preview.demoState.requestQuery = v),
          onBodyChange: (v) => (Preview.demoState.requestBody = v),
          onExecute: () => Toast.show("Request sent!", "success"),
        }),
      ]),
    ]);
  },

  /**
   * Renders diff viewer component previews.
   * @returns {Object[]} Array of Mithril vnodes
   */
  renderDiffViewer: () => {
    const sampleDiff = [
      {
        oldLine: 1,
        newLine: 1,
        type: LineType.UNCHANGED,
        content: "function handle(request)",
      },
      {
        oldLine: 2,
        newLine: 2,
        type: LineType.UNCHANGED,
        content: "  local response = {}",
      },
      {
        oldLine: 3,
        newLine: 0,
        type: LineType.REMOVED,
        content: '  response.body = "Hello"',
      },
      {
        oldLine: 0,
        newLine: 3,
        type: LineType.ADDED,
        content: '  response.body = "Hello, World!"',
      },
      {
        oldLine: 0,
        newLine: 4,
        type: LineType.ADDED,
        content: "  response.status = 200",
      },
      {
        oldLine: 4,
        newLine: 5,
        type: LineType.UNCHANGED,
        content: "  return response",
      },
      { oldLine: 5, newLine: 6, type: LineType.UNCHANGED, content: "end" },
    ];

    return [
      m(".preview-section", [
        m("h3", "Version Labels"),
        m(VersionLabels, {
          oldLabel: "v1",
          newLabel: "v2",
          oldMeta: "2 days ago",
          newMeta: "just now",
          additions: 2,
          deletions: 1,
        }),
      ]),

      m(".preview-section", [
        m("h3", "Diff Viewer"),
        m(DiffViewer, {
          lines: sampleDiff,
          maxHeight: "300px",
        }),
      ]),
    ];
  },

  /**
   * Renders AI request viewer component previews.
   * @returns {Object} Mithril vnode
   */
  renderAIRequestViewer: () => {
    const sampleRequests = [
      {
        id: "aireq-1",
        execution_id: "exec-123",
        provider: "openai",
        model: "gpt-4",
        endpoint: "/v1/chat/completions",
        request_json: JSON.stringify({
          model: "gpt-4",
          messages: [
            { role: "system", content: "You are a helpful assistant." },
            { role: "user", content: "Hello, how are you?" },
          ],
          max_tokens: 100,
        }),
        response_json: JSON.stringify({
          id: "chatcmpl-abc123",
          choices: [
            {
              message: {
                role: "assistant",
                content: "I'm doing well, thank you for asking!",
              },
            },
          ],
          usage: { prompt_tokens: 25, completion_tokens: 12, total_tokens: 37 },
        }),
        status: "success",
        input_tokens: 25,
        output_tokens: 12,
        duration_ms: 845,
        created_at: Math.floor(Date.now() / 1000) - 60,
      },
      {
        id: "aireq-2",
        execution_id: "exec-123",
        provider: "anthropic",
        model: "claude-3-5-sonnet-20241022",
        endpoint: "/v1/messages",
        request_json: JSON.stringify({
          model: "claude-3-5-sonnet-20241022",
          messages: [{ role: "user", content: "What is 2+2?" }],
          max_tokens: 50,
        }),
        response_json: JSON.stringify({
          id: "msg_abc123",
          content: [{ type: "text", text: "2 + 2 equals 4." }],
          usage: { input_tokens: 15, output_tokens: 8 },
        }),
        status: "success",
        input_tokens: 15,
        output_tokens: 8,
        duration_ms: 523,
        created_at: Math.floor(Date.now() / 1000) - 30,
      },
      {
        id: "aireq-3",
        execution_id: "exec-123",
        provider: "openai",
        model: "gpt-4",
        endpoint: "/v1/chat/completions",
        request_json: JSON.stringify({
          model: "gpt-4",
          messages: [{ role: "user", content: "Generate a long response." }],
        }),
        response_json: null,
        status: "error",
        error_message: "Rate limit exceeded. Please retry after 60 seconds.",
        input_tokens: null,
        output_tokens: null,
        duration_ms: 150,
        created_at: Math.floor(Date.now() / 1000) - 10,
      },
    ];

    return m(".preview-section", [
      m("h3", "AI Request Viewer"),
      m(Card, { style: "max-width: 800px; margin-bottom: 1rem;" }, [
        m(CardHeader, {
          title: "AI Requests",
          subtitle: "3 API calls",
          icon: "network",
        }),
        m(CardContent, { noPadding: true }, [
          m(AIRequestViewer, {
            requests: sampleRequests,
            maxHeight: "400px",
            noBorder: true,
          }),
        ]),
      ]),

      m("h3", "Empty State"),
      m(Card, { style: "max-width: 800px;" }, [
        m(CardContent, { noPadding: true }, [
          m(AIRequestViewer, { requests: [], noBorder: true }),
        ]),
      ]),
    ]);
  },
};
