import { test, expect } from "@playwright/test";
import { execFileSync } from "node:child_process";
import { resolve } from "node:path";
import ts from "typescript";

// Execute only freshly built first-party factories, just as a native consumer does.
let script = "window.nativeControls = {};";
test.beforeAll(() => {
	for (const name of ["checkbox", "radio", "switch", "input", "textarea", "slider", "select", "field", "fieldset"]) {
		const source = execFileSync(resolve(".test-output/html-ui-browser"), [name], { encoding: "utf8" });
		const { outputText } = ts.transpileModule(source, {
			compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
		});
		script += `(function(exports) { ${outputText}\n })(window.nativeControls[${JSON.stringify(name)}] = {});`;
	}
});
test.beforeEach(async ({ page }) => { await page.addScriptTag({ content: script }); });

test("native controls accept external names and ARIA attributes without a label slot", async ({ page }) => {
	await page.evaluate(() => {
		const modules = (window as any).nativeControls;
		for (const name of ["checkbox", "radio", "switch", "input", "textarea", "slider", "select"]) {
			const title = document.createElement("span"); title.id = `${name}-label`; title.textContent = `${name} external`;
			const help = document.createElement("span"); help.id = `${name}-help`; help.textContent = "Helpful information";
			const props = { name, value: name === "slider" ? 50 : "value", id: `${name}-id`, ariaLabel: "Fallback name",
				ariaLabelledby: title.id, ariaDescribedby: help.id, ariaInvalid: "false" };
			const slots = name === "select" ? { options: () => Object.assign(document.createElement("option"), { value: "value", textContent: "Choice" }) } : {};
			const factory = modules[name][name === "switch" ? "switchControl" : name];
			document.body.append(title, help, factory(props, slots));
		}
	});
	for (const name of ["checkbox", "radio", "switch", "input", "textarea", "slider", "select"]) {
		const control = page.locator(`#${name}-id`);
		await expect(control).toHaveAccessibleName(`${name} external`);
		await expect(control).toHaveAccessibleDescription("Helpful information");
		await expect(control).toHaveAttribute("aria-label", "Fallback name");
		await expect(control).toHaveAttribute("aria-invalid", "false");
		await expect(control.locator("..")).not.toHaveAttribute("aria-labelledby");
	}
});

test("checkbox indicator preserves native mixed state, validation, activation and reset", async ({ page }) => {
	await page.evaluate(() => {
		const form = document.createElement("form");
		const root = (window as any).nativeControls.checkbox.checkbox(
			{ name: "agreement", value: "yes", checked: false, indeterminate: true, required: true, ariaLabel: "Agreement" }, {},
		);
		form.append(root); document.body.append(form);
	});
	const checkbox = page.getByRole("checkbox", { name: "Agreement" });
	await expect(checkbox).toHaveJSProperty("indeterminate", true);
	await expect(checkbox).not.toHaveAttribute("indeterminate");
	await expect(page.locator('[data-ui="checkbox"][data-ui-part="indicator"]')).toHaveAttribute("aria-hidden", "true");
	expect(await checkbox.evaluate((node: HTMLInputElement) => node.matches(":indeterminate") && !node.checkValidity())).toBe(true);
	await page.locator('[data-ui="checkbox"][data-ui-part="indicator"]').evaluate(node => (node as HTMLElement).click());
	await expect(checkbox).toBeChecked();
	await expect(checkbox).toHaveJSProperty("indeterminate", false);
	expect(await page.locator("form").evaluate((node: HTMLFormElement) => new FormData(node).get("agreement"))).toBe("yes");
	await page.locator("form").evaluate((node: HTMLFormElement) => node.reset());
	await expect(checkbox).not.toBeChecked();
	expect(await checkbox.evaluate((node: HTMLInputElement) => node.checkValidity())).toBe(false);
});

test("switch track and thumb leave keyboard, disabled and submission behavior native", async ({ page }) => {
	await page.evaluate(() => {
		const form = document.createElement("form");
		form.append((window as any).nativeControls.switch.switchControl({ name: "updates", value: "yes", ariaLabel: "Updates" }, {}));
		document.body.append(form);
	});
	const control = page.getByRole("switch", { name: "Updates" });
	await expect(page.locator('[data-ui="switch"][data-ui-part="track"] > [data-ui-part="thumb"]')).toHaveAttribute("aria-hidden", "true");
	await control.focus(); await control.press("Space"); await expect(control).toBeChecked();
	expect(await page.locator("form").evaluate((form: HTMLFormElement) => new FormData(form).get("updates"))).toBe("yes");
	await control.evaluate((node: HTMLInputElement) => node.disabled = true);
	await page.locator('[data-ui="switch"][data-ui-part="track"]').evaluate(node => (node as HTMLElement).click());
	await expect(control).toBeChecked();
	expect(await page.locator("form").evaluate((form: HTMLFormElement) => new FormData(form).has("updates"))).toBe(false);
});

test("radio indicators preserve group exclusivity, arrow navigation and reset baseline", async ({ page }) => {
	await page.evaluate(() => {
		const form = document.createElement("form");
		for (const value of ["one", "two"]) form.append((window as any).nativeControls.radio.radio(
			{ name: "choice", value, defaultChecked: value === "one", ariaLabel: value, required: true }, {},
		));
		document.body.append(form);
	});
	const one = page.getByRole("radio", { name: "one" }); const two = page.getByRole("radio", { name: "two" });
	await one.focus(); await one.press("ArrowRight");
	await expect(two).toBeChecked(); await expect(one).not.toBeChecked();
	expect(await page.locator("form").evaluate((form: HTMLFormElement) => new FormData(form).get("choice"))).toBe("two");
	await page.locator("form").evaluate((form: HTMLFormElement) => form.reset());
	await expect(one).toBeChecked(); await expect(two).not.toBeChecked();
	await expect(page.locator('[data-ui="radio"][data-ui-part="indicator"]').first()).toHaveAttribute("aria-hidden", "true");
});

test("select chevron preserves real options and native required validation", async ({ page }) => {
	await page.evaluate(() => {
		const group = document.createElement("optgroup"); group.label = "Choices";
		for (const value of ["one", "two"]) group.append(Object.assign(document.createElement("option"), { value, textContent: value }));
		const root = (window as any).nativeControls.select.select({ required: true, ariaLabel: "Selection" }, { options: () => group });
		root.querySelector("select")!.selectedIndex = -1;
		document.body.append(root);
	});
	const select = page.getByRole("combobox", { name: "Selection" });
	await expect(select.locator("option")).toHaveCount(2);
	await expect(page.locator('[data-ui="select"][data-ui-part="chevron"]')).toHaveAttribute("aria-hidden", "true");
	expect(await select.evaluate((node: HTMLSelectElement) => node.checkValidity())).toBe(false);
	await select.selectOption("two"); await expect(select).toHaveValue("two");
	expect(await select.evaluate((node: HTMLSelectElement) => node.checkValidity())).toBe(true);
});

test("Field links its label, description and error while disabled stays presentation only", async ({ page }) => {
	await page.evaluate(() => {
		const text = (value: string) => Object.assign(document.createElement("span"), { textContent: value });
		const root = (window as any).nativeControls.field.field(
			{ id: "email", descriptionId: "email-help", errorId: "email-error", invalid: true, disabled: true }, {
				label: () => text("Email"), description: () => text("Use your work email"), error: () => text("Enter a valid address"),
				control: ({ id }: { id: string }) => Object.assign(document.createElement("input"), { id }),
			},
		);
		root.querySelector("input")!.setAttribute("aria-describedby", "email-help email-error");
		document.body.append(root);
	});
	const input = page.getByRole("textbox", { name: "Email" });
	await expect(input).toHaveAccessibleDescription("Use your work email Enter a valid address");
	await expect(input).toBeEnabled();
	await expect(page.locator('[data-ui="field"][data-ui-part="root"]')).toHaveAttribute("data-invalid", "true");
	await expect(page.locator('[data-ui="field"][data-ui-part="root"]')).toHaveAttribute("data-disabled", "true");
	await expect(page.locator('[data-ui="field"][data-ui-part="content"] > [data-ui-part="error"]')).toHaveAttribute("id", "email-error");
});

test("Fieldset description follows the native legend without changing descendant disabling", async ({ page }) => {
	await page.evaluate(() => {
		const span = (textContent: string) => Object.assign(document.createElement("span"), { textContent });
		const root = (window as any).nativeControls.fieldset.fieldset({ disabled: true }, {
			legend: () => span("Account"), description: () => span("Account preferences"),
			content: () => Object.assign(document.createElement("input"), { id: "account-input" }),
		});
		document.body.append(root);
	});
	await expect(page.getByRole("group", { name: "Account" })).toHaveCount(1);
	await expect(page.locator("fieldset > legend:first-child")).toHaveCount(1);
	await expect(page.locator('fieldset > legend + [data-ui-part="description"]')).toHaveText("Account preferences");
	await expect(page.locator("#account-input")).toBeDisabled();
});
