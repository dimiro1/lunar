/**
 * @fileoverview Language-agnostic Monaco hover + completion registration for
 * Lunar functions. The per-language API data lives in editor-<lang>-api.js;
 * this module wires any of those configs into Monaco.
 */

import { luaEditorApi } from "./editor-lua-api.js";
import { starlarkEditorApi } from "./editor-starlark-api.js";

/**
 * Editor configs keyed by Lunar function language.
 * @type {Object.<string, import('./editor-lua-api.js').EditorLanguageConfig>}
 */
const CONFIGS = {
  [luaEditorApi.language]: luaEditorApi,
  [starlarkEditorApi.language]: starlarkEditorApi,
};

/**
 * Resolves a Lunar function language to its editor config, defaulting to Lua.
 * @param {string} language
 * @returns {import('./editor-lua-api.js').EditorLanguageConfig}
 */
const configFor = (language) => CONFIGS[language] || luaEditorApi;

/**
 * Returns the Monaco language id to use for a Lunar function language. Starlark
 * has no native Monaco grammar, so it maps to "python".
 * @param {string} language - "lua" or "starlark"
 * @returns {string} Monaco language id
 */
export function monacoLanguageFor(language) {
  return configFor(language).monacoLanguage;
}

// Monaco languages that already have providers registered, so we register each
// grammar's hover/completion providers exactly once per page.
const registered = new Set();

/**
 * Registers hover and completion providers for the given function language.
 * Safe to call repeatedly; registration happens once per Monaco grammar.
 * Requires window.monaco to be loaded.
 * @param {string} language - "lua" or "starlark"
 */
export function registerEditorProviders(language) {
  if (!window.monaco) return;

  const config = configFor(language);
  if (registered.has(config.monacoLanguage)) return;
  registered.add(config.monacoLanguage);

  monaco.languages.registerHoverProvider(
    config.monacoLanguage,
    buildHoverProvider(config),
  );
  monaco.languages.registerCompletionItemProvider(
    config.monacoLanguage,
    buildCompletionProvider(config),
  );
}

/**
 * Builds a Monaco hover provider that shows the signature and description for a
 * `module.member` symbol under the cursor.
 * @param {import('./editor-lua-api.js').EditorLanguageConfig} config
 * @returns {Object} Monaco HoverProvider
 */
function buildHoverProvider(config) {
  return {
    provideHover: (model, position) => {
      const word = model.getWordAtPosition(position);
      if (!word) return null;

      const line = model.getLineContent(position.lineNumber);
      const beforeWord = line.substring(0, word.startColumn - 1);
      const match = beforeWord.match(/(\w+)\.$/);
      if (!match) return null;

      const fullName = `${match[1]}.${word.word}`;
      const doc = config.docs[fullName];
      if (!doc) return null;

      return {
        contents: [
          { value: `**${fullName}**` },
          { value: `\`\`\`${config.hoverFence}\n${doc.signature}\n\`\`\`` },
          { value: doc.description },
        ],
      };
    },
  };
}

/**
 * Builds a Monaco completion provider from the config's API docs and snippets.
 * @param {import('./editor-lua-api.js').EditorLanguageConfig} config
 * @returns {Object} Monaco CompletionItemProvider
 */
function buildCompletionProvider(config) {
  return {
    provideCompletionItems: () => {
      const asSnippet =
        monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet;

      const suggestions = Object.entries(config.docs).map(([name, doc]) => ({
        label: name,
        kind: monaco.languages.CompletionItemKind.Method,
        insertText: doc.snippet,
        insertTextRules: asSnippet,
        documentation: doc.description,
      }));

      for (const snippet of config.snippets) {
        suggestions.push({
          label: snippet.label,
          kind: monaco.languages.CompletionItemKind.Snippet,
          insertText: snippet.insertText,
          insertTextRules: asSnippet,
          documentation: snippet.documentation,
        });
      }

      return { suggestions };
    },
  };
}
