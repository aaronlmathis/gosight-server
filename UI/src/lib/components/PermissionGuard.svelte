<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	export let requiredPermission: string;
	export let fallbackUrl: string = '/';
	export let showError: boolean = true;

	let hasPermission = false;
	let isLoading = true;

	onMount(() => {
		const unsubscribe = auth.subscribe((state) => {
			isLoading = state.isLoading;

			if (!state.isLoading) {
				if (!state.isAuthenticated) {
					// Redirect to login if not authenticated
					window.location.href = `/login?redirect=${encodeURIComponent($page.url.pathname)}`;
					return;
				}

				hasPermission = state.user?.permissions?.includes(requiredPermission) || false;

				if (!hasPermission && !showError) {
					// Redirect to fallback URL if no permission and not showing error
					goto(fallbackUrl);
				}
			}
		});

		return unsubscribe;
	});
</script>

{#if isLoading}
	<div class="flex min-h-screen items-center justify-center">
		<div class="h-32 w-32 animate-spin rounded-full border-b-2 border-blue-600"></div>
	</div>
{:else if hasPermission}
	<slot />
{:else if showError}
	<div class="flex min-h-screen items-center justify-center bg-gray-50">
		<div class="text-center">
			<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
				<svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.98-.833-2.75 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
					/>
				</svg>
			</div>
			<h3 class="mt-4 text-lg font-medium text-gray-900">Access Denied</h3>
			<p class="mt-2 text-sm text-gray-500">You don't have permission to access this page.</p>
			<div class="mt-6">
				<button
					on:click={() => goto(fallbackUrl)}
					class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
				>
					Go Back
				</button>
			</div>
		</div>
	</div>
{/if}
