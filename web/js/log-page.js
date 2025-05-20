import { gosightFetch } from "./api.js";

// Wait for DOM content to be fully loaded
document.addEventListener("DOMContentLoaded", () => {
  // Initialize state variables
  const state = {
    expandedLogKeys: new Set(),
    currentCursor: null,
    previousCursors: [],
    isLoading: false,
    hasMore: false,
    nextCursor: null,
    cursorStack: [], // Stack to keep track of cursor history
    currentPage: 1,
    firstVisibleTime: null, // Track the timestamp of first visible log
    lastVisibleTime: null   // Track the timestamp of last visible log
  };

  // Define mappings for form fields to URL parameters
  const mappings = {
    contains: "filter-keyword",
    source: "filter-source",
    container: "container-name",
    endpoint: "endpoint-name",
    app: "app-name",
    start: "start-time",
    end: "end-time"
  };

  // Get DOM elements
  const elements = {
    searchForm: document.getElementById("log-search-form"),
    resultsTable: document.getElementById("log-results"),
    cursorDisplay: document.getElementById("cursor-time"),
    tagFilters: document.getElementById("tag-filters"),
    resetButton: document.getElementById("reset-search"),
    searchButton: document.getElementById("search-submit"),
    prevPageBtn: document.getElementById("prev-page"),
    nextPageBtn: document.getElementById("next-page"),
    prevPageMobileBtn: document.getElementById("prev-page-mobile"),
    nextPageMobileBtn: document.getElementById("next-page-mobile"),
    logCount: document.getElementById("log-count")
  };

  // Verify required elements exist
  if (!elements.searchForm || !elements.resultsTable) {
    console.error("Required DOM elements not found. Check if the page is properly loaded.");
    return;
  }

  // Initialize the page
  function initializePage() {
    try {
      console.log('Initializing page...');
      
      // First, set up all event listeners
      setupEventListeners();
      
      // Then load form values from URL and render pills
      const urlParams = new URLSearchParams(window.location.search);
      console.log('URL Parameters:', urlParams.toString());
      
      if (urlParams.toString()) {
        console.log('Loading filters from URL parameters');
        loadFormFromURL();
      }
      
      // Initialize endpoint dropdown
      populateEndpointDropdown();
      
      // Finally, fetch initial logs
      fetchLogs();
      
    } catch (err) {
      console.error("Error initializing log page:", err);
    }
  }

  function setupEventListeners() {
    // Add form submit handler
    elements.searchForm.addEventListener("submit", (e) => {
      e.preventDefault();
      handleFilterChange();
    });

    // Add reset button handler
    if (elements.resetButton) {
      elements.resetButton.addEventListener("click", (e) => {
        e.preventDefault();
        elements.searchForm.reset();
        state.currentCursor = null;
        state.previousCursors = [];
        state.nextCursor = null;
        state.hasMore = false;
        state.cursorStack = [];
        
        const tagFilters = document.getElementById("tag-filters");
        if (tagFilters) tagFilters.innerHTML = "";
        
        renderActiveFiltersFromForm();
        updateURLFromForm();
        fetchLogs();
      });
    }

    // Add navigation event listener
    window.addEventListener("popstate", () => {
      console.log('Navigation occurred, reloading from URL');
      loadFormFromURL();
      renderActiveFiltersFromForm();
      fetchLogs();
    });

    // Setup form field event handlers
    bindFormAutoTags();
    
    // Setup pagination
    setupPaginationListeners();

    // Setup dropdown handlers
    setupDropdownHandlers();

    // Add event listeners for table interactions using event delegation
    elements.resultsTable.addEventListener("click", (e) => {
      // Handle copy button clicks
      if (e.target.classList.contains("copy-btn")) {
        const value = e.target.dataset.copy;
        if (value) {
          navigator.clipboard.writeText(value).then(() => {
            e.target.textContent = "✅";
            setTimeout(() => e.target.textContent = "📋", 1000);
          });
        }
        return;
      }

      // Handle tag clicks
      const tag = e.target?.dataset?.tag;
      if (tag) {
        const [key, value] = tag.split(":");
        if (!key || !value) return;

        const tagContainer = document.getElementById("tag-filters");
        if ([...tagContainer.querySelectorAll("span")].some(span => span.textContent.includes(`${key}:${value}`))) return;

        const span = document.createElement("span");
        span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
        span.innerHTML = `${key}:${value} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
        tagContainer.appendChild(span);
        updateURLFromForm();
        return;
      }

      // Handle filter key clicks
      if (e.target.matches("[data-filter-key]")) {
        e.preventDefault();

        const key = e.target.dataset.filterKey;
        const value = e.target.dataset.filterValue;
        if (!key || !value) return;

        // Check if this is a regular form field
        const formFields = ["contains", "source", "endpoint", "container", "app", "level", "category", "start", "end"];
        if (formFields.includes(key)) {
          // Handle as form field
          const mappings = {
            contains: "filter-keyword",
            source: "filter-source",
            endpoint: "endpoint-name",
            container: "container-name",
            app: "app-name",
            start: "start-time",
            end: "end-time"
          };

          const id = mappings[key];
          if (id) {
            const input = document.getElementById(id);
            if (input) {
              input.value = value;
              // Trigger change event to update filters
              input.dispatchEvent(new Event("change"));
            }
          } else if (key === "level" || key === "category") {
            // Handle checkbox fields
            document.querySelectorAll(`.filter-${key}-option`).forEach(cb => {
              if (cb.value === value) {
                cb.checked = true;
                // Trigger change event to update filters
                cb.dispatchEvent(new Event("change"));
              }
            });
          }
        } else {
          // Handle as custom tag
          addTagFilter(key, value);
          // Update URL and trigger search
          updateURLFromForm();
          fetchLogs();
        }
      }
    });

    // Use event delegation for tag filters
    document.getElementById("tag-filters").addEventListener("click", (e) => {
      if (!e.target.classList.contains("remove-tag")) return;

      const pill = e.target.closest("span");
      if (!pill) return;

      // Get the key and value from the data attributes
      const key = pill.dataset.key;
      const value = pill.dataset.value;
      if (!key || !value) return;

      console.log('Removing pill:', { key, value, isCustomTag: pill.dataset.customTag });

      // Remove matching form values
      if (key === "level") {
        document.querySelectorAll(".filter-level-option:checked").forEach(cb => {
          if (cb.value === value) {
            cb.checked = false;
            cb.dispatchEvent(new Event('change'));
          }
        });
      } else if (key === "category") {
        document.querySelectorAll(".filter-category-option:checked").forEach(cb => {
          if (cb.value === value) {
            cb.checked = false;
            cb.dispatchEvent(new Event('change'));
          }
        });
      } else {
        const mappings = {
          contains: "filter-keyword",
          source: "filter-source",
          container: "container-name",
          endpoint: "endpoint-name",
          app: "app-name",
          start: "start-time",
          end: "end-time"
        };
        
        const id = mappings[key];
        if (id) {
          const input = document.getElementById(id);
          if (input) {
            input.value = "";
            // Trigger both input and change events to ensure state updates
            input.dispatchEvent(new Event('input', { bubbles: true }));
            input.dispatchEvent(new Event('change', { bubbles: true }));
          }
        }
      }

      // Remove the pill visually
      pill.remove();

      // Reset pagination state
      state.currentCursor = null;
      state.previousCursors = [];
      state.nextCursor = null;
      state.hasMore = false;
      state.cursorStack = [];

      // Update URL and trigger new search
      updateURLFromForm();
      fetchLogs();
    });
  }

  function setupDropdownHandlers() {
    // Handle level dropdown
    const levelButton = document.getElementById("filter-level");
    const levelDropdown = document.getElementById("filter-level-dropdown");
    
    if (levelButton && levelDropdown) {
        levelButton.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            levelDropdown.classList.toggle("hidden");
        });
    }

    // Handle category dropdown
    const categoryButton = document.getElementById("filter-category");
    const categoryDropdown = document.getElementById("filter-category-dropdown");
    
    if (categoryButton && categoryDropdown) {
        categoryButton.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            categoryDropdown.classList.toggle("hidden");
        });
    }

    // Close dropdowns when clicking outside
    document.addEventListener("click", (e) => {
        if (levelDropdown && !levelButton?.contains(e.target) && !levelDropdown.contains(e.target)) {
            levelDropdown.classList.add("hidden");
        }
        if (categoryDropdown && !categoryButton?.contains(e.target) && !categoryDropdown.contains(e.target)) {
            categoryDropdown.classList.add("hidden");
        }
    });
  }

  // Centralized filter change handler
  function handleFilterChange() {
    // Reset pagination state
    state.currentCursor = null;
    state.previousCursors = [];
    state.nextCursor = null;
    state.hasMore = false;
    state.cursorStack = [];

    // Save existing custom tag filters
    const existingCustomTags = Array.from(document.querySelectorAll("#tag-filters span[data-custom-tag='true']"))
        .map(span => ({
            key: span.dataset.key,
            value: span.dataset.value
        }));

    // Render active filters from form
    renderActiveFiltersFromForm();

    // Re-add custom tags
    existingCustomTags.forEach(tag => {
        addTagFilter(tag.key, tag.value, false);
    });

    updateURLFromForm();
    fetchLogs();
  }

  function bindFormAutoTags() {
    const inputs = [
        "filter-keyword", "filter-source", "container-name", "endpoint-name",
        "app-name", "start-time", "end-time"
    ];

    // Add change handlers to all inputs
    inputs.forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            console.log(`Setting up handlers for input: ${id}`);
            
            // Add change event listener
            el.addEventListener("change", () => {
                console.log(`Change event triggered for ${id}`);
                handleFilterChange();
            });

            // Add input event for real-time filtering with debounce
            const debouncedHandler = debounce(() => {
                console.log(`Debounced input event triggered for ${id}`);
                handleFilterChange();
            }, 300);

            el.addEventListener("input", debouncedHandler);

            // Add Enter key handler
            el.addEventListener("keypress", (e) => {
                if (e.key === "Enter") {
                    console.log(`Enter pressed in ${id}`);
                    e.preventDefault();
                    handleFilterChange();
                }
            });
        }
    });

    // Add change handlers to checkboxes
    document.querySelectorAll(".filter-level-option, .filter-category-option").forEach(cb => {
        cb.addEventListener("change", () => {
            console.log('Checkbox changed:', cb.value);
            handleFilterChange();
        });
    });
  }

  // Start initialization
  initializePage();

  // Function to update pagination button states
  function updatePaginationButtons() {
    console.log('Updating pagination buttons:', {
      hasMore: state.hasMore,
      nextCursor: state.nextCursor,
      cursorStackLength: state.cursorStack.length
    });

    const updateButton = (button, enabled) => {
      if (button) {
        button.disabled = !enabled;
        button.classList.toggle('opacity-50', !enabled);
        button.classList.toggle('cursor-not-allowed', !enabled);
        button.classList.toggle('hover:bg-gray-50', enabled);
        button.classList.toggle('dark:hover:bg-gray-700', enabled);
      }
    };

    // Update desktop buttons
    const canGoPrev = state.cursorStack.length > 1;
    updateButton(elements.prevPageBtn, canGoPrev);
    updateButton(elements.nextPageBtn, state.hasMore);

    // Update mobile buttons
    updateButton(elements.prevPageMobileBtn, canGoPrev);
    updateButton(elements.nextPageMobileBtn, state.hasMore);

    // Update the cursor time display
    if (elements.cursorDisplay) {
      if (state.firstVisibleTime) {
        const displayTime = new Date(state.firstVisibleTime).toLocaleString();
        elements.cursorDisplay.textContent = displayTime;
      } else {
        elements.cursorDisplay.textContent = '-';
      }
    }
  }

  // Function to handle pagination
  function setupPaginationListeners() {
    const handlePrevPage = async () => {
      if (state.cursorStack.length > 1) {
        // Remove current cursor
        state.cursorStack.pop();
        // Get the previous cursor
        const prevCursor = state.cursorStack[state.cursorStack.length - 1];
        console.log('Navigating to previous page with cursor:', prevCursor);
        state.nextCursor = prevCursor;
        await fetchLogs(false);
      }
    };

    const handleNextPage = async () => {
      if (state.hasMore && state.nextCursor) {
        console.log('Navigating to next page with cursor:', state.nextCursor);
        await fetchLogs(true);
      }
    };

    // Desktop pagination
    if (elements.prevPageBtn) {
      elements.prevPageBtn.addEventListener('click', handlePrevPage);
    }
    if (elements.nextPageBtn) {
      elements.nextPageBtn.addEventListener('click', handleNextPage);
    }

    // Mobile pagination
    if (elements.prevPageMobileBtn) {
      elements.prevPageMobileBtn.addEventListener('click', handlePrevPage);
    }
    if (elements.nextPageMobileBtn) {
      elements.nextPageMobileBtn.addEventListener('click', handleNextPage);
    }
  }

  async function fetchLogs(isNextPage = false) {
    console.log('fetchLogs called with params:', { isNextPage, currentState: { ...state } });
    
    if (state.isLoading) {
      console.log('Fetch skipped - already loading');
      return;
    }
    state.isLoading = true;

    const LOGS_PER_PAGE = 50;

    try {
      const params = new URLSearchParams();

      // Get all filter values
      const filters = {
        keyword: document.getElementById("filter-keyword")?.value?.trim(),
        levels: Array.from(document.querySelectorAll(".filter-level-option:checked")).map(el => el.value),
        categories: Array.from(document.querySelectorAll(".filter-category-option:checked")).map(el => el.value),
        source: document.getElementById("filter-source")?.value?.trim(),
        container: document.getElementById("container-name")?.value?.trim(),
        endpoint: document.getElementById("endpoint-name")?.value?.trim(),
        app: document.getElementById("app-name")?.value?.trim(),
        start: document.getElementById("start-time")?.value,
        end: document.getElementById("end-time")?.value,
      };

      console.log('Current filters:', filters);

      // Add non-empty filters to params
      if (filters.keyword) params.set("contains", filters.keyword);
      filters.levels.forEach(level => params.append("level", level));
      filters.categories.forEach(cat => params.append("category", cat));
      if (filters.source) params.set("source", filters.source);
      if (filters.container) params.set("container", filters.container);
      if (filters.endpoint) params.set("endpoint", filters.endpoint);
      if (filters.app) params.set("app", filters.app);
      
      // Handle datetime filters
      if (filters.start) {
        const startDate = new Date(filters.start);
        if (!isNaN(startDate.getTime())) {
          params.set("start", startDate.toISOString());
        }
      }
      if (filters.end) {
        const endDate = new Date(filters.end);
        if (!isNaN(endDate.getTime())) {
          params.set("end", endDate.toISOString());
        }
      }

      // Add custom tag filters
      document.querySelectorAll("#tag-filters span[data-custom-tag='true']").forEach(span => {
        const key = span.dataset.key;
        const value = span.dataset.value;
        if (key && value) {
          params.append(`tag_${key.trim()}`, value.trim());
        }
      });

      // Handle cursor for pagination
      const cursorToUse = isNextPage ? state.nextCursor : state.currentCursor;
      if (cursorToUse) {
        console.log('Using cursor for request:', cursorToUse);
        params.set("cursor", cursorToUse);
      }

      params.set("limit", String(LOGS_PER_PAGE));
      params.set("order", "desc");

      console.log('Fetching with params:', params.toString());

      // Show loading state
      elements.resultsTable.innerHTML = `
        <tr>
          <td colspan="7" class="text-center py-8">
            <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white"></div>
            <div class="mt-2 text-sm text-gray-600 dark:text-gray-400">Loading logs...</div>
          </td>
        </tr>
      `;

      const res = await gosightFetch(`/api/v1/logs?${params.toString()}`);
      console.log('API Response status:', res.status);
      
      if (!res.ok) {
        throw new Error(`HTTP error! status: ${res.status}`);
      }
      
      const data = await res.json();
      console.log('API Response data:', data);
      
      const logs = data?.logs || [];
      
      if (logs.length === 0) {
        elements.resultsTable.innerHTML = `
          <tr>
            <td colspan="7" class="text-center py-4">
              <div class="text-gray-500 dark:text-gray-400">No logs found</div>
            </td>
          </tr>
        `;
        state.hasMore = false;
        state.nextCursor = null;
        state.currentCursor = null;
        updatePaginationButtons();
        updateURLFromForm();
        return;
      }

      // Always clear the results table when loading a new page
      elements.resultsTable.innerHTML = "";
      
      renderLogs(logs, false);

      // Update pagination state
      state.hasMore = data.has_more;
      
      if (logs.length > 0) {
        const firstLog = logs[0];
        const lastLog = logs[logs.length - 1];
        
        state.firstVisibleTime = firstLog.timestamp;
        state.lastVisibleTime = lastLog.timestamp;

        // Update cursors
        if (isNextPage) {
          // When going to next page, current becomes what was next
          state.currentCursor = state.nextCursor;
          // And next becomes the new next cursor from response
          state.nextCursor = data.next_cursor || lastLog.timestamp;
        } else {
          // For initial load or prev page, current is what we used to fetch
          state.currentCursor = cursorToUse;
          // And next is the new next cursor
          state.nextCursor = data.next_cursor || lastLog.timestamp;
        }
        
        console.log('Updated cursors:', {
          current: state.currentCursor,
          next: state.nextCursor,
          isNextPage
        });

        // Update cursor stack for navigation
        if (isNextPage) {
          state.cursorStack.push(state.currentCursor);
        } else if (!cursorToUse) {
          // Reset cursor stack when doing a new search
          state.cursorStack = [state.currentCursor];
        }
      }

      // Update pagination buttons
      updatePaginationButtons();

      // Update log count
      if (elements.logCount) {
        elements.logCount.textContent = `${logs.length} results`;
      }

      // Update URL after everything is done
      updateURLFromForm();

    } catch (error) {
      console.error('Error fetching logs:', error);
      elements.resultsTable.innerHTML = `
        <tr>
          <td colspan="7" class="text-center py-4">
            <div class="text-red-500">Error loading logs: ${error.message}</div>
          </td>
        </tr>
      `;
    } finally {
      state.isLoading = false;
    }
  }

  function getLogKey(log) {
    return `${log.timestamp}_${log.message}`;
  }
  function truncateMessage(message, maxLines = 3) {
    if (!message) return message;
    
    // First split by newlines
    const lines = message.split('\n');
    
    if (lines.length <= maxLines) {
      // If it's a single line but very long, truncate it
      if (lines.length === 1 && message.length > 150) {
        return message.substring(0, 150) + '...';
      }
      return message;
    }
    
    // Take first few lines and add indication of how many more lines exist
    const truncated = lines.slice(0, maxLines).join('\n');
    return `${truncated}\n[... ${lines.length - maxLines} more lines ...]`;
  }
  function renderLogs(logs, append = false) {
    if (!append) {
      elements.resultsTable.innerHTML = "";
    }

    if (logs.length === 0 && !append) {
      elements.resultsTable.innerHTML = `
        <tr>
          <td colspan="7" class="text-center py-4 text-gray-500 dark:text-gray-400">
            No logs found matching your criteria
          </td>
        </tr>
      `;
      return;
    }

    // Create a document fragment to batch DOM operations
    const fragment = document.createDocumentFragment();

    logs.forEach((log, i) => {
      const timestamp = new Date(log.timestamp).toLocaleString();
      const logKey = getLogKey(log);
      const isExpanded = state.expandedLogKeys.has(logKey);
      const truncatedMessage = truncateMessage(log.message);

      // Main log row
      const row = document.createElement("tr");
      row.className = `divide-y divide-gray-100 dark:divide-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800 ${isExpanded ? 'bg-gray-50 dark:bg-gray-800' : ''}`;
      row.dataset.logKey = logKey;
      
      row.innerHTML = `
        <td class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">${timestamp}</td>
        <td class="px-4 py-2">
          <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getColor(log.level)}">${sanitize(log.level || 'info').toLowerCase()}</span>
        </td>
        <td class="px-4 py-2 text-sm">${sanitize(log.source || '')}</td>
        <td class="px-4 py-2 text-sm">
          ${sanitize(log.tags?.container_name || log.tags?.hostname || '')}
        </td>
        <td class="px-4 py-2 text-xs text-gray-700 dark:text-gray-200 font-mono max-w-[400px] w-full">
          ${sanitize(truncatedMessage).split('\n').map(line => 
            `<div class="truncate">${sanitize(line)}</div>`
          ).join('')}
        </td>
        <td class="px-4 py-2 text-sm">${sanitize(log.meta?.user || '')}</td>
        <td class="px-4 py-2 text-sm">
          <button class="text-blue-600 hover:underline text-xs expand-log" data-log-key="${sanitize(logKey)}">
            ${isExpanded ? 'Hide Details' : 'Show Details'}
          </button>
        </td>
      `;

      // Details row
      const detailsRow = document.createElement("tr");
      detailsRow.className = `details-row ${isExpanded ? '' : 'hidden'} bg-white dark:bg-gray-900 divide-y divide-gray-100 dark:divide-gray-800`;
      detailsRow.dataset.logKey = logKey;
      
      detailsRow.innerHTML = `
        <td colspan="7" class="px-6 py-4">
          <div class="space-y-4">
            <!-- Full Message Section -->
            <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-100 dark:border-gray-700 shadow-sm">
              <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
                <h3 class="text-xs uppercase font-medium text-gray-500 dark:text-gray-400">Full Message</h3>
              </div>
              <div class="p-4">
                <pre class="text-xs text-gray-700 dark:text-gray-300 font-mono whitespace-pre-wrap break-words">${sanitize(log.message)}</pre>
              </div>
            </div>

            <!-- Metadata Sections -->
            <div class="grid grid-cols-3 gap-4">
              ${renderSection("Tags", log.tags)}
              ${renderSection("Meta", filterOutExtra(log.meta))}
              ${renderSection("Fields", log.fields)}
            </div>
          </div>
        </td>
      `;

      // Add click handler for expand/collapse
      const expandButton = row.querySelector(".expand-log");
      expandButton.addEventListener("click", (e) => {
        e.preventDefault();
        e.stopPropagation();

        const logKey = e.target.dataset.logKey;
        const isCurrentlyExpanded = state.expandedLogKeys.has(logKey);
        const mainRow = e.target.closest('tr');
        const detailsRow = mainRow.nextElementSibling;

        if (!detailsRow || !detailsRow.classList.contains('details-row')) {
          console.error('Details row not found for log:', logKey);
          return;
        }
        
        if (isCurrentlyExpanded) {
          state.expandedLogKeys.delete(logKey);
          mainRow.classList.remove("bg-gray-50", "dark:bg-gray-800");
          e.target.textContent = "Show Details";
          detailsRow.classList.add("hidden");
        } else {
          state.expandedLogKeys.add(logKey);
          mainRow.classList.add("bg-gray-50", "dark:bg-gray-800");
          e.target.textContent = "Hide Details";
          detailsRow.classList.remove("hidden");
        }
      });

      fragment.appendChild(row);
      fragment.appendChild(detailsRow);
    });

    // Batch append all rows at once
    elements.resultsTable.appendChild(fragment);

    // Update cursor display
    if (logs.length > 0) {
      state.currentCursor = state.nextCursor;
      if (elements.cursorDisplay) {
        elements.cursorDisplay.textContent = new Date(state.currentCursor).toLocaleString();
      }
    }
  }
  function filterOutExtra(meta) {
    if (!meta || typeof meta !== "object") return {};
    const copy = { ...meta };
    delete copy.extra;
    return copy;
  }
  function formatTimestampForDisplay(isoString) {
    const date = new Date(isoString);
    return date.toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short"
    });
  }
  function updateURLFromForm() {
    const params = new URLSearchParams();

    // Get all filter values from form fields
    const filters = {
      keyword: document.getElementById("filter-keyword")?.value?.trim(),
      levels: Array.from(document.querySelectorAll(".filter-level-option:checked")).map(el => el.value),
      categories: Array.from(document.querySelectorAll(".filter-category-option:checked")).map(el => el.value),
      source: document.getElementById("filter-source")?.value?.trim(),
      container: document.getElementById("container-name")?.value?.trim() || "",
      endpoint: document.getElementById("endpoint-name")?.value?.trim() || "",
      app: document.getElementById("app-name")?.value?.trim() || "",
      start: document.getElementById("start-time")?.value,
      end: document.getElementById("end-time")?.value,
    };

    // Add non-empty filters to URL
    if (filters.keyword) params.set("contains", filters.keyword);
    filters.levels.forEach(level => params.append("level", level));
    filters.categories.forEach(cat => params.append("category", cat));
    if (filters.source) params.set("source", filters.source);
    if (filters.container) params.set("container", filters.container);
    if (filters.endpoint) params.set("endpoint", filters.endpoint);
    if (filters.app) params.set("app", filters.app);
    if (filters.start) params.set("start", new Date(filters.start).toISOString());
    if (filters.end) params.set("end", new Date(filters.end).toISOString());

    // Add all tag filters
    document.querySelectorAll("#tag-filters span").forEach(span => {
      const key = span.dataset.key;
      const value = span.dataset.value;
      const isCustomTag = span.dataset.customTag === "true";
      
      if (key && value) {
        if (isCustomTag) {
          // For custom tags, add with tag_ prefix
          params.append(`tag_${key.trim()}`, value.trim());
        }
      }
    });

    // Add cursor if exists and is not null
    if (state.currentCursor) {
      params.set("cursor", state.currentCursor);
    }

    // Update URL without reloading page
    const newURL = `${window.location.pathname}?${params.toString()}`;
    console.log('Updating URL with params:', params.toString());
    history.replaceState(null, "", newURL);
  }

  function renderSection(title, obj) {
    if (!obj || typeof obj !== "object" || Object.keys(obj).length === 0) return "";

    const entries = Object.entries(obj);
    const rows = entries.map(([key, val]) => {
      const stringVal = typeof val === "object" ? JSON.stringify(val, null, 2) : String(val);
      const safeKey = sanitize(key);
      const displayVal = sanitize(stringVal);
      const safeFilterValue = stringVal.replace(/"/g, '&quot;'); // Escape quotes for data attributes

      return `
        <tr>
          <td class="text-xs font-semibold text-gray-600 dark:text-gray-300 py-1 min-w-[100px]">${safeKey}</td>
          <td class="text-xs text-gray-700 dark:text-gray-300 font-mono py-1 pl-4">${displayVal}</td>
          <td class="text-xs py-1 pl-2 whitespace-nowrap">
            <button class="text-blue-600 hover:underline text-xs font-medium add-filter"
              data-filter-key="${safeKey}"
              data-filter-value="${safeFilterValue}">
              + Add Filter
            </button>
          </td>
        </tr>
      `;
    }).join("");

    return `
      <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-100 dark:border-gray-700 shadow-sm">
        <div class="px-4 py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
          <h3 class="text-xs font-semibold text-gray-600 dark:text-gray-300">${sanitize(title)}</h3>
        </div>
        <div class="p-4">
          <table class="w-full">
            <tbody>
              ${rows}
            </tbody>
          </table>
        </div>
      </div>
    `;
  }

  function loadFormFromURL() {
    const params = new URLSearchParams(window.location.search);
    console.log('Loading form from URL params:', params.toString());

    // Helper function to safely set input value
    const setInputValue = (id, value) => {
      const el = document.getElementById(id);
      if (el) el.value = value || '';
    };

    // Helper function to safely set checkbox state
    const setCheckboxState = (selector, values) => {
      document.querySelectorAll(selector).forEach(checkbox => {
        // Handle both 'warn' and 'warning' cases for loading
        const checkboxValue = checkbox.value;
        const matches = values.some(v => {
          const value = v.toLowerCase();
          return value === checkboxValue.toLowerCase() || 
                 (value === 'warn' && checkboxValue.toLowerCase() === 'warning') ||
                 (value === 'warning' && checkboxValue.toLowerCase() === 'warn');
        });
        checkbox.checked = matches;
      });
    };

    // Reset all form fields
    setInputValue("filter-keyword", "");
    setInputValue("filter-source", "");
    setInputValue("endpoint-name", "");
    setInputValue("container-name", "");
    setInputValue("app-name", "");
    setInputValue("start-time", "");
    setInputValue("end-time", "");

    // Clear all checkboxes
    document.querySelectorAll(".filter-level-option, .filter-category-option").forEach(cb => cb.checked = false);

    // Clear tag filters
    const tagFilters = document.getElementById("tag-filters");
    if (tagFilters) tagFilters.innerHTML = "";

    // Load values from URL parameters
    if (params.has("contains")) setInputValue("filter-keyword", params.get("contains"));
    if (params.has("source")) setInputValue("filter-source", params.get("source"));
    if (params.has("endpoint")) setInputValue("endpoint-name", params.get("endpoint"));
    if (params.has("container")) setInputValue("container-name", params.get("container"));
    if (params.has("app")) setInputValue("app-name", params.get("app"));

    // Handle datetime inputs
    if (params.has("start")) {
      const startDate = new Date(params.get("start"));
      if (!isNaN(startDate.getTime())) {
        setInputValue("start-time", startDate.toISOString().slice(0, 16));
      }
    }
    
    if (params.has("end")) {
      const endDate = new Date(params.get("end"));
      if (!isNaN(endDate.getTime())) {
        setInputValue("end-time", endDate.toISOString().slice(0, 16));
      }
    }

    // Set checkboxes for levels and categories
    setCheckboxState(".filter-level-option", params.getAll("level"));
    setCheckboxState(".filter-category-option", params.getAll("category"));

    // Handle custom tag filters
    const tagFiltersToAdd = [];
    for (const [key, value] of params.entries()) {
      if (key.startsWith("tag_")) {
        const tagKey = key.slice(4); // Remove 'tag_' prefix
        // Skip if this is a regular form field
        if (!["contains", "source", "endpoint", "container", "app", "level", "category", "start", "end"].includes(tagKey)) {
          tagFiltersToAdd.push({ key: tagKey, value });
        }
      }
    }

    // Add all tag filters after clearing
    tagFiltersToAdd.forEach(({ key, value }) => {
      const span = document.createElement("span");
      span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
      span.dataset.customTag = "true";
      span.dataset.key = key;
      span.dataset.value = value;
      span.innerHTML = `${sanitize(key)}:${sanitize(value)} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
      tagFilters.appendChild(span);
    });

    // Handle cursor
    const cursor = params.get("cursor");
    if (cursor) {
      state.currentCursor = cursor;
    }

    // Render active filters after all form fields and tags are set
    renderActiveFiltersFromForm();
  }

  function sanitize(input) {
    if (input === null || input === undefined) {
      return '';
    }
    
    // Convert to string if not already
    const str = String(input);
    
    // Create a map for HTML entities
    const entityMap = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;',
      '/': '&#x2F;',
      '`': '&#x60;',
      '=': '&#x3D;'
    };
    
    // Replace special characters with HTML entities
    return str.replace(/[&<>"'`=\/]/g, function(s) {
      return entityMap[s];
    });
  }

  function getColor(level) {
    switch ((level || "").toLowerCase()) {
      case "emergency": return "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-200";
      case "alert": return "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-200";
      case "critical": return "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-200";
      case "error": return "bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-200";
      case "warn":
      case "warning": return "bg-orange-100 text-orange-800 dark:bg-orange-800 dark:text-orange-100";
      case "info": return "bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100";
      case "debug": return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300";
      case "notice": return "bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-200";
      default: return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";
    }
  }

  // Add debounce function at the top level
  function debounce(func, wait, immediate = false) {
    let timeout;
    
    return function executedFunction(...args) {
      const context = this;
      
      const later = () => {
        timeout = null;
        if (!immediate) func.apply(context, args);
      };
      
      const callNow = immediate && !timeout;
      
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
      
      if (callNow) func.apply(context, args);
    };
  }

  // Modify addTagFilter to accept optional parameter to control fetchLogs
  function addTagFilter(key, value, shouldFetchLogs = true) {
    // Don't add regular search parameters as tags
    const searchParams = [
      "contains", "source", "container", "endpoint", 
      "app", "start", "end", "level", "category"
    ];
    if (searchParams.includes(key)) return;

    const tagContainer = document.getElementById("tag-filters");
    if (!tagContainer) return;
    
    // Check if filter already exists
    const existingFilter = Array.from(tagContainer.children).find(span => {
      return span.dataset.key === key && span.dataset.value === value;
    });
    
    if (existingFilter) return;

    const span = document.createElement("span");
    span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
    span.dataset.customTag = "true";
    span.dataset.key = key;
    span.dataset.value = value;
    span.innerHTML = `${sanitize(key)}:${sanitize(value)} <button class="ml-1 text-xs remove-tag" type="button">&times;</button>`;
    tagContainer.appendChild(span);

    // Update URL and optionally trigger search
    updateURLFromForm();
    if (shouldFetchLogs) {
      fetchLogs();
    }
  }

  async function populateEndpointDropdown() {
    const dropdown = document.getElementById("endpoint-dropdown");
    const input = document.getElementById("endpoint-name");
    let allItems = [];
    let isLoading = false;

    async function fetchEndpoints() {
      if (isLoading) return;
      isLoading = true;

      dropdown.innerHTML = `
        <div class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
          Loading endpoints...
        </div>
      `;
      dropdown.classList.remove("hidden");

      try {
        const [hosts, containers] = await Promise.all([
          gosightFetch("/api/v1/endpoints/hosts").then(res => res.ok ? res.json() : []),
          gosightFetch("/api/v1/endpoints/containers").then(res => res.ok ? res.json() : []),
        ]);

        allItems = [
          { label: "Hosts", items: hosts.map(h => h.hostname).filter(Boolean) },
          { label: "Containers", items: containers.map(c => c.Name ?? "").filter(Boolean) },
        ];

        updateDropdown(input.value.toLowerCase());
      } catch (err) {
        console.error("Failed to load endpoints:", err);
        dropdown.innerHTML = `
          <div class="px-4 py-2 text-sm text-red-500">
            Failed to load endpoints. Please try again.
          </div>
        `;
      } finally {
        isLoading = false;
      }
    }

    function updateDropdown(searchValue) {
      dropdown.innerHTML = "";
      dropdown.classList.remove("hidden");

      const matchedItems = allItems.map(group => ({
        label: group.label,
        items: group.items.filter(item => 
          item.toLowerCase().includes(searchValue)
        )
      })).filter(group => group.items.length > 0);

      if (matchedItems.length === 0) {
        dropdown.innerHTML = `
          <div class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
            No matching endpoints found
          </div>
        `;
        return;
      }

      matchedItems.forEach(group => {
        // Add group label
        const groupLabel = document.createElement("div");
        groupLabel.className = "px-3 py-1 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase bg-gray-50 dark:bg-gray-800";
        groupLabel.textContent = group.label;
        dropdown.appendChild(groupLabel);

        // Add group items
        group.items.forEach(item => {
          const itemEl = document.createElement("div");
          itemEl.className = "cursor-pointer px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-800 dark:text-white";
          
          // Highlight matching text
          const itemText = item;
          const matchIndex = itemText.toLowerCase().indexOf(searchValue);
          if (matchIndex >= 0) {
            const before = itemText.slice(0, matchIndex);
            const match = itemText.slice(matchIndex, matchIndex + searchValue.length);
            const after = itemText.slice(matchIndex + searchValue.length);
            itemEl.innerHTML = `${before}<span class="bg-yellow-200 dark:bg-yellow-900">${match}</span>${after}`;
          } else {
            itemEl.textContent = itemText;
          }

          itemEl.addEventListener("click", () => {
            input.value = item;
            dropdown.classList.add("hidden");
            renderActiveFiltersFromForm();
            updateURLFromForm();
            fetchLogs();
          });
          
          dropdown.appendChild(itemEl);
        });
      });
    }

    // Initial load
    fetchEndpoints();

    // Setup input handlers
    let debounceTimer;
    input.addEventListener("input", (e) => {
      const searchValue = e.target.value.toLowerCase();
      clearTimeout(debounceTimer);
      
      if (!allItems.length) {
        fetchEndpoints();
        return;
      }

      debounceTimer = setTimeout(() => {
        updateDropdown(searchValue);
      }, 150);
    });

    // Handle focus
    input.addEventListener("focus", () => {
      if (!allItems.length) {
        fetchEndpoints();
      } else {
        updateDropdown(input.value.toLowerCase());
      }
    });

    // Handle click outside
    document.addEventListener("click", (e) => {
      if (!input.contains(e.target) && !dropdown.contains(e.target)) {
        dropdown.classList.add("hidden");
      }
    });

    // Handle keyboard navigation
    input.addEventListener("keydown", (e) => {
      const items = dropdown.querySelectorAll(".cursor-pointer");
      const currentIndex = Array.from(items).findIndex(item => item.classList.contains("bg-gray-100"));

      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          if (items.length === 0) return;
          
          if (currentIndex === -1) {
            items[0].classList.add("bg-gray-100", "dark:bg-gray-700");
          } else {
            items[currentIndex].classList.remove("bg-gray-100", "dark:bg-gray-700");
            const nextIndex = (currentIndex + 1) % items.length;
            items[nextIndex].classList.add("bg-gray-100", "dark:bg-gray-700");
            items[nextIndex].scrollIntoView({ block: "nearest" });
          }
          break;

        case "ArrowUp":
          e.preventDefault();
          if (items.length === 0) return;
          
          if (currentIndex === -1) {
            items[items.length - 1].classList.add("bg-gray-100", "dark:bg-gray-700");
          } else {
            items[currentIndex].classList.remove("bg-gray-100", "dark:bg-gray-700");
            const prevIndex = (currentIndex - 1 + items.length) % items.length;
            items[prevIndex].classList.add("bg-gray-100", "dark:bg-gray-700");
            items[prevIndex].scrollIntoView({ block: "nearest" });
          }
          break;

        case "Enter":
          e.preventDefault();
          const selectedItem = dropdown.querySelector(".bg-gray-100, .dark:bg-gray-700");
          if (selectedItem) {
            selectedItem.click();
          }
          break;

        case "Escape":
          dropdown.classList.add("hidden");
          input.blur();
          break;
      }
    });
  }

  function renderActiveFiltersFromForm() {
    const tagContainer = document.getElementById("tag-filters");
    
    // Save existing custom tags before clearing
    const existingCustomTags = Array.from(tagContainer.querySelectorAll("span[data-custom-tag='true']"))
      .map(span => ({
        key: span.dataset.key,
        value: span.dataset.value,
        element: span.cloneNode(true)
      }));
    
    tagContainer.innerHTML = ""; // clear all

    // Add pills for regular search inputs
    const mappings = {
      contains: "filter-keyword",
      source: "filter-source",
      container: "container-name",
      endpoint: "endpoint-name",
      app: "app-name",
      start: "start-time",
      end: "end-time"
    };

    for (const [key, id] of Object.entries(mappings)) {
      const val = document.getElementById(id)?.value?.trim();
      if (val) {
        const span = document.createElement("span");
        span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
        span.dataset.customTag = "false"; // Mark as not a custom tag
        span.dataset.key = key;
        span.dataset.value = val;

        const display = (key === "start" || key === "end")
          ? `${key}: ${formatTimestampForDisplay(val)}`
          : `${key}:${val}`;

        const textSpan = document.createElement("span");
        textSpan.className = "mr-1";
        textSpan.textContent = display;

        const button = document.createElement("button");
        button.type = "button";
        button.className = "ml-1 text-xs remove-tag";
        button.title = "Remove filter";
        button.textContent = "×";

        span.appendChild(textSpan);
        span.appendChild(button);

        tagContainer.appendChild(span);
      }
    }

    // Add level and category filters as pills
    const getChecked = cls =>
      Array.from(document.querySelectorAll(`.${cls}:checked`)).map(el => el.value);

    getChecked("filter-level-option").forEach(v => {
      const span = document.createElement("span");
      span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
      span.dataset.customTag = "false"; // Mark as not a custom tag
      span.dataset.key = "level";
      span.dataset.value = v;
      const textSpan = document.createElement("span");
      textSpan.className = "mr-1";
      textSpan.textContent = `level:${v}`;

      const button = document.createElement("button");
      button.type = "button";
      button.className = "ml-1 text-xs remove-tag";
      button.title = "Remove filter";
      button.textContent = "×";

      span.appendChild(textSpan);
      span.appendChild(button);
      tagContainer.appendChild(span);
    });

    getChecked("filter-category-option").forEach(v => {
      const span = document.createElement("span");
      span.className = "inline-flex items-center px-3 py-1 rounded-sm text-sm font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100";
      span.dataset.customTag = "false"; // Mark as not a custom tag
      span.dataset.key = "category";
      span.dataset.value = v;
      const textSpan = document.createElement("span");
      textSpan.className = "mr-1";
      textSpan.textContent = `category:${v}`;

      const button = document.createElement("button");
      button.type = "button";
      button.className = "ml-1 text-xs remove-tag";
      button.title = "Remove filter";
      button.textContent = "×";

      span.appendChild(textSpan);
      span.appendChild(button);
      tagContainer.appendChild(span);
    });

    // Re-add saved custom tags
    existingCustomTags.forEach(tag => {
      tagContainer.appendChild(tag.element);
    });
  }
});
