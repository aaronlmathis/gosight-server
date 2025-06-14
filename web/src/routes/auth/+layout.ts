import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	// Return minimal data for auth pages - don't load user data
	// since auth pages handle their own authentication state
	return {
		title: 'Authentication - GoSight'
	};
};
