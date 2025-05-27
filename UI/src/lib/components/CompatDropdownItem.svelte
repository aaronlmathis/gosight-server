<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	// Basic props for dropdown item
	export let href: string | undefined = undefined;
	export let disabled: boolean = false;

	// Handle the 'class' prop using a different name
	let className: string | undefined = undefined;
	export { className as class };

	const dispatch = createEventDispatcher<{
		click: MouseEvent;
	}>();

	function handleClick(event: MouseEvent) {
		if (!disabled) {
			dispatch('click', event);
		}
	}
</script>

{#if href}
	<a
		{href}
		class="block px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white {className}"
		class:opacity-50={disabled}
		class:pointer-events-none={disabled}
		on:click={handleClick}
		{...$$restProps}
	>
		<slot />
	</a>
{:else}
	<button
		type="button"
		class="block w-full px-4 py-2 text-left hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white {className}"
		class:opacity-50={disabled}
		{disabled}
		on:click={handleClick}
		{...$$restProps}
	>
		<slot />
	</button>
{/if}
