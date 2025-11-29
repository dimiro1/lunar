/**
 * @fileoverview Request builder component for testing HTTP endpoints.
 */

import { Button, ButtonVariant } from "./button.js";
import { Card, CardContent, CardHeader } from "./card.js";
import {
  CopyInput,
  FormGroup,
  FormInput,
  FormLabel,
  FormSelect,
  FormTextarea,
} from "./form.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {Object} HttpMethod
 * @property {string} value - HTTP method value
 * @property {string} label - Display label
 */

/**
 * @typedef {Object} CodeExamples
 * @property {string} curl - cURL command example
 * @property {string} javascript - JavaScript fetch example
 * @property {string} python - Python requests example
 * @property {string} go - Go net/http example
 */

/**
 * Request a builder component for testing functions.
 * Provides a form interface for building and sending HTTP requests.
 * @type {Object}
 */
export const RequestBuilder = {
  /**
   * Renders the request builder component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.url=''] - Base URL for the request
   * @param {string} [vnode.attrs.method='GET'] - HTTP method
   * @param {string} [vnode.attrs.query=''] - Query string parameters
   * @param {string} [vnode.attrs.headers] - Request headers as JSON string
   * @param {string} [vnode.attrs.body=''] - Request body
   * @param {function(string): void} [vnode.attrs.onMethodChange] - Callback when method changes
   * @param {function(string): void} [vnode.attrs.onQueryChange] - Callback when query changes
   * @param {function(string): void} [vnode.attrs.onHeadersChange] - Callback when headers change
   * @param {function(string): void} [vnode.attrs.onBodyChange] - Callback when body changes
   * @param {function} [vnode.attrs.onExecute] - Callback when execute button is clicked
   * @param {boolean} [vnode.attrs.loading=false] - Whether request is in progress
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      url = "",
      method = "GET",
      query = "",
      headers = '{"Content-Type": "application/json"}',
      body = "",
      onMethodChange,
      onQueryChange,
      onHeadersChange,
      onBodyChange,
      onExecute,
      loading = false,
    } = vnode.attrs;

    const methods = [
      { value: "GET", label: "GET" },
      { value: "POST", label: "POST" },
      { value: "PUT", label: "PUT" },
      { value: "DELETE", label: "DELETE" },
      { value: "PATCH", label: "PATCH" },
    ];

    return m(Card, [
      m(CardHeader, { title: t("requestBuilder.request") }),
      m(CardContent, [
        // URL display with method selector
        m(".request-builder__url", [
          m(FormSelect, {
            options: methods,
            selected: method,
            onchange: (e) => onMethodChange && onMethodChange(e.target.value),
          }),
          m(CopyInput, {
            value: url + (query ? `?${query}` : ""),
            mono: true,
          }),
        ]),

        // Query params
        m(FormGroup, [
          m(FormLabel, { text: t("requestBuilder.queryParams") }),
          m(FormInput, {
            value: query,
            placeholder: "key=value&other=value",
            mono: true,
            oninput: (e) => onQueryChange && onQueryChange(e.target.value),
          }),
        ]),

        // Headers
        m(FormGroup, [
          m(FormLabel, { text: t("requestBuilder.headers") }),
          m(FormTextarea, {
            value: headers,
            rows: 2,
            mono: true,
            oninput: (e) => onHeadersChange && onHeadersChange(e.target.value),
          }),
        ]),

        // Body
        m(FormGroup, [
          m(FormLabel, { text: t("requestBuilder.requestBody") }),
          m(FormTextarea, {
            value: body,
            rows: 4,
            mono: true,
            oninput: (e) => onBodyChange && onBodyChange(e.target.value),
          }),
        ]),

        // Execute button
        m(
          Button,
          {
            variant: ButtonVariant.PRIMARY,
            icon: "play",
            fullWidth: true,
            onclick: onExecute,
            disabled: loading,
            loading: loading,
          },
          loading ? t("requestBuilder.executing") : t("requestBuilder.execute"),
        ),
      ]),
    ]);
  },
};

/**
 * Generates code examples for different programming languages.
 * Creates ready-to-use code snippets for cURL, JavaScript, Python, and Go.
 * @param {string} url - Base URL
 * @param {string} method - HTTP method
 * @param {string} query - Query string parameters
 * @param {string} headers - Headers as JSON string
 * @param {string} body - Request body
 * @returns {CodeExamples} Object with code examples for each language
 */
export function generateCodeExamples(url, method, query, headers, body) {
  const fullUrl = url + (query ? `?${query}` : "");
  const hasBody = body && ["POST", "PUT", "PATCH"].includes(method);

  let headersList = [];
  try {
    if (headers && headers.trim()) {
      headersList = Object.entries(JSON.parse(headers)).map(([k, v]) => ({
        key: k,
        value: String(v),
      }));
    }
  } catch (e) {}

  // cURL
  let curl = `curl -X ${method}`;
  headersList.forEach((h) => (curl += ` \\\n  -H '${h.key}: ${h.value}'`));
  if (hasBody) curl += ` \\\n  -d '${body.replace(/'/g, "'\\''")}'`;
  curl += ` \\\n  '${fullUrl}'`;

  // JavaScript
  let js = headersList.length || hasBody
    ? `const response = await fetch('${fullUrl}', {\n  method: '${method}',\n`
    : `const response = await fetch('${fullUrl}');\n\n`;
  if (headersList.length || hasBody) {
    if (headersList.length) {
      js += `  headers: {\n${
        headersList.map((h) => `    '${h.key}': '${h.value}'`).join(",\n")
      }\n  },\n`;
    }
    if (hasBody) {
      js += `  body: '${
        body.replace(/\\/g, "\\\\").replace(/'/g, "\\'").replace(/\n/g, "\\n")
      }'\n`;
    }
    js += `});\n\n`;
  }
  js += `const data = await response.json();\nconsole.log(data);`;

  // Python
  let py = `import requests\n\n`;
  if (headersList.length) {
    py += `headers = {\n${
      headersList.map((h) => `    '${h.key}': '${h.value}'`).join(",\n")
    }\n}\n\n`;
  }
  py += `response = requests.${method.toLowerCase()}(\n    '${fullUrl}'`;
  if (headersList.length) py += `,\n    headers=headers`;
  if (hasBody) py += `,\n    data='${body.replace(/'/g, "\\'")}'`;
  py += `\n)\n\nprint(response.json())`;

  // Go
  let go = `package main\n\nimport (\n\t"fmt"\n\t"io"\n\t"net/http"`;
  if (hasBody) go += `\n\t"strings"`;
  go += `\n)\n\nfunc main() {\n`;
  if (hasBody) {
    go += `\tbody := strings.NewReader("${
      body.replace(/"/g, '\\"').replace(/\n/g, "\\n")
    }")\n`;
    go += `\treq, err := http.NewRequest("${method}", "${fullUrl}", body)\n`;
  } else {
    go += `\treq, err := http.NewRequest("${method}", "${fullUrl}", nil)\n`;
  }
  go += `\tif err != nil {\n\t\tpanic(err)\n\t}\n\n`;
  headersList.forEach(
    (h) => (go += `\treq.Header.Set("${h.key}", "${h.value}")\n`),
  );
  if (headersList.length) go += `\n`;
  go += `\tclient := &http.Client{}\n\tresp, err := client.Do(req)\n`;
  go += `\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer resp.Body.Close()\n\n`;
  go +=
    `\tresBody, _ := io.ReadAll(resp.Body)\n\tfmt.Println(string(resBody))\n}`;

  return { curl, javascript: js, python: py, go };
}
