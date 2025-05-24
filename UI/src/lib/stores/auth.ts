import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface User {
	id: string;
	username: string;
	email: string;
	permissions: string[];
}

export interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

const initialState: AuthState = {
	user: null,
	isAuthenticated: false,
	isLoading: true
};

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,
		init: () => {
			if (browser) {
				// Get user data injected by the server
				const userData = (window as any).__USER_DATA__;
				if (userData && userData.user) {
					set({
						user: userData.user,
						isAuthenticated: true,
						isLoading: false
					});
				} else {
					set({
						user: null,
						isAuthenticated: false,
						isLoading: false
					});
				}
			}
		},
		logout: () => {
			if (browser) {
				window.location.href = '/logout';
			}
		},
		hasPermission: (permission: string): boolean => {
			let hasPermission = false;
			const unsubscribe = subscribe((state) => {
				hasPermission = state.user?.permissions?.includes(permission) || false;
			});
			unsubscribe();
			return hasPermission;
		},
		setLoading: (loading: boolean) => {
			update(state => ({ ...state, isLoading: loading }));
		}
	};
}

export const auth = createAuthStore();
