// Copyright (C) 2025 Aaron Mathis
// This file is part of GoSight Server.
//
// GoSight Server is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// GoSight Server is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with GoSight Server.  If not, see <https://www.gnu.org/licenses/>.

/**
 * Dashboard store for managing dashboard state, widgets, and layout.
 * Provides CRUD operations for widgets and dashboard management.
 */

import { writable, get, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import type { Dashboard, Widget, WidgetPosition, WidgetType } from '$lib/types/dashboard';
import { getWidgetSize, getCurrentBreakpoint, validateWidgetSize } from '$lib/configs/widget-sizing';

// Multi-dashboard store structure
interface DashboardStore {
  dashboards: Dashboard[];
  activeDashboardId: string;
}

const DEFAULT_DASHBOARD_STORE: DashboardStore = {
  dashboards: [
    {
      id: 'main',
      name: 'Main Dashboard',
      widgets: [],
      layout: { columns: 12, rowHeight: 120 },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
  ],
  activeDashboardId: 'main'
};

function createDashboardStore() {
  const { subscribe, set, update } = writable<DashboardStore>(DEFAULT_DASHBOARD_STORE);

  return {
    subscribe,

    /**
     * Get the currently active dashboard
     */
    getActiveDashboard: () => {
      const store = get({ subscribe });
      return store.dashboards.find(d => d.id === store.activeDashboardId) || store.dashboards[0];
    },

    /**
     * Switch to a different dashboard
     */
    setActiveDashboard: (dashboardId: string) => {
      update(store => ({
        ...store,
        activeDashboardId: dashboardId
      }));
      
      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
      }
    },

    /**
     * Create a new dashboard
     */
    createDashboard: (name: string) => {
      const newDashboard: Dashboard = {
        id: name.toLowerCase().replace(/\s+/g, '-') + '-' + Date.now(),
        name,
        widgets: [],
        layout: { columns: 12, rowHeight: 120 },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      update(store => ({
        ...store,
        dashboards: [...store.dashboards, newDashboard],
        activeDashboardId: newDashboard.id
      }));

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
        toast.success(`Dashboard "${name}" created`);
      }

      return newDashboard.id;
    },

    /**
     * Delete a dashboard
     */
    deleteDashboard: (dashboardId: string) => {
      update(store => {
        if (store.dashboards.length <= 1) {
          toast.error('Cannot delete the last dashboard');
          return store;
        }

        const dashboardToDelete = store.dashboards.find(d => d.id === dashboardId);
        if (!dashboardToDelete) return store;

        const newDashboards = store.dashboards.filter(d => d.id !== dashboardId);
        let newActiveDashboardId = store.activeDashboardId;

        // If we're deleting the active dashboard, switch to the first remaining one
        if (dashboardId === store.activeDashboardId) {
          newActiveDashboardId = newDashboards[0].id;
        }

        toast.success(`Dashboard "${dashboardToDelete.name}" deleted`);

        return {
          ...store,
          dashboards: newDashboards,
          activeDashboardId: newActiveDashboardId
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
      }
    },

    /**
     * Add a widget to the active dashboard
     */
    addWidget: (widget: Omit<Widget, 'id' | 'createdAt' | 'updatedAt'>) => {
      const newWidget: Widget = {
        ...widget,
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      update(store => {
        const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
        if (!activeDashboard) return store;

        const updatedDashboard = {
          ...activeDashboard,
          widgets: [...activeDashboard.widgets, newWidget],
          updatedAt: new Date().toISOString()
        };

        return {
          ...store,
          dashboards: store.dashboards.map(d => 
            d.id === store.activeDashboardId ? updatedDashboard : d
          )
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
        toast.success(`Added ${widget.title} widget`);
      }

      return newWidget.id;
    },

    /**
     * Add a widget with smart sizing based on widget type and current screen size
     */
    addWidgetWithSmartSizing: (widgetType: string, title: string, config: Record<string, any> = {}) => {
      // Get appropriate size for current breakpoint
      const breakpoint = getCurrentBreakpoint();
      const defaultSize = getWidgetSize(widgetType, breakpoint);

      // Find empty position using inline logic to avoid circular reference
      const store = get({ subscribe });
      const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);

      let foundPosition: WidgetPosition;
      if (!activeDashboard) {
        foundPosition = { x: 0, y: 0, width: defaultSize.width, height: defaultSize.height };
      } else {
        // Inline findEmptyPosition logic
        const { columns } = activeDashboard.layout;
        const widgets = activeDashboard.widgets;

        // Create a grid to track occupied cells
        const maxRows = Math.max(10, ...widgets.map(w => w.position.y + w.position.height));
        const grid = Array(maxRows).fill(null).map(() => Array(columns).fill(false));

        // Mark occupied cells
        widgets.forEach(widget => {
          for (let y = widget.position.y; y < widget.position.y + widget.position.height; y++) {
            for (let x = widget.position.x; x < widget.position.x + widget.position.width; x++) {
              if (y < maxRows && x < columns) {
                grid[y][x] = true;
              }
            }
          }
        });

        // Find first available position
        foundPosition = { x: 0, y: maxRows, width: defaultSize.width, height: defaultSize.height };
        for (let y = 0; y <= maxRows - defaultSize.height; y++) {
          for (let x = 0; x <= columns - defaultSize.width; x++) {
            let canPlace = true;

            // Check if the area is free
            for (let dy = 0; dy < defaultSize.height && canPlace; dy++) {
              for (let dx = 0; dx < defaultSize.width && canPlace; dx++) {
                if (y + dy >= maxRows || grid[y + dy][x + dx]) {
                  canPlace = false;
                }
              }
            }

            if (canPlace) {
              foundPosition = { x, y, width: defaultSize.width, height: defaultSize.height };
              y = maxRows; // Break outer loop
              break;
            }
          }
        }
      }

      const widget = {
        type: widgetType,
        title,
        config,
        position: foundPosition
      };

      // Create the new widget
      const newWidget: Widget = {
        ...widget,
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };

      // Add to store
      update(store => {
        const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
        if (!activeDashboard) return store;

        const updatedDashboard = {
          ...activeDashboard,
          widgets: [...activeDashboard.widgets, newWidget],
          updatedAt: new Date().toISOString()
        };

        return {
          ...store,
          dashboards: store.dashboards.map(d =>
            d.id === store.activeDashboardId ? updatedDashboard : d
          )
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
        toast.success(`Added ${widget.title} widget`);
      }

      return newWidget.id;
    },

    /**
     * FIXED: Move widget to new position
     */
    moveWidget: (widgetId: string, position: WidgetPosition) => {
      console.log('moveWidget called:', { widgetId, position });

      update(store => {
        const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
        if (!activeDashboard) return store;

        const widget = activeDashboard.widgets.find(w => w.id === widgetId);
        if (!widget) {
          console.warn('Widget not found:', widgetId);
          return store;
        }

        console.log('Updating widget position:', {
          widgetId,
          oldPosition: widget.position,
          newPosition: position
        });

        const updatedWidget = {
          ...widget,
          position: { ...position },
          updatedAt: new Date().toISOString()
        };

        const updatedDashboard = {
          ...activeDashboard,
          widgets: activeDashboard.widgets.map(w => 
            w.id === widgetId ? updatedWidget : w
          ),
          updatedAt: new Date().toISOString()
        };

        console.log('Dashboard after move:', updatedDashboard.widgets.find(w => w.id === widgetId)?.position);

        // CRITICAL FIX: Return the new store state
        return {
          ...store,
          dashboards: store.dashboards.map(d => 
            d.id === store.activeDashboardId ? updatedDashboard : d
          )
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
        console.log('Widget move saved to localStorage');
      }
    },

    /**
     * Remove widget from active dashboard
     */
    removeWidget: (widgetId: string) => {
      update(store => {
        const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
        if (!activeDashboard) return store;

        const widget = activeDashboard.widgets.find(w => w.id === widgetId);
        
        const updatedDashboard = {
          ...activeDashboard,
          widgets: activeDashboard.widgets.filter(w => w.id !== widgetId),
          updatedAt: new Date().toISOString()        };

        if (browser && widget) {
          toast.success(`Removed ${widget.title}`);
        }

        return {
          ...store,
          dashboards: store.dashboards.map(d => 
            d.id === store.activeDashboardId ? updatedDashboard : d
          )
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
      }
    },

    /**
     * Resize widget in active dashboard
     */
    resizeWidget: (widgetId: string, size: { width: number; height: number }) => {
      update(store => {
        const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
        if (!activeDashboard) return store;

        const widget = activeDashboard.widgets.find(w => w.id === widgetId);
        if (!widget) return store;

        const updatedWidget = {
          ...widget,
          position: { 
            ...widget.position, 
            width: size.width, 
            height: size.height 
          },
          updatedAt: new Date().toISOString()
        };

        const updatedDashboard = {
          ...activeDashboard,
          widgets: activeDashboard.widgets.map(w => 
            w.id === widgetId ? updatedWidget : w
          ),
          updatedAt: new Date().toISOString()
        };

        return {
          ...store,
          dashboards: store.dashboards.map(d => 
            d.id === store.activeDashboardId ? updatedDashboard : d
          )
        };
      });

      if (browser) {
        const store = get({ subscribe });
        localStorage.setItem('dashboards', JSON.stringify(store));
      }
    },

    /**
     * Find empty position in active dashboard
     */
    findEmptyPosition: (width: number = 3, height: number = 2): WidgetPosition => {
      const store = get({ subscribe });
      const activeDashboard = store.dashboards.find(d => d.id === store.activeDashboardId);
      if (!activeDashboard) return { x: 0, y: 0, width, height };

      const { columns } = activeDashboard.layout;
      const widgets = activeDashboard.widgets;
      
      // Create a grid to track occupied cells
      const maxRows = Math.max(10, ...widgets.map(w => w.position.y + w.position.height));
      const grid = Array(maxRows).fill(null).map(() => Array(columns).fill(false));
      
      // Mark occupied cells
      widgets.forEach(widget => {
        for (let y = widget.position.y; y < widget.position.y + widget.position.height; y++) {
          for (let x = widget.position.x; x < widget.position.x + widget.position.width; x++) {
            if (y < maxRows && x < columns) {
              grid[y][x] = true;
            }
          }
        }
      });
      
      // Find first available position
      for (let y = 0; y <= maxRows - height; y++) {
        for (let x = 0; x <= columns - width; x++) {
          let canPlace = true;
          
          // Check if the area is free
          for (let dy = 0; dy < height && canPlace; dy++) {
            for (let dx = 0; dx < width && canPlace; dx++) {
              if (y + dy >= maxRows || grid[y + dy][x + dx]) {
                canPlace = false;
              }
            }
          }
          
          if (canPlace) {
            return { x, y, width, height };
          }
        }
      }
      
      // If no space found, add to the bottom
      return { x: 0, y: maxRows, width, height };
    },

    /**
     * Load dashboards from localStorage
     */
    load: async () => {
      try {
        if (browser) {
          const stored = localStorage.getItem('dashboards');
          if (stored) {
            const dashboardStore = JSON.parse(stored);
            set(dashboardStore);
            console.log('Loaded dashboards from localStorage:', dashboardStore);
          } else {
            set(DEFAULT_DASHBOARD_STORE);
            console.log('No saved dashboards found, starting with default');
          }
        }
      } catch (error) {
        console.warn('Failed to load dashboards:', error);
        set(DEFAULT_DASHBOARD_STORE);
      }
    },

    /**
     * Save all dashboards to localStorage
     */
    save: async () => {
      const store = get({ subscribe });
      
      try {
        if (browser) {
          localStorage.setItem('dashboards', JSON.stringify(store));
          console.log('Dashboards saved to localStorage:', store);
          toast.success('Dashboard saved');
        }
      } catch (error) {
        console.warn('Failed to save dashboards:', error);
      }
    },

    /**
     * Reset all dashboards
     */
    reset: () => {
      set(DEFAULT_DASHBOARD_STORE);
      if (browser) {
        localStorage.removeItem('dashboards');
        toast.success('All dashboards reset');
      }
    }
  };
}

export const dashboardStore = createDashboardStore();

// Initialize dashboard on store creation
if (browser) {
  dashboardStore.load();
}

// UI state stores
export const isEditMode = writable(false);
export const draggedWidget = writable<Widget | string | null>(null);
export const selectedWidget = writable<string | null>(null);
export const showGridLines = writable(true);

// FIXED: Replace custom derived store with proper Svelte derived store
export const activeDashboard = derived(
  dashboardStore,
  ($store) => {
    const active = $store.dashboards.find(d => d.id === $store.activeDashboardId) || $store.dashboards[0];
    console.log('🔄 activeDashboard derived updated:', active.id, 'widgets:', active.widgets.length);
    return active;
  }
);

// DEVELOPMENT: Expose stores to window for testing (after all stores are declared)
if (browser && import.meta.env.DEV) {
  console.log('🔧 Exposing dashboard stores to window...');
  (window as any).dashboardStore = dashboardStore;
  (window as any).activeDashboard = activeDashboard;
  (window as any).isEditMode = isEditMode;
  (window as any).draggedWidget = draggedWidget;
  console.log('🔧 Dashboard stores exposed to window for testing');
  console.log('Available stores:', Object.keys(window).filter(key => key.includes('dashboard') || key.includes('edit') || key.includes('drag')));
}
