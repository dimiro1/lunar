/**
 * @fileoverview Code examples component showing API calls in multiple languages.
 */

import { Card, CardContent, CardHeader } from "./card.js";
import { icons } from "../icons.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {('curl'|'javascript'|'python'|'go')} SupportedLanguage
 */

/**
 * Code examples component.
 * Shows how to call a function from different programming languages.
 * @type {Object}
 */
export const CodeExamples = {
  /**
   * Currently selected language.
   * @type {SupportedLanguage}
   */
  selectedLang: "curl",

  /**
   * Whether copy was successful (for feedback).
   * @type {boolean}
   */
  copied: false,

  /**
   * Copies text to clipboard and shows feedback.
   * @param {string} text - Text to copy
   */
  copyToClipboard: (text) => {
    navigator.clipboard.writeText(text).then(() => {
      CodeExamples.copied = true;
      m.redraw();
      setTimeout(() => {
        CodeExamples.copied = false;
        m.redraw();
      }, 2000);
    });
  },

  /**
   * Generates a code example for the selected language.
   * @param {string} functionId - Function ID
   * @param {string} method - HTTP method
   * @param {string} query - Query string
   * @param {string} body - Request body
   * @returns {string} Generated code example
   */
  generateCodeExample: (functionId, method, query, body) => {
    const host = window.location.origin;
    const url = `${host}/fn/${functionId}${query ? "?" + query : ""}`;
    const lang = CodeExamples.selectedLang;

    /** @type {Object.<SupportedLanguage, string>} */
    const examples = {
      curl: `curl -X ${method} '${url}'${
        body
          ? ` \\\n  -H 'Content-Type: application/json' \\\n  -d '${
            body.replace(/'/g, "'\\''")
          }'`
          : ""
      }`,

      javascript: `fetch('${url}', {
  method: '${method}',${
        body
          ? `\n  headers: {\n    'Content-Type': 'application/json'\n  },\n  body: '${
            body.replace(/\\/g, "\\\\").replace(/'/g, "\\'")
          }'`
          : ""
      }
})
  .then(response => response.text())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`,

      python: `import requests

response = requests.${method.toLowerCase()}('${url}'${
        body ? `,\n    json=${body}` : ""
      })
print(response.text)`,

      go: `package main

import (
    "fmt"
    "io"
    "net/http"${body ? '\n    "strings"' : ""}
)

func main() {
    ${
        body
          ? `body := strings.NewReader(\`${body}\`)
    req, err := http.NewRequest("${method}", "${url}", body)`
          : `req, err := http.NewRequest("${method}", "${url}", nil)`
      }
    if err != nil {
        panic(err)
    }
    ${body ? 'req.Header.Set("Content-Type", "application/json")\n    ' : ""}
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    bodyBytes, _ := io.ReadAll(resp.Body)
    fmt.Println(string(bodyBytes))
}`,
    };

    return examples[lang] || examples.curl;
  },

  /**
   * Gets the highlight.js language class for the selected language.
   * @returns {string} Language class name
   */
  getLanguageClass: () => {
    /** @type {Object.<SupportedLanguage, string>} */
    const langMap = {
      curl: "bash",
      javascript: "javascript",
      python: "python",
      go: "go",
    };
    return langMap[CodeExamples.selectedLang] || "bash";
  },

  /**
   * Renders the code examples component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.functionId - Function ID
   * @param {string} vnode.attrs.method - HTTP method
   * @param {string} vnode.attrs.query - Query string
   * @param {string} vnode.attrs.body - Request body
   * @returns {Object} Mithril vnode
   */
  view: (vnode) => {
    const { functionId, method, query, body } = vnode.attrs;

    return m(Card, { style: "margin-bottom: 1.5rem" }, [
      m(
        CardHeader,
        {
          title: t("codeExamples.title"),
          subtitle: t("codeExamples.subtitle"),
        },
        [
          m(
            "select.form-select",
            {
              "aria-label": `${
                t("codeExamples.selectLanguage")
              } ${CodeExamples.selectedLang}`,
              value: CodeExamples.selectedLang,
              onchange: (e) => {
                CodeExamples.selectedLang = e.target.value;
              },
              style: "width: auto;",
            },
            [
              m("option", { value: "curl" }, "cURL"),
              m("option", { value: "javascript" }, "JavaScript (fetch)"),
              m("option", { value: "python" }, "Python (requests)"),
              m("option", { value: "go" }, "Go (net/http)"),
            ],
          ),
        ],
      ),
      m(CardContent, { noPadding: true }, [
        m(".code-examples-wrapper", { style: "position: relative;" }, [
          m(
            "button.form-copy-button",
            {
              type: "button",
              title: CodeExamples.copied
                ? t("codeExamples.copied")
                : t("codeExamples.copyToClipboard"),
              style: "position: absolute; top: 2rem; right: 1rem; z-index: 10;",
              onclick: () => {
                const code = CodeExamples.generateCodeExample(
                  functionId,
                  method,
                  query,
                  body,
                );
                CodeExamples.copyToClipboard(code);
              },
            },
            CodeExamples.copied
              ? m(
                "span",
                { style: "color: var(--color-success)" },
                m.trust(icons.check()),
              )
              : m.trust(icons.copy()),
          ),
          m(
            "pre",
            {
              style:
                "overflow-x: auto; margin: 0; padding: 1rem; border-radius: 0 0 var(--radius-md) var(--radius-md); background: #0d1117; color: #c9d1d9;",
            },
            [
              m("code", {
                class: `language-${CodeExamples.getLanguageClass()}`,
                oncreate: (codeVnode) => {
                  const code = CodeExamples.generateCodeExample(
                    functionId,
                    method,
                    query,
                    body,
                  );
                  codeVnode.dom.textContent = code;
                  if (window.hljs) {
                    window.hljs.highlightElement(codeVnode.dom);
                  }
                },
                onupdate: (codeVnode) => {
                  const code = CodeExamples.generateCodeExample(
                    functionId,
                    method,
                    query,
                    body,
                  );
                  // Clear previous content and highlighting
                  codeVnode.dom.textContent = code;
                  delete codeVnode.dom.dataset.highlighted;
                  if (window.hljs) {
                    window.hljs.highlightElement(codeVnode.dom);
                  }
                },
              }),
            ],
          ),
        ]),
      ]),
    ]);
  },
};
