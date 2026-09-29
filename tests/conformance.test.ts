/// <reference lib="dom" />
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "bun:test";
import ts from "typescript";
import { projectAccordion } from "./project-accordion.mjs";

const output = resolve(".test-output");
mkdirSync(output, { recursive: true });
const bin = resolve(output, "html-ui");
execFileSync("go", ["build", "-o", bin, "./cmd/html-ui"]);
const run = (...args) => execFileSync(bin, args, { encoding: "utf8" });
const names = run("--list").trim().split("\n");
const files = names.map((name) => {
	const path = resolve(output, `${name}.ts`);
	writeFileSync(path, run(name));
	return path;
});
writeFileSync(
	resolve(output, "consumer.ts"),
	`
import { accordion, type AccordionSlots, type AccordionProps } from './accordion.js';
import type { SliderProps } from './slider.js';
// @ts-expect-error required does not apply to range inputs
const invalidRange: SliderProps = { required: true };
type RendererContent = string | { render(): void } | readonly RendererContent[];
const slots: AccordionSlots<RendererContent> = {
  summary: () => 'Heading',
  content: ({ open }) => [open ? 'Expanded' : 'Collapsed'],
};
const props: AccordionProps = {};
// @ts-expect-error required summary slot must remain required
const missing: AccordionSlots = {};
// @ts-expect-error native factory requires HTMLElement content
accordion(props, { summary: () => 'Heading' });
// @ts-expect-error open must be boolean
const bad: AccordionProps = { open: 'false' };
const nativeSlots: AccordionSlots = { summary: () => document.createElement('span') };
const root: HTMLDetailsElement = accordion({ open: false }, nativeSlots);
void [slots, root, missing, bad];
`,
);
const options = {
	strict: true,
	types: [], // Check generated browser modules without ambient Node/Bun dependency types.
	target: ts.ScriptTarget.ES2022,
	module: ts.ModuleKind.ES2022,
	lib: ["lib.es2022.d.ts", "lib.dom.d.ts"],
	declaration: true,
	emitDeclarationOnly: true,
	outDir: resolve(output, "compiled"),
	noEmitOnError: true,
};
const roots = [...files, resolve(output, "consumer.ts")];
function diagnostics(program) {
	return ts.formatDiagnosticsWithColorAndContext(
		ts.getPreEmitDiagnostics(program),
		{
			getCanonicalFileName: (x) => x,
			getCurrentDirectory: () => process.cwd(),
			getNewLine: () => "\n",
		},
	);
}

test("all generated modules and declarations type-check, including slot specialization", () => {
	assert.equal(names.length, JSON.parse(readFileSync("internal/catalog/catalog.json", "utf8")).length);
	const program = ts.createProgram(roots, options);
	assert.equal(diagnostics(program), "");
	assert.equal(program.emit().emitSkipped, false);
	const declarationFiles = names.map((name) =>
		resolve(output, "compiled", `${name}.d.ts`),
	);
	const declarations = ts.createProgram(declarationFiles, {
		...options,
		noEmit: true,
	});
	assert.equal(diagnostics(declarations), "");
});

test("legacy accordion AST preserves defaults, optional props, slot scope, events and child order", () => {
	const expected = JSON.parse(
		readFileSync("tests/fixtures/accordion.json", "utf8"),
	);
	assert.deepEqual(projectAccordion(run("accordion", "--contract-version=1")), expected);
});

test("source consumer rejects unknown calls and syntax without executing source", () => {
	const source = run("accordion", "--contract-version=1");
	const attacks = [
		source.replace(
			"  return root;",
			"  globalThis.htmlUiExecuted = true;\n  return root;",
		),
		source.replace('document.createElement("details")', "getTag()"),
		source.replace(
			"props: AccordionProps,",
			"props: AccordionProps = arbitraryCall(),",
		),
		source.replace(
			"export function accordion",
			"export async function accordion",
		),
		source.replace("props: AccordionProps,", "...props: AccordionProps,"),
		source.replace(
			"if (name !== undefined) root.setAttribute",
			"root.setAttribute",
		),
		source.replace(
			"slots.summary()",
			"slots.summary(globalThis.htmlUiExecuted = true)",
		),
		source.replace("  return root;", "  for (;;) {}\n  return root;"),
		source + "\nglobalThis.htmlUiExecuted = true;\n",
		source.replace(
			"export const contractVersion = 1",
			"export const contractVersion = 999",
		),
	];
	for (const input of attacks)
		assert.throws(() => projectAccordion(input), /accordion\.ts:\d+:\d+:/);
	assert.equal(globalThis.htmlUiExecuted, undefined);
});

test("range constraints initialize before value for browser sanitization", () => {
	// Protect the source conversion contract as well as browser behavior.
	const source = ts.createSourceFile("slider.ts", run("slider"), ts.ScriptTarget.Latest, true);
	const attributes: string[] = [];
	function visit(node: ts.Node) {
		if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) &&
			node.expression.name.text === "setAttribute" && ts.isStringLiteral(node.arguments[0])) {
			attributes.push(node.arguments[0].text);
		}
		ts.forEachChild(node, visit);
	}
	visit(source);
	for (const name of ["min", "max", "step"]) {
		expect(attributes.indexOf(name)).toBeGreaterThanOrEqual(0);
		expect(attributes.indexOf("value")).toBeGreaterThan(attributes.indexOf(name));
	}
});
