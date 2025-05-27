import type { Widget, WidgetTemplate, WidgetType, WidgetConfig, WidgetPosition, Dashboard, DashboardPreferences } from '$lib/types/dashboard';
import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

// Default dashboard configuration
const DEFAULT_DASHBOARD: Dashboard = {
	id: 'default',
	name: 'Main Dashboard',
	isDefault: true,
	widgets: [],
	layout: {
		columns: 12,
		rowHeight: 60,
		margin: [16, 16],
		padding: [20, 20]
	},
	createdAt: new Date().toISOString(),
	updatedAt: new Date().toISOString()
};

// Default dashboard preferences
const DEFAULT_PREFERENCES: DashboardPreferences = {
	dashboards: [DEFAULT_DASHBOARD],
	defaultDashboardId: 'default',
	globalSettings: {
		autoRefresh: true,
		refreshInterval: 30,
		showGrid: false,
		snapToGrid: true,
		theme: 'auto'
	}
};

// Dashboard store
function createDashboardStore() {
	const { subscribe, set, update } = writable<DashboardPreferences>(DEFAULT_PREFERENCES);

	// Helper function to save preferences
	async function savePreferences(preferences: DashboardPreferences) {
		if (!browser) return;
		
		try {
			// Save to localStorage as backup
			localStorage.setItem('gosight-dashboard-preferences', JSON.stringify(preferences));
			
			// Save to API
			await fetch('/api/v1/users/preferences', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					dashboard: preferences
				})
			});
		} catch (error) {
			console.error('Failed to save dashboard preferences:', error);
		}
	}

	return {
		subscribe,
		
		// Load dashboard preferences from API/localStorage
		async load() {
			if (!browser) return;
			
			try {
				// Try to load from API first
				const response = await fetch('/api/v1/users/preferences');
				if (response.ok) {
					const preferences = await response.json();
					if (preferences.dashboard && preferences.dashboard.dashboards) {
						const dashboardPrefs = typeof preferences.dashboard === 'string' 
							? JSON.parse(preferences.dashboard)
							: preferences.dashboard;
						
						set({ ...DEFAULT_PREFERENCES, ...dashboardPrefs });
						return;
					}
				}
				
				// Fallback to localStorage
				const stored = localStorage.getItem('gosight-dashboard-preferences');
				if (stored) {
					const parsed = JSON.parse(stored);
					set({ ...DEFAULT_PREFERENCES, ...parsed });
				}
			} catch (error) {
				console.warn('Failed to load dashboard preferences:', error);
				set(DEFAULT_PREFERENCES);
			}
		},

		// Save dashboard preferences to API and localStorage
		async save(preferences: DashboardPreferences) {
			set(preferences);
			await savePreferences(preferences);
		},

		// Save specific dashboard (alias for save with current preferences)
		async saveDashboard(dashboardId: string) {
			let current: DashboardPreferences;
			const unsubscribe = subscribe(prefs => current = prefs);
			unsubscribe();
			await savePreferences(current!);
		},

		// Add a new dashboard
		addDashboard(dashboard: Omit<Dashboard, 'id' | 'createdAt' | 'updatedAt'>) {
			const newDashboard: Dashboard = {
				...dashboard,
				id: crypto.randomUUID(),
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString()
			};
			
			update(prefs => {
				const updated = {
					...prefs,
					dashboards: [...prefs.dashboards, newDashboard]
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
			
			return newDashboard;
		},

		// Update dashboard
		updateDashboard(dashboardId: string, updates: Partial<Dashboard>) {
			update(prefs => {
				const updated = {
					...prefs,
					dashboards: prefs.dashboards.map(d => 
						d.id === dashboardId 
							? { ...d, ...updates, updatedAt: new Date().toISOString() }
							: d
					)
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},		// Delete dashboard
		deleteDashboard(dashboardId: string) {
			update(prefs => {
				const updated = {
					...prefs,
					dashboards: prefs.dashboards.filter(d => d.id !== dashboardId),
					defaultDashboardId: prefs.defaultDashboardId === dashboardId
						? prefs.dashboards.find(d => d.id !== dashboardId)?.id || 'default'
						: prefs.defaultDashboardId
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},

		// Add widget to dashboard
		addWidget(dashboardId: string, widget: Omit<Widget, 'id' | 'createdAt' | 'updatedAt'>) {
			const newWidget: Widget = {
				...widget,
				id: crypto.randomUUID(),
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString()
			};

			update(prefs => {
				const updated = {
					...prefs,
					dashboards: prefs.dashboards.map(d => 
						d.id === dashboardId 
							? { 
								...d, 
								widgets: [...d.widgets, newWidget],
								updatedAt: new Date().toISOString()
							}
							: d
					)
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});

			return newWidget;
		},

		// Update widget
		updateWidget(dashboardId: string, widgetId: string, updates: Partial<Widget>) {
			update(prefs => {
				const updated = {
					...prefs,
					dashboards: prefs.dashboards.map(d => 
						d.id === dashboardId 
							? {
								...d,
								widgets: d.widgets.map(w => 
									w.id === widgetId 
										? { ...w, ...updates, updatedAt: new Date().toISOString() }
										: w
								),
								updatedAt: new Date().toISOString()
							}
							: d
					)
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},

		// Remove widget
		removeWidget(dashboardId: string, widgetId: string) {
			update(prefs => {
				const updated = {
					...prefs,
					dashboards: prefs.dashboards.map(d => 
						d.id === dashboardId 
							? {
								...d,
								widgets: d.widgets.filter(w => w.id !== widgetId),
								updatedAt: new Date().toISOString()
							}
							: d
					)
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},

		// Update global settings
		updateGlobalSettings(settings: Partial<DashboardPreferences['globalSettings']>) {
			update(prefs => {
				const updated = {
					...prefs,
					globalSettings: { ...prefs.globalSettings, ...settings }
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},

		// Set default dashboard
		setDefaultDashboard(dashboardId: string) {
			update(prefs => {
				const updated = {
					...prefs,
					defaultDashboardId: dashboardId
				};
				// Save in background without blocking the update
				savePreferences(updated);
				return updated;
			});
		},

		// Set active dashboard
		setActiveDashboard(dashboardId: string) {
			const dashboard = this.getDashboard(dashboardId);
			if (dashboard) {
				currentDashboard.set(dashboardId);
			}
		},

		// Move widget to new position
		moveWidget(dashboardId: string, widgetId: string, position: Pick<WidgetPosition, 'x' | 'y'>) {
			// Get current preferences
			let currentPrefs: DashboardPreferences;
			const unsubscribe = subscribe(prefs => currentPrefs = prefs);
			unsubscribe();
			
			const widget = currentPrefs!.dashboards
				.find(d => d.id === dashboardId)?.widgets
				.find(w => w.id === widgetId);
			
			if (widget) {
				this.updateWidget(dashboardId, widgetId, { 
					position: {
						...widget.position || { x: 0, y: 0, width: 2, height: 2 },
						...position
					}
				});
			}
		},

		// Resize widget
		resizeWidget(dashboardId: string, widgetId: string, position: WidgetPosition) {
			this.updateWidget(dashboardId, widgetId, { position });
		},

		// Duplicate widget
		duplicateWidget(dashboardId: string, widgetId: string, position: WidgetPosition) {
			// Get current preferences
			let currentPrefs: DashboardPreferences;
			const unsubscribe = subscribe(prefs => currentPrefs = prefs);
			unsubscribe();
			
			const original = currentPrefs!.dashboards
				.find(d => d.id === dashboardId)?.widgets
				.find(w => w.id === widgetId);
			
			if (!original) return;

			const newWidget = {
				...original,
				title: `${original.title} (Copy)`,
				position,
				config: { ...original.config }
			};

			// Remove id, createdAt, updatedAt so addWidget generates new ones
			const { id, createdAt, updatedAt, ...widgetData } = newWidget;
			this.addWidget(dashboardId, widgetData);
		},

		// Get widget by ID (helper method)
		getWidget(dashboardId: string, widgetId: string): Widget | undefined {
			let currentPrefs: DashboardPreferences;
			const unsubscribe = subscribe(prefs => currentPrefs = prefs);
			unsubscribe();
			
			const dashboard = currentPrefs!.dashboards.find(d => d.id === dashboardId);
			return dashboard?.widgets.find(w => w.id === widgetId);
		},

		// Get dashboard by ID (helper method)
		getDashboard(dashboardId: string): Dashboard | undefined {
			let currentPrefs: DashboardPreferences;
			const unsubscribe = subscribe(prefs => currentPrefs = prefs);
			unsubscribe();
			
			return currentPrefs!.dashboards.find(d => d.id === dashboardId);
		},

		// Get current preferences (helper method)
		getCurrentPreferences(): DashboardPreferences {
			let current: DashboardPreferences;
			const unsubscribe = subscribe(prefs => current = prefs);
			unsubscribe();
			return current!;
		}
	};
}

export const dashboardStore = createDashboardStore();

// Current active dashboard
export const currentDashboard = writable<string>('default');

// Derived stores
export const activeDashboard = derived(
	[dashboardStore, currentDashboard],
	([prefs, currentId]) => prefs.dashboards.find(d => d.id === currentId) || prefs.dashboards[0]
);

export const isEditMode = writable<boolean>(false);
export const draggedWidget = writable<Widget | null>(null);
export const selectedWidget = writable<Widget | null>(null);

// Derived store for unsaved changes tracking
export const hasUnsavedChanges = derived(dashboardStore, ($prefs) => {
	// This is a simplified implementation - in a real app you'd track changes more precisely
	return false; // For now, assume always saved since we auto-save
});
