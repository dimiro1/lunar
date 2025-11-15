export const CodeExamples = {
  selectedLang: "curl",
  copied: false,

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

  generateCodeExample: (functionId, method, query, body) => {
    const host = window.location.origin;
    const url = `${host}/fn/${functionId}${query ? '?' + query : ''}`;
    const lang = CodeExamples.selectedLang;

    const examples = {
      curl: `curl -X ${method} '${url}'${body ? ` \\\n  -H 'Content-Type: application/json' \\\n  -d '${body.replace(/'/g, "'\\''")}'` : ''}`,

      javascript: `fetch('${url}', {
  method: '${method}',${body ? `\n  headers: {\n    'Content-Type': 'application/json'\n  },\n  body: '${body.replace(/\\/g, '\\\\').replace(/'/g, "\\'")}'` : ''}
})
  .then(response => response.text())
  .then(data => console.log(data))
  .catch(error => console.error('Error:', error));`,

      python: `import requests

response = requests.${method.toLowerCase()}('${url}'${body ? `,\n    json=${body}` : ''})
print(response.text)`,

      go: `package main

import (
    "fmt"
    "io"
    "net/http"${body ? '\n    "strings"' : ''}
)

func main() {
    ${body ? `body := strings.NewReader(\`${body}\`)
    req, err := http.NewRequest("${method}", "${url}", body)` : `req, err := http.NewRequest("${method}", "${url}", nil)`}
    if err != nil {
        panic(err)
    }
    ${body ? 'req.Header.Set("Content-Type", "application/json")\n    ' : ''}
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    bodyBytes, _ := io.ReadAll(resp.Body)
    fmt.Println(string(bodyBytes))
}`
    };

    return examples[lang] || examples.curl;
  },

  getLanguageClass: () => {
    const langMap = {
      curl: "bash",
      javascript: "javascript",
      python: "python",
      go: "go"
    };
    return langMap[CodeExamples.selectedLang] || "bash";
  },

  view: (vnode) => {
    const { functionId, method, query, body } = vnode.attrs;

    return m(".card.mb-24", [
      m(".card-header", [
        m("div", [
          m(".card-title", "Code Examples"),
          m(".card-subtitle", "Call this function from your application"),
        ]),
        m(
          "select.form-select",
          {
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
      ]),
      m("div", { style: "padding: 16px; position: relative;" }, [
        m(
          "button.btn",
          {
            style: "position: absolute; top: 24px; right: 24px; padding: 6px 12px; font-size: 12px; z-index: 10;",
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
          CodeExamples.copied ? "Copied!" : "Copy",
        ),
        m(
          "pre",
          { style: "border-radius: 6px; overflow-x: auto; margin: 0;" },
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
    ]);
  },
};
