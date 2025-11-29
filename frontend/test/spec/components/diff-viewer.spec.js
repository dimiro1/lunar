/**
 * @fileoverview Tests for DiffViewer components - focused on critical functionality.
 */

import {
  DiffViewer,
  LineType,
  VersionLabels,
} from "../../../js/components/diff-viewer.js";
import { t } from "../../../js/i18n/index.js";

/**
 * Check if vnode has a specific class (handles both tag selector and attrs.class)
 */
function vnodeHasClass(node, className) {
  if (!node || typeof node !== "object") return false;
  // Check tag selector (e.g., "div.class-name")
  if (typeof node.tag === "string" && node.tag.includes(className)) return true;
  // Check attrs.class
  return getVnodeClass(node).includes(className);
}

/**
 * Deep search for a node matching predicate
 */
function findNode(node, predicate) {
  if (!node) return null;
  if (typeof node !== "object") return null;
  if (Array.isArray(node)) {
    for (const child of node) {
      const found = findNode(child, predicate);
      if (found) return found;
    }
    return null;
  }
  if (predicate(node)) return node;
  if (node.children) return findNode(node.children, predicate);
  return null;
}

/**
 * Find all nodes matching predicate
 */
function findAllNodes(node, predicate, results = []) {
  if (!node) return results;
  if (typeof node !== "object") return results;
  if (Array.isArray(node)) {
    node.forEach((child) => findAllNodes(child, predicate, results));
    return results;
  }
  if (predicate(node)) results.push(node);
  if (node.children) findAllNodes(node.children, predicate, results);
  return results;
}

describe("LineType", () => {
  it("has correct enum values", () => {
    expect(LineType.ADDED).toBe("added");
    expect(LineType.REMOVED).toBe("removed");
    expect(LineType.UNCHANGED).toBe("unchanged");
  });
});

describe("VersionLabels", () => {
  it("shows additions count", () => {
    const vnode = {
      attrs: { additions: 5, deletions: 0 },
      children: [],
    };
    const result = VersionLabels.view(vnode);

    const addedSpan = findNode(
      result,
      (n) => vnodeHasClass(n, "diff-stats-added"),
    );
    expect(addedSpan).toBeTruthy();
  });

  it("shows deletions count", () => {
    const vnode = {
      attrs: { additions: 0, deletions: 3 },
      children: [],
    };
    const result = VersionLabels.view(vnode);

    const removedSpan = findNode(
      result,
      (n) => vnodeHasClass(n, "diff-stats-removed"),
    );
    expect(removedSpan).toBeTruthy();
  });

  it("hides additions when zero", () => {
    const vnode = {
      attrs: { additions: 0, deletions: 3 },
      children: [],
    };
    const result = VersionLabels.view(vnode);

    const addedSpan = findNode(
      result,
      (n) => vnodeHasClass(n, "diff-stats-added"),
    );
    expect(addedSpan).toBeFalsy();
  });

  it("hides deletions when zero", () => {
    const vnode = {
      attrs: { additions: 5, deletions: 0 },
      children: [],
    };
    const result = VersionLabels.view(vnode);

    const removedSpan = findNode(
      result,
      (n) => vnodeHasClass(n, "diff-stats-removed"),
    );
    expect(removedSpan).toBeFalsy();
  });
});

describe("DiffViewer", () => {
  it("renders table structure", () => {
    const vnode = {
      attrs: { lines: [] },
      children: [],
    };
    const result = DiffViewer.view(vnode);

    const table = findNode(result, (n) => vnodeHasClass(n, "diff-table"));
    expect(table).toBeTruthy();
  });

  it("renders correct number of diff lines", () => {
    const lines = [
      { type: LineType.UNCHANGED, content: "line 1", oldLine: 1, newLine: 1 },
      { type: LineType.REMOVED, content: "line 2", oldLine: 2, newLine: 0 },
      { type: LineType.ADDED, content: "line 3", oldLine: 0, newLine: 2 },
    ];
    const vnode = {
      attrs: { lines },
      children: [],
    };
    const result = DiffViewer.view(vnode);

    // Find tbody and check children (each line is a component rendered)
    const tbody = findNode(result, (n) => n && n.tag === "tbody");
    expect(tbody).toBeTruthy();
    expect(tbody.children.length).toBe(3);
  });

  it("applies maxHeight style", () => {
    const vnode = {
      attrs: { lines: [], maxHeight: "400px" },
      children: [],
    };
    const result = DiffViewer.view(vnode);

    const scroll = findNode(result, (n) => vnodeHasClass(n, "diff-scroll"));
    expect(scroll).toBeTruthy();
    expect(scroll.attrs.style).toContain("max-height: 400px");
  });

  it("applies no-border class when noBorder is true", () => {
    const vnode = {
      attrs: { lines: [], noBorder: true },
      children: [],
    };
    const result = DiffViewer.view(vnode);

    // Class could be in attrs.class or appended to existing class
    const hasNoBorder = vnodeHasClass(result, "diff-container--no-border");
    expect(hasNoBorder).toBe(true);
  });

  it("has accessibility attributes", () => {
    const vnode = {
      attrs: { lines: [] },
      children: [],
    };
    const result = DiffViewer.view(vnode);

    expect(result.attrs.role).toBe("region");
    expect(result.attrs["aria-label"]).toBe(t("diff.codeDiff"));
  });
});
