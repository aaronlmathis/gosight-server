<script lang="ts">
	import { DropdownItem } from 'flowbite-svelte';
	import { createEventDispatcher } from 'svelte';

	// Re-export all DropdownItem props
	export let href: string | undefined = undefined;
	export let disabled: boolean = false;
	export let divider: boolean = false;
	export let defaultClass: string | undefined = undefined;
	export let class: string | undefined = undefined;

	const dispatch = createEventDispatcher<{
		click: MouseEvent;
	}>();

	function handleClick(event: MouseEvent) {
		dispatch('click', event);
	}
</script>

<DropdownItem
	{href}
	{disabled}
	{divider}
	{defaultClass}
	class={class}
	on:click={handleClick}
	{...$$restProps}
>
	<slot />
</DropdownItem>
