import { defineConfig } from "rolldown";

export default defineConfig({
  input: "client/src/extension.ts",
  external: ["vscode"],
  output: {
    file: "dist/extension.js",
    format: "cjs",
  },
});
