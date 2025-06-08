// Final Comprehensive Drag-Drop Test Script
console.log('🧪 Final Drag-Drop Test Suite Starting...');

// Test 1: Check Edit Mode Status
function testEditMode() {
  console.log('\n1️⃣ Testing Edit Mode Status:');
  
  // Check if isEditMode store exists and is accessible
  if (typeof window.isEditMode !== 'undefined') {
    console.log('✅ isEditMode store is accessible');
    
    // Get current value
    window.isEditMode.subscribe(value => {
      console.log(`   Current edit mode: ${value}`);
    });
  } else {
    console.log('❌ isEditMode store not accessible');
  }
}

// Test 2: Toggle Edit Mode and Check Widget Draggability
function testEditModeToggle() {
  console.log('\n2️⃣ Testing Edit Mode Toggle:');
  
  // Find edit mode toggle button - try multiple selectors
  let editButton = document.querySelector('button:has(.lucide-edit-3)');
  if (!editButton) {
    editButton = document.querySelector('button:has(.lucide-edit)');
  }
  if (!editButton) {
    editButton = document.querySelector('button[aria-label*="edit" i]');
  }
  if (!editButton) {
    editButton = document.querySelector('button[title*="edit" i]');
  }
  if (!editButton) {
    editButton = Array.from(document.querySelectorAll('button')).find(btn => 
      btn.textContent.toLowerCase().includes('edit') || 
      btn.querySelector('svg') && btn.querySelector('svg').classList.toString().includes('edit')
    );
  }
  if (!editButton) {
    // Log all buttons to help debug
    console.log('   Available buttons:');
    document.querySelectorAll('button').forEach((btn, i) => {
      const hasIcon = btn.querySelector('svg') ? ' (has icon)' : '';
      const iconClasses = btn.querySelector('svg')?.classList.toString() || '';
      console.log(`     ${i}: "${btn.textContent.trim()}"${hasIcon} (classes: ${btn.className}) (icon: ${iconClasses})`);
    });
  }
  
  if (editButton) {
    console.log('✅ Edit mode toggle button found');
    console.log('   Button text:', editButton.textContent.trim());
    
    // Click to toggle edit mode
    editButton.click();
    
    setTimeout(() => {
      // Check widget draggable attributes after toggle
      const widgets = document.querySelectorAll('.widget-container');
      console.log(`   Found ${widgets.length} widgets`);
      
      widgets.forEach((widget, index) => {
        const draggable = widget.getAttribute('draggable');
        console.log(`   Widget ${index + 1} draggable: ${draggable}`);
      });
    }, 100);
    
  } else {
    console.log('❌ Edit mode toggle button not found');
  }
}

// Test 3: Check Widget Container Configuration
function testWidgetContainers() {
  console.log('\n3️⃣ Testing Widget Containers:');
  
  const widgets = document.querySelectorAll('.widget-container');
  console.log(`   Found ${widgets.length} widget containers`);
  
  widgets.forEach((widget, index) => {
    const id = widget.querySelector('[data-widget-id]')?.getAttribute('data-widget-id') || `widget-${index}`;
    const draggable = widget.getAttribute('draggable');
    const hasRing = widget.classList.contains('ring-2');
    const hasEditIndicators = widget.querySelector('.absolute') !== null;
    
    console.log(`   Widget ${id}:`);
    console.log(`     - Draggable: ${draggable}`);
    console.log(`     - Has edit ring: ${hasRing}`);
    console.log(`     - Has edit indicators: ${hasEditIndicators}`);
  });
}

// Test 4: Simulate Drag and Drop
function testDragAndDrop() {
  console.log('\n4️⃣ Testing Drag and Drop Simulation:');
  
  const widgets = document.querySelectorAll('.widget-container[draggable="true"]');
  const grid = document.querySelector('.dashboard-grid');
  
  if (widgets.length === 0) {
    console.log('❌ No draggable widgets found');
    return;
  }
  
  if (!grid) {
    console.log('❌ Dashboard grid not found');
    return;
  }
  
  const widget = widgets[0];
  console.log('✅ Starting drag simulation...');
  
  // Simulate drag start
  const dragStartEvent = new DragEvent('dragstart', {
    bubbles: true,
    cancelable: true,
    dataTransfer: new DataTransfer()
  });
  
  widget.dispatchEvent(dragStartEvent);
  console.log('   ✅ Drag start event dispatched');
  
  // Simulate drag over grid
  const rect = grid.getBoundingClientRect();
  const dragOverEvent = new DragEvent('dragover', {
    bubbles: true,
    cancelable: true,
    clientX: rect.left + 200,
    clientY: rect.top + 200,
    dataTransfer: new DataTransfer()
  });
  
  grid.dispatchEvent(dragOverEvent);
  console.log('   ✅ Drag over event dispatched');
  
  // Check for drop preview
  setTimeout(() => {
    const dropPreview = document.querySelector('.drop-preview');
    console.log(`   Drop preview visible: ${dropPreview ? 'Yes' : 'No'}`);
    
    // Simulate drop
    const dropEvent = new DragEvent('drop', {
      bubbles: true,
      cancelable: true,
      clientX: rect.left + 200,
      clientY: rect.top + 200,
      dataTransfer: new DataTransfer()
    });
    
    grid.dispatchEvent(dropEvent);
    console.log('   ✅ Drop event dispatched');
    
    // Simulate drag end
    const dragEndEvent = new DragEvent('dragend', {
      bubbles: true,
      cancelable: true
    });
    
    widget.dispatchEvent(dragEndEvent);
    console.log('   ✅ Drag end event dispatched');
    
  }, 100);
}

// Test 5: Check Store Values
function testStoreValues() {
  console.log('\n5️⃣ Testing Store Values:');
  
  // Check draggedWidget store
  if (window.draggedWidget) {
    window.draggedWidget.subscribe(value => {
      console.log(`   draggedWidget: ${value ? (typeof value === 'object' ? value.id : value) : 'null'}`);
    });
  }
  
  // Check activeDashboard store  
  if (window.activeDashboard) {
    window.activeDashboard.subscribe(value => {
      console.log(`   activeDashboard: ${value ? value.name : 'null'} (${value ? value.widgets.length : 0} widgets)`);
    });
  }
}

// Run all tests
function runAllTests() {
  console.log('🚀 Running Final Drag-Drop Test Suite...');
  
  testEditMode();
  testStoreValues();
  testWidgetContainers();
  
  setTimeout(() => {
    testEditModeToggle();
    
    setTimeout(() => {
      testWidgetContainers();
      testDragAndDrop();
    }, 500);
  }, 200);
}

// Export for console use
window.finalDragDropTest = {
  runAll: runAllTests,
  testEditMode,
  testEditModeToggle,
  testWidgetContainers,
  testDragAndDrop,
  testStoreValues
};

console.log('📝 Test functions available:');
console.log('   window.finalDragDropTest.runAll() - Run complete test suite');
console.log('   window.finalDragDropTest.testEditMode() - Test edit mode status');
console.log('   window.finalDragDropTest.testEditModeToggle() - Test edit mode toggle');
console.log('   window.finalDragDropTest.testWidgetContainers() - Test widget containers');
console.log('   window.finalDragDropTest.testDragAndDrop() - Test drag and drop simulation');
console.log('   window.finalDragDropTest.testStoreValues() - Test store values');

// Auto-run on load
runAllTests();
