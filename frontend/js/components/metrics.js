/**
 * @fileoverview Metric visualization components for the function metrics
 * dashboard. All charts are hand-rolled inline SVG (rendered through Mithril's
 * hyperscript) so the dashboard needs no charting dependency. They are
 * intentionally small and presentational — callers pass already-aggregated,
 * display-ready data.
 */

/**
 * A single summary statistic: a big number with a label and optional sublabel.
 * @type {Object}
 */
export const MetricCard = {
  /**
   * @param {Object} vnode - Mithril vnode
   * @param {string} vnode.attrs.label - Stat label (e.g. "Total executions")
   * @param {string|number} vnode.attrs.value - The headline value
   * @param {string} [vnode.attrs.sublabel] - Optional secondary line
   * @param {"default"|"success"|"danger"} [vnode.attrs.variant="default"] - Accent
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { label, value, sublabel, variant = "default" } = vnode.attrs;
    return m(`.metric-card.metric-card--${variant}`, [
      m(".metric-card__label", label),
      m(".metric-card__value", String(value)),
      sublabel ? m(".metric-card__sub", sublabel) : null,
    ]);
  },
};

// Shared chart geometry. The SVG uses an abstract coordinate space and is
// stretched to its container width via CSS (preserveAspectRatio="none"), so the
// absolute numbers here only set proportions, not pixels.
const CHART_HEIGHT = 100;
const STEP = 10; // horizontal space allotted to each data point
const BAR_WIDTH = 7;

/**
 * A vertical bar chart of execution volume over time. Each bar is split into a
 * success portion (bottom) and an error portion (top) so volume and error
 * spikes both read at a glance. Hovering a bar shows its native SVG tooltip.
 * @type {Object}
 */
export const BarChart = {
  /**
   * @param {Object} vnode - Mithril vnode
   * @param {Array<{count:number,errorCount:number,title:string}>} vnode.attrs.bars
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const bars = vnode.attrs.bars || [];
    const maxCount = Math.max(1, ...bars.map((b) => b.count));
    const width = Math.max(bars.length * STEP, STEP);

    return m(".chart.bar-chart", [
      m(
        "svg.chart__svg",
        {
          viewBox: `0 0 ${width} ${CHART_HEIGHT}`,
          preserveAspectRatio: "none",
          role: "img",
        },
        bars.map((b, i) => {
          const totalH = (b.count / maxCount) * CHART_HEIGHT;
          const errorH = (b.errorCount / maxCount) * CHART_HEIGHT;
          const successH = totalH - errorH;
          const x = i * STEP + (STEP - BAR_WIDTH) / 2;
          return m("g.bar-chart__group", [
            // Success portion (bottom of the bar).
            successH > 0
              ? m("rect.bar-chart__bar.bar-chart__bar--success", {
                x,
                y: CHART_HEIGHT - successH,
                width: BAR_WIDTH,
                height: successH,
              })
              : null,
            // Error portion, stacked on top.
            errorH > 0
              ? m("rect.bar-chart__bar.bar-chart__bar--error", {
                x,
                y: CHART_HEIGHT - totalH,
                width: BAR_WIDTH,
                height: errorH,
              })
              : null,
            // Invisible full-height hit area so the tooltip works even for
            // empty buckets and tiny bars.
            m("rect.bar-chart__hit", {
              x: i * STEP,
              y: 0,
              width: STEP,
              height: CHART_HEIGHT,
            }, m("title", b.title)),
          ]);
        }),
      ),
    ]);
  },
};

/**
 * A compact line chart (sparkline) of a single series over time, e.g. average
 * duration. Uses a non-scaling stroke so the line stays crisp regardless of how
 * far the SVG is stretched.
 * @type {Object}
 */
export const Sparkline = {
  /**
   * @param {Object} vnode - Mithril vnode
   * @param {Array<{value:number,title:string}>} vnode.attrs.points
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const points = vnode.attrs.points || [];
    const maxVal = Math.max(1, ...points.map((p) => p.value));
    const width = Math.max((points.length - 1) * STEP, STEP);

    const coords = points.map((p, i) => {
      const x = i * STEP;
      const y = CHART_HEIGHT - (p.value / maxVal) * CHART_HEIGHT;
      return [x, y];
    });

    const line = coords.map(([x, y]) => `${x},${y}`).join(" ");
    // Close the area down to the baseline for a subtle fill.
    const area = coords.length > 0
      ? `${coords[0][0]},${CHART_HEIGHT} ${line} ${
        coords[coords.length - 1][0]
      },${CHART_HEIGHT}`
      : "";

    return m(".chart.sparkline", [
      m(
        "svg.chart__svg",
        {
          viewBox: `0 0 ${width} ${CHART_HEIGHT}`,
          preserveAspectRatio: "none",
          role: "img",
        },
        [
          coords.length > 0
            ? m("polygon.sparkline__area", { points: area })
            : null,
          coords.length > 0
            ? m("polyline.sparkline__line", {
              points: line,
              "vector-effect": "non-scaling-stroke",
            })
            : null,
          // Per-point hover targets for tooltips.
          points.map((p, i) =>
            m("rect.sparkline__hit", {
              x: i * STEP - STEP / 2,
              y: 0,
              width: STEP,
              height: CHART_HEIGHT,
            }, m("title", p.title))
          ),
        ],
      ),
    ]);
  },
};
