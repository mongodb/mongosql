import { defineConfig } from 'vitest/config';
import wasm from 'vite-plugin-wasm';

export default defineConfig({
    plugins: [
        wasm(), // Enable WebAssembly
    ],
    test: {
        environment: 'node', // Works in Node.js
    },
})
