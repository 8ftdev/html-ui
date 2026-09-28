import { test, expect } from "@playwright/test";
import { execFileSync } from "node:child_process";
import { resolve } from "node:path";
import ts from "typescript";

declare global {
	interface Window {
		primitives: {
			accordion: typeof import("../.test-output/accordion");
			checkbox: typeof import("../.test-output/checkbox");
			input: typeof import("../.test-output/input");
			slider: typeof import("../.test-output/slider");
			progress: typeof import("../.test-output/progress");
			dialog: typeof import("../.test-output/dialog");
		};
		content(text: string): HTMLSpanElement;
		toggles: number;
	}
}

// Only execute this project's fresh CLI output. Adapter AST checks never execute input.
let script: string;
test.beforeAll(() => {
	script = "window.primitives = {};";
	for (const name of ["accordion", "checkbox", "input", "slider", "progress", "dialog"]) {
		const source = execFileSync(resolve(".test-output/html-ui-browser"), [name], { encoding: "utf8" });
		const { outputText } = ts.transpileModule(source, {
			compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
		});
		script += `(function(exports) { ${outputText}\n })(window.primitives[${JSON.stringify(name)}] = {});`;
	}
	script += `window.content = (text) => Object.assign(document.createElement('span'), { textContent: text });`;
});

test.beforeEach(async ({ page }) => {
	await page.addScriptTag({ content: script });
});

test("accordion initializes booleans, slot scope and child order", async ({ page }) => {
	const result = await page.evaluate(() => {
		let scope;
		const root = window.primitives.accordion.accordion({ open: false }, {
			summary: () => window.content("Heading"),
			content: (value) => { scope = value; return window.content("Body"); },
		});
		return { tag: root.tagName, open: root.hasAttribute("open"), name: root.hasAttribute("name"), scope,
			children: Array.from(root.children, (child) => [child.tagName, child.textContent]) };
	});
	expect(result).toEqual({ tag: "DETAILS", open: false, name: false, scope: { open: false },
		children: [["SUMMARY", "Heading"], ["SPAN", "Body"]] });
});

test("accordion preserves group and open state with optional content omitted", async ({ page }) => {
	await page.evaluate(() => document.body.append(window.primitives.accordion.accordion(
		{ name: "group", open: true }, { summary: () => window.content("Heading") },
	)));
	await expect(page.locator("details")).toHaveAttribute("name", "group");
	await expect(page.locator("details")).toHaveJSProperty("open", true);
	await expect(page.locator("details > *")).toHaveCount(1);
});

test("named accordion group keeps only the activated disclosure open", async ({ page }) => {
	await page.evaluate(() => {
		for (const title of ["First", "Second"]) document.body.append(window.primitives.accordion.accordion(
			{ name: "group" }, { summary: () => window.content(title) },
		));
	});
	await page.locator("summary").nth(0).click();
	await expect(page.locator("details").nth(0)).toHaveJSProperty("open", true);
	await page.locator("summary").nth(1).click();
	await expect(page.locator("details").nth(0)).toHaveJSProperty("open", false);
	await expect(page.locator("details").nth(1)).toHaveJSProperty("open", true);
});

test("summary keyboard activation toggles and notifies listeners", async ({ page }) => {
	await page.evaluate(() => {
		window.toggles = 0;
		const root = window.primitives.accordion.accordion({}, { summary: () => window.content("Heading") });
		root.addEventListener("toggle", () => window.toggles++);
		document.body.append(root);
	});
	await page.locator("summary").focus();
	await page.keyboard.press("Enter");
	await expect(page.locator("details")).toHaveJSProperty("open", true);
	await expect.poll(() => page.evaluate(() => window.toggles)).toBeGreaterThan(0);
});

test("checkbox false booleans and native activation", async ({ page }) => {
	await page.evaluate(() => document.body.append(window.primitives.checkbox.checkbox(
		{ checked: false, disabled: false }, { label: () => window.content("Check") },
	)));
	const input = page.getByRole("checkbox");
	await expect(input).not.toBeChecked();
	await expect(input).toBeEnabled();
	await input.click();
	await expect(input).toBeChecked();
});

test("omitted progress value stays indeterminate", async ({ page }) => {
	await page.evaluate(() => document.body.append(window.primitives.progress.progress({ label: "Loading" }, {})));
	const progress = page.getByRole("progressbar", { name: "Loading" });
	await expect(progress).not.toHaveAttribute("value");
	await expect(progress).toHaveJSProperty("position", -1);
});

test("dialog commands open modally, focus, close and restore focus", async ({ page }) => {
	await page.evaluate(() => document.body.append(window.primitives.dialog.dialog(
		{ id: "example-dialog", titleId: "example-title" }, {
			trigger: () => window.content("Open"), title: () => window.content("Title"), close: () => window.content("Close"),
		},
	)));
	const trigger = page.getByRole("button", { name: "Open", exact: true });
	const modal = page.locator("dialog");
	const close = modal.getByRole("button", { name: "Close", includeHidden: true });
	await expect(trigger).toHaveAttribute("command", "show-modal");
	await expect(trigger).toHaveAttribute("commandfor", "example-dialog");
	await expect(close).toHaveAttribute("command", "close");
	await expect(close).toHaveAttribute("commandfor", "example-dialog");
	await expect(modal).toHaveAttribute("aria-labelledby", "example-title");
	await trigger.focus();
	await trigger.press("Enter");
	await expect(page.getByRole("dialog", { name: "Title" })).toBeVisible();
	await expect(page.locator("dialog:modal")).toHaveCount(1);
	await expect(close).toBeFocused();
	await close.click();
	await expect(modal).not.toBeVisible();
	await expect(trigger).toBeFocused();
	await trigger.press("Enter");
	await page.keyboard.press("Escape");
	await expect(modal).not.toBeVisible();
});

test("slider preserves initial value within custom bounds", async ({ page }) => {
	await page.evaluate(() => document.body.append(window.primitives.slider.slider(
		{ value: 150, min: 100, max: 200, step: 10 }, { label: () => window.content("Range") },
	)));
	const input = page.getByRole("slider");
	for (const [name, value] of Object.entries({ min: "100", max: "200", step: "10" })) await expect(input).toHaveAttribute(name, value);
	await expect(input).toHaveJSProperty("valueAsNumber", 150);
	await input.focus();
	await input.press("ArrowRight");
	await expect(input).toHaveJSProperty("valueAsNumber", 160);
});

test("form reset restores supplied checkbox and text defaults", async ({ page }) => {
	await page.evaluate(() => {
		const form = document.createElement("form");
		form.append(
			window.primitives.checkbox.checkbox({ checked: true }, { label: () => window.content("Check") }),
			window.primitives.input.input({ value: "seed" }, { label: () => window.content("Text") }),
		);
		const reset = document.createElement("button");
		reset.type = "reset"; reset.textContent = "Reset"; form.append(reset);
		document.body.append(form);
	});
	await page.getByRole("checkbox").uncheck();
	await page.getByRole("textbox").fill("edited");
	await page.getByRole("button", { name: "Reset" }).click();
	await expect(page.getByRole("checkbox")).toBeChecked();
	await expect(page.getByRole("textbox")).toHaveValue("seed");
});
