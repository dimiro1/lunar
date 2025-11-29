/**
 * @fileoverview Tests for Table components.
 */

import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableHeaderRow,
  TableRow,
} from "../../../js/components/table.js";
import { t } from "../../../js/i18n/index.js";

describe("Table", () => {
  describe("view()", () => {
    it("renders table wrapper structure", () => {
      const vnode = { attrs: {}, children: [] };
      const result = Table.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("table-wrapper");
    });

    it("renders with hoverable by default", () => {
      const vnode = { attrs: {}, children: [] };
      const result = Table.view(vnode);

      // Find the inner table element
      const tableResponsive = result.children[0];
      const table = tableResponsive.children[0];
      expect(table.attrs["data-table-hoverable"]).toBe(true);
    });

    it("disables hoverable when specified", () => {
      const vnode = { attrs: { hoverable: false }, children: [] };
      const result = Table.view(vnode);

      const tableResponsive = result.children[0];
      const table = tableResponsive.children[0];
      expect(table.attrs["data-table-hoverable"]).toBeFalsy();
    });

    it("enables striped when specified", () => {
      const vnode = { attrs: { striped: true }, children: [] };
      const result = Table.view(vnode);

      const tableResponsive = result.children[0];
      const table = tableResponsive.children[0];
      expect(table.attrs["data-table-striped"]).toBe(true);
    });

    it("passes custom class to wrapper", () => {
      const vnode = { attrs: { class: "my-table" }, children: [] };
      const result = Table.view(vnode);

      expect(result).toHaveClass("my-table");
    });
  });
});

describe("TableHeader", () => {
  describe("view()", () => {
    it("renders thead element", () => {
      const vnode = { attrs: {}, children: [] };
      const result = TableHeader.view(vnode);

      expect(result.tag).toBe("thead");
      expect(result).toHaveClass("table__header");
    });

    it("passes custom class", () => {
      const vnode = { attrs: { class: "custom-header" }, children: [] };
      const result = TableHeader.view(vnode);

      expect(result).toHaveClass("custom-header");
    });
  });
});

describe("TableBody", () => {
  describe("view()", () => {
    it("renders tbody element", () => {
      const vnode = { attrs: {}, children: [] };
      const result = TableBody.view(vnode);

      expect(result.tag).toBe("tbody");
      expect(result).toHaveClass("table__body");
    });
  });
});

describe("TableRow", () => {
  describe("view()", () => {
    it("renders tr element", () => {
      const vnode = { attrs: {}, children: [] };
      const result = TableRow.view(vnode);

      expect(result.tag).toBe("tr");
      expect(result).toHaveClass("table__row");
    });

    it("applies selected class when selected", () => {
      const vnode = { attrs: { selected: true }, children: [] };
      const result = TableRow.view(vnode);

      expect(result).toHaveClass("table__row--selected");
      expect(result.attrs["aria-selected"]).toBe(true);
    });

    it("does not apply selected class when not selected", () => {
      const vnode = { attrs: { selected: false }, children: [] };
      const result = TableRow.view(vnode);

      expect(getVnodeClass(result)).not.toContain("table__row--selected");
    });

    it("passes onclick handler", () => {
      const clickHandler = jasmine.createSpy("onclick");
      const vnode = { attrs: { onclick: clickHandler }, children: [] };
      const result = TableRow.view(vnode);

      expect(result.attrs.onclick).toBe(clickHandler);
    });
  });
});

describe("TableHead", () => {
  describe("view()", () => {
    it("renders th element", () => {
      const vnode = { attrs: {}, children: ["Header"] };
      const result = TableHead.view(vnode);

      expect(result.tag).toBe("th");
      expect(result).toHaveClass("table__head");
      expect(result.attrs.scope).toBe("col");
    });

    it("sets width style when width is provided", () => {
      const vnode = { attrs: { width: "200px" }, children: ["Header"] };
      const result = TableHead.view(vnode);

      expect(result.attrs.style).toEqual({ width: "200px" });
    });

    it("does not set style when width is not provided", () => {
      const vnode = { attrs: {}, children: ["Header"] };
      const result = TableHead.view(vnode);

      expect(result.attrs.style).toBeUndefined();
    });
  });
});

describe("TableCell", () => {
  describe("view()", () => {
    it("renders td element", () => {
      const vnode = { attrs: {}, children: ["Cell content"] };
      const result = TableCell.view(vnode);

      expect(result.tag).toBe("td");
      expect(result).toHaveClass("table__cell");
    });

    it("applies mono class when mono is true", () => {
      const vnode = { attrs: { mono: true }, children: ["Code"] };
      const result = TableCell.view(vnode);

      expect(result).toHaveClass("table__cell--mono");
    });

    it("applies center alignment class", () => {
      const vnode = { attrs: { align: "center" }, children: ["Centered"] };
      const result = TableCell.view(vnode);

      expect(result).toHaveClass("table__cell--center");
    });

    it("applies right alignment class", () => {
      const vnode = { attrs: { align: "right" }, children: ["Right"] };
      const result = TableCell.view(vnode);

      expect(result).toHaveClass("table__cell--right");
    });

    it("does not apply alignment class for left alignment", () => {
      const vnode = { attrs: { align: "left" }, children: ["Left"] };
      const result = TableCell.view(vnode);

      expect(getVnodeClass(result)).not.toContain("table__cell--center");
      expect(getVnodeClass(result)).not.toContain("table__cell--right");
    });
  });
});

describe("TableEmpty", () => {
  describe("view()", () => {
    it("renders tr with td spanning columns", () => {
      const vnode = { attrs: { colspan: 4 } };
      const result = TableEmpty.view(vnode);

      expect(result.tag).toBe("tr");
      const td = result.children[0];
      expect(td.tag).toBe("td");
      expect(td.attrs.colspan).toBe(4);
    });

    it("uses default colspan of 1", () => {
      const vnode = { attrs: {} };
      const result = TableEmpty.view(vnode);

      const td = result.children[0];
      expect(td.attrs.colspan).toBe(1);
    });

    it("displays default message", () => {
      const vnode = { attrs: {} };
      const result = TableEmpty.view(vnode);

      const td = result.children[0];
      const message = td.children.find((child) => child.tag === "p");
      // Extract text from children (may be array of text nodes or string)
      const text = Array.isArray(message.children)
        ? (message.children[0]?.children || message.children[0])
        : message.children;
      expect(text).toBe(t("table.noData"));
    });

    it("displays custom message", () => {
      const vnode = { attrs: { message: "No items found" } };
      const result = TableEmpty.view(vnode);

      const td = result.children[0];
      const message = td.children.find((child) => child.tag === "p");
      const text = Array.isArray(message.children)
        ? (message.children[0]?.children || message.children[0])
        : message.children;
      expect(text).toBe("No items found");
    });
  });
});

describe("TableHeaderRow", () => {
  describe("view()", () => {
    it("renders row with string columns", () => {
      const vnode = { attrs: { columns: ["Name", "Age", "City"] } };
      const result = TableHeaderRow.view(vnode);

      expect(result.tag).toBe(TableRow);
      expect(result.children.length).toBe(3);
    });

    it("renders row with column objects", () => {
      const vnode = {
        attrs: {
          columns: [
            { name: "Name", width: "200px" },
            { name: "Age", width: "100px" },
          ],
        },
      };
      const result = TableHeaderRow.view(vnode);

      expect(result.children.length).toBe(2);
      expect(result.children[0].attrs.width).toBe("200px");
      expect(result.children[1].attrs.width).toBe("100px");
    });

    it("handles empty columns array", () => {
      const vnode = { attrs: { columns: [] } };
      const result = TableHeaderRow.view(vnode);

      expect(result.children.length).toBe(0);
    });

    it("handles missing columns attribute", () => {
      const vnode = { attrs: {} };
      const result = TableHeaderRow.view(vnode);

      expect(result.children.length).toBe(0);
    });
  });
});
