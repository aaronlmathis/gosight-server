// Simple test script to verify drag and drop functionality
// This script can be run in the browser console

console.log('🧪 Testing Drag and Drop Functionality...');

// First, let's check the current dashboard state
const checkDashboardState = () => {
  console.log('📊 Current Dashboard State:');
  
  // Check what's available on window
  console.log('Window keys:', Object.keys(window).filter(key => key.includes('dashboard') || key.includes('edit') || key.includes('drag')));
  
  // Try to access the Svelte 5 stores
  if (window.activeDashboard) {
    const dashboard = window.activeDashboard;
    // In Svelte 5, stores are state objects with get() method
    const dashboardValue = typeof dashboard.get === 'function' ? dashboard.get() : dashboard;
    console.log('Active Dashboard:', dashboardValue);
    console.log('Widgets count:', dashboardValue?.widgets?.length || 0);
  } else {
    console.log('❌ activeDashboard not found on window');
  }
  
  if (window.dashboardStore) {
    console.log('✅ dashboardStore found');
  } else {
    console.log('❌ dashboardStore not found on window');
  }
  
  if (window.isEditMode) {
    const editMode = window.isEditMode;
    const editModeValue = typeof editMode.get === 'function' ? editMode.get() : editMode;
    console.log('Edit Mode:', editModeValue);
  } else {
    console.log('❌ isEditMode not found on window');
  }
  
  if (window.draggedWidget) {
    const draggedWidget = window.draggedWidget;
    const draggedValue = typeof draggedWidget.get === 'function' ? draggedWidget.get() : draggedWidget;
    console.log('Dragged Widget:', draggedValue);
  } else {
    console.log('❌ draggedWidget not found on window');
  }
  
  // Also check DOM elements
  const dashboardGrids = document.querySelectorAll('.dashboard-grid');
  const widgets = document.querySelectorAll('.widget-container');
  const allDivs = document.querySelectorAll('div[class*="dashboard"], div[class*="grid"]');
  
  console.log('Dashboard grids found:', dashboardGrids.length);
  console.log('Widgets found:', widgets.length);
  console.log('Dashboard-related divs:', allDivs.length);
  
  // Log some classes to help debug
  if (allDivs.length > 0) {
    console.log('Sample div classes:', Array.from(allDivs).slice(0, 3).map(div => div.className));
  }
};

// Function to enable edit mode
const enableEditMode = () => {
  if (window.isEditMode) {
    const store = window.isEditMode;
    if (typeof store.set === 'function') {
      store.set(true);
      console.log('✅ Edit mode enabled via store');
    } else if (typeof store.update === 'function') {
      store.update(() => true);
      console.log('✅ Edit mode enabled via update');
    }
  } else {
    // Find the edit mode button with more specific selectors
    const editButtons = Array.from(document.querySelectorAll('button')).filter(btn => {
      const text = btn.textContent || btn.innerText || '';
      return text.includes('Edit') || text.includes('edit');
    });
    
    if (editButtons.length > 0) {
      editButtons[0].click();
      console.log('✅ Edit mode button clicked');
    } else {
      console.log('❌ Edit mode button not found');
      console.log('Available buttons:', Array.from(document.querySelectorAll('button')).map(btn => btn.textContent));
    }
  }
};

// Function to add a test widget
const addTestWidget = () => {
  if (window.dashboardStore) {
    const position = { x: 0, y: 0, width: 3, height: 2 };
    const widgetId = window.dashboardStore.addWidget({
      type: 'metric',
      title: 'Test Widget',
      position,
      config: {}
    });
    console.log('✅ Test widget added:', widgetId);
    return widgetId;
  } else {
    // Try using the Add Widget button in the UI
    const addButton = document.querySelector('button:contains("Add Widget"), [class*="Add"]');
    if (addButton) {
      addButton.click();
      console.log('✅ Add widget button clicked');
    } else {
      console.log('❌ Could not find add widget method');
    }
  }
};

// Function to simulate widget move
const moveTestWidget = (widgetId, newPosition) => {
  if (window.dashboardStore) {
    window.dashboardStore.moveWidget(widgetId, newPosition);
    console.log('🚀 Widget moved to:', newPosition);
    
    // Check if the move persisted
    setTimeout(() => {
      const dashboard = window.activeDashboard?.get?.();
      const widget = dashboard?.widgets?.find(w => w.id === widgetId);
      if (widget) {
        console.log('✅ Widget position after move:', widget.position);
        if (widget.position.x === newPosition.x && widget.position.y === newPosition.y) {
          console.log('🎉 Move successful and persisted!');
        } else {
          console.log('❌ Move did not persist properly');
        }
      }
    }, 100);
  }
};

// Run the test
const runTest = () => {
  console.log('🚀 Starting Drag & Drop Test...');
  checkDashboardState();
  enableEditMode();
  
  setTimeout(() => {
    const widgetId = addTestWidget();
    
    setTimeout(() => {
      checkDashboardState();
      moveTestWidget(widgetId, { x: 2, y: 1, width: 3, height: 2 });
    }, 500);
  }, 500);
};

// Export to window for manual testing
if (typeof window !== 'undefined') {
  window.testDragDrop = {
    checkDashboardState,
    enableEditMode,
    addTestWidget,
    moveTestWidget,
    runTest
  };
  
  console.log('🔧 Test utilities available at window.testDragDrop');
  console.log('Run window.testDragDrop.runTest() to start testing');
}
