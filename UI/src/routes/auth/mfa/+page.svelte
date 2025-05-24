<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { user } from '$lib/stores';
	import { Shield, ArrowLeft } from 'lucide-svelte';

	let code = '';
	let remember = false;
	let loading = false;
	let error = '';
	let next = '';

	onMount(() => {
		// Get redirect parameter
		next = $page.url.searchParams.get('next') || '/';

		// Redirect if already logged in
		user.subscribe((user) => {
			if (user) {
				goto(next);
			}
		});
	});

	async function handleMFAVerify() {
		if (!code) {
			error = 'Please enter the 6-digit verification code';
			return;
		}

		if (code.length !== 6 || !/^\d+$/.test(code)) {
			error = 'Please enter a valid 6-digit code';
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await api.auth.verifyMFA({ code, remember });

			if (response.success) {
				// Set user in store
				user.set(response.user);

				// Redirect to original page or dashboard
				goto(next);
			} else {
				error = response.message || 'MFA verification failed';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'MFA verification failed. Please try again.';
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleMFAVerify();
		}
	}

	function handleBackToLogin() {
		goto(`/auth/login?next=${encodeURIComponent(next)}`);
	}

	// Auto-format code input
	function formatCode(event: Event) {
		const target = event.target as HTMLInputElement;
		let value = target.value.replace(/\D/g, '');
		if (value.length > 6) {
			value = value.slice(0, 6);
		}
		code = value;
	}
</script>

<svelte:head>
	<title>Two-Factor Authentication - GoSight</title>
</svelte:head>

<div
	class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8 dark:bg-gray-900"
>
	<div class="w-full max-w-md space-y-8">
		<div>
			<div
				class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900"
			>
				<Shield class="h-6 w-6 text-blue-600 dark:text-blue-400" />
			</div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				Two-Factor Authentication
			</h2>
			<p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
				Enter the 6-digit code from your authenticator app
			</p>
		</div>

		<div class="mt-8 space-y-6">
			{#if error}
				<div class="rounded-md bg-red-50 p-4 dark:bg-red-900/50">
					<div class="text-sm text-red-800 dark:text-red-200">
						{error}
					</div>
				</div>
			{/if}

			<div class="space-y-6">
				<div>
					<label for="code" class="sr-only">Verification Code</label>
					<input
						id="code"
						name="code"
						type="text"
						inputmode="numeric"
						pattern="[0-9]*"
						maxlength="6"
						required
						class="relative block w-full appearance-none rounded-md border border-gray-300 bg-white px-3 py-2 text-center text-2xl tracking-widest text-gray-900 placeholder-gray-500 focus:z-10 focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:placeholder-gray-400"
						placeholder="000000"
						bind:value={code}
						on:input={formatCode}
						on:keypress={handleKeyPress}
						disabled={loading}
					/>
				</div>

				<div class="flex items-center">
					<input
						id="remember"
						name="remember"
						type="checkbox"
						class="h-4 w-4 rounded border-gray-300 bg-white text-blue-600 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800"
						bind:checked={remember}
						disabled={loading}
					/>
					<label for="remember" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
						Remember this device for 30 days
					</label>
				</div>

				<div class="space-y-3">
					<button
						type="button"
						class="group relative flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:bg-blue-600 dark:hover:bg-blue-700"
						on:click={handleMFAVerify}
						disabled={loading || !code}
					>
						{#if loading}
							<svg
								class="mr-3 -ml-1 h-5 w-5 animate-spin text-white"
								xmlns="http://www.w3.org/2000/svg"
								fill="none"
								viewBox="0 0 24 24"
							>
								<circle
									class="opacity-25"
									cx="12"
									cy="12"
									r="10"
									stroke="currentColor"
									stroke-width="4"
								></circle>
								<path
									class="opacity-75"
									fill="currentColor"
									d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
								></path>
							</svg>
							Verifying...
						{:else}
							Verify Code
						{/if}
					</button>

					<button
						type="button"
						class="group relative flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={handleBackToLogin}
						disabled={loading}
					>
						<ArrowLeft class="mr-2 h-4 w-4" />
						Back to Login
					</button>
				</div>
			</div>
		</div>

		<div class="text-center">
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Having trouble? Contact your administrator for assistance.
			</p>
		</div>
	</div>
</div>
