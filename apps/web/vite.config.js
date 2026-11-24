import {defineConfig} from 'vite';
import {readFileSync} from 'node:fs';

export default defineConfig({
  define: {__AISTD_CLI_VERSION__: JSON.stringify(readFileSync(new URL('../cli/VERSION', import.meta.url), 'utf8').trim())},
  build: {outDir: '../../dist', emptyOutDir: true}
});
