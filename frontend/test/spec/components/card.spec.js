/**
 * @fileoverview Tests for Card components.
 */

import {
  Card,
  CardContent,
  CardDivider,
  CardFooter,
  CardHeader,
  CardVariant,
  MaximizableCard,
} from "../../../js/components/card.js";

describe("Card", () => {
  describe("view()", () => {
    it("renders a div with card class", () => {
      const vnode = { attrs: {}, children: [] };
      const result = Card.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("card");
    });

    it("applies default variant (no variant class)", () => {
      const vnode = { attrs: {}, children: [] };
      const result = Card.view(vnode);

      // Default variant doesn't add a modifier class
      expect(getVnodeClass(result)).not.toContain("card--default");
    });

    it("applies danger variant class", () => {
      const vnode = { attrs: { variant: CardVariant.DANGER }, children: [] };
      const result = Card.view(vnode);

      expect(result).toHaveClass("card--danger");
    });

    it("applies success variant class", () => {
      const vnode = { attrs: { variant: CardVariant.SUCCESS }, children: [] };
      const result = Card.view(vnode);

      expect(result).toHaveClass("card--success");
    });

    it("applies padded class when padded is true", () => {
      const vnode = { attrs: { padded: true }, children: [] };
      const result = Card.view(vnode);

      expect(result).toHaveClass("card--padded");
    });

    it("preserves custom class names", () => {
      const vnode = { attrs: { class: "my-card" }, children: [] };
      const result = Card.view(vnode);

      expect(result).toHaveClass("my-card");
      expect(result).toHaveClass("card");
    });
  });
});

describe("CardVariant", () => {
  it("exports all expected variants", () => {
    expect(CardVariant.DEFAULT).toBe("default");
    expect(CardVariant.DANGER).toBe("danger");
    expect(CardVariant.SUCCESS).toBe("success");
    expect(CardVariant.WARNING).toBe("warning");
    expect(CardVariant.INFO).toBe("info");
  });
});

describe("CardHeader", () => {
  describe("view()", () => {
    it("renders a div with card__header class", () => {
      const vnode = { attrs: { title: "Test Title" }, children: [] };
      const result = CardHeader.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("card__header");
    });

    it("renders title in h3 element", () => {
      const vnode = { attrs: { title: "My Title" }, children: [] };
      const result = CardHeader.view(vnode);

      // Find h3 in the nested structure
      const wrapper = result.children[0];
      const titleElement = wrapper.children.find((c) => c && c.tag === "h3");
      expect(titleElement).toBeTruthy();
    });

    it("renders subtitle when provided", () => {
      const vnode = {
        attrs: { title: "Title", subtitle: "Subtitle text" },
        children: [],
      };
      const result = CardHeader.view(vnode);

      // When subtitle is present, a title group is created
      const wrapper = result.children[0];
      const titleGroup = wrapper.children.find((c) => c && c.tag === "div");
      expect(titleGroup).toBeTruthy();
    });

    it("applies danger variant classes", () => {
      const vnode = {
        attrs: { title: "Danger", variant: CardVariant.DANGER },
        children: [],
      };
      const result = CardHeader.view(vnode);

      expect(result).toHaveClass("card__header--danger");
    });
  });
});

describe("CardContent", () => {
  describe("view()", () => {
    it("renders a div with card__content class", () => {
      const vnode = { attrs: {}, children: ["Content here"] };
      const result = CardContent.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("card__content");
    });

    it("applies dark class when dark is true", () => {
      const vnode = { attrs: { dark: true }, children: [] };
      const result = CardContent.view(vnode);

      expect(result).toHaveClass("card__content--dark");
    });

    it("applies large class when large is true", () => {
      const vnode = { attrs: { large: true }, children: [] };
      const result = CardContent.view(vnode);

      expect(result).toHaveClass("card__content--large");
    });

    it("applies no-padding class when noPadding is true", () => {
      const vnode = { attrs: { noPadding: true }, children: [] };
      const result = CardContent.view(vnode);

      expect(result).toHaveClass("card__content--no-padding");
    });

    it("preserves custom class names", () => {
      const vnode = { attrs: { class: "custom-content" }, children: [] };
      const result = CardContent.view(vnode);

      expect(result).toHaveClass("custom-content");
    });
  });
});

describe("CardFooter", () => {
  describe("view()", () => {
    it("renders a div with card__footer class", () => {
      const vnode = { attrs: {}, children: ["Footer content"] };
      const result = CardFooter.view(vnode);

      expect(result.tag).toBe("div");
      expect(result).toHaveClass("card__footer");
    });

    it("preserves custom class names", () => {
      const vnode = { attrs: { class: "my-footer" }, children: [] };
      const result = CardFooter.view(vnode);

      expect(result).toHaveClass("my-footer");
    });
  });
});

describe("CardDivider", () => {
  describe("view()", () => {
    it("renders an hr with card__divider class", () => {
      const result = CardDivider.view();

      expect(result.tag).toBe("hr");
      expect(result).toHaveClass("card__divider");
    });
  });
});

describe("MaximizableCard", () => {
  function createVnode(attrs = {}) {
    return {
      attrs: { title: "Test", icon: "code", ...attrs },
      state: {},
      children: ["Content"],
    };
  }

  describe("view()", () => {
    it("renders Card component when not maximized", () => {
      const vnode = createVnode();
      const result = MaximizableCard.view(vnode);

      // result.tag is the Card component (an object with view method)
      expect(result.tag).toBe(Card);
    });

    it("uses internal state by default", () => {
      const vnode = createVnode();
      vnode.state.isMaximized = false;
      const result = MaximizableCard.view(vnode);

      // Should render Card component, not maximized overlay
      expect(result.tag).toBe(Card);
    });

    it("initializes internal state if undefined", () => {
      const vnode = createVnode();
      // state.isMaximized is undefined
      MaximizableCard.view(vnode);

      expect(vnode.state.isMaximized).toBe(false);
      expect(vnode.state.escapeHandler).toBeNull();
    });
  });

  describe("external state control", () => {
    it("uses isMaximized from attrs when provided", () => {
      const vnode = createVnode({ isMaximized: true });
      vnode.state.isMaximized = false; // Internal state should be ignored
      const result = MaximizableCard.view(vnode);

      // Should render maximized overlay (a div with class selector)
      const hasOverlay =
        getVnodeClass(result).includes("card-maximized-overlay") ||
        (result.tag && result.tag.includes("card-maximized-overlay"));
      expect(hasOverlay).toBe(true);
    });

    it("uses internal state when isMaximized attr is undefined", () => {
      const vnode = createVnode();
      vnode.state.isMaximized = false;
      const result = MaximizableCard.view(vnode);

      // Should render Card component, not overlay
      expect(result.tag).toBe(Card);
    });

    it("calls onToggleMaximize when provided and maximize button clicked", () => {
      const onToggleMaximize = jasmine.createSpy("onToggleMaximize");
      const vnode = createVnode({ isMaximized: false, onToggleMaximize });
      const result = MaximizableCard.view(vnode);

      // Find maximize button
      const header = result.children.find(
        (c) => c && c.tag === CardHeader,
      );
      expect(header).toBeTruthy();

      // The actions are passed to CardHeader
      const actions = header.attrs.actions;
      expect(actions).toBeTruthy();

      // Find the CardMaximizeBtn in actions
      const maximizeBtn = actions.find(
        (a) => a && a.tag && a.attrs && a.attrs.onclick,
      );
      expect(maximizeBtn).toBeTruthy();

      // Simulate click
      maximizeBtn.attrs.onclick();
      expect(onToggleMaximize).toHaveBeenCalledWith(true);
    });

    it("calls onToggleMaximize with false when restoring from maximized", () => {
      const onToggleMaximize = jasmine.createSpy("onToggleMaximize");
      const vnode = createVnode({ isMaximized: true, onToggleMaximize });
      const result = MaximizableCard.view(vnode);

      // In maximized state, find the close button
      const maximizedContent = result.children.find(
        (c) => c && getVnodeClass(c).includes("card-maximized"),
      );
      expect(maximizedContent).toBeTruthy();

      // Find close button in header
      const header = maximizedContent.children.find(
        (c) => c && getVnodeClass(c).includes("card__header"),
      );
      const closeBtn = header.children.find(
        (c) => c && c.tag === "button",
      );
      expect(closeBtn).toBeTruthy();

      // Simulate click
      closeBtn.attrs.onclick();
      expect(onToggleMaximize).toHaveBeenCalledWith(false);
    });

    it("modifies internal state when onToggleMaximize is not provided", () => {
      const vnode = createVnode();
      vnode.state.isMaximized = false;
      const result = MaximizableCard.view(vnode);

      // Find maximize button
      const header = result.children.find(
        (c) => c && c.tag === CardHeader,
      );
      const actions = header.attrs.actions;
      const maximizeBtn = actions.find(
        (a) => a && a.tag && a.attrs && a.attrs.onclick,
      );

      // Simulate click - should modify internal state
      maximizeBtn.attrs.onclick();
      expect(vnode.state.isMaximized).toBe(true);
    });
  });
});
