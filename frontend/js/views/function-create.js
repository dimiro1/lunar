/**
 * @fileoverview Function creation view with template selection.
 */

import { API } from "../api.js";
import { t } from "../i18n/index.js";
import { Toast } from "../components/toast.js";
import { BackButton, Button, ButtonVariant } from "../components/button.js";
import {
  FormGroup,
  FormHelp,
  FormInput,
  FormLabel,
  FormSelect,
} from "../components/form.js";
import {
  FunctionTemplates,
  getTemplateCode,
  getTemplateDescription,
  getTemplateName,
  TemplateCard,
  TemplateCards,
} from "../components/template-card.js";

/**
 * @typedef {Object} FormData
 * @property {string} name - Function name
 * @property {string} description - Function description
 * @property {string} code - Initial function code
 * @property {string} language - Function language ("lua" or "starlark")
 */

/**
 * @typedef {Object} ParsedError
 * @property {string} field - Field name that has the error
 * @property {string} message - Error message
 */

/**
 * Function creation view component.
 * Allows creating new functions with template selection.
 * @type {Object}
 */
export const FunctionCreate = {
  /**
   * Form data state.
   * @type {FormData}
   */
  formData: {
    name: "",
    description: "",
    code: "",
    language: "lua",
  },

  /**
   * Field-specific errors.
   * @type {Object.<string, string>}
   */
  errors: {},

  /**
   * Currently selected template ID.
   * @type {string}
   */
  selectedTemplate: "http",

  /**
   * Initializes the view and resets form state.
   */
  oninit: () => {
    FunctionCreate.formData = {
      name: "",
      description: "",
      code: "",
      language: "lua",
    };
    FunctionCreate.errors = {};
    FunctionCreate.selectedTemplate = "http";
    // Set initial code from the default template in the default language.
    FunctionCreate.formData.code = getTemplateCode("http", "lua");
  },

  /**
   * Selects a template and updates the code for the current language.
   * @param {string} templateId - Template ID to select
   */
  selectTemplate: (templateId) => {
    FunctionCreate.selectedTemplate = templateId;
    FunctionCreate.formData.code = getTemplateCode(
      templateId,
      FunctionCreate.formData.language,
    );
  },

  /**
   * Switches the function language and refreshes the starter code to the
   * selected template's snippet in the new language.
   * @param {string} language - "lua" or "starlark"
   */
  selectLanguage: (language) => {
    FunctionCreate.formData.language = language;
    FunctionCreate.formData.code = getTemplateCode(
      FunctionCreate.selectedTemplate,
      language,
    );
  },

  /**
   * Parses an error message into field and message.
   * @param {string} message - Error message in "field: message" format
   * @returns {ParsedError|null} Parsed error or null if a format doesn't match
   */
  parseErrorMessage: (message) => {
    const match = message.match(/^(\w+):\s*(.+)$/);
    if (match) {
      return { field: match[1], message: match[2] };
    }
    return null;
  },

  /**
   * Creates a new function via the API.
   * @returns {Promise<void>}
   */
  createFunction: async () => {
    FunctionCreate.errors = {};
    try {
      const payload = {
        name: FunctionCreate.formData.name,
        description: FunctionCreate.formData.description,
        code: FunctionCreate.formData.code,
        language: FunctionCreate.formData.language,
      };

      await API.functions.create(payload);
      m.route.set("/functions");
    } catch (e) {
      const error = FunctionCreate.parseErrorMessage(e.message);
      if (error) {
        FunctionCreate.errors[error.field] = error.message;
        m.redraw();
      } else {
        Toast.show(t("create.failedToCreate") + ": " + e.message, "error");
      }
    }
  },

  /**
   * Renders the function creation view.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    return m(".create-function-page.fade-in", [
      m(".create-function-header", [
        m(".create-function-back", [m(BackButton, { href: "#!/functions" })]),
        m("h1.create-function-title", t("create.title")),
        m(
          "p.create-function-subtitle",
          t("create.subtitle"),
        ),
      ]),

      m(".create-function-form", [
        // Function Name
        m(FormGroup, [
          m(FormLabel, {
            text: t("create.functionName"),
            for: "function-name",
          }),
          m(FormInput, {
            id: "function-name",
            placeholder: t("create.functionNamePlaceholder"),
            value: FunctionCreate.formData.name,
            error: !!FunctionCreate.errors.name,
            mono: true,
            oninput: (e) => {
              FunctionCreate.formData.name = e.target.value;
              delete FunctionCreate.errors.name;
            },
          }),
          FunctionCreate.errors.name &&
          m(FormHelp, { error: true, text: FunctionCreate.errors.name }),
        ]),

        // Language
        m(FormGroup, [
          m(FormLabel, {
            text: t("create.functionLanguage"),
            for: "function-language",
          }),
          m(FormSelect, {
            id: "function-language",
            selected: FunctionCreate.formData.language,
            options: [
              { value: "lua", label: "Lua" },
              { value: "starlark", label: "Starlark" },
            ],
            onchange: (e) => FunctionCreate.selectLanguage(e.target.value),
          }),
          m(FormHelp, { text: t("create.functionLanguageHelp") }),
        ]),

        // Starter Template
        m(FormGroup, [
          m(FormLabel, { text: t("create.starterTemplate") }),
          m(
            TemplateCards,
            FunctionTemplates.map((template) =>
              m(TemplateCard, {
                key: template.id,
                name: getTemplateName(template.id),
                description: getTemplateDescription(template.id),
                icon: template.icon,
                selected: FunctionCreate.selectedTemplate === template.id,
                onclick: () => FunctionCreate.selectTemplate(template.id),
              })
            ),
          ),
        ]),

        // Actions
        m(".create-function-actions", [
          m(
            Button,
            {
              variant: ButtonVariant.GHOST,
              href: "#!/functions",
            },
            t("common.cancel"),
          ),
          m(
            Button,
            {
              variant: ButtonVariant.PRIMARY,
              onclick: FunctionCreate.createFunction,
            },
            t("create.createButton"),
          ),
        ]),
      ]),
    ]);
  },
};
