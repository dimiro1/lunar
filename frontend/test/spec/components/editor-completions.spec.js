/**
 * @fileoverview Tests for the language-agnostic editor completion wiring and the
 * per-language API doc configs.
 */

import { monacoLanguageFor } from "../../../js/components/editor-completions.js";
import { luaEditorApi } from "../../../js/components/editor-lua-api.js";
import { starlarkEditorApi } from "../../../js/components/editor-starlark-api.js";

describe("editor completions", () => {
  describe("monacoLanguageFor", () => {
    it("maps lua to the lua Monaco grammar", () => {
      expect(monacoLanguageFor("lua")).toBe("lua");
    });

    it("maps starlark to the python Monaco grammar", () => {
      expect(monacoLanguageFor("starlark")).toBe("python");
    });

    it("defaults unknown languages to lua", () => {
      expect(monacoLanguageFor(undefined)).toBe("lua");
      expect(monacoLanguageFor("cobol")).toBe("lua");
    });
  });

  describe("language configs", () => {
    const configs = [luaEditorApi, starlarkEditorApi];

    it("expose the same set of documented symbols", () => {
      const luaKeys = Object.keys(luaEditorApi.docs).sort();
      const starlarkKeys = Object.keys(starlarkEditorApi.docs).sort();
      expect(starlarkKeys).toEqual(luaKeys);
    });

    configs.forEach((config) => {
      it(`(${config.language}) documents core modules`, () => {
        for (const key of ["log.info", "kv.get", "http.get", "json.encode"]) {
          expect(config.docs[key]).toBeDefined();
          expect(config.docs[key].signature).toEqual(jasmine.any(String));
          expect(config.docs[key].snippet).toEqual(jasmine.any(String));
        }
      });

      it(`(${config.language}) ships a handler snippet`, () => {
        const labels = config.snippets.map((s) => s.label);
        expect(labels).toContain("handler");
      });
    });

    it("uses dialect-appropriate handler snippets", () => {
      const lua = luaEditorApi.snippets.find((s) => s.label === "handler");
      const star = starlarkEditorApi.snippets.find((s) =>
        s.label ===
          "handler"
      );
      expect(lua.insertText).toContain("function handler(ctx, event)");
      expect(star.insertText).toContain("def handler(ctx, event):");
    });
  });
});
