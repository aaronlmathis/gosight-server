<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { user } from '$lib/stores';
	import { Eye, EyeOff, LogIn } from 'lucide-svelte';

	let username = '';
	let password = '';
	let rememberMe = false;
	let showPassword = false;
	let loading = false;
	let error = '';

	onMount(() => {
		// Redirect if already logged in
		user.subscribe(user => {
			if (user) {
				goto('/');
			}
		});

		// Check for saved credentials
		const savedUsername = localStorage.getItem('gosight_username');
		if (savedUsername) {
			username = savedUsername;
			rememberMe = true;
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
				// Save credentials if remember me is checked
				if (rememberMe) {
					localStorage.setItem('gosight_username', username);
				} else {
					localStorage.removeItem('gosight_username');
				}

				// Redirect to dashboard
				goto('/');
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
</script>

<svelte:head>
	<title>Login - GoSight</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md w-full space-y-8">
		<div>
			<div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
				<LogIn class="h-6 w-6 text-blue-600 dark:text-blue-400" />
			</div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				Sign in to GoSight
			</h2>
			<p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
				Monitor and manage your infrastructure
			</p>
		</div>

		<form class="mt-8 space-y-6" on:submit|preventDefault={handleLogin}>
			{#if error}
				<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
					<div class="text-sm text-red-800 dark:text-red-200">{error}</div>
				</div>
			{/if}

			<div class="space-y-4">
				<div>
					<label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Username
					</label>
					<div class="mt-1">
						<input
							id="username"
							name="username"
							type="text"
							autocomplete="username"
							required
							bind:value={username}
							on:keypress={handleKeyPress}
							class="appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
							placeholder="Enter your username"
						/>
					</div>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Password
					</label>
					<div class="mt-1 relative">
						<input
							id="password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="current-password"
							required
							bind:value={password}
							on:keypress={handleKeyPress}
							class="appearance-none relative block w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
							placeholder="Enter your password"
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center"
							on:click={() => showPassword = !showPassword}
						>
							{#if showPassword}
								<EyeOff class="h-4 w-4 text-gray-400 hover:text-gray-500" />
							{:else}
								<Eye class="h-4 w-4 text-gray-400 hover:text-gray-500" />
							{/if}
						</button>
					</div>
				</div>
			</div>

			<div class="flex items-center justify-between">
				<div class="flex items-center">
					<input
						id="remember-me"
						name="remember-me"
						type="checkbox"
						bind:checked={rememberMe}
						class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded dark:bg-gray-800"
					/>
					<label for="remember-me" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
						Remember me
					</label>
				</div>

				<div class="text-sm">
					<a href="/auth/forgot-password" class="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300">
						Forgot your password?
					</a>
				</div>
			</div>

			<div>
				<button
					type="submit"
					disabled={loading}
					class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-blue-700 dark:hover:bg-blue-600"
				>
					{#if loading}
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
					{/if}
					Sign in
				</button>
			</div>

			<div class="text-center">
				<p class="text-sm text-gray-600 dark:text-gray-400">
					Don't have an account?
					<a href="/auth/register" class="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300">
						Sign up
					</a>
				</p>
			</div>
		</form>
	</div>
</div>
