/// <reference lib="dom" />
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { test } from "bun:test";
import ts from "typescript";

const output = resolve(".test-output/ui-v2");
mkdirSync(output, { recursive: true });
const bin = resolve(output, "html-ui");
execFileSync("go", ["build", "-o", bin, "./cmd/html-ui"]);
const run = (name: string, version = 2) => execFileSync(bin, [name, `--contract-version=${version}`], { encoding: "utf8" });
const parse = (source: string) => ts.createSourceFile("contract.ts", source, ts.ScriptTarget.Latest, true);

// Read only literal metadata; the consumer never evaluates source or a factory.
function literal(node: ts.Expression): any {
  if (ts.isAsExpression(node)) return literal(node.expression);
  if (ts.isStringLiteral(node)) return node.text;
  if (node.kind === ts.SyntaxKind.TrueKeyword) return true;
  if (node.kind === ts.SyntaxKind.FalseKeyword) return false;
  if (ts.isArrayLiteralExpression(node)) return node.elements.map(item => literal(item as ts.Expression));
  if (ts.isObjectLiteralExpression(node)) {
    const result = Object.create(null);
    for (const item of node.properties) {
      assert(ts.isPropertyAssignment(item));
      assert(ts.isIdentifier(item.name) || ts.isStringLiteral(item.name));
      result[item.name.text] = literal(item.initializer);
    }
    return result;
  }
  throw new Error(`non-literal UI metadata: ${node.getText()}`);
}
function ui(source: string): any {
  for (const statement of parse(source).statements) {
    if (!ts.isVariableStatement(statement)) continue;
    for (const declaration of statement.declarationList.declarations) {
      if (declaration.name.getText() === "ui") {
        assert(declaration.initializer);
        return JSON.parse(JSON.stringify(literal(declaration.initializer)));
      }
    }
  }
  throw new Error("missing ui metadata");
}

test("accordion exposes actual owned parts and cross-node expansion without fabricating disabled", () => {
  const metadata = ui(run("accordion"));
  assert.equal(metadata.component, "accordion");
  assert.equal(metadata.behavior.kind, "native");
  assert.deepEqual(metadata.parts, {
    root: { node: "root", styleRole: "disclosure", state: {
      expanded: { source: { node: "root", attribute: "open", present: true } },
    } },
    content: { node: "content", styleRole: "disclosure-content", state: {
      expanded: { source: { node: "root", attribute: "open", present: true } },
    } },
    trigger: { node: "summary", styleRole: "disclosure-trigger", state: {
      expanded: { source: { node: "root", attribute: "open", present: true } },
      focusVisible: { source: { node: "summary", pseudo: "focus-visible" } },
      hover: { source: { node: "summary", pseudo: "hover" } },
    } },
  });
});

test("button disabled is native and adapter-required tabs expose only owned nodes", () => {
  assert.deepEqual(ui(run("button")).parts.root.state.disabled, { source: { node: "root", pseudo: "disabled" } });
  const tabs = ui(run("tabs"));
  assert.equal(tabs.behavior.kind, "adapter-required");
  assert.deepEqual(Object.keys(tabs.parts).sort(), ["list", "root"]);
  assert(tabs.behavior.requirements.some((text: string) => text.includes("roving focus")));
});

test("v2 preserves every native factory except stable styling markers", () => {
  const names = execFileSync(bin, ["--list"], { encoding: "utf8" }).trim().split("\n");
  for (const name of names) {
    const before = parse(run(name, 1)).statements.find(ts.isFunctionDeclaration)!;
    const after = parse(run(name)).statements.find(ts.isFunctionDeclaration)!;
    const statements = (node: ts.FunctionDeclaration) => node.body!.statements.map(item => item.getText());
    const original = statements(before);
    const updated = statements(after);
    const markers = updated.filter(text => /\.setAttribute\("data-ui(?:-part)?",/.test(text));
    const metadata = ui(run(name));
    assert.equal(markers.length, Object.keys(metadata.parts).length * 2);
    for (const [part, value] of Object.entries(metadata.parts) as [string, any][]) {
      assert(markers.includes(`${value.node}.setAttribute("data-ui", "${name}");`));
      assert(markers.includes(`${value.node}.setAttribute("data-ui-part", "${part}");`));
    }
    assert.deepEqual(updated.filter(text => !markers.includes(text)), original, name);
  }
});

test("all v2 sources and declarations type-check and reject unsupported style targets", () => {
  const names = execFileSync(bin, ["--list"], { encoding: "utf8" }).trim().split("\n");
  const sources = names.map(name => {
    const path = resolve(output, `${name}.ts`);
    writeFileSync(path, run(name));
    return path;
  });
  const consumer = resolve(output, "consumer.ts");
  writeFileSync(consumer, `
import type { AccordionClasses } from './accordion';
import type { ButtonClasses } from './button';
const classes: AccordionClasses = { trigger: { base: 'p-4', state: { expanded: 'bg-accent', hover: { mode: 'replace', value: 'bg-muted' }, focusVisible: { mode: 'omit' } } } };
const objects: AccordionClasses<{ color: string }> = { trigger: { state: { hover: { color: 'red' } }, unstyled: true } };
const disabled: ButtonClasses = { root: { state: { disabled: 'opacity-50' } } };
const content: AccordionClasses = { content: { state: { expanded: 'p-4' } } };
// @ts-expect-error unknown part
const missingPart: AccordionClasses = { panel: { base: 'p-4' } };
// @ts-expect-error native summary has no disabled behavior
const missingState: AccordionClasses = { trigger: { state: { disabled: 'opacity-50' } } };
// @ts-expect-error replacement requires a payload
const badReplace: AccordionClasses = { root: { base: { mode: 'replace' } } };
// @ts-expect-error default payload is a class string
const badPayload: AccordionClasses = { trigger: { base: 42 } };
void [classes, content, objects, disabled, missingPart, missingState, badReplace, badPayload];
`);
  const options: ts.CompilerOptions = {
    strict: true, types: [], target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ES2022,
    lib: ["lib.es2022.d.ts", "lib.dom.d.ts"], declaration: true, emitDeclarationOnly: true,
    noEmitOnError: true, outDir: resolve(output, "types"),
  };
  const program = ts.createProgram([...sources, consumer], options);
  const diagnostics = ts.getPreEmitDiagnostics(program);
  assert.equal(diagnostics.length, 0, ts.formatDiagnosticsWithColorAndContext(diagnostics, {
    getCanonicalFileName: x => x, getCurrentDirectory: () => process.cwd(), getNewLine: () => "\n",
  }));
  assert.equal(program.emit().emitSkipped, false);
  const declarations = ts.createProgram(names.map(name => resolve(output, "types", `${name}.d.ts`)), { ...options, noEmit: true });
  assert.equal(ts.getPreEmitDiagnostics(declarations).length, 0);
});
