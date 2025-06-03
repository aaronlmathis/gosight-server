<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api/api';
	import { auth } from '$lib/stores/authStore';

	let username = '';
	let password = '';
	let loading = false;
	let error = '';
	let providers: { name: string; display_name: string }[] = [];
	let next = '';
	let shakeError = false;

	onMount(async () => {
		// Get redirect parameter
		next = $page.url.searchParams.get('next') || '/';

		// Check for error parameter from OAuth callback
		const errorParam = $page.url.searchParams.get('error');
		if (errorParam) {
			switch (errorParam) {
				case 'invalid_provider':
					error = 'Invalid authentication provider selected';
					break;
				case 'auth_failed':
					error = 'Authentication failed. Please try again.';
					break;
				case 'user_load_failed':
					error = 'Failed to load user information. Please contact support.';
					break;
				default:
					error = 'Authentication error occurred';
			}
		}

		// Redirect if already logged in
		auth.subscribe((authState) => {
			if (authState.isAuthenticated && authState.user) {
				goto(next);
			}
		});

		// Load authentication providers
		try {
			const providerData = await api.auth.getProviders();
			providers = providerData.providers || [];
		} catch (err) {
			console.error('Failed to load providers:', err);
		}
	});

	async function handleLogin() {
		if (!username || !password) {
			error = 'Please enter both username and password';
			triggerShake();
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await api.login({ username, password });

			if (response.success) {
				// Set user in auth store
				auth.setUser(response.user);
				// Redirect to original page or dashboard
				goto(next);
			} else if (response.mfa_required) {
				// MFA is required - redirect to MFA page with next parameter
				goto(`/auth/mfa?next=${encodeURIComponent(next)}`);
			} else {
				error = response.message || 'Login failed';
				triggerShake();
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed. Please try again.';
			triggerShake();
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleLogin();
		}
	}

	function handleSSOLogin(provider: string) {
		window.location.href = `/api/v1/auth/login/start?provider=${provider}&next=${encodeURIComponent(next)}`;
	}

	function getProviderIcon(provider: string): string {
		const icons: Record<string, string> = {
			google: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/google.svg',
			github: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/github.svg',
			azure: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/microsoftazure.svg',
			aws: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/amazonaws.svg'
		};
		return icons[provider] || `https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/${provider}.svg`;
	}

	function capitalizeProvider(provider: string): string {
		if (provider === 'aws') return 'AWS';
		if (provider === 'azure') return 'Azure';
		return provider.charAt(0).toUpperCase() + provider.slice(1);
	}

	function triggerShake() {
		shakeError = true;
		setTimeout(() => {
			shakeError = false;
		}, 800);
	}
</script>

<svelte:head>
	<title>Login – GoSight</title>
</svelte:head>

<div class="mx-auto w-full max-w-md transition-all" class:animate-shake={shakeError}>
	<div
		class="w-full max-w-md rounded-lg border border-gray-200 bg-white p-6 shadow-md dark:border-gray-700 dark:bg-gray-800"
	>
		<div class="mb-6 text-center">
			<h1 class="text-3xl font-bold tracking-tight text-blue-500">GoSight</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">Sign in to your account</p>
		</div>

		{#if error}
			<div
				class="mb-4 rounded border border-red-300 bg-red-100 p-2 text-center text-sm text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
			>
				{error}
			</div>
		{/if}

		<form on:submit|preventDefault={handleLogin} class="space-y-4">
			<div>
				<label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
					Username
				</label>
				<input
					type="text"
					id="username"
					name="username"
					required
					bind:value={username}
					on:keypress={handleKeyPress}
					class="mt-1 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-900 dark:text-white"
				/>
			</div>
			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
					Password
				</label>
				<input
					type="password"
					id="password"
					name="password"
					required
					bind:value={password}
					on:keypress={handleKeyPress}
					class="mt-1 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-900 dark:text-white"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded-lg bg-blue-600 px-4 py-2.5 text-center text-sm font-medium text-white hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 disabled:opacity-50 dark:bg-blue-500 dark:hover:bg-blue-600 dark:focus:ring-blue-800"
			>
				{#if loading}
					<div
						class="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-b-2 border-white"
					></div>
				{/if}
				Sign in
			</button>
		</form>

		<div class="mt-6 text-center text-sm text-gray-400 dark:text-gray-500">or sign in with</div>

		<div class="mt-4 space-y-2">
			{#each providers as provider}
				{#if provider.name !== 'local'}
					<button
						type="button"
						on:click={() => handleSSOLogin(provider.name)}
						class="inline-flex w-full items-center justify-center rounded-md border border-gray-300 bg-gray-200 px-4 py-2 text-sm font-medium text-gray-900 shadow-sm hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800"
					>
						<img
							src={getProviderIcon(provider.name)}
							class="mr-2 h-5 w-5"
							alt={provider.display_name}
						/>
						Sign in with {provider.display_name}
					</button>
				{/if}
			{/each}
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
