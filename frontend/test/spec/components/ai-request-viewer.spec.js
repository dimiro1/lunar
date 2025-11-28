/**
 * @fileoverview Tests for AIRequestViewer component.
 */

import { AIRequestViewer } from "../../../js/components/ai-request-viewer.js";
import {
  Table,
  TableBody,
  TableEmpty,
  TableRow,
} from "../../../js/components/table.js";
import { Badge, BadgeVariant } from "../../../js/components/badge.js";
import { CodeViewer } from "../../../js/components/code-viewer.js";

describe("AIRequestViewer", () => {
  // Reset expanded rows before each test
  beforeEach(() => {
    AIRequestViewer.expandedRows.clear();
  });

  describe("view()", () => {
    it("renders empty state when no requests", () => {
      const vnode = { attrs: { requests: [] } };
      const result = AIRequestViewer.view(vnode);

      expect(result.tag).toBe(Table);
      const tableBody = result.children[0];
      expect(tableBody.tag).toBe(TableBody);
      const emptyRow = tableBody.children[0];
      expect(emptyRow.tag).toBe(TableEmpty);
      expect(emptyRow.attrs.colspan).toBe(7);
      expect(emptyRow.attrs.message).toBe(
        "No AI requests recorded for this execution.",
      );
    });

    it("renders table with requests", () => {
      const requests = [
        {
          id: "req-1",
          provider: "openai",
          model: "gpt-4",
          status: "success",
          input_tokens: 100,
          output_tokens: 50,
          duration_ms: 1500,
          created_at: 1700000000,
          endpoint: "https://api.openai.com/v1/chat/completions",
          request_json: '{"model":"gpt-4"}',
          response_json: '{"content":"Hello"}',
        },
      ];

      const vnode = { attrs: { requests } };
      const result = AIRequestViewer.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("ai-request-viewer");
    });

    it("applies noBorder class when specified", () => {
      const requests = [createMockRequest()];
      const vnode = { attrs: { requests, noBorder: true } };
      const result = AIRequestViewer.view(vnode);

      expect(result).toHaveClass("ai-request-viewer--no-border");
    });

    it("applies maxHeight style", () => {
      const requests = [createMockRequest()];
      const vnode = { attrs: { requests, maxHeight: "500px" } };
      const result = AIRequestViewer.view(vnode);

      expect(result.attrs.style).toContain("max-height: 500px");
    });

    it("uses default maxHeight of 400px", () => {
      const requests = [createMockRequest()];
      const vnode = { attrs: { requests } };
      const result = AIRequestViewer.view(vnode);

      expect(result.attrs.style).toContain("max-height: 400px");
    });
  });

  describe("toggleRow()", () => {
    it("adds row to expanded set when not expanded", () => {
      expect(AIRequestViewer.expandedRows.has("req-1")).toBe(false);

      AIRequestViewer.toggleRow("req-1");

      expect(AIRequestViewer.expandedRows.has("req-1")).toBe(true);
    });

    it("removes row from expanded set when already expanded", () => {
      AIRequestViewer.expandedRows.add("req-1");

      AIRequestViewer.toggleRow("req-1");

      expect(AIRequestViewer.expandedRows.has("req-1")).toBe(false);
    });
  });

  describe("formatJSON()", () => {
    it("formats valid JSON string with indentation", () => {
      const input = '{"name":"test","value":123}';
      const { formatted, truncated } = AIRequestViewer.formatJSON(input);

      expect(formatted).toContain('"name": "test"');
      expect(formatted).toContain('"value": 123');
      expect(truncated).toBe(false);
    });

    it("returns empty string for null input", () => {
      const { formatted, truncated } = AIRequestViewer.formatJSON(null);

      expect(formatted).toBe("");
      expect(truncated).toBe(false);
    });

    it("returns empty string for undefined input", () => {
      const { formatted, truncated } = AIRequestViewer.formatJSON(undefined);

      expect(formatted).toBe("");
      expect(truncated).toBe(false);
    });

    it("handles object input by stringifying it", () => {
      const input = { name: "test", value: 123 };
      const { formatted, truncated } = AIRequestViewer.formatJSON(input);

      expect(formatted).toContain('"name": "test"');
      expect(truncated).toBe(false);
    });

    it("returns original string for invalid JSON", () => {
      const input = "not valid json";
      const { formatted, truncated } = AIRequestViewer.formatJSON(input);

      expect(formatted).toBe("not valid json");
      expect(truncated).toBe(false);
    });

    it("truncates long JSON when truncate is true", () => {
      const longValue = "x".repeat(6000);
      const input = JSON.stringify({ data: longValue });
      const { formatted, truncated } = AIRequestViewer.formatJSON(input, true);

      expect(truncated).toBe(true);
      expect(formatted).toContain("... (truncated)");
      expect(formatted.length).toBeLessThan(input.length);
    });

    it("does not truncate when truncate is false", () => {
      const longValue = "x".repeat(6000);
      const input = JSON.stringify({ data: longValue });
      const { formatted, truncated } = AIRequestViewer.formatJSON(input, false);

      expect(truncated).toBe(false);
      expect(formatted).not.toContain("... (truncated)");
    });
  });

  describe("renderRequestRows()", () => {
    it("renders main row with request data", () => {
      const req = createMockRequest();
      const result = AIRequestViewer.renderRequestRows(req);

      // Result should be a fragment
      expect(result.tag).toBe("[");
      expect(result.attrs.key).toBe(req.id);

      // First child is the main row
      const mainRow = result.children[0];
      expect(mainRow.tag).toBe(TableRow);
      expect(mainRow.attrs.key).toBe(req.id);
    });

    it("applies expanded class when row is expanded", () => {
      const req = createMockRequest();
      AIRequestViewer.expandedRows.add(req.id);

      const result = AIRequestViewer.renderRequestRows(req);
      const mainRow = result.children[0];

      expect(mainRow.attrs.class).toContain("ai-request-viewer__row--expanded");
    });

    it("does not apply expanded class when row is collapsed", () => {
      const req = createMockRequest();

      const result = AIRequestViewer.renderRequestRows(req);
      const mainRow = result.children[0];

      expect(mainRow.attrs.class).not.toContain(
        "ai-request-viewer__row--expanded",
      );
    });

    it("renders expanded row when row is expanded", () => {
      const req = createMockRequest();
      AIRequestViewer.expandedRows.add(req.id);

      const result = AIRequestViewer.renderRequestRows(req);

      // Should have 2 children: main row and expanded row
      expect(result.children.length).toBe(2);
      const expandedRow = result.children[1];
      expect(expandedRow.tag).toBe("tr");
      expect(expandedRow.attrs.key).toBe(req.id + "-expanded");
    });

    it("does not render expanded row when collapsed", () => {
      const req = createMockRequest();

      const result = AIRequestViewer.renderRequestRows(req);

      expect(result.children.length).toBe(1);
    });

    it("displays tokens as dash when not available", () => {
      const req = createMockRequest();
      req.input_tokens = null;
      req.output_tokens = null;

      const result = AIRequestViewer.renderRequestRows(req);
      const mainRow = result.children[0];
      // Tokens cell is index 4 (after chevron, provider, model, status)
      const tokensCell = mainRow.children[4];
      const tokenText = Array.isArray(tokensCell.children)
        ? tokensCell.children[0]
        : tokensCell.children;

      expect(tokenText).toBe("-");
    });

    it("displays formatted tokens when available", () => {
      const req = createMockRequest();
      req.input_tokens = 100;
      req.output_tokens = 50;

      const result = AIRequestViewer.renderRequestRows(req);
      const mainRow = result.children[0];
      const tokensCell = mainRow.children[4];

      // Tokens are now displayed as [input, " in ", output, " out"]
      expect(Array.isArray(tokensCell.children)).toBe(true);
      expect(tokensCell.children.length).toBe(4);
      expect(tokensCell.children[0].children).toBe(100);
      expect(tokensCell.children[2].children).toBe(50);
    });
  });

  describe("renderJSONPanel()", () => {
    it("renders CodeViewer with formatted JSON", () => {
      const jsonStr = '{"test":"value"}';
      const result = AIRequestViewer.renderJSONPanel("Request", jsonStr);

      expect(result.tag).toBe(CodeViewer);
      expect(result.attrs.language).toBe("json");
      expect(result.attrs.title).toBe("Request");
      expect(result.attrs.maxHeight).toBe("200px");
      expect(result.attrs.padded).toBe(true);
    });

    it("passes formatted code to CodeViewer", () => {
      const jsonStr = '{"test":"value"}';
      const result = AIRequestViewer.renderJSONPanel("Response", jsonStr);

      // Code should be formatted (indented)
      expect(result.attrs.code).toContain('"test": "value"');
    });
  });
});

/**
 * Creates a mock AI request for testing.
 * @returns {Object} Mock AI request
 */
function createMockRequest() {
  return {
    id: "req-123",
    provider: "openai",
    model: "gpt-4",
    status: "success",
    input_tokens: 100,
    output_tokens: 50,
    duration_ms: 1500,
    created_at: 1700000000,
    endpoint: "https://api.openai.com/v1/chat/completions",
    request_json: '{"model":"gpt-4","messages":[]}',
    response_json: '{"content":"Hello"}',
    error_message: null,
  };
}
