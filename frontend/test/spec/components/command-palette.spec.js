/**
 * @fileoverview Tests for CommandPalette component - focused on critical functionality.
 */

import { CommandPalette } from "../../../js/components/command-palette.js";

describe("CommandPalette", () => {
  // Store original state to restore after each test
  let originalState;

  beforeEach(() => {
    originalState = {
      isOpen: CommandPalette.isOpen,
      query: CommandPalette.query,
      results: CommandPalette.results,
      selectedIndex: CommandPalette.selectedIndex,
      functions: CommandPalette.functions,
      loading: CommandPalette.loading,
      customItems: CommandPalette.customItems,
    };
    // Reset state before each test
    CommandPalette.isOpen = false;
    CommandPalette.query = "";
    CommandPalette.results = [];
    CommandPalette.selectedIndex = 0;
    CommandPalette.functions = [];
    CommandPalette.loading = false;
    CommandPalette.customItems = [];
  });

  afterEach(() => {
    // Restore original state
    Object.assign(CommandPalette, originalState);
  });

  describe("view", () => {
    it("returns null when not open", () => {
      CommandPalette.isOpen = false;
      const result = CommandPalette.view();
      expect(result).toBeNull();
    });

    it("returns overlay when open", () => {
      CommandPalette.isOpen = true;
      CommandPalette.results = [];
      const result = CommandPalette.view();
      expect(result).toBeTruthy();
      // Class can be in tag selector or attrs.class
      const hasOverlayClass =
        (result.tag && result.tag.includes("command-palette-overlay")) ||
        (getVnodeClass(result).includes("command-palette-overlay"));
      expect(hasOverlayClass).toBe(true);
    });
  });

  describe("close", () => {
    it("sets isOpen to false and clears state", () => {
      CommandPalette.isOpen = true;
      CommandPalette.query = "test";
      CommandPalette.results = [{ label: "item" }];

      CommandPalette.close();

      expect(CommandPalette.isOpen).toBe(false);
      expect(CommandPalette.query).toBe("");
      expect(CommandPalette.results).toEqual([]);
    });
  });

  describe("updateResults", () => {
    beforeEach(() => {
      // Set up some test functions
      CommandPalette.functions = [
        { id: "1", name: "my-function", disabled: false },
        { id: "2", name: "other-func", disabled: true },
      ];
    });

    it("shows all items when query is empty", () => {
      CommandPalette.query = "";
      CommandPalette.updateResults();

      // Should include nav items + function items
      expect(CommandPalette.results.length).toBeGreaterThan(2);
    });

    it("filters by label", () => {
      CommandPalette.query = "my-function";
      CommandPalette.updateResults();

      // Should only have items matching "my-function"
      const matching = CommandPalette.results.filter((r) =>
        r.label.toLowerCase().includes("my-function")
      );
      expect(matching.length).toBe(CommandPalette.results.length);
    });

    it("filters by description", () => {
      CommandPalette.query = "Create";
      CommandPalette.updateResults();

      // Should include nav item with description containing "Create"
      const createItem = CommandPalette.results.find((r) =>
        r.description?.toLowerCase().includes("create")
      );
      expect(createItem).toBeTruthy();
    });

    it("resets selectedIndex when out of bounds", () => {
      CommandPalette.query = "";
      CommandPalette.updateResults();
      CommandPalette.selectedIndex = 100;

      CommandPalette.updateResults();

      expect(CommandPalette.selectedIndex).toBeLessThan(100);
    });
  });

  describe("handleKeyDown", () => {
    beforeEach(() => {
      CommandPalette.results = [
        { label: "Item 1", path: "/1" },
        { label: "Item 2", path: "/2" },
        { label: "Item 3", path: "/3" },
      ];
      CommandPalette.selectedIndex = 0;
      CommandPalette.isOpen = true;
    });

    it("ArrowDown increments selectedIndex", () => {
      const event = { key: "ArrowDown", preventDefault: jasmine.createSpy() };
      CommandPalette.handleKeyDown(event);

      expect(event.preventDefault).toHaveBeenCalled();
      expect(CommandPalette.selectedIndex).toBe(1);
    });

    it("ArrowDown does not exceed results length", () => {
      CommandPalette.selectedIndex = 2;
      const event = { key: "ArrowDown", preventDefault: jasmine.createSpy() };
      CommandPalette.handleKeyDown(event);

      expect(CommandPalette.selectedIndex).toBe(2);
    });

    it("ArrowUp decrements selectedIndex", () => {
      CommandPalette.selectedIndex = 2;
      const event = { key: "ArrowUp", preventDefault: jasmine.createSpy() };
      CommandPalette.handleKeyDown(event);

      expect(event.preventDefault).toHaveBeenCalled();
      expect(CommandPalette.selectedIndex).toBe(1);
    });

    it("ArrowUp does not go below 0", () => {
      CommandPalette.selectedIndex = 0;
      const event = { key: "ArrowUp", preventDefault: jasmine.createSpy() };
      CommandPalette.handleKeyDown(event);

      expect(CommandPalette.selectedIndex).toBe(0);
    });

    it("Escape closes palette", () => {
      const event = { key: "Escape", preventDefault: jasmine.createSpy() };
      CommandPalette.handleKeyDown(event);

      expect(CommandPalette.isOpen).toBe(false);
    });
  });

  describe("custom items registration", () => {
    it("registerItems adds items with source tag", () => {
      const items = [
        { type: "custom", label: "Test", icon: "bolt", onSelect: () => {} },
      ];

      CommandPalette.registerItems("test-source", items);

      expect(CommandPalette.customItems.length).toBe(1);
      expect(CommandPalette.customItems[0].source).toBe("test-source");
      expect(CommandPalette.customItems[0].label).toBe("Test");
    });

    it("registerItems replaces items from same source", () => {
      CommandPalette.registerItems("source-a", [
        { type: "custom", label: "First" },
      ]);
      CommandPalette.registerItems("source-a", [
        { type: "custom", label: "Second" },
      ]);

      expect(CommandPalette.customItems.length).toBe(1);
      expect(CommandPalette.customItems[0].label).toBe("Second");
    });

    it("registerItems keeps items from different sources", () => {
      CommandPalette.registerItems("source-a", [
        { type: "custom", label: "A" },
      ]);
      CommandPalette.registerItems("source-b", [
        { type: "custom", label: "B" },
      ]);

      expect(CommandPalette.customItems.length).toBe(2);
    });

    it("unregisterItems removes all items from source", () => {
      CommandPalette.registerItems("source-a", [
        { type: "custom", label: "A" },
      ]);
      CommandPalette.registerItems("source-b", [
        { type: "custom", label: "B" },
      ]);

      CommandPalette.unregisterItems("source-a");

      expect(CommandPalette.customItems.length).toBe(1);
      expect(CommandPalette.customItems[0].label).toBe("B");
    });

    it("unregisterItems handles non-existent source gracefully", () => {
      CommandPalette.registerItems("source-a", [
        { type: "custom", label: "A" },
      ]);

      CommandPalette.unregisterItems("non-existent");

      expect(CommandPalette.customItems.length).toBe(1);
    });
  });

  describe("updateResults with custom items", () => {
    beforeEach(() => {
      CommandPalette.functions = [];
      CommandPalette.customItems = [
        {
          type: "custom",
          label: "Custom Action",
          icon: "bolt",
          source: "test",
        },
      ];
    });

    it("includes custom items in results", () => {
      CommandPalette.query = "";
      CommandPalette.updateResults();

      const customItem = CommandPalette.results.find(
        (r) => r.type === "custom",
      );
      expect(customItem).toBeTruthy();
      expect(customItem.label).toBe("Custom Action");
    });

    it("filters custom items by query", () => {
      CommandPalette.query = "Custom Action";
      CommandPalette.updateResults();

      const customItems = CommandPalette.results.filter(
        (r) => r.type === "custom",
      );
      expect(customItems.length).toBe(1);
      expect(customItems[0].label).toBe("Custom Action");
    });

    it("custom items appear at the beginning of results", () => {
      CommandPalette.query = "";
      CommandPalette.updateResults();

      expect(CommandPalette.results[0].type).toBe("custom");
    });
  });

  describe("selectItem with custom type", () => {
    it("calls onSelect callback for custom type", () => {
      const onSelect = jasmine.createSpy("onSelect");
      const item = { type: "custom", label: "Test", onSelect };

      CommandPalette.isOpen = true;
      CommandPalette.selectItem(item);

      expect(onSelect).toHaveBeenCalled();
      expect(CommandPalette.isOpen).toBe(false);
    });

    it("does not call onSelect if not provided", () => {
      const item = { type: "custom", label: "Test" };

      CommandPalette.isOpen = true;
      CommandPalette.selectItem(item);

      // Should not throw and should still close
      expect(CommandPalette.isOpen).toBe(false);
    });
  });
});
