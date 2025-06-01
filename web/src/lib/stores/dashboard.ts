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

import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import type { Dashboard, Widget, WidgetPosition, WidgetType } from '$lib/types/dashboard';

const DEFAULT_DASHBOARD: Dashboard = {
  id: 'default',
  name: 'Main Dashboard',
  widgets: [],
  layout: { 
    columns: 12, 
    rowHeight: 120 
  },
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString()
};

function createDashboardStore() {
  const { subscribe, set, update } = writable<Dashboard>(DEFAULT_DASHBOARD);

  return {
    subscribe,
    
    /**
     * Add a new widget to the dashboard
     */
    addWidget: (widget: Omit<Widget, 'id' | 'createdAt' | 'updatedAt'>) => {
      const newWidget: Widget = {
        ...widget,
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };
      
      update(dashboard => ({
        ...dashboard,
        widgets: [...dashboard.widgets, newWidget],
        updatedAt: new Date().toISOString()
      }));
      
      if (browser) {
        toast.success(`Added ${widget.title} widget`, {
          description: 'Widget has been added to your dashboard'
        });
      }
      
      return newWidget.id;
    },
    
    /**
     * Move a widget to a new position
     */
    moveWidget: (widgetId: string, position: Partial<WidgetPosition>) => {
      console.log('moveWidget called:', { widgetId, position });
      
      update(dashboard => {
        const updatedDashboard = {
          ...dashboard,
          widgets: dashboard.widgets.map(w => {
            if (w.id === widgetId) {
              const newPosition = { ...w.position, ...position };
              console.log('Updating widget position:', { 
                widgetId: w.id, 
                oldPosition: w.position, 
                newPosition 
              });
              return { 
                ...w, 
                position: newPosition,
                updatedAt: new Date().toISOString()
              };
            }
            return w;
          }),
          updatedAt: new Date().toISOString()
        };
        
        console.log('Dashboard after move:', updatedDashboard.widgets.find(w => w.id === widgetId)?.position);
        return updatedDashboard;
      });
    },
    
    /**
     * Resize a widget
     */
    resizeWidget: (widgetId: string, size: Pick<WidgetPosition, 'width' | 'height'>) => {
      update(dashboard => ({
        ...dashboard,
        widgets: dashboard.widgets.map(w => 
          w.id === widgetId 
            ? { 
                ...w, 
                position: { ...w.position, ...size },
                updatedAt: new Date().toISOString()
              }
            : w
        ),
        updatedAt: new Date().toISOString()
      }));
    },
    
    /**
     * Remove a widget from the dashboard
     */
    removeWidget: (widgetId: string) => {
      const currentDashboard = get({ subscribe });
      const widget = currentDashboard.widgets.find(w => w.id === widgetId);
      
      update(dashboard => ({
        ...dashboard,
        widgets: dashboard.widgets.filter(w => w.id !== widgetId),
        updatedAt: new Date().toISOString()
      }));
      
      if (browser && widget) {
        toast.success(`Removed ${widget.title}`, {
          description: 'Widget has been removed from your dashboard'
        });
      }
    },
    
    /**
     * Update widget configuration
     */
    updateWidgetConfig: (widgetId: string, config: Record<string, any>) => {
      update(dashboard => ({
        ...dashboard,
        widgets: dashboard.widgets.map(w => 
          w.id === widgetId 
            ? { 
                ...w, 
                config: { ...w.config, ...config },
                updatedAt: new Date().toISOString()
              }
            : w
        ),
        updatedAt: new Date().toISOString()
      }));
    },
    
    /**
     * Find an empty position for a new widget
     */
    findEmptyPosition: (width: number = 3, height: number = 2): WidgetPosition => {
      const currentDashboard = get({ subscribe });
      const { columns } = currentDashboard.layout;
      const widgets = currentDashboard.widgets;
      
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
     * Update dashboard layout settings
     */
    updateLayout: (layout: Partial<Dashboard['layout']>) => {
      update(dashboard => ({
        ...dashboard,
        layout: { ...dashboard.layout, ...layout },
        updatedAt: new Date().toISOString()
      }));
      
      if (browser) {
        toast.success('Layout updated', {
          description: 'Dashboard layout has been updated'
        });
      }
    },
    
    /**
     * Load dashboard from API or localStorage
     */
    load: async (dashboardId?: string) => {
      try {
        // Try to load from API first
        // const response = await fetch(`/api/dashboards/${dashboardId || 'default'}`);
        // if (response.ok) {
        //   const dashboard = await response.json();
        //   set(dashboard);
        //   return;
        // }
        
        // Fallback to localStorage
        if (browser) {
          const stored = localStorage.getItem(`dashboard-${dashboardId || 'default'}`);
          if (stored) {
            const dashboard = JSON.parse(stored);
            set(dashboard);
            console.log('Loaded dashboard from localStorage:', dashboard);
            return;
          }
          
          // If no saved data, initialize with sample widgets for demo
          const sampleDashboard: Dashboard = {
            ...DEFAULT_DASHBOARD,
            widgets: [
              {
                id: crypto.randomUUID(),
                type: 'metric',
                title: 'Total Users',
                position: { x: 0, y: 0, width: 2, height: 1 },
                config: {},
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              },
              {
                id: crypto.randomUUID(),
                type: 'gauge',
                title: 'CPU Usage',
                position: { x: 2, y: 0, width: 2, height: 1 },
                config: {},
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              },
              {
                id: crypto.randomUUID(),
                type: 'chart',
                title: 'Performance Metrics',
                position: { x: 0, y: 1, width: 4, height: 2 },
                config: {},
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              }
            ]
          };
          
          set(sampleDashboard);
          console.log('Initialized dashboard with sample widgets');
          
          // Save the initial sample data
          localStorage.setItem(`dashboard-${dashboardId || 'default'}`, JSON.stringify(sampleDashboard));
        }
      } catch (error) {
        console.warn('Failed to load dashboard:', error);
      }
    },
    
    /**
     * Save dashboard to API and localStorage
     */
    save: async () => {
      const dashboard = get({ subscribe });
      
      try {
        // Try to save to API first
        // await fetch(`/api/dashboards/${dashboard.id}`, {
        //   method: 'PUT',
        //   headers: { 'Content-Type': 'application/json' },
        //   body: JSON.stringify(dashboard)
        // });
        
        // Always save to localStorage as backup
        if (browser) {
          localStorage.setItem(`dashboard-${dashboard.id}`, JSON.stringify(dashboard));
          toast.success('Dashboard saved', {
            description: 'Your dashboard layout has been saved successfully'
          });
        }
      } catch (error) {
        console.warn('Failed to save dashboard to API:', error);
        // Still save to localStorage
        if (browser) {
          localStorage.setItem(`dashboard-${dashboard.id}`, JSON.stringify(dashboard));
          toast.warning('Dashboard saved locally', {
            description: 'Could not sync to server, but saved locally'
          });
        }
      }
    },
    
    /**
     * Reset dashboard to default state
     */
    reset: () => {
      set(DEFAULT_DASHBOARD);
      if (browser) {
        toast.success('Dashboard reset', {
          description: 'Your dashboard has been reset to default state'
        });
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

// Auto-save when dashboard changes
dashboardStore.subscribe(() => {
  // Debounce saves to avoid excessive API calls
  clearTimeout((globalThis as any).dashboardSaveTimeout);
  (globalThis as any).dashboardSaveTimeout = setTimeout(() => {
    dashboardStore.save();
  }, 1000);
});
