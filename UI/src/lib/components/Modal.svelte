<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { X } from 'lucide-svelte';

	export let show = false;
	export let title = '';
	export let size: 'sm' | 'md' | 'lg' | 'xl' | '2xl' = 'md';
	export let closeOnBackdrop = true;
	export let showCloseButton = true;

	const dispatch = createEventDispatcher();

	let modalElement: HTMLElement;

	const sizeClasses = {
		sm: 'max-w-md',
		md: 'max-w-lg',
		lg: 'max-w-2xl',
		xl: 'max-w-4xl',
		'2xl': 'max-w-6xl'
	};

	onMount(() => {
		const handleEscape = (event: KeyboardEvent) => {
			if (event.key === 'Escape' && show) {
				close();
			}
		};

		document.addEventListener('keydown', handleEscape);

		return () => {
			document.removeEventListener('keydown', handleEscape);
		};
	});

	function close() {
		show = false;
		dispatch('close');
	}

	function handleBackdropClick(event: MouseEvent) {
		if (closeOnBackdrop && event.target === event.currentTarget) {
			close();
		}
	}

	// Prevent body scroll when modal is open
	$: if (typeof document !== 'undefined') {
		if (show) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}
	}
</script>

{#if show}
	<!-- Backdrop -->
	<div
		role="presentation"
		class="bg-opacity-75 dark:bg-opacity-75 fixed inset-0 z-40 bg-gray-500 transition-opacity dark:bg-gray-900"
		on:click={handleBackdropClick}
		on:keydown={() => {}}
	></div>

	<!-- Modal -->
	<div
		role="dialog"
		aria-modal="true"
		class="fixed inset-0 z-50 overflow-y-auto"
		on:click={handleBackdropClick}
		on:keydown={() => {}}
	>
		<div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
			<div
				bind:this={modalElement}
				class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full dark:bg-gray-800 {sizeClasses[
					size
				]}"
			>
				{#if title || showCloseButton}
					<div
						class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-700"
					>
						{#if title}
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">
								{title}
							</h3>
						{/if}
						{#if showCloseButton}
							<button
								type="button"
								class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:ring-2 focus:ring-blue-500 focus:outline-none dark:bg-gray-800 dark:hover:text-gray-300"
								on:click={close}
							>
								<span class="sr-only">Close</span>
								<X class="h-6 w-6" />
							</button>
						{/if}
					</div>
				{/if}

				<div class="px-6 py-4">
					<slot />
				</div>

				<slot name="footer">
					<div
						class="flex justify-end space-x-3 border-t border-gray-200 bg-gray-50 px-6 py-4 dark:border-gray-700 dark:bg-gray-900/50"
					>
						<button
							type="button"
							class="inline-flex justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
							on:click={close}
						>
							Cancel
						</button>
					</div>
				</slot>
			</div>
		</div>
	</div>
{/if}
