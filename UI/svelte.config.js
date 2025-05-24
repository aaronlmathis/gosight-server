import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Use static adapter for Go backend integration
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',  // Use index.html as fallback for client-side routing
			precompress: false,
			strict: false  // Don't fail on dynamic routes
		}),
		// Don't try to prerender any routes except index page
		prerender: {
			entries: ['/'],
			handleHttpError: ({ path, referrer, message }) => {
				// Ignore prerendering errors
				console.warn(`Ignoring prerender error for path: ${path}`);
				return;
			}
		},
		paths: {
			base: '',
			assets: ''
		}
	}
};

export default config;
