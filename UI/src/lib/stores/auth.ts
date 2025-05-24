import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { User } from '$lib/types';

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
		init: async () => {
			if (browser) {
				try {
					// Check if user is authenticated by calling the API
					const response = await fetch('/api/v1/auth/me', {
						credentials: 'include', // Include cookies
						headers: {
							'Accept': 'application/json',
						}
					});

					if (response.ok) {
						const userData = await response.json();
						set({
							user: userData,
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
				} catch (error) {
					console.error('Failed to check authentication status:', error);
					set({
						user: null,
						isAuthenticated: false,
						isLoading: false
					});
				}
			}
		},
		logout: async () => {
			if (browser) {
				try {
					await fetch('/api/v1/auth/logout', {
						method: 'POST',
						credentials: 'include',
						headers: {
							'Accept': 'application/json',
							'Content-Type': 'application/json'
						}
					});
				} catch (error) {
					console.error('Logout error:', error);
				} finally {
					// Clear auth state regardless of API call result
					set({
						user: null,
						isAuthenticated: false,
						isLoading: false
					});
					// Redirect to login
					window.location.href = '/auth/login';
				}
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
		setUser: (userData: User) => {
			set({
				user: userData,
				isAuthenticated: true,
				isLoading: false
			});
		},
		setLoading: (loading: boolean) => {
			update(state => ({ ...state, isLoading: loading }));
		}
	};
}

export const auth = createAuthStore();
