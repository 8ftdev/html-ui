import { defineConfig } from "@playwright/test";

export default defineConfig({
	testDir: "./tests",
	testMatch: "**/*.pw.ts", // Bun's test discovery leaves browser tests to Playwright.
	globalSetup: "./tests/browser-setup.mjs",
	workers: 1,
	use: { headless: true, trace: "retain-on-failure" },
	projects: ["chromium", "firefox", "webkit"].map((browserName) => ({
		name: browserName,
		use: { browserName: browserName as "chromium" | "firefox" | "webkit" },
	})),
});
