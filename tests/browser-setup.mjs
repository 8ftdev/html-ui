import { execFileSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { resolve } from "node:path";

export default function setup() {
	mkdirSync(resolve(".test-output"), { recursive: true });
	execFileSync("go", ["build", "-o", resolve(".test-output/html-ui-browser"), "./cmd/html-ui"]);
}
