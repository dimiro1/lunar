/**
 * @fileoverview Tests for Badge components.
 */

import {
  Badge,
  BadgeSize,
  BadgeVariant,
  IDBadge,
  LogLevelBadge,
  MethodBadges,
  StatusBadge,
  TriggerBadge,
} from "../../../js/components/badge.js";
import { t } from "../../../js/i18n/index.js";

describe("Badge", () => {
  describe("view()", () => {
    it("renders a span element by default", () => {
      const vnode = { attrs: {}, children: ["Label"] };
      const result = Badge.view(vnode);

      expect(result.tag).toBe("span");
    });

    it("renders an anchor when href is provided", () => {
      const vnode = { attrs: { href: "/link" }, children: ["Link"] };
      const result = Badge.view(vnode);

      expect(result.tag).toBe("a");
      expect(result.attrs.href).toBe("/link");
    });

    it("applies primary variant class by default", () => {
      const vnode = { attrs: {}, children: ["Badge"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--primary");
    });

    it("applies specified variant class", () => {
      const vnode = {
        attrs: { variant: BadgeVariant.SUCCESS },
        children: ["Success"],
      };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--success");
    });

    it("applies info variant class", () => {
      const vnode = {
        attrs: { variant: BadgeVariant.INFO },
        children: ["Info"],
      };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--info");
    });

    it("applies default size class by default", () => {
      const vnode = { attrs: {}, children: ["Badge"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--default");
    });

    it("applies specified size class", () => {
      const vnode = { attrs: { size: BadgeSize.SM }, children: ["Small"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--sm");
    });

    it("applies uppercase class when uppercase is true", () => {
      const vnode = { attrs: { uppercase: true }, children: ["LOUD"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--uppercase");
    });

    it("applies mono class when mono is true", () => {
      const vnode = { attrs: { mono: true }, children: ["CODE"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("badge--mono");
    });

    it("renders dot element when dot is true", () => {
      const vnode = { attrs: { dot: true }, children: ["Status"] };
      const result = Badge.view(vnode);

      // Find dot in children array
      const hasDot = result.children.some(
        (child) =>
          child && child.tag === "span" &&
          getVnodeClass(child).includes("badge__dot"),
      );
      expect(hasDot).toBe(true);
    });

    it("renders dot with glow when dotGlow is true", () => {
      const vnode = {
        attrs: { dot: true, dotGlow: true },
        children: ["Active"],
      };
      const result = Badge.view(vnode);

      const dotChild = result.children.find(
        (child) =>
          child && child.tag === "span" &&
          getVnodeClass(child).includes("badge__dot"),
      );
      expect(dotChild).toBeTruthy();
      expect(getVnodeClass(dotChild)).toContain("badge__dot--glow");
    });

    it("preserves custom class names", () => {
      const vnode = { attrs: { class: "my-badge" }, children: ["Custom"] };
      const result = Badge.view(vnode);

      expect(result).toHaveClass("my-badge");
      expect(result).toHaveClass("badge");
    });
  });
});

describe("BadgeVariant", () => {
  it("exports all expected variants", () => {
    expect(BadgeVariant.PRIMARY).toBe("primary");
    expect(BadgeVariant.SECONDARY).toBe("secondary");
    expect(BadgeVariant.DESTRUCTIVE).toBe("destructive");
    expect(BadgeVariant.OUTLINE).toBe("outline");
    expect(BadgeVariant.SUCCESS).toBe("success");
    expect(BadgeVariant.WARNING).toBe("warning");
    expect(BadgeVariant.INFO).toBe("info");
  });
});

describe("BadgeSize", () => {
  it("exports all expected sizes", () => {
    expect(BadgeSize.SM).toBe("sm");
    expect(BadgeSize.DEFAULT).toBe("default");
    expect(BadgeSize.LG).toBe("lg");
  });
});

describe("StatusBadge", () => {
  describe("view()", () => {
    it("renders enabled state correctly", () => {
      const vnode = { attrs: { enabled: true } };
      const result = StatusBadge.view(vnode);

      // StatusBadge returns a Badge vnode
      expect(result.tag).toBe(Badge);
      expect(result.attrs.variant).toBe(BadgeVariant.SUCCESS);
      expect(result.attrs.dot).toBe(true);
      expect(result.attrs.uppercase).toBe(true);
      expect(result.children).toContain(t("badge.enabled"));
    });

    it("renders disabled state correctly", () => {
      const vnode = { attrs: { enabled: false } };
      const result = StatusBadge.view(vnode);

      expect(result.tag).toBe(Badge);
      expect(result.attrs.variant).toBe(BadgeVariant.WARNING);
      expect(result.children).toContain(t("badge.disabled"));
    });

    it("applies glow effect when glow is true and enabled", () => {
      const vnode = { attrs: { enabled: true, glow: true } };
      const result = StatusBadge.view(vnode);

      expect(result.attrs.dotGlow).toBe(true);
      expect(result.attrs.size).toBe(BadgeSize.DEFAULT);
    });

    it("does not apply glow when disabled", () => {
      const vnode = { attrs: { enabled: false, glow: true } };
      const result = StatusBadge.view(vnode);

      expect(result.attrs.dotGlow).toBe(false);
    });
  });
});

describe("IDBadge", () => {
  describe("view()", () => {
    it("renders with ID text", () => {
      const vnode = { attrs: { id: "abc123" } };
      const result = IDBadge.view(vnode);

      expect(result.tag).toBe(Badge);
      // children contains the ID (may be string or array)
      const text = Array.isArray(result.children)
        ? result.children[0]
        : result.children;
      expect(text).toBe("abc123");
    });

    it("uses secondary variant without href", () => {
      const vnode = { attrs: { id: "test" } };
      const result = IDBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.SECONDARY);
    });

    it("uses outline variant with href", () => {
      const vnode = { attrs: { id: "test", href: "/link" } };
      const result = IDBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.OUTLINE);
      expect(result.attrs.href).toBe("/link");
    });

    it("renders with hashtag icon and mono font", () => {
      const vnode = { attrs: { id: "test" } };
      const result = IDBadge.view(vnode);

      expect(result.attrs.icon).toBe("hashtag");
      expect(result.attrs.mono).toBe(true);
      expect(result.attrs.size).toBe(BadgeSize.SM);
    });
  });
});

describe("MethodBadges", () => {
  describe("view()", () => {
    it("renders a badge for each method", () => {
      const vnode = { attrs: { methods: ["GET", "POST", "DELETE"] } };
      const result = MethodBadges.view(vnode);

      expect(result.tag).toBe("div");
      expect(result.children.length).toBe(3);
    });

    it("renders empty when no methods provided", () => {
      const vnode = { attrs: {} };
      const result = MethodBadges.view(vnode);

      expect(result.children.length).toBe(0);
    });

    it("renders badges with correct method text", () => {
      const vnode = { attrs: { methods: ["GET"] } };
      const result = MethodBadges.view(vnode);

      const badge = result.children[0];
      expect(badge.tag).toBe(Badge);
      const text = Array.isArray(badge.children)
        ? badge.children[0]
        : badge.children;
      expect(text).toBe("GET");
    });
  });
});

describe("LogLevelBadge", () => {
  describe("view()", () => {
    it("renders INFO with success variant", () => {
      const vnode = { attrs: { level: "INFO" } };
      const result = LogLevelBadge.view(vnode);

      expect(result.tag).toBe(Badge);
      expect(result.attrs.variant).toBe(BadgeVariant.SUCCESS);
      const text = Array.isArray(result.children)
        ? result.children[0]
        : result.children;
      expect(text).toBe("INFO");
    });

    it("renders WARN with warning variant", () => {
      const vnode = { attrs: { level: "WARN" } };
      const result = LogLevelBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.WARNING);
    });

    it("renders ERROR with destructive variant", () => {
      const vnode = { attrs: { level: "ERROR" } };
      const result = LogLevelBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.DESTRUCTIVE);
    });

    it("renders DEBUG with secondary variant", () => {
      const vnode = { attrs: { level: "DEBUG" } };
      const result = LogLevelBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.SECONDARY);
    });

    it("renders unknown levels with secondary variant", () => {
      const vnode = { attrs: { level: "TRACE" } };
      const result = LogLevelBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.SECONDARY);
    });
  });
});

describe("TriggerBadge", () => {
  describe("view()", () => {
    it("renders cron trigger with info variant", () => {
      const vnode = { attrs: { trigger: "cron" } };
      const result = TriggerBadge.view(vnode);

      expect(result.tag).toBe(Badge);
      expect(result.attrs.variant).toBe(BadgeVariant.INFO);
      expect(result.attrs.size).toBe(BadgeSize.SM);
      expect(result.attrs.uppercase).toBe(true);
      expect(result.attrs.mono).toBe(true);
      expect(result.children).toContain(t("executions.triggers.cron"));
    });

    it("renders http trigger with success variant", () => {
      const vnode = { attrs: { trigger: "http" } };
      const result = TriggerBadge.view(vnode);

      expect(result.tag).toBe(Badge);
      expect(result.attrs.variant).toBe(BadgeVariant.SUCCESS);
      expect(result.children).toContain(t("executions.triggers.http"));
    });

    it("renders unknown trigger as http (success variant)", () => {
      const vnode = { attrs: { trigger: "unknown" } };
      const result = TriggerBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.SUCCESS);
      expect(result.children).toContain(t("executions.triggers.http"));
    });

    it("renders undefined trigger as http", () => {
      const vnode = { attrs: {} };
      const result = TriggerBadge.view(vnode);

      expect(result.attrs.variant).toBe(BadgeVariant.SUCCESS);
      expect(result.children).toContain(t("executions.triggers.http"));
    });
  });
});
