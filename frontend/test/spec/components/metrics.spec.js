/**
 * @fileoverview Tests for the metric visualization components.
 */

import {
  BarChart,
  MetricCard,
  Sparkline,
} from "../../../js/components/metrics.js";

// Unwraps a vnode's text child, which Mithril may store as a raw string or as a
// normalized "#" text vnode.
const text = (vnode) => {
  const c = vnode.children;
  const first = Array.isArray(c) ? c[0] : c;
  return first && first.tag === "#" ? first.children : first;
};

describe("MetricCard", () => {
  it("renders the label and value", () => {
    const result = MetricCard.view({
      attrs: { label: "Total executions", value: "1,234" },
    });
    expect(result).toHaveClass("metric-card");
    // children: [label, value, sublabel|null]
    expect(text(result.children[0])).toBe("Total executions");
    expect(text(result.children[1])).toBe("1,234");
  });

  it("applies the variant modifier class", () => {
    const result = MetricCard.view({
      attrs: { label: "Error rate", value: "5%", variant: "danger" },
    });
    expect(result).toHaveClass("metric-card--danger");
  });

  it("omits the sublabel when not provided", () => {
    const result = MetricCard.view({ attrs: { label: "x", value: "1" } });
    expect(result.children[2]).toBe(null);
  });
});

describe("BarChart", () => {
  it("renders an svg with one group per bar", () => {
    const result = BarChart.view({
      attrs: {
        bars: [
          { count: 3, errorCount: 1, title: "a" },
          { count: 0, errorCount: 0, title: "b" },
        ],
      },
    });
    expect(result).toHaveClass("bar-chart");
    const svg = result.children[0];
    expect(svg.tag).toBe("svg");
    expect(svg.children.length).toBe(2);
  });

  it("handles an empty series without throwing", () => {
    const result = BarChart.view({ attrs: { bars: [] } });
    expect(result.children[0].tag).toBe("svg");
    expect(result.children[0].children.length).toBe(0);
  });
});

describe("Sparkline", () => {
  it("renders a polyline for the series", () => {
    const result = Sparkline.view({
      attrs: {
        points: [
          { value: 10, title: "a" },
          { value: 20, title: "b" },
        ],
      },
    });
    expect(result).toHaveClass("sparkline");
    const svg = result.children[0];
    expect(svg.tag).toBe("svg");
    // [area, line, hit-rects]
    const line = svg.children[1];
    expect(line.tag).toBe("polyline");
    // value 10 → y=50, value 20 (max) → y=0, over a width of 10.
    expect(line.attrs.points).toBe("0,50 10,0");
  });

  it("handles an empty series without throwing", () => {
    const result = Sparkline.view({ attrs: { points: [] } });
    expect(result.children[0].tag).toBe("svg");
  });
});
