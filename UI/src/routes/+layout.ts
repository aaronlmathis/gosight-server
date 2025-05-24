import type { LayoutLoad } from './$types';

// Force SPA mode - no SSR and only prerender index
export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = async ({ fetch, url }) => {
	// This will be populated with actual user data from the Go backend
	// For now, returning default structure
	return {
		title: 'GoSight',
		user: null,
		permissions: {},
		meta: {}
	};
};
