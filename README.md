# html-ui

An unstyled primitive **code generator written in Go**. It emits standalone TypeScript describing props, defaults, named/scoped slots, native events, and native DOM structure. Other tools can turn that source into framework components.

```sh
html-ui accordion | html-ui-vue > Accordion.vue
```

## Portable UI contracts

Version 2 is the default and exposes owned styling parts, native state bindings, semantic style roles, and generated TypeScript override types:

```sh
html-ui accordion > accordion.ts
```

The source additionally exports `ui`, `AccordionStyle<Style>`, and `AccordionClasses<Style>`. Owned elements receive `data-ui`/`data-ui-part` markers. No CSS, tokens, framework imports, wrappers, or interaction code are added. All 50 primitives expose their owned parts; slot-provided elements are not assumed to belong to the producer.

The intended downstream pipeline separates framework and theme selection:

```sh
# Proposed downstream tools, not included in this repository:
html-ui accordion | html-ui-react | html-ui-shadcn > accordion.tsx
```

Framework converters must preserve UI metadata and styling targets for the theme stage. The Vue Vapor converter consumes version 2 and preserves its metadata/types and element markers. Explicit `--contract-version=1` remains available for legacy consumers. See [the v2 contract](docs/contract-v2.md) for the full schema and intermediate-output requirements.

## Stdout

```typescript
export interface AccordionProps {
  name?: string;
  open?: boolean;
}
export const defaults = {
  open: false,
} as const satisfies Partial<AccordionProps>;

export interface AccordionSlots<Content = HTMLElement> {
  summary: () => Content;
  content?: (scope: { open: boolean }) => Content;
}
```

The full module also exports `contractVersion`, `AccordionEvents`, `nativeEvents`, and an `accordion(props, slots): HTMLDetailsElement` factory. The factory describes a small, predictable DOM construction tree. It initializes once; it does not implement reactive rendering. Calling it in a browser moves any supplied content elements into the result.

Converters read the TypeScript AST. They map slot names, optionality, scope parameters, and placement to their own rendering representation. The content generic has a default, **not an `extends HTMLElement` constraint**. Native invocation specializes it to `HTMLElement`; a framework converter can use another type.

## Declarations

You can emit declarations from the generated module:

```sh
html-ui accordion > accordion.ts
npx tsc accordion.ts --strict --target ES2022 --lib ES2022,DOM \
  --declaration --emitDeclarationOnly --outDir types
```

`.d.ts` output is useful for consumers; it cannot replace the `.ts` factory for source conversion.

## Testing

To build an offline preview of all 50 primitives:

```sh
bun run preview
```

Open `preview.html` in a browser. Each row shows the native example on the left and its initial rendered HTML on the right, highlighted with Speed Highlight. The page includes sample props and slots; adapter-dependent behaviors are labeled. Rebuild after changing the generator. The preview is optional and is not used by the test suite.

```sh
bun install --frozen-lockfile
bunx playwright install chromium firefox webkit
bun run test
# Include Go tests, race checks, and vet:
make check
```

`bun test` runs the TypeScript and AST conformance checks. `bun run test:browser` runs native interaction tests in Chromium, Firefox, and WebKit; `bun run test` runs both. Playwright builds the CLI and transpiles its output in memory, then injects it into isolated browser pages. No generated HTML fixture or server is needed. Happy DOM is no longer a dependency.

Generated TypeScript and declaration artifacts remain under `.test-output/`. Browser failures retain traces under `test-results/`.

See [verification and coverage](docs/verification.md).

### Local UI building blocks

The catalog includes `card` and `grid` native containers and an `icon` decorative span with a default slot for local SVG content. Their appearance belongs to UI plugins. `button` supports `ariaLabel` for icon-only accessible names. `input` binds `id`, `type` (text/email/password), and `autocomplete` to its native input, retaining its enclosing label and model/reset semantics.
