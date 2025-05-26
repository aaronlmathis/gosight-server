<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Dropdown, Badge } from 'flowbite-svelte';
	import CompatButton from '$lib/components/CompatButton.svelte';
	import CompatDropdownItem from '$lib/components/CompatDropdownItem.svelte';
	import { dashboardStore, isEditMode, hasUnsavedChanges } from '$lib/stores/dashboard';
	import {
		Edit3,
		Save,
		Plus,
		MoreVertical,
		Copy,
		Trash2,
		Settings,
		Grid3x3,
		Eye
	} from 'lucide-svelte';

	export let activeDashboardId: string;

	const dispatch = createEventDispatcher<{
		addWidget: void;
		createDashboard: void;
		deleteDashboard: { id: string };
		duplicateDashboard: { id: string };
		configureDashboard: { id: string };
	}>();

	$: activeDashboard = $dashboardStore.dashboards.find((d) => d.id === activeDashboardId);
	$: unsavedChanges = $hasUnsavedChanges;

	function toggleEditMode() {
		isEditMode.update((mode) => !mode);
	}

	function saveDashboard() {
		if (activeDashboard) {
			dashboardStore.saveDashboard(activeDashboard.id);
		}
	}

	function handleAddWidget() {
		dispatch('addWidget');
	}

	function handleCreateDashboard() {
		dispatch('createDashboard');
	}

	function handleDeleteDashboard() {
		if (activeDashboard && confirm(`Delete dashboard "${activeDashboard.name}"?`)) {
			dispatch('deleteDashboard', { id: activeDashboard.id });
		}
	}

	function handleDuplicateDashboard() {
		if (activeDashboard) {
			dispatch('duplicateDashboard', { id: activeDashboard.id });
		}
	}

	function handleConfigureDashboard() {
		if (activeDashboard) {
			dispatch('configureDashboard', { id: activeDashboard.id });
		}
	}
</script>

<div class="flex items-center justify-between border-b border-gray-200 bg-white p-4">
	<!-- Left Side - Dashboard Info -->
	<div class="flex items-center gap-4">
		<div>
			<h1 class="flex items-center gap-2 text-xl font-semibold text-gray-900">
				{activeDashboard?.name || 'Dashboard'}
				{#if unsavedChanges}
					<Badge color="yellow" class="text-xs">Unsaved</Badge>
				{/if}
			</h1>
			{#if activeDashboard?.description}
				<p class="mt-1 text-sm text-gray-600">
					{activeDashboard.description}
				</p>
			{/if}
		</div>
	</div>

	<!-- Right Side - Actions -->
	<div class="flex items-center gap-2">
		<!-- Save Button -->
		{#if $isEditMode && unsavedChanges}
			<CompatButton on:click={saveDashboard} color="green" size="sm">
				<Save class="mr-2 h-4 w-4" />
				Save Changes
			</CompatButton>
		{/if}

		<!-- Add Widget Button -->
		{#if $isEditMode}
			<CompatButton on:click={handleAddWidget} color="blue" size="sm">
				<Plus class="mr-2 h-4 w-4" />
				Add Widget
			</CompatButton>
		{/if}

		<!-- Edit Mode Toggle -->
		<CompatButton on:click={toggleEditMode} color={$isEditMode ? 'red' : 'light'} size="sm">
			{#if $isEditMode}
				<Eye class="mr-2 h-4 w-4" />
				View Mode
			{:else}
				<Edit3 class="mr-2 h-4 w-4" />
				Edit Mode
			{/if}
		</CompatButton>

		<!-- Dashboard Menu -->
		<CompatButton color="light" size="sm" class="!p-2">
			<MoreVertical class="h-4 w-4" />
		</CompatButton>
		<Dropdown>
			<CompatDropdownItem on:click={handleCreateDashboard}>
				<Grid3x3 class="mr-2 h-4 w-4" />
				New Dashboard
			</CompatDropdownItem>
			<CompatDropdownItem on:click={handleDuplicateDashboard}>
				<Copy class="mr-2 h-4 w-4" />
				Duplicate Dashboard
			</CompatDropdownItem>
			<CompatDropdownItem on:click={handleConfigureDashboard}>
				<Settings class="mr-2 h-4 w-4" />
				Dashboard Settings
			</CompatDropdownItem>
			<CompatDropdownItem class="border-t" on:click={handleDeleteDashboard}>
				<Trash2 class="mr-2 h-4 w-4 text-red-600" />
				<span class="text-red-600">Delete Dashboard</span>
			</CompatDropdownItem>
		</Dropdown>
	</div>
</div>

<!-- Dashboard Tabs -->
{#if $dashboardStore.dashboards.length > 1}
	<div class="flex items-center gap-1 border-b bg-gray-50 px-4 py-2">
		{#each $dashboardStore.dashboards as dashboard}
			<button
				class="rounded-md px-3 py-1 text-sm transition-colors {dashboard.id === activeDashboardId
					? 'bg-blue-100 font-medium text-blue-700'
					: 'text-gray-600 hover:bg-gray-100'}"
				on:click={() => dashboardStore.setActiveDashboard(dashboard.id)}
			>
				{dashboard.name}
				{#if dashboard.widgets.length === 0}
					<span class="ml-1 text-xs opacity-60">(empty)</span>
				{/if}
			</button>
		{/each}
	</div>
{/if}
