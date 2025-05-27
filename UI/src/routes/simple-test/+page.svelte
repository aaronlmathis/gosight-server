<script lang="ts">
	import CompatButton from '$lib/components/CompatButton.svelte';
	import { Card } from 'flowbite-svelte';

	let clickCount = 0;

	function handleClick() {
		console.log('Button clicked!');
		clickCount++;
		alert(`Button clicked ${clickCount} times!`);
	}
</script>

<svelte:head>
	<title>Simple Button Test - GoSight</title>
</svelte:head>

<div class="p-8">
	<h1 class="mb-4 text-2xl font-bold">Simple Button Test</h1>

	<Card class="mb-4 p-4">
		<p class="mb-4">This is a simple test to verify buttons work correctly.</p>
		<p class="mb-4">Click count: {clickCount}</p>

		<div class="space-y-2">
			<CompatButton color="blue" on:click={handleClick}>CompatButton Test</CompatButton>

			<button
				type="button"
				class="rounded bg-green-500 px-4 py-2 text-white"
				on:click={handleClick}
			>
				Native Button Test
			</button>
		</div>
	</Card>
</div>
