<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { user } from '$lib/stores';

	let username = '';
	let password = '';
	let loading = false;
	let error = '';
	let providers: string[] = [];
	let next = '';

	onMount(async () => {
		// Get redirect parameter
		next = $page.url.searchParams.get('next') || '/';

		// Redirect if already logged in
		user.subscribe((user) => {
			if (user) {
				goto(next);
			}
		});

		// Load authentication providers
		try {
			const providerData = await api.getProviders();
			providers = providerData.providers || [];
		} catch (err) {
			console.error('Failed to load providers:', err);
		}
	});

	async function handleLogin() {
		if (!username || !password) {
			error = 'Please enter both username and password';
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await api.login({ username, password });

			if (response.success) {
				// Set user in store
				user.set(response.user);
				// Redirect to original page or dashboard
				goto(next);
			} else if (response.mfa_required) {
				// MFA is required - redirect to MFA page with next parameter
				goto(`/auth/mfa?next=${encodeURIComponent(next)}`);
			} else {
				error = response.message || 'Login failed';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Login failed. Please try again.';
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
		window.location.href = `/login/start?provider=${provider}&next=${encodeURIComponent(next)}`;
	}

	function getProviderIcon(provider: string): string {
		const icons: Record<string, string> = {
			google: 'https://simpleicons.org/icons/google.svg',
			github: 'https://simpleicons.org/icons/github.svg',
			azure: 'https://simpleicons.org/icons/microsoftazure.svg',
			aws: 'https://simpleicons.org/icons/amazonaws.svg'
		};
		return icons[provider] || `https://simpleicons.org/icons/${provider}.svg`;
	}

	function capitalizeProvider(provider: string): string {
		if (provider === 'aws') return 'AWS';
		if (provider === 'azure') return 'Azure';
		return provider.charAt(0).toUpperCase() + provider.slice(1);
	}
</script>

<svelte:head>
	<title>Login – GoSight</title>
</svelte:head>

<div class="mx-auto w-full max-w-md transition-all">
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
				{#if provider !== 'local'}
					<button
						type="button"
						on:click={() => handleSSOLogin(provider)}
						class="inline-flex w-full items-center justify-center rounded-md border border-gray-300 bg-gray-200 px-4 py-2 text-sm font-medium text-gray-900 shadow-sm hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:hover:bg-gray-800"
					>
						<img
							src={getProviderIcon(provider)}
							class="mr-2 h-5 w-5"
							alt={capitalizeProvider(provider)}
						/>
						Sign in with {capitalizeProvider(provider)}
					</button>
				{/if}
			{/each}
		</div>
	</div>
</div>
