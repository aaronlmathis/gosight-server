<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { user } from '$lib/stores';
	import { Eye, EyeOff, UserPlus } from 'lucide-svelte';

	let formData = {
		username: '',
		email: '',
		password: '',
		confirmPassword: '',
		firstName: '',
		lastName: ''
	};
	let showPassword = false;
	let showConfirmPassword = false;
	let loading = false;
	let error = '';
	let validationErrors: Record<string, string> = {};

	onMount(() => {
		// Redirect if already logged in
		user.subscribe(user => {
			if (user) {
				goto('/');
			}
		});
	});

	function validateForm() {
		validationErrors = {};

		if (!formData.username) {
			validationErrors.username = 'Username is required';
		} else if (formData.username.length < 3) {
			validationErrors.username = 'Username must be at least 3 characters';
		}

		if (!formData.email) {
			validationErrors.email = 'Email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
			validationErrors.email = 'Please enter a valid email address';
		}

		if (!formData.password) {
			validationErrors.password = 'Password is required';
		} else if (formData.password.length < 6) {
			validationErrors.password = 'Password must be at least 6 characters';
		}

		if (!formData.confirmPassword) {
			validationErrors.confirmPassword = 'Please confirm your password';
		} else if (formData.password !== formData.confirmPassword) {
			validationErrors.confirmPassword = 'Passwords do not match';
		}

		if (!formData.firstName) {
			validationErrors.firstName = 'First name is required';
		}

		if (!formData.lastName) {
			validationErrors.lastName = 'Last name is required';
		}

		return Object.keys(validationErrors).length === 0;
	}

	async function handleRegister() {
		if (!validateForm()) {
			return;
		}

		try {
			loading = true;
			error = '';

			const response = await api.register({
				username: formData.username,
				email: formData.email,
				password: formData.password,
				first_name: formData.firstName,
				last_name: formData.lastName
			});
			
			if (response.success) {
				// Redirect to login with success message
				goto('/auth/login?message=Registration successful. Please log in.');
			} else {
				error = response.message || 'Registration failed';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Registration failed. Please try again.';
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleRegister();
		}
	}
</script>

<svelte:head>
	<title>Register - GoSight</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
	<div class="max-w-md w-full space-y-8">
		<div>
			<div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-green-100 dark:bg-green-900">
				<UserPlus class="h-6 w-6 text-green-600 dark:text-green-400" />
			</div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				Create your account
			</h2>
			<p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
				Join GoSight to monitor your infrastructure
			</p>
		</div>

		<form class="mt-8 space-y-6" on:submit|preventDefault={handleRegister}>
			{#if error}
				<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
					<div class="text-sm text-red-800 dark:text-red-200">{error}</div>
				</div>
			{/if}

			<div class="space-y-4">
				<!-- Name Fields -->
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label for="firstName" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
							First Name
						</label>
						<input
							id="firstName"
							name="firstName"
							type="text"
							required
							bind:value={formData.firstName}
							on:keypress={handleKeyPress}
							class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.firstName ? 'border-red-500' : ''}"
							placeholder="First name"
						/>
						{#if validationErrors.firstName}
							<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.firstName}</p>
						{/if}
					</div>

					<div>
						<label for="lastName" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
							Last Name
						</label>
						<input
							id="lastName"
							name="lastName"
							type="text"
							required
							bind:value={formData.lastName}
							on:keypress={handleKeyPress}
							class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.lastName ? 'border-red-500' : ''}"
							placeholder="Last name"
						/>
						{#if validationErrors.lastName}
							<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.lastName}</p>
						{/if}
					</div>
				</div>

				<!-- Username -->
				<div>
					<label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Username
					</label>
					<input
						id="username"
						name="username"
						type="text"
						autocomplete="username"
						required
						bind:value={formData.username}
						on:keypress={handleKeyPress}
						class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.username ? 'border-red-500' : ''}"
						placeholder="Choose a username"
					/>
					{#if validationErrors.username}
						<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.username}</p>
					{/if}
				</div>

				<!-- Email -->
				<div>
					<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Email Address
					</label>
					<input
						id="email"
						name="email"
						type="email"
						autocomplete="email"
						required
						bind:value={formData.email}
						on:keypress={handleKeyPress}
						class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.email ? 'border-red-500' : ''}"
						placeholder="Enter your email"
					/>
					{#if validationErrors.email}
						<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.email}</p>
					{/if}
				</div>

				<!-- Password -->
				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Password
					</label>
					<div class="mt-1 relative">
						<input
							id="password"
							name="password"
							type={showPassword ? 'text' : 'password'}
							autocomplete="new-password"
							required
							bind:value={formData.password}
							on:keypress={handleKeyPress}
							class="appearance-none relative block w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.password ? 'border-red-500' : ''}"
							placeholder="Create a password"
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
					{#if validationErrors.password}
						<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.password}</p>
					{/if}
				</div>

				<!-- Confirm Password -->
				<div>
					<label for="confirmPassword" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
						Confirm Password
					</label>
					<div class="mt-1 relative">
						<input
							id="confirmPassword"
							name="confirmPassword"
							type={showConfirmPassword ? 'text' : 'password'}
							autocomplete="new-password"
							required
							bind:value={formData.confirmPassword}
							on:keypress={handleKeyPress}
							class="appearance-none relative block w-full px-3 py-2 pr-10 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white bg-white dark:bg-gray-800 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm {validationErrors.confirmPassword ? 'border-red-500' : ''}"
							placeholder="Confirm your password"
						/>
						<button
							type="button"
							class="absolute inset-y-0 right-0 pr-3 flex items-center"
							on:click={() => showConfirmPassword = !showConfirmPassword}
						>
							{#if showConfirmPassword}
								<EyeOff class="h-4 w-4 text-gray-400 hover:text-gray-500" />
							{:else}
								<Eye class="h-4 w-4 text-gray-400 hover:text-gray-500" />
							{/if}
						</button>
					</div>
					{#if validationErrors.confirmPassword}
						<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationErrors.confirmPassword}</p>
					{/if}
				</div>
			</div>

			<div>
				<button
					type="submit"
					disabled={loading}
					class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-green-700 dark:hover:bg-green-600"
				>
					{#if loading}
						<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
					{/if}
					Create Account
				</button>
			</div>

			<div class="text-center">
				<p class="text-sm text-gray-600 dark:text-gray-400">
					Already have an account?
					<a href="/auth/login" class="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400 dark:hover:text-blue-300">
						Sign in
					</a>
				</p>
			</div>
		</form>
	</div>
</div>
