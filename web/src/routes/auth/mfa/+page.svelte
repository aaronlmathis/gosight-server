<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { Shield, ArrowLeft } from 'lucide-svelte';

	let code = '';
	let remember = false;
	let loading = false;
	let error = '';
	let next = '';
	let shakeError = false;

	onMount(() => {
		// Get redirect parameter
		next = $page.url.searchParams.get('next') || '/';

		// Redirect if already logged in
		auth.subscribe((authState) => {
			if (authState.isAuthenticated && authState.user) {
				goto(next);
			}
		});
	});

	async function handleMFAVerify() {
		if (!code) {
			error = 'Please enter the 6-digit verification code';
			triggerShake();
			return;
		}

		if (code.length !== 6 || !/^\d+$/.test(code)) {
			error = 'Please enter a valid 6-digit code';
			triggerShake();
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await api.auth.verifyMFA({ code, remember });

			if (response.success) {
				// Set user in auth store
				auth.setUser(response.user);

				// Redirect to original page or dashboard
				window.location.href = next;
			} else {
				error = response.message || 'MFA verification failed';
				triggerShake();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'MFA verification failed. Please try again.';
			triggerShake();
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

	function triggerShake() {
		shakeError = true;
		setTimeout(() => {
			shakeError = false;
		}, 800);
	}
</script>

<svelte:head>
	<title>Two-Factor Authentication - GoSight</title>
</svelte:head>

<div class="mx-auto w-full max-w-md transition-all" class:animate-shake={shakeError}>
	<div
		class="w-full max-w-md rounded-lg border border-gray-200 bg-white p-6 shadow-md dark:border-gray-700 dark:bg-gray-800"
	>
		<div class="mb-6 text-center">
			<div
				class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900"
			>
				<Shield class="h-6 w-6 text-blue-600 dark:text-blue-400" />
			</div>
			<h2 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">
				Two-Factor Authentication
			</h2>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				Enter the 6-digit code from your authenticator app
			</p>
		</div>

		{#if error}
			<div
				class="mb-4 rounded border border-red-300 bg-red-100 p-2 text-center text-sm text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
			>
				{error}
			</div>
		{/if}

		<form on:submit|preventDefault={handleMFAVerify} class="space-y-4">
			<div>
				<label for="code" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
					Verification Code
				</label>
				<input
					id="code"
					name="code"
					type="text"
					inputmode="numeric"
					pattern="[0-9]*"
					maxlength="6"
					required
					class="mt-1 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-center text-lg tracking-widest text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-900 dark:text-white"
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
				<label for="remember" class="ml-2 block text-sm text-gray-700 dark:text-gray-300">
					Remember this device for 30 days
				</label>
			</div>

			<button
				type="submit"
				disabled={loading || !code}
				class="w-full rounded-lg bg-blue-600 px-4 py-2.5 text-center text-sm font-medium text-white hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-blue-500 dark:hover:bg-blue-600 dark:focus:ring-blue-800"
			>
				{#if loading}
					<div
						class="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-b-2 border-white"
					></div>
				{/if}
				{loading ? 'Verifying...' : 'Verify Code'}
			</button>

			<button
				type="button"
				class="w-full rounded-lg border border-gray-300 bg-gray-100 px-4 py-2.5 text-center text-sm font-medium text-gray-700 hover:bg-gray-200 focus:ring-4 focus:ring-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
				on:click={handleBackToLogin}
				disabled={loading}
			>
				<ArrowLeft class="mr-2 inline h-4 w-4" />
				Back to Login
			</button>
		</form>

		<div class="mt-6 text-center">
			<p class="text-xs text-gray-500 dark:text-gray-400">
				Having trouble? Contact your administrator for assistance.
			</p>
		</div>
	</div>
</div>

<style>
	@keyframes shake {
		10%,
		90% {
			transform: translateX(-1px);
		}
		20%,
		80% {
			transform: translateX(2px);
		}
		30%,
		50%,
		70% {
			transform: translateX(-4px);
		}
		40%,
		60% {
			transform: translateX(4px);
		}
	}

	:global(.animate-shake) {
		animation: shake 0.8s ease-in-out;
	}
</style>
