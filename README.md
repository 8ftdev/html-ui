# html-ui

An unstyled primitive **code generator written in Go**. It emits standalone TypeScript describing props, defaults, named/scoped slots, native events, and native DOM structure. Other tools can turn that source into framework components.

```sh
html-ui accordion | html-ui-vue > Accordion.vue
```

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

To build an offline preview of all 37 primitives:

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
