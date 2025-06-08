import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	base: '/',
	plugins: [tailwindcss(), sveltekit()],
	resolve: {
		alias: {
			'$UI$': './src/lib/components/ui'
		// or wherever your UI components are located
		},
		extensions: ['.js', '.ts', '.svelte', '.json']
	},
	server: {
		fs: {
			allow: ['..', '../uploads']
		}
	}
});
