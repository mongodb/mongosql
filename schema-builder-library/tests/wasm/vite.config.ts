import { defineConfig, type Plugin } from 'vitest/config';
import wasm from 'vite-plugin-wasm';

export default defineConfig({
    plugins: [
        wasm() as Plugin, // Enable WebAssembly
    ],
    test: {
        environment: 'node', // Works in Node.js
    },
})
