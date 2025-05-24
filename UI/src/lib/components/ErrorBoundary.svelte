<script lang="ts">
	import { AlertTriangle, RefreshCw } from 'lucide-svelte';

	export let error: Error | null = null;
	export let showDetails = false;
	export let onRetry: (() => void) | null = null;

	function toggleDetails() {
		showDetails = !showDetails;
	}

	function handleRetry() {
		if (onRetry) {
			error = null;
			onRetry();
		}
	}
</script>

{#if error}
	<div class="min-h-64 flex items-center justify-center p-6">
		<div class="text-center max-w-md">
			<div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-red-100 dark:bg-red-900/20">
				<AlertTriangle class="h-6 w-6 text-red-600 dark:text-red-400" />
			</div>
			
			<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">
				Something went wrong
			</h3>
			
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
				We encountered an unexpected error. Please try again.
			</p>

			{#if showDetails}
				<div class="mt-4 p-3 bg-gray-50 dark:bg-gray-800 rounded-md text-left">
					<p class="text-xs font-mono text-gray-700 dark:text-gray-300 break-words">
						{error.message}
					</p>
					{#if error.stack}
						<details class="mt-2">
							<summary class="text-xs text-gray-500 dark:text-gray-400 cursor-pointer">
								Stack trace
							</summary>
							<pre class="mt-1 text-xs text-gray-600 dark:text-gray-400 whitespace-pre-wrap overflow-auto max-h-40">
								{error.stack}
							</pre>
						</details>
					{/if}
				</div>
			{/if}

			<div class="mt-6 flex flex-col sm:flex-row gap-3 justify-center">
				{#if onRetry}
					<button
						type="button"
						class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						on:click={handleRetry}
					>
						<RefreshCw class="h-4 w-4 mr-2" />
						Try Again
					</button>
				{/if}
				
				<button
					type="button"
					class="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
					on:click={toggleDetails}
				>
					{showDetails ? 'Hide' : 'Show'} Details
				</button>
			</div>
		</div>
	</div>
{:else}
	<slot />
{/if}
