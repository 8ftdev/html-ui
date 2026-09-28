import { highlightHTML } from "@speed-highlight/core";
import entries from "../.test-output/preview/registry";

function element(markup: string): HTMLElement {
	const template = document.createElement("template");
	template.innerHTML = markup;
	return template.content.firstElementChild as HTMLElement;
}
function text(value: string): HTMLElement {
	const span = document.createElement("span");
	span.textContent = value;
	return span;
}

// Format only the displayed source; keep the live example's whitespace intact.
function formatHTML(node: Node, depth = 0): string {
	const indent = "  ".repeat(depth);
	if (!(node instanceof Element)) {
		const container = document.createElement("div");
		container.append(node.cloneNode(true));
		return indent + container.innerHTML;
	}
	// Whitespace inside these elements is meaningful; retain their serialization.
	if (["PRE", "TEXTAREA", "SCRIPT", "STYLE"].includes(node.tagName)) return indent + node.outerHTML;
	const shell = (node.cloneNode(false) as Element).outerHTML;
	const closing = `</${node.localName}>`;
	if (!shell.endsWith(closing)) return indent + shell; // HTML void element.
	return [
		indent + shell.slice(0, -closing.length),
		...Array.from(node.childNodes, (child) => formatHTML(child, depth + 1)),
		indent + closing,
	].join("\n");
}

function sample(recipe) {
	const name = recipe.Name;
	const id = `sample-${name}`;
	const props: Record<string, unknown> = {};
	for (const prop of recipe.Props) {
		if (prop.Default !== undefined) props[prop.Name] = prop.Default;
		else if (prop.Required) props[prop.Name] = prop.Type === "number" ? 0.65 : `${id}-${prop.Name}`;
	}
	Object.assign(props, {
		id, titleId: `${id}-title`, descriptionId: `${id}-description`, inputId: `${id}-input`, listId: `${id}-list`,
		label: recipe.Title, name: id,
	});
	if (name === "avatar") Object.assign(props, { alt: "Sample avatar", src: "data:image/svg+xml," + encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" width="48" height="48"><rect width="48" height="48" rx="24" fill="#dce6de"/><text x="24" y="30" text-anchor="middle" font-family="sans-serif" font-size="18" fill="#264834">UI</text></svg>') });
	if (["progress", "meter"].includes(name)) props.value = 0.65;
	if (name === "slider") Object.assign(props, { value: 40, min: 0, max: 100, step: 10 });
	if (name === "input") props.placeholder = "Type something…";
	if (name === "number-field") props.value = 3;
	if (name === "otp-field") Object.assign(props, { placeholder: "123456", maxLength: 6 });
	const slots = {};
	for (const slot of recipe.Slots) slots[slot.Name] = () => {
		if (slot.Name === "options") {
			if (name === "select") return element('<optgroup label="Fruit"><option>Apple</option><option>Pear</option><option>Plum</option></optgroup>');
			if (name === "autocomplete") return element('<option value="Apple"></option>');
			return element('<div role="option" id="sample-combobox-option">Apple</div>');
		}
		if (slot.Name === "control") return element(`<input id="${id}" aria-describedby="${id}-description" placeholder="Your name">`);
		if (slot.Name === "tabs") return element('<button role="tab" id="sample-tabs-tab" aria-controls="sample-tabs-panel" aria-selected="true">Overview</button>');
		if (slot.Name === "panels") return element('<section role="tabpanel" id="sample-tabs-panel" aria-labelledby="sample-tabs-tab">Panel content</section>');
		if (slot.Name === "content") {
			if (["checkbox-group", "radio-group"].includes(name)) return element(`<label><input type="${name === "radio-group" ? "radio" : "checkbox"}" name="${id}"> First choice</label>`);
			if (name === "form") return element('<label>Email <input type="email" placeholder="you@example.com"></label>');
			if (["menu", "context-menu"].includes(name)) return element('<button role="menuitem" type="button">Edit item</button>');
			if (name === "menubar") return element('<button role="menuitem" type="button">File</button>');
			if (["toolbar", "toggle-group"].includes(name)) return element('<button type="button" aria-pressed="false">Bold</button>');
			if (name === "navigation-menu") return element('<a href="#accordion">Accordion example</a>');
			const body = text("Example content supplied through a slot.");
			if (name === "alert-dialog") body.id = `${id}-description`;
			return body;
		}
		return text({ summary: "Show details", trigger: `Open ${recipe.Title.toLowerCase()}`, title: recipe.Title,
			close: "Close", label: recipe.Title, legend: "Choose an option", description: "A short description of this field." }[slot.Name] ?? slot.Name);
	};
	return { props, slots };
}

async function render() {
	const list = document.querySelector("main")!;
	for (const { recipe, factory } of entries) {
		const row = document.createElement("section");
		row.className = "preview-row";
		row.id = recipe.Name;
		const example = document.createElement("div");
		example.className = "example-column";
		const heading = document.createElement("h2");
		heading.textContent = recipe.Title;
		const description = document.createElement("p");
		description.className = "description";
		description.textContent = recipe.Summary;
		const status = document.createElement("p");
		status.className = "behavior";
		status.textContent = recipe.Status === "native" ? "Native behavior" : "Structure only · adapter behavior required";
		const stage = document.createElement("div");
		stage.className = "sample";
		const { props, slots } = sample(recipe);
		const root = factory(props, slots);
		stage.append(root);
		example.append(heading, description, status, stage);
		const codeColumn = document.createElement("div");
		codeColumn.className = "code-column";
		const code = document.createElement("div");
		code.className = "shj-lang-html shj-block";
		code.setAttribute("tabindex", "0");
		code.setAttribute("aria-label", `${recipe.Title} initial HTML`);
		code.innerHTML = await highlightHTML(formatHTML(root), "html");
		codeColumn.append(code);
		row.append(example, codeColumn);
		list.append(row);
	}
	document.querySelector("#count")!.textContent = `${entries.length} primitives`;
	document.body.dataset.ready = "true";
}
render().catch((error) => {
	document.querySelector("#count")!.textContent = `Preview failed: ${error.message}`;
	console.error(error);
});
