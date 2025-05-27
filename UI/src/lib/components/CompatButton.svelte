<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import { createEventDispatcher } from 'svelte';

	// Re-export all Button props
	export let color: string = 'primary';
	export let size: string = 'md';
	export let type: string = 'button';
	export let disabled: boolean = false;
	export let pill: boolean = false;
	export let outline: boolean = false;
	export let gradient: boolean = false;

	// Additional props that might be used
	export let href: string | undefined = undefined;
	export let role: string | undefined = undefined;
	export let id: string | undefined = undefined;

	const dispatch = createEventDispatcher<{
		click: MouseEvent;
	}>();

	function handleClick(event: MouseEvent) {
		dispatch('click', event);
	}
</script>

<Button
	{color}
	{size}
	{type}
	{disabled}
	{pill}
	{outline}
	{gradient}
	{href}
	{role}
	{id}
	on:click={handleClick}
	{...$$restProps}
>
	<slot />
</Button>
