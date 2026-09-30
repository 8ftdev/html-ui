import { test, expect } from "@playwright/test";
import { execFileSync } from "node:child_process";
import { resolve } from "node:path";
import ts from "typescript";

// Only the freshly built first-party factory is executed.
let script = "";
test.beforeAll(() => {
	const source = execFileSync(resolve(".test-output/html-ui-browser"), ["select-list"], { encoding: "utf8" });
	const { outputText } = ts.transpileModule(source, {
		compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
	});
	script = `(function(exports) { ${outputText}\n })(window.selectList = {});`;
});
test.beforeEach(async ({ page }) => { await page.addScriptTag({ content: script }); });

test("external label names the visible combobox while the native proxy submits selection", async ({ page }) => {
	await page.evaluate(() => {
		const label = Object.assign(document.createElement("label"), { htmlFor: "choice", textContent: "Choice" });
		const help = Object.assign(document.createElement("span"), { id: "choice-help", textContent: "Choose a size" });
		const form = document.createElement("form");
		const group = Object.assign(document.createElement("optgroup"), { label: "Sizes" });
		for (const value of ["small", "large"]) group.append(Object.assign(document.createElement("option"), { value, textContent: value }));
		form.append((window as any).selectList.selectList(
			{ id: "choice", popupId: "choice-popup", name: "size", value: "large", required: true, ariaDescribedby: help.id, ariaInvalid: "false" },
			{ options: () => group },
		));
		document.body.append(label, help, form);
	});
	const trigger = page.getByRole("combobox", { name: "Choice" });
	await expect(trigger).toHaveAccessibleDescription("Choose a size");
	await expect(trigger).toHaveAttribute("aria-controls", "choice-popup");
	await expect(trigger).toHaveAttribute("aria-required", "true");
	await expect(trigger).toHaveAttribute("aria-invalid", "false");
	await expect(page.getByRole("combobox")).toHaveCount(1);
	const control = page.locator('[data-ui="select-list"][data-ui-part="control"]');
	await expect(control).toHaveAttribute("aria-hidden", "true");
	await expect(control).toHaveAttribute("tabindex", "-1");
	await expect(control).not.toHaveAttribute("hidden");
	await expect(control).toHaveValue("large");
	expect(await control.evaluate((node: HTMLSelectElement) => node.checkValidity())).toBe(true);
	await control.evaluate((node: HTMLSelectElement) => node.selectedIndex = -1);
	expect(await control.evaluate((node: HTMLSelectElement) => node.matches(":invalid") && !node.checkValidity())).toBe(true);
	await control.selectOption("large");
	expect(await page.locator("form").evaluate((form: HTMLFormElement) => new FormData(form).get("size"))).toBe("large");
	await trigger.click();
	expect(await page.locator("#choice-popup").evaluate(node => node.matches(":popover-open"))).toBe(true);
	// Visibility is native; the factory does not claim adapter-managed ARIA or option projection.
	await expect(trigger).toHaveAttribute("aria-expanded", "false");
	await expect(page.locator('#choice-popup > [data-ui-part="option"]')).toHaveAttribute("hidden", "");
	await expect(page.locator('#choice-popup > [data-ui-part="option"]')).toBeEmpty();
	await page.locator("form").evaluate((form: HTMLFormElement) => form.reset());
	await expect(control).toHaveValue("small");
});

test("omitted value preserves native selected options and reset while disabled applies to both controls", async ({ page }) => {
	await page.evaluate(() => {
		const form = document.createElement("form");
		const group = Object.assign(document.createElement("optgroup"), { label: "Sizes" });
		for (const value of ["small", "large"]) {
			const option = Object.assign(document.createElement("option"), { value, textContent: value });
			option.defaultSelected = value === "large";
			group.append(option);
		}
		form.append((window as any).selectList.selectList({ id: "choice", popupId: "choice-popup", name: "size" }, {
			label: () => Object.assign(document.createElement("span"), { textContent: "Size" }), options: () => group,
		}));
		form.append((window as any).selectList.selectList({ id: "disabled", popupId: "disabled-popup", name: "disabled", disabled: true, ariaLabel: "Disabled" }, {
			options: () => Object.assign(document.createElement("option"), { value: "one", textContent: "One" }),
		}));
		document.body.append(form);
	});
	await expect(page.getByRole("combobox", { name: "Size" })).toHaveAttribute("id", "choice");
	const control = page.locator('select[name="size"]');
	await expect(control).toHaveValue("large");
	await control.selectOption("small");
	await page.locator("form").evaluate((form: HTMLFormElement) => form.reset());
	await expect(control).toHaveValue("large");
	await expect(page.getByRole("combobox", { name: "Disabled" })).toBeDisabled();
	await expect(page.locator('select[name="disabled"]')).toBeDisabled();
	expect(await page.locator("form").evaluate((form: HTMLFormElement) => new FormData(form).has("disabled"))).toBe(false);
});
