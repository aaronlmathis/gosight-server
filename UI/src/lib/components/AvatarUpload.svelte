<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Camera, Upload, X, RotateCcw, Check } from 'lucide-svelte';
	import { api } from '$lib/api';
	import Modal from './Modal.svelte';

	export let currentAvatar = '';
	export let size: 'sm' | 'md' | 'lg' = 'md';

	const dispatch = createEventDispatcher<{
		uploaded: { avatar_url: string };
		deleted: void;
	}>();

	let showModal = false;
	let selectedFile: File | null = null;
	let previewUrl = '';
	let uploading = false;
	let uploadError = '';
	let draggedOver = false;
	let cropData = { x: 0, y: 0, width: 0, height: 0 };
	let showCropper = false;
	let canvasElement: HTMLCanvasElement;
	let imageElement: HTMLImageElement;
	let uploadLimits: any = null;

	const sizeClasses = {
		sm: 'w-16 h-16',
		md: 'w-24 h-24',
		lg: 'w-32 h-32'
	};

	// Load upload limits on component mount
	$: if (showModal && !uploadLimits) {
		loadUploadLimits();
	}

	async function loadUploadLimits() {
		try {
			uploadLimits = await api.getUploadLimits();
		} catch (err) {
			console.error('Failed to load upload limits:', err);
		}
	}

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			processFile(target.files[0]);
		}
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		draggedOver = false;

		if (event.dataTransfer?.files && event.dataTransfer.files[0]) {
			processFile(event.dataTransfer.files[0]);
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		draggedOver = true;
	}

	function handleDragLeave() {
		draggedOver = false;
	}

	function processFile(file: File) {
		uploadError = '';

		// Validate file type
		if (!file.type.startsWith('image/')) {
			uploadError = 'Please select an image file';
			return;
		}

		// Validate file size
		if (uploadLimits && file.size > uploadLimits.max_file_size) {
			uploadError = `File size must be less than ${uploadLimits.max_file_size_mb}MB`;
			return;
		}

		selectedFile = file;

		// Create preview URL
		if (previewUrl) {
			URL.revokeObjectURL(previewUrl);
		}
		previewUrl = URL.createObjectURL(file);

		// Show cropper for image adjustment
		setTimeout(() => {
			showCropperInterface();
		}, 100);
	}

	function showCropperInterface() {
		if (!imageElement || !canvasElement) return;

		// Set up basic crop area (center square)
		const img = imageElement;
		const size = Math.min(img.naturalWidth, img.naturalHeight);
		const x = (img.naturalWidth - size) / 2;
		const y = (img.naturalHeight - size) / 2;

		cropData = { x, y, width: size, height: size };
		showCropper = true;
	}

	async function uploadAvatar() {
		if (!selectedFile) return;

		try {
			uploading = true;
			uploadError = '';

			let result;
			if (showCropper && cropData.width > 0) {
				// Use crop endpoint
				result = await api.cropAvatar(selectedFile, cropData);
			} else {
				// Use simple upload endpoint
				result = await api.uploadAvatar(selectedFile);
			}

			if (result && result.success && result.avatar_url) {
				dispatch('uploaded', { avatar_url: result.avatar_url });
				closeModal();
			} else {
				uploadError = result?.message || 'Upload failed';
			}
		} catch (err) {
			uploadError = err instanceof Error ? err.message : 'Upload failed';
		} finally {
			uploading = false;
		}
	}

	async function deleteAvatar() {
		try {
			uploading = true;
			const result = await api.deleteAvatar();

			if (result && result.success) {
				dispatch('deleted');
				closeModal();
			} else {
				uploadError = result?.message || 'Delete failed';
			}
		} catch (err) {
			uploadError = err instanceof Error ? err.message : 'Delete failed';
		} finally {
			uploading = false;
		}
	}

	function closeModal() {
		showModal = false;
		selectedFile = null;
		if (previewUrl) {
			URL.revokeObjectURL(previewUrl);
			previewUrl = '';
		}
		uploadError = '';
		showCropper = false;
		cropData = { x: 0, y: 0, width: 0, height: 0 };
	}

	function resetSelection() {
		selectedFile = null;
		if (previewUrl) {
			URL.revokeObjectURL(previewUrl);
			previewUrl = '';
		}
		showCropper = false;
		uploadError = '';
	}
</script>

<!-- Avatar Display -->
<div class="relative">
	<div
		class="relative {sizeClasses[size]} overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700"
	>
		{#if currentAvatar}
			<img src={currentAvatar} alt="Profile picture" class="h-full w-full object-cover" />
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<Camera class="h-8 w-8 text-gray-400" />
			</div>
		{/if}

		<!-- Overlay button -->
		<button
			on:click={() => (showModal = true)}
			class="bg-opacity-50 absolute inset-0 flex items-center justify-center bg-black opacity-0 transition-opacity duration-200 hover:opacity-100"
		>
			<Camera class="h-6 w-6 text-white" />
		</button>
	</div>
</div>

<!-- Upload Modal -->
<Modal bind:show={showModal} title="Update Profile Picture" size="lg" on:close={closeModal}>
	<div class="space-y-6">
		{#if uploadError}
			<div
				class="rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
			>
				<div class="text-sm text-red-800 dark:text-red-200">{uploadError}</div>
			</div>
		{/if}

		{#if !selectedFile}
			<!-- File Selection Area -->
			<div
				class="rounded-lg border-2 border-dashed p-8 text-center transition-colors {draggedOver
					? 'border-blue-400 bg-blue-50 dark:border-blue-600 dark:bg-blue-900/20'
					: 'border-gray-300 hover:border-gray-400 dark:border-gray-700 dark:hover:border-gray-600'}"
				on:drop={handleDrop}
				on:dragover={handleDragOver}
				on:dragleave={handleDragLeave}
				role="button"
				tabindex="0"
				on:keydown={() => {}}
			>
				<Upload class="mx-auto h-12 w-12 text-gray-400" />
				<div class="mt-4">
					<p class="text-lg font-medium text-gray-900 dark:text-white">
						Drop your image here, or
						<label class="cursor-pointer text-blue-600 hover:text-blue-500 dark:text-blue-400">
							browse
							<input type="file" accept="image/*" class="hidden" on:change={handleFileSelect} />
						</label>
					</p>
					<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
						{#if uploadLimits}
							Maximum size: {uploadLimits.max_file_size_mb}MB
						{/if}
						Supported formats: JPEG, PNG, GIF
					</p>
				</div>
			</div>
		{:else}
			<!-- Image Preview and Cropper -->
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<h4 class="text-lg font-medium text-gray-900 dark:text-white">Preview</h4>
					<button
						on:click={resetSelection}
						class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
					>
						<RotateCcw class="h-5 w-5" />
					</button>
				</div>

				<div class="flex gap-6">
					<!-- Original Image -->
					<div class="flex-1">
						<p class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">Original</p>
						<div class="overflow-hidden rounded-lg border bg-gray-50 dark:bg-gray-800">
							<img
								bind:this={imageElement}
								src={previewUrl}
								alt="Preview"
								class="h-48 max-w-full object-contain"
							/>
						</div>
					</div>

					<!-- Preview -->
					<div class="w-32">
						<p class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">Preview</p>
						<div
							class="h-24 w-24 overflow-hidden rounded-full border-2 border-gray-300 bg-gray-200 dark:border-gray-600 dark:bg-gray-700"
						>
							<img src={previewUrl} alt="Avatar preview" class="h-full w-full object-cover" />
						</div>
					</div>
				</div>

				{#if showCropper}
					<div class="text-sm text-gray-600 dark:text-gray-400">
						<p>
							The image will be automatically cropped to a square and resized for optimal display.
						</p>
					</div>
				{/if}
			</div>
		{/if}

		{#if currentAvatar}
			<!-- Current Avatar Actions -->
			<div class="border-t pt-4 dark:border-gray-700">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-700 dark:text-gray-300">
							Current Profile Picture
						</p>
						<p class="text-xs text-gray-500 dark:text-gray-400">
							Remove your current profile picture
						</p>
					</div>
					<button
						on:click={deleteAvatar}
						disabled={uploading}
						class="px-3 py-2 text-sm font-medium text-red-600 hover:text-red-700 disabled:opacity-50 dark:text-red-400 dark:hover:text-red-300"
					>
						Remove
					</button>
				</div>
			</div>
		{/if}

		<canvas bind:this={canvasElement} class="hidden"></canvas>
	</div>

	<svelte:fragment slot="footer">
		<div
			class="flex justify-end space-x-3 border-t border-gray-200 bg-gray-50 px-6 py-4 dark:border-gray-700 dark:bg-gray-900/50"
		>
			<button
				type="button"
				class="inline-flex justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
				on:click={closeModal}
				disabled={uploading}
			>
				Cancel
			</button>
			{#if selectedFile}
				<button
					type="button"
					on:click={uploadAvatar}
					disabled={uploading}
					class="inline-flex items-center justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:outline-none disabled:opacity-50"
				>
					{#if uploading}
						<div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
					{:else}
						<Check class="mr-2 h-4 w-4" />
					{/if}
					{uploading ? 'Uploading...' : 'Upload'}
				</button>
			{/if}
		</div>
	</svelte:fragment>
</Modal>
