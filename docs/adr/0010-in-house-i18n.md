# 0010. In-house i18n with locale modules

- Status: Accepted
- Date: 2026-06-02

## Context

The dashboard ships in more than one language (currently English and Brazilian
Portuguese) and needs to translate hundreds of UI strings, interpolate values
("{{count}} functions total"), pick a sensible default from the browser, and
remember the user's choice. The reflex solution is an off-the-shelf i18n
library, but the project's constraints push back:

- **No Node.js / no bundler** (see [ADR-0004](0004-no-nodejs-vendored-js.md)), so
  any library would have to be vendored as a buildless `<script>`/ES module.
- The need is genuinely small: a flat-ish dictionary, dot-keyed lookup, simple
  `{{param}}` interpolation, locale persistence, and a fallback chain. A general
  i18n framework brings pluralisation rules, ICU message syntax, async catalog
  loading, and other machinery we don't need.

## Decision

We will implement i18n ourselves as a small module under `frontend/js/i18n/`.

- Each locale is a plain ES module default-exporting a nested dictionary
  (`locales/en.js`, `locales/pt-BR.js`). English is the source of truth.
- `index.js` exposes an `i18n` object and a `t(key, params)` shorthand.
  Translation uses **dot-notation keys** (`nav.logout`) resolved against the
  nested object, with `{{param}}` interpolation.
- Lookup has a defined fallback chain: current locale → `DEFAULT_LOCALE` (`en`)
  → the key itself (with a `console.warn`), so a missing translation degrades
  visibly but never crashes the UI.
- The active locale is auto-detected from `navigator.language` (with a `pt*` →
  `pt-BR` heuristic) on first run, persisted in `localStorage`, and triggers
  `m.redraw()` on change so Mithril re-renders.

Adding a language is: create a `locales/<code>.js`, register it in `index.js`,
and add its display name to `localeNames`.

## Consequences

- Zero runtime dependency for translation; the i18n layer is a handful of files
  we fully control and can read end to end.
- Translations live in version control as plain data modules, reviewable in
  normal diffs, and editable by translators without tooling.
- The fallback chain makes missing keys safe and discoverable (warned in the
  console) rather than fatal.
- Trade-off: we don't get advanced features like CLDR pluralisation or gender
  rules. If a future locale needs them we will revisit rather than retrofit them
  awkwardly. There is also no automated check that locale files are *complete* —
  the runtime fallback hides gaps — so completeness currently relies on review
  (a candidate for a small lint task later).
- The i18n module is covered by the frontend Jasmine suite (see
  [ADR-0011](0011-testing-strategy.md)).
