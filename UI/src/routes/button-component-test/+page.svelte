<script lang="ts">
	import { Button } from 'flowbite-svelte';
	import CompatButton from '$lib/components/CompatButton.svelte';

	let message = 'No clicks yet';
	let clickCount = 0;

	function testClick() {
		clickCount++;
		message = `Button clicked ${clickCount} times at ${new Date().toLocaleTimeString()}`;
		console.log('Test click function called, count:', clickCount);
	}
</script>

<svelte:head>
	<title>Button Component Test</title>
</svelte:head>

<div class="container mx-auto p-8">
	<h1 class="mb-6 text-2xl font-bold">Button Component Comparison Test</h1>

	<div class="space-y-4">
		<p class="text-lg">
			Current message: <span class="rounded bg-gray-100 px-2 py-1 font-mono">{message}</span>
		</p>

		<!-- Test 1: Native HTML button -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 1: Native HTML button</h2>
			<button
				class="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
				on:click={testClick}
			>
				Native HTML button
			</button>
		</div>

		<!-- Test 2: Direct Flowbite Button -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 2: Direct Flowbite Button</h2>
			<Button color="blue" on:click={testClick}>Direct Flowbite Button</Button>
		</div>

		<!-- Test 3: CompatButton component -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 3: CompatButton component</h2>
			<CompatButton color="blue" on:click={testClick}>CompatButton component</CompatButton>
		</div>

		<!-- Test 4: CompatButton with debugging -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 4: CompatButton with event debugging</h2>
			<CompatButton
				color="green"
				on:click={(e) => {
					console.log('CompatButton click event:', e);
					testClick();
				}}
				on:mousedown={() => console.log('CompatButton mousedown')}
				on:mouseup={() => console.log('CompatButton mouseup')}
			>
				CompatButton with debug
			</CompatButton>
		</div>

		<!-- Test 5: Flowbite Button with debugging -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 5: Direct Flowbite Button with debugging</h2>
			<Button
				color="purple"
				on:click={(e) => {
					console.log('Direct Flowbite Button click event:', e);
					testClick();
				}}
				on:mousedown={() => console.log('Direct Flowbite Button mousedown')}
				on:mouseup={() => console.log('Direct Flowbite Button mouseup')}
			>
				Direct Flowbite with debug
			</Button>
		</div>
	</div>
</div>
