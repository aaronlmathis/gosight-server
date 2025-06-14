// Debug utilities for development
import { browser } from '$app/environment';
import { dashboardStore, activeDashboard, isEditMode, draggedWidget } from '$lib/stores/dashboardStore';

// Expose stores to window in development mode
if (browser && import.meta.env.DEV) {
  console.log('🔧 [DEBUG] Exposing dashboard stores to window...');
  
  // Add to window with error handling
  try {
    (window as any).dashboardStore = dashboardStore;
    (window as any).activeDashboard = activeDashboard;
    (window as any).isEditMode = isEditMode;
    (window as any).draggedWidget = draggedWidget;
    
    console.log('🔧 [DEBUG] Dashboard stores exposed to window for testing');
    console.log('🔧 [DEBUG] Available stores:', Object.keys(window).filter(key => 
      key.includes('dashboard') || key.includes('edit') || key.includes('drag')));
    
    // Test store access
    console.log('🔧 [DEBUG] Store test:', {
      dashboardStore: typeof (window as any).dashboardStore,
      activeDashboard: typeof (window as any).activeDashboard,
      isEditMode: typeof (window as any).isEditMode,
      draggedWidget: typeof (window as any).draggedWidget
    });
  } catch (error) {
    console.error('❌ [DEBUG] Failed to expose stores:', error);
  }
}
