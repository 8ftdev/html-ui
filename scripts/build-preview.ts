import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";

const output = resolve(".test-output/preview");
mkdirSync(output, { recursive: true });
const bin = resolve(output, "html-ui");
execFileSync("go", ["build", "-o", bin, "./cmd/html-ui"]);
const catalog = JSON.parse(readFileSync("internal/catalog/catalog.json", "utf8"));
const imports: string[] = [];
const entries: string[] = [];
for (const [index, recipe] of catalog.entries()) {
	const source = execFileSync(bin, [recipe.Name], { encoding: "utf8" });
	writeFileSync(resolve(output, `${recipe.Name}.ts`), source);
	const factory = source.match(/export function (\w+)\(/)?.[1];
	if (!factory) throw new Error(`Missing factory: ${recipe.Name}`);
	imports.push(`import { ${factory} as factory${index} } from './${recipe.Name}';`);
	entries.push(`{ recipe: ${JSON.stringify(recipe)}, factory: factory${index} }`);
}
writeFileSync(resolve(output, "registry.ts"), `${imports.join("\n")}\nexport default [${entries.join(",\n")}];`);
const result = await Bun.build({ entrypoints: ["preview/main.ts"], target: "browser", format: "iife", minify: true });
if (!result.success) throw new AggregateError(result.logs, "Preview bundle failed");
const script = (await result.outputs[0].text()).replaceAll("</script", "<\\/script");
const theme = readFileSync(import.meta.resolve("@speed-highlight/core/themes/github-dark.css").replace("file://", ""), "utf8");
const template = readFileSync("preview/template.html", "utf8");
writeFileSync("preview.html", template.replace("/* HIGHLIGHT_THEME */", () => theme).replace("/* PREVIEW_SCRIPT */", () => script));
console.log(`Built ${resolve("preview.html")} (${catalog.length} primitives, fully offline)`);
