/**
 * @fileoverview Tests for utility functions - focused on critical functionality.
 */

import { formatUnixTimestamp, getFunctionTabs } from "../../js/utils.js";

describe("getFunctionTabs", () => {
  it("returns all 7 tabs", () => {
    const tabs = getFunctionTabs("test-id");
    expect(tabs.length).toBe(7);
  });

  it("includes correct tab ids", () => {
    const tabs = getFunctionTabs("test-id");
    const ids = tabs.map((t) => t.id);
    expect(ids).toContain("code");
    expect(ids).toContain("versions");
    expect(ids).toContain("executions");
    expect(ids).toContain("metrics");
    expect(ids).toContain("data");
    expect(ids).toContain("settings");
    expect(ids).toContain("test");
  });

  it("generates hrefs with function id", () => {
    const funcId = "my-func-123";
    const tabs = getFunctionTabs(funcId);

    tabs.forEach((tab) => {
      expect(tab.href).toContain(funcId);
    });
  });

  it("each tab has label and href", () => {
    const tabs = getFunctionTabs("test-id");

    tabs.forEach((tab) => {
      expect(tab.label).toBeTruthy();
      expect(tab.href).toBeTruthy();
    });
  });
});

describe("formatUnixTimestamp", () => {
  // Use a known timestamp: Nov 14, 2023 at 22:13:20 UTC
  const testTimestamp = 1700000000;

  it("returns N/A for zero timestamp", () => {
    expect(formatUnixTimestamp(0)).toBe("N/A");
  });

  it("returns N/A for null/undefined/NaN timestamp", () => {
    expect(formatUnixTimestamp(null)).toBe("N/A");
    expect(formatUnixTimestamp(undefined)).toBe("N/A");
    expect(formatUnixTimestamp(NaN)).toBe("N/A");
  });

  it("formats datetime by default", () => {
    const result = formatUnixTimestamp(testTimestamp);
    // Should contain both date and time components
    expect(result).not.toBe("N/A");
    expect(result).not.toBe("Invalid Date");
    // Result should be a non-empty string with date info
    expect(result.length).toBeGreaterThan(5);
  });

  it("formats time only", () => {
    const result = formatUnixTimestamp(testTimestamp, "time");
    expect(result).not.toBe("N/A");
    // Time format typically shorter than full datetime
    expect(result.length).toBeGreaterThan(0);
  });

  it("formats date only", () => {
    const result = formatUnixTimestamp(testTimestamp, "date");
    expect(result).not.toBe("N/A");
    expect(result.length).toBeGreaterThan(0);
  });

  it("handles explicit datetime format", () => {
    const result = formatUnixTimestamp(testTimestamp, "datetime");
    expect(result).not.toBe("N/A");
    expect(result.length).toBeGreaterThan(0);
  });

  it("converts from seconds not milliseconds", () => {
    // Timestamp in seconds = Nov 14, 2023
    // If wrongly interpreted as milliseconds, would be Jan 1970
    const result = formatUnixTimestamp(testTimestamp, "date");
    // Should NOT be from 1970
    expect(result).not.toContain("1970");
  });
});
