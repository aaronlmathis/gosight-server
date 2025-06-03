<!--
 * Auth Layout
 * This layout provides the centered auth container with dark background
 * matching the original design from web/templates/layouts/layout_auth.html
 * This layout completely replaces the main layout for auth pages
-->
<script lang="ts">
	import '../../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/authStore';

	onMount(async () => {
		// Initialize auth store
		await auth.init();

		// Set dark theme for auth pages
		document.documentElement.classList.add('dark');
		document.body.style.backgroundColor = '#111827';
	});
</script>

<svelte:head>
	<title>Authentication – GoSight</title>
</svelte:head>

<!-- Full screen auth layout that replaces the main layout -->
<div class="flex min-h-screen items-center justify-center bg-gray-900 px-4">
	<slot />
</div>

<style>
	:global(html) {
		background-color: #111827 !important;
	}
	:global(body) {
		background-color: #111827 !important;
		overflow: hidden !important;
	}

	/* Hide any main layout elements that might leak through */
	:global(nav),
	:global(aside),
	:global(.sidebar),
	:global(.navbar) {
		display: none !important;
	}
</style>
