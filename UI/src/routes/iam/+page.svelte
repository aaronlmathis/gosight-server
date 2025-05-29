<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		goto('/iam/users');
	});
</script>

<svelte:head>
	<title>Identity and Access Management</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center">
	<div class="text-center">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">
			Redirecting to Identity and Access Management...
		</h1>
		<p class="mt-2 text-gray-600 dark:text-gray-400">Taking you to the users management page</p>
	</div>
</div>
