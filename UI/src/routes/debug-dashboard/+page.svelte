<script lang="ts">
	// This is a copy of the DashboardManager but without the Modal wrapper
	// to test if the Modal is causing the button issues

	import { createEventDispatcher, onMount } from 'svelte';
	import { Card, Input, Label, Select, Textarea, Badge, Spinner } from 'flowbite-svelte';
	import CompatButton from '$lib/components/CompatButton.svelte';
	import {
		Plus,
		Download,
		Upload,
		Copy,
		Trash2,
		Edit3,
		Save,
		FileText,
		Folder
	} from 'lucide-svelte';

	let activeTab: 'manage' | 'templates' | 'import-export' = 'manage';

	function openCreateModal() {
		console.log('openCreateModal called from debug page');
		alert('Create Modal function called from debug page!');
	}

	function testClick() {
		console.log('Test click function called from debug page');
		alert('Test click works from debug page!');
	}

	onMount(() => {
		console.log('Debug Dashboard Manager mounted!');
	});
</script>

<div class="p-8">
	<h1 class="mb-6 text-2xl font-bold">Debug Dashboard Manager (No Modal)</h1>

	<!-- Test buttons directly without Modal wrapper -->
	<div class="space-y-4">
		<div class="flex gap-2">
			<!-- Debug test button -->
			<button
				type="button"
				class="rounded bg-red-500 px-4 py-2 text-white"
				on:click={() => alert('Simple button works!')}
			>
				Test Button
			</button>

			<!-- Another test with named function -->
			<button type="button" class="rounded bg-green-500 px-4 py-2 text-white" on:click={testClick}>
				Named Function Test
			</button>

			<!-- CompatButton test -->
			<CompatButton color="blue" size="sm" on:click={openCreateModal}>
				<Plus class="mr-2 h-4 w-4" />
				Create Dashboard
			</CompatButton>
		</div>

		<!-- Tab navigation test -->
		<nav class="-mb-px flex space-x-8">
			<button
				class="border-b-2 px-1 py-2 text-sm font-medium {activeTab === 'manage'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
				on:click={() => {
					console.log('Switching to manage tab');
					activeTab = 'manage';
					alert('Switched to manage tab');
				}}
			>
				<Folder class="mr-2 inline h-4 w-4" />
				Manage Dashboards
			</button>
			<button
				class="border-b-2 px-1 py-2 text-sm font-medium {activeTab === 'templates'
					? 'border-blue-500 text-blue-600'
					: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
				on:click={() => {
					console.log('Switching to templates tab');
					activeTab = 'templates';
					alert('Switched to templates tab');
				}}
			>
				<FileText class="mr-2 inline h-4 w-4" />
				Templates
			</button>
		</nav>
	</div>
</div>
