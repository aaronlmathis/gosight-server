<script lang="ts">
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let color: 'blue' | 'gray' | 'white' = 'blue';
	export let text = '';
	export let fullPage = false;

	const sizeClasses = {
		sm: 'h-4 w-4',
		md: 'h-8 w-8',
		lg: 'h-12 w-12'
	};

	const colorClasses = {
		blue: 'border-blue-600',
		gray: 'border-gray-600',
		white: 'border-white'
	};
</script>

{#if fullPage}
	<div class="fixed inset-0 bg-white dark:bg-gray-900 bg-opacity-75 dark:bg-opacity-75 flex items-center justify-center z-50">
		<div class="text-center">
			<div class="animate-spin rounded-full {sizeClasses[size]} border-b-2 {colorClasses[color]} mx-auto"></div>
			{#if text}
				<p class="mt-4 text-sm text-gray-600 dark:text-gray-400">{text}</p>
			{/if}
		</div>
	</div>
{:else}
	<div class="flex items-center justify-center {text ? 'flex-col' : ''}">
		<div class="animate-spin rounded-full {sizeClasses[size]} border-b-2 {colorClasses[color]}"></div>
		{#if text}
			<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{text}</p>
		{/if}
	</div>
{/if}
