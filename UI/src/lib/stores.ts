/**
 * Global application stores
 */
import { writable, derived, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

// User and authentication state
export interface User {
	id: string;
	email: string;
	roles: string[];
	permissions: string[];
}

export const user: Writable<User | null> = writable(null);
export const isAuthenticated = derived(user, ($user) => $user !== null);

// Navigation state
export const currentPath = writable('/');
export const breadcrumbs: Writable<Record<string, string>> = writable({});

// UI state
export const sidebarCollapsed = writable(false);
export const darkMode = writable(false);

// Load dark mode preference from localStorage
if (browser) {
	const stored = localStorage.getItem('darkMode');
	if (stored) {
		darkMode.set(JSON.parse(stored));
	}
}

// Save dark mode preference to localStorage
darkMode.subscribe((value) => {
	if (browser) {
		localStorage.setItem('darkMode', JSON.stringify(value));
		if (value) {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	}
});

// Alert state
export interface Alert {
	id: string;
	level: 'info' | 'warning' | 'error' | 'success';
	message: string;
	timestamp: Date;
	dismissible?: boolean;
}

export const alerts: Writable<Alert[]> = writable([]);

export function addAlert(alert: Omit<Alert, 'id' | 'timestamp'>) {
	const newAlert: Alert = {
		...alert,
		id: Math.random().toString(36).substr(2, 9),
		timestamp: new Date(),
		dismissible: alert.dismissible ?? true
	};

	alerts.update(list => [newAlert, ...list]);

	// Auto-dismiss after 5 seconds for non-error alerts
	if (alert.level !== 'error' && newAlert.dismissible) {
		setTimeout(() => {
			dismissAlert(newAlert.id);
		}, 5000);
	}

	return newAlert.id;
}

export function dismissAlert(id: string) {
	alerts.update(list => list.filter(alert => alert.id !== id));
}

// Active alerts from server
export const activeAlerts: Writable<any[]> = writable([]);
export const activeAlertsCount = derived(activeAlerts, ($alerts) => $alerts.length);

// Real-time data counters
export const realtimeCounters = writable({
	alerts: 0,
	events: 0,
	logs: 0,
	endpoints: { online: 0, offline: 0 },
	containers: { running: 0, stopped: 0 }
});

// Individual counters for dashboard
export const alertCountStore = writable(0);
export const endpointCountStore = writable(0);
export const containerCountStore = writable(0);
export const eventCountStore = writable(0);

// Search state
export const globalSearchQuery = writable('');
export const searchStore = writable('');
export const searchResults: Writable<any[]> = writable([]);

// Loading states
export const loadingStates: Writable<Record<string, boolean>> = writable({});

export function setLoading(key: string, loading: boolean) {
	loadingStates.update(states => ({
		...states,
		[key]: loading
	}));
}

// Modal state
export interface Modal {
	id: string;
	component: any;
	props?: Record<string, any>;
}

export const modals: Writable<Modal[]> = writable([]);

export function openModal(component: any, props?: Record<string, any>) {
	const id = Math.random().toString(36).substr(2, 9);
	modals.update(list => [...list, { id, component, props }]);
	return id;
}

export function closeModal(id: string) {
	modals.update(list => list.filter(modal => modal.id !== id));
}

// Filter states for different pages
export const filters = writable({
	alerts: {
		state: '',
		level: '',
		search: ''
	},
	events: {
		level: '',
		type: '',
		scope: '',
		search: ''
	},
	logs: {
		level: '',
		source: '',
		search: ''
	},
	endpoints: {
		status: '',
		type: '',
		search: ''
	}
});
