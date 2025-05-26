<script lang="ts">
	import { onMount, createEventDispatcher, tick } from 'svelte';

	export let endpoint: any = null;

	const dispatch = createEventDispatcher();

	let command = '';
	let responses: Array<{ text: string; timestamp?: Date; type?: 'command' | 'output' | 'error' }> =
		[];
	let inputEl: HTMLInputElement;
	let outputEl: HTMLElement;

	onMount(() => {
		// Add welcome message
		addResponse(
			`Welcome to ${endpoint?.hostname || 'remote'} console. Type commands below:`,
			new Date(),
			'output'
		);
		// Focus the input
		if (inputEl) {
			inputEl.focus();
		}
	});

	function addResponse(
		text: string,
		timestamp: Date = new Date(),
		type: 'command' | 'output' | 'error' = 'output'
	) {
		responses = [...responses, { text, timestamp, type }];
		// Scroll to bottom after DOM update
		tick().then(() => {
			if (outputEl) {
				outputEl.scrollTop = outputEl.scrollHeight;
			}
		});
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			const cmd = command.trim();
			if (!cmd) return;

			// Add command to output
			addResponse(`$ ${cmd}`, new Date(), 'command');

			// Clear input
			command = '';

			// Add loading indicator
			addResponse('Command sent...', new Date(), 'output');

			// Emit event to parent component
			dispatch('run', cmd);
		}
	}

	// Public method to add command response from parent
	export function addCommandResponse(response: string, isError = false) {
		addResponse(response, new Date(), isError ? 'error' : 'output');
	}
</script>

<div
	class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
	id="console"
	role="tabpanel"
	aria-labelledby="console-tab"
>
	<div class="rounded-lg bg-black p-4 font-mono text-sm text-green-400 shadow-lg">
		<div class="mb-4">
			<div class="mb-2">
				<span class="text-blue-400">user</span>@<span class="text-purple-400"
					>{endpoint?.hostname || 'host'}</span
				>:<span class="text-red-400">~</span><span class="text-white">$</span><span
					class="blink-cursor"
				></span>
			</div>
		</div>
		<div bind:this={outputEl} class="mb-4 h-96 overflow-y-auto">
			<div class="space-y-1">
				{#each responses as response}
					<div
						class={response.type === 'command'
							? 'text-green-400'
							: response.type === 'error'
								? 'text-red-400'
								: 'text-gray-400'}
					>
						{response.text}
					</div>
				{/each}
			</div>
		</div>
		<div class="flex items-center">
			<span class="text-blue-400">user</span>@<span class="text-purple-400"
				>{endpoint?.hostname || 'host'}</span
			>:<span class="text-red-400">~</span><span class="text-white">$ </span>
			<input
				bind:this={inputEl}
				bind:value={command}
				on:keydown={handleKeyDown}
				class="ml-1 flex-1 border-none bg-transparent text-green-400 outline-none"
				type="text"
				placeholder="Enter command..."
				autocomplete="off"
				id="console-command"
			/>
		</div>
	</div>
</div>

<style>
	.blink-cursor {
		animation: blink 1s infinite;
	}

	@keyframes blink {
		0%,
		50% {
			opacity: 1;
		}
		51%,
		100% {
			opacity: 0;
		}
	}
</style>
