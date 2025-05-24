<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let currentPage = 1;
	export let totalPages = 1;
	export let totalItems = 0;
	export let pageSize = 10;
	export let showPerPage = true;
	export let showInfo = true;
	export let maxVisiblePages = 7;
	export let pageSizeOptions = [10, 20, 50, 100];

	const dispatch = createEventDispatcher();

	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages && page !== currentPage) {
			dispatch('pageChange', page);
		}
	}

	function goToPrevious() {
		goToPage(currentPage - 1);
	}

	function goToNext() {
		goToPage(currentPage + 1);
	}

	function changePageSize(newSize: number) {
		dispatch('pageSizeChange', newSize);
	}

	function getVisiblePages(): number[] {
		const pages: number[] = [];
		const half = Math.floor(maxVisiblePages / 2);
		let start = Math.max(1, currentPage - half);
		let end = Math.min(totalPages, start + maxVisiblePages - 1);

		// Adjust start if we're near the end
		if (end - start + 1 < maxVisiblePages) {
			start = Math.max(1, end - maxVisiblePages + 1);
		}

		for (let i = start; i <= end; i++) {
			pages.push(i);
		}

		return pages;
	}

	$: visiblePages = getVisiblePages();
	$: startItem = (currentPage - 1) * pageSize + 1;
	$: endItem = Math.min(currentPage * pageSize, totalItems);
</script>

{#if totalPages > 1}
	<div
		class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6 dark:border-gray-700 dark:bg-gray-800"
	>
		<!-- Mobile pagination -->
		<div class="flex flex-1 justify-between sm:hidden">
			<button
				on:click={goToPrevious}
				disabled={currentPage === 1}
				class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
			>
				Previous
			</button>
			<button
				on:click={goToNext}
				disabled={currentPage === totalPages}
				class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
			>
				Next
			</button>
		</div>

		<!-- Desktop pagination -->
		<div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
			<div class="flex items-center gap-4">
				{#if showInfo}
					<p class="text-sm text-gray-700 dark:text-gray-300">
						Showing
						<span class="font-medium">{startItem}</span>
						to
						<span class="font-medium">{endItem}</span>
						of
						<span class="font-medium">{totalItems}</span>
						results
					</p>
				{/if}

				{#if showPerPage}
					<div class="flex items-center gap-2">
						<label for="pageSize" class="text-sm text-gray-700 dark:text-gray-300">
							Per page:
						</label>
						<select
							id="pageSize"
							value={pageSize}
							on:change={(e) => changePageSize(parseInt((e.target as HTMLSelectElement).value))}
							class="rounded-md border border-gray-300 bg-white px-2 py-1 text-sm text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						>
							{#each pageSizeOptions as option}
								<option value={option}>{option}</option>
							{/each}
						</select>
					</div>
				{/if}
			</div>

			<nav
				class="relative z-0 inline-flex -space-x-px rounded-md shadow-sm"
				aria-label="Pagination"
			>
				<!-- Previous button -->
				<button
					on:click={goToPrevious}
					disabled={currentPage === 1}
					class="relative inline-flex items-center rounded-l-md border border-gray-300 bg-white px-2 py-2 text-sm font-medium text-gray-500 hover:bg-gray-50 focus:z-10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					aria-label="Previous page"
				>
					<i class="fas fa-chevron-left"></i>
				</button>

				<!-- First page -->
				{#if visiblePages[0] > 1}
					<button
						on:click={() => goToPage(1)}
						class="relative inline-flex items-center border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:z-10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					>
						1
					</button>
					{#if visiblePages[0] > 2}
						<span
							class="relative inline-flex items-center border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300"
						>
							...
						</span>
					{/if}
				{/if}

				<!-- Visible pages -->
				{#each visiblePages as page}
					<button
						on:click={() => goToPage(page)}
						class="relative inline-flex items-center border px-4 py-2 text-sm font-medium focus:z-10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none {page ===
						currentPage
							? 'z-10 border-blue-500 bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400'
							: 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
						aria-current={page === currentPage ? 'page' : undefined}
					>
						{page}
					</button>
				{/each}

				<!-- Last page -->
				{#if visiblePages[visiblePages.length - 1] < totalPages}
					{#if visiblePages[visiblePages.length - 1] < totalPages - 1}
						<span
							class="relative inline-flex items-center border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300"
						>
							...
						</span>
					{/if}
					<button
						on:click={() => goToPage(totalPages)}
						class="relative inline-flex items-center border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:z-10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					>
						{totalPages}
					</button>
				{/if}

				<!-- Next button -->
				<button
					on:click={goToNext}
					disabled={currentPage === totalPages}
					class="relative inline-flex items-center rounded-r-md border border-gray-300 bg-white px-2 py-2 text-sm font-medium text-gray-500 hover:bg-gray-50 focus:z-10 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					aria-label="Next page"
				>
					<i class="fas fa-chevron-right"></i>
				</button>
			</nav>
		</div>
	</div>
{/if}
