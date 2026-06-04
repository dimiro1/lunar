/**
 * @fileoverview Function metrics view: a per-function dashboard of execution
 * count, error rate, and duration, trended over a selectable time range. Data
 * comes from pre-aggregated metric buckets (see the GraphQL Function.metrics
 * field), so it survives the much shorter execution-retention window.
 */

import { icons } from "../icons.js";
import { API } from "../api.js";
import { formatUnixTimestamp, getFunctionTabs } from "../utils.js";
import { routes } from "../routes.js";
import { BackButton } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  StatusBadge,
} from "../components/badge.js";
import { TabContent, Tabs } from "../components/tabs.js";
import { BarChart, MetricCard, Sparkline } from "../components/metrics.js";
import { t } from "../i18n/index.js";

/**
 * @typedef {import('../types.js').LunarFunction} LunarFunction
 */

const DAY = 24 * 60 * 60;
const HOUR = 60 * 60;

/**
 * Selectable time ranges. Each maps to a window length, the granularity the
 * buckets are aggregated at, and the bucket size in seconds (used both for the
 * GraphQL query and for zero-filling gaps in the series client-side).
 * @type {Object.<string, {rangeSeconds:number, granularity:string, bucketSeconds:number}>}
 */
const RANGES = {
  "24h": { rangeSeconds: 24 * HOUR, granularity: "hour", bucketSeconds: HOUR },
  "7d": { rangeSeconds: 7 * DAY, granularity: "day", bucketSeconds: DAY },
  "30d": { rangeSeconds: 30 * DAY, granularity: "day", bucketSeconds: DAY },
};

/** Formats a millisecond duration for display. */
const formatMs = (ms) => {
  if (ms == null) return t("common.na");
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)} s`;
  return `${Math.round(ms)} ms`;
};

/** Formats a 0–1 ratio as a percentage. */
const formatPct = (rate) => `${(rate * 100).toFixed(1)}%`;

/** Formats an integer with thousands separators. */
const formatInt = (n) => Number(n || 0).toLocaleString();

/**
 * Function metrics view component.
 * @type {Object}
 */
export const FunctionMetrics = {
  /** @type {LunarFunction|null} */
  func: null,
  /** @type {Object|null} Raw metrics payload from the API. */
  metrics: null,
  /** @type {Array<Object>} Zero-filled time series for the charts. */
  series: [],
  /** @type {boolean} */
  loading: true,
  /** @type {string} Active range key into RANGES. */
  range: "7d",

  oninit: (vnode) => {
    FunctionMetrics.loadData(vnode.attrs.id);
  },

  /**
   * Loads the function and its metrics for the active range.
   * @param {string} id - Function ID
   * @returns {Promise<void>}
   */
  loadData: async (id) => {
    FunctionMetrics.loading = true;
    const preset = RANGES[FunctionMetrics.range];
    const now = Math.floor(Date.now() / 1000);
    // Align the window start to a bucket boundary so the from/to we query match
    // the keys the server groups by (integer division of unix seconds).
    const from =
      Math.floor((now - preset.rangeSeconds) / preset.bucketSeconds) *
      preset.bucketSeconds;
    try {
      const { func, metrics } = await API.functions.getWithMetrics(
        id,
        from,
        now,
        preset.granularity,
      );
      FunctionMetrics.func = func;
      FunctionMetrics.metrics = metrics;
      FunctionMetrics.series = FunctionMetrics.buildSeries(
        metrics,
        from,
        now,
        preset.bucketSeconds,
      );
    } catch (e) {
      console.error("Failed to load metrics:", e);
    } finally {
      FunctionMetrics.loading = false;
      m.redraw();
    }
  },

  /**
   * Expands the sparse bucket list returned by the API into a dense, ordered
   * series covering every bucket in [from, to), filling gaps (hours/days with no
   * executions) with zeros so the charts have a continuous time axis.
   * @param {Object} metrics - API metrics payload
   * @param {number} from - Window start (unix seconds, bucket-aligned)
   * @param {number} to - Window end (unix seconds)
   * @param {number} bucketSeconds - Bucket size
   * @returns {Array<Object>} Dense series, oldest first
   */
  buildSeries: (metrics, from, to, bucketSeconds) => {
    const byStart = {};
    ((metrics && metrics.buckets) || []).forEach((b) => {
      byStart[b.bucketStart] = b;
    });
    const series = [];
    for (let s = from; s < to; s += bucketSeconds) {
      const b = byStart[s];
      series.push({
        start: s,
        count: b ? b.count : 0,
        errorCount: b ? b.errorCount : 0,
        avgDurationMs: b ? b.avgDurationMs : 0,
      });
    }
    return series;
  },

  /**
   * Switches the active range and reloads.
   * @param {string} range - Range key into RANGES
   */
  selectRange: (range) => {
    if (range === FunctionMetrics.range) return;
    FunctionMetrics.range = range;
    FunctionMetrics.loadData(FunctionMetrics.func.id);
  },

  /**
   * Renders the range selector segmented control.
   * @returns {Object} Mithril vnode
   */
  rangeSelector: () =>
    m(
      ".metrics-range",
      { role: "group", "aria-label": t("metrics.rangeLabel") },
      Object.keys(RANGES).map((key) =>
        m(
          "button.metrics-range__btn",
          {
            class: key === FunctionMetrics.range
              ? "metrics-range__btn--active"
              : "",
            onclick: () => FunctionMetrics.selectRange(key),
          },
          t(`metrics.range.${key}`),
        )
      ),
    ),

  /**
   * Renders the function metrics view.
   * @returns {Object} Mithril vnode
   */
  view: () => {
    if (FunctionMetrics.loading && !FunctionMetrics.func) {
      return m(".loading", [
        m.trust(icons.spinner()),
        m("p", t("functions.loadingFunction")),
      ]);
    }

    if (!FunctionMetrics.func) {
      return m(
        ".fade-in",
        m(Card, m(CardContent, t("common.functionNotFound"))),
      );
    }

    const func = FunctionMetrics.func;
    const metrics = FunctionMetrics.metrics;
    const summary = (metrics && metrics.summary) || {
      count: 0,
      errorCount: 0,
      errorRate: 0,
      avgDurationMs: 0,
      maxDurationMs: 0,
    };
    const granularity = (metrics && metrics.granularity) ||
      RANGES[FunctionMetrics.range].granularity;
    const tsFormat = granularity === "hour" ? "datetime" : "date";
    const hasData = summary.count > 0;

    return m(".fade-in", [
      // Header (mirrors the other function detail tabs).
      m(".function-details-header", [
        m(".function-details-left", [
          m(BackButton, { href: routes.functions() }),
          m(".function-details-divider"),
          m(".function-details-info", [
            m("h1.function-details-title", [
              func.name,
              m(IDBadge, { id: func.id }),
              m(
                Badge,
                {
                  variant: BadgeVariant.OUTLINE,
                  size: BadgeSize.SM,
                  mono: true,
                },
                `v${func.active_version.version}`,
              ),
            ]),
            m(
              "p.function-details-description",
              func.description || t("common.noDescription"),
            ),
          ]),
        ]),
        m(".function-details-actions", [
          m(StatusBadge, { enabled: !func.disabled, glow: true }),
        ]),
      ]),

      m(Tabs, { tabs: getFunctionTabs(func.id), activeTab: "metrics" }),

      m(TabContent, [
        m(".metrics-tab-container", [
          // Range selector.
          m(".metrics-toolbar", [
            m("h2.metrics-toolbar__title", t("metrics.title")),
            FunctionMetrics.rangeSelector(),
          ]),

          // Summary cards.
          m(".metrics-summary", [
            m(MetricCard, {
              label: t("metrics.totalExecutions"),
              value: formatInt(summary.count),
            }),
            m(MetricCard, {
              label: t("metrics.errorRate"),
              value: formatPct(summary.errorRate),
              sublabel: t("metrics.errorCount", { count: summary.errorCount }),
              variant: summary.errorCount > 0 ? "danger" : "success",
            }),
            m(MetricCard, {
              label: t("metrics.avgDuration"),
              value: formatMs(summary.avgDurationMs),
            }),
            m(MetricCard, {
              label: t("metrics.maxDuration"),
              value: formatMs(summary.maxDurationMs),
            }),
          ]),

          // Charts, or an empty state when nothing has run in the window.
          hasData
            ? [
              m(Card, [
                m(CardHeader, { title: t("metrics.volumeTitle") }),
                m(CardContent, [
                  m(BarChart, {
                    bars: FunctionMetrics.series.map((b) => ({
                      count: b.count,
                      errorCount: b.errorCount,
                      title: t("metrics.bucketTooltip", {
                        time: formatUnixTimestamp(b.start, tsFormat),
                        count: b.count,
                        errors: b.errorCount,
                      }),
                    })),
                  }),
                  m(".chart__legend", [
                    m(".chart__legend-item", [
                      m(".chart__legend-swatch.chart__legend-swatch--success"),
                      t("metrics.legend.success"),
                    ]),
                    m(".chart__legend-item", [
                      m(".chart__legend-swatch.chart__legend-swatch--error"),
                      t("metrics.legend.error"),
                    ]),
                  ]),
                ]),
              ]),
              m(Card, [
                m(CardHeader, { title: t("metrics.durationTitle") }),
                m(CardContent, [
                  m(Sparkline, {
                    points: FunctionMetrics.series.map((b) => ({
                      value: b.avgDurationMs,
                      title: t("metrics.durationTooltip", {
                        time: formatUnixTimestamp(b.start, tsFormat),
                        value: formatMs(b.avgDurationMs),
                      }),
                    })),
                  }),
                ]),
              ]),
            ]
            : m(Card, [
              m(CardContent, m(".metrics-empty", t("metrics.emptyState"))),
            ]),
        ]),
      ]),
    ]);
  },
};
