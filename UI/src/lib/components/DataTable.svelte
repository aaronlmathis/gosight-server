<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let data: any[] = [];
	export let columns: Array<{
		key: string;
		label: string;
		sortable?: boolean;
		width?: string;
		align?: 'left' | 'center' | 'right';
		render?: (value: any, row: any) => string;
	}> = [];
	export let loading = false;
	export let error = '';
	export let emptyMessage = 'No data available';
	export let emptyIcon = 'fas fa-inbox';
	export let sortBy = '';
	export let sortOrder: 'asc' | 'desc' = 'desc';
	export let selectable = false;
	export let selectedRows: any[] = [];
	export let stickyHeader = false;
	export let hoverable = true;
	export let striped = false;

	const dispatch = createEventDispatcher();

	function handleSort(column: any) {
		if (!column.sortable) return;
		
		if (sortBy === column.key) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortBy = column.key;
			sortOrder = 'desc';
		}
		
		dispatch('sort', { column: column.key, order: sortOrder });
	}

	function handleRowClick(row: any, index: number) {
		dispatch('rowClick', { row, index });
	}

	function toggleRowSelection(row: any) {
		const isSelected = selectedRows.some(selected => selected.id === row.id);
		if (isSelected) {
			selectedRows = selectedRows.filter(selected => selected.id !== row.id);
		} else {
			selectedRows = [...selectedRows, row];
		}
		dispatch('selectionChange', selectedRows);
	}

	function toggleAllSelection() {
		if (selectedRows.length === data.length) {
			selectedRows = [];
		} else {
			selectedRows = [...data];
		}
		dispatch('selectionChange', selectedRows);
	}

	function isRowSelected(row: any): boolean {
		return selectedRows.some(selected => selected.id === row.id);
	}

	function getCellValue(row: any, column: any): any {
		const value = row[column.key];
		return column.render ? column.render(value, row) : value;
	}

	function getColumnClass(column: any): string {
		let classes = 'px-6 py-3 text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider';
		
		if (column.align === 'center') classes += ' text-center';
		else if (column.align === 'right') classes += ' text-right';
		else classes += ' text-left';
		
		if (column.sortable) classes += ' cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600';
		
		return classes;
	}

	function getCellClass(column: any): string {
		let classes = 'px-6 py-4 whitespace-nowrap text-sm';
		
		if (column.align === 'center') classes += ' text-center';
		else if (column.align === 'right') classes += ' text-right';
		else classes += ' text-left';
		
		return classes;
	}
</script>

<div class="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow">
	{#if error}
		<div class="p-6 bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800">
			<div class="flex">
				<i class="fas fa-exclamation-triangle text-red-500 mr-3 mt-0.5"></i>
				<div>
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<p class="text-sm text-red-600 dark:text-red-300 mt-1">{error}</p>
				</div>
			</div>
		</div>
	{/if}

	<div class="overflow-x-auto">
		<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
			<thead class="bg-gray-50 dark:bg-gray-700 {stickyHeader ? 'sticky top-0 z-10' : ''}">
				<tr>
					{#if selectable}
						<th class="px-6 py-3 text-left">
							<input
								type="checkbox"
								checked={data.length > 0 && selectedRows.length === data.length}
								indeterminate={selectedRows.length > 0 && selectedRows.length < data.length}
								on:change={toggleAllSelection}
								class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
							/>
						</th>
					{/if}
					{#each columns as column}
						<th
							class="{getColumnClass(column)} {column.width ? `w-${column.width}` : ''}"
							on:click={() => handleSort(column)}
						>
							<div class="flex items-center gap-1">
								<span>{column.label}</span>
								{#if column.sortable}
									{#if sortBy === column.key}
										<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} text-xs"></i>
									{:else}
										<i class="fas fa-sort text-xs opacity-50"></i>
									{/if}
								{/if}
							</div>
						</th>
					{/each}
				</tr>
			</thead>
			<tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
				{#if loading}
					<tr>
						<td colspan={columns.length + (selectable ? 1 : 0)} class="px-6 py-12 text-center">
							<div class="flex justify-center items-center">
								<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
								<span class="ml-3 text-gray-600 dark:text-gray-400">Loading...</span>
							</div>
						</td>
					</tr>
				{:else if data.length === 0}
					<tr>
						<td colspan={columns.length + (selectable ? 1 : 0)} class="px-6 py-12 text-center">
							<div class="flex flex-col items-center">
								<i class="{emptyIcon} text-4xl text-gray-400 mb-4"></i>
								<p class="text-gray-600 dark:text-gray-400">{emptyMessage}</p>
							</div>
						</td>
					</tr>
				{:else}
					{#each data as row, index}
						<tr
							class="{hoverable ? 'hover:bg-gray-50 dark:hover:bg-gray-700' : ''} {striped && index % 2 === 1 ? 'bg-gray-50 dark:bg-gray-700/50' : ''} {isRowSelected(row) ? 'bg-blue-50 dark:bg-blue-900/20' : ''} cursor-pointer"
							on:click={() => handleRowClick(row, index)}
						>
							{#if selectable}
								<td class="px-6 py-4">
									<input
										type="checkbox"
										checked={isRowSelected(row)}
										on:change|stopPropagation={() => toggleRowSelection(row)}
										class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
									/>
								</td>
							{/if}
							{#each columns as column}
								<td class="{getCellClass(column)} text-gray-900 dark:text-white">
									{@html getCellValue(row, column)}
								</td>
							{/each}
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>
