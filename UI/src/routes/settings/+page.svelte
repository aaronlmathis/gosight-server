<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { api } from '$lib/api';
	import { formatDate } from '$lib/utils';
	import { User, Settings, Save, Eye, EyeOff, AlertCircle } from 'lucide-svelte';
	import type { User as UserType } from '$lib/types';

	let currentUser: UserType | null = null;
	let activeTab = 'profile';
	let loading = false;
	let saving = false;
	let message = '';
	let error = '';

	// Profile form data
	let profileData = {
		first_name: '',
		last_name: '',
		email: '',
		username: ''
	};

	// Password form data
	let passwordData = {
		current_password: '',
		new_password: '',
		confirm_password: ''
	};

	let showCurrentPassword = false;
	let showNewPassword = false;
	let showConfirmPassword = false;

	// Preferences
	let preferences = {
		theme: 'light',
		notifications: {
			email_alerts: true,
			push_alerts: true,
			alert_frequency: 'immediate'
		},
		dashboard: {
			refresh_interval: 30,
			default_time_range: '1h',
			show_system_metrics: true
		}
	};

	onMount(async () => {
		if ($auth.user) {
			currentUser = $auth.user;
			profileData = {
				first_name: $auth.user.first_name || '',
				last_name: $auth.user.last_name || '',
				email: $auth.user.email || '',
				username: $auth.user.username || ''
			};
		}

		await loadUserPreferences();
	});

	async function loadUserPreferences() {
		try {
			const response = await api.getUserPreferences();
			if (response && typeof response === 'object' && 'data' in response && response.data) {
				preferences = { ...preferences, ...response.data };
			}
		} catch (err) {
			console.error('Failed to load user preferences:', err);
		}
	}

	async function updateProfile() {
		try {
			saving = true;
			error = '';
			message = '';
			const response = await api.updateProfile(profileData);
			if (response && typeof response === 'object' && 'success' in response && response.success) {
				message = 'Profile updated successfully';
				// Update auth store - ensure we have all required fields
				if ($auth.user && $auth.user.id) {
					auth.setUser({
						...$auth.user,
						...profileData,
						id: $auth.user.id // Ensure id is preserved
					});
				}
			} else {
				const errorMsg =
					response && typeof response === 'object' && 'message' in response
						? response.message
						: 'Failed to update profile';
				error = typeof errorMsg === 'string' ? errorMsg : 'Failed to update profile';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update profile';
		} finally {
			saving = false;
		}
	}

	async function updatePassword() {
		if (
			!passwordData.current_password ||
			!passwordData.new_password ||
			!passwordData.confirm_password
		) {
			error = 'All password fields are required';
			return;
		}

		if (passwordData.new_password !== passwordData.confirm_password) {
			error = 'New passwords do not match';
			return;
		}

		if (passwordData.new_password.length < 6) {
			error = 'Password must be at least 6 characters';
			return;
		}

		try {
			saving = true;
			error = '';
			message = '';
			const response = await api.updatePassword({
				current_password: passwordData.current_password,
				new_password: passwordData.new_password
			});

			if (response && typeof response === 'object' && 'success' in response && response.success) {
				message = 'Password updated successfully';
				passwordData = {
					current_password: '',
					new_password: '',
					confirm_password: ''
				};
			} else {
				const errorMsg =
					response && typeof response === 'object' && 'message' in response
						? response.message
						: 'Failed to update password';
				error = typeof errorMsg === 'string' ? errorMsg : 'Failed to update password';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update password';
		} finally {
			saving = false;
		}
	}

	async function updatePreferences() {
		try {
			saving = true;
			error = '';
			message = '';
			const response = await api.updateUserPreferences(preferences);
			if (response && typeof response === 'object' && 'success' in response && response.success) {
				message = 'Preferences updated successfully';
			} else {
				const errorMsg =
					response && typeof response === 'object' && 'message' in response
						? response.message
						: 'Failed to update preferences';
				error = typeof errorMsg === 'string' ? errorMsg : 'Failed to update preferences';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update preferences';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Settings - GoSight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white shadow dark:bg-gray-800">
		<div class="px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 items-center justify-between">
				<div class="flex items-center">
					<Settings class="mr-3 h-6 w-6 text-gray-400" />
					<h1 class="text-xl font-semibold text-gray-900 dark:text-white">Settings</h1>
				</div>
			</div>
		</div>
	</div>

	{#if message}
		<div class="mx-4 mt-4 sm:mx-6 lg:mx-8">
			<div
				class="rounded-md border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20"
			>
				<p class="text-green-800 dark:text-green-200">{message}</p>
			</div>
		</div>
	{/if}

	{#if error}
		<div class="mx-4 mt-4 sm:mx-6 lg:mx-8">
			<div
				class="rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
			>
				<div class="flex">
					<AlertCircle class="mr-2 h-5 w-5 text-red-400" />
					<p class="text-red-800 dark:text-red-200">{error}</p>
				</div>
			</div>
		</div>
	{/if}

	<div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
		<div class="px-4 py-6 sm:px-0">
			<div class="lg:grid lg:grid-cols-12 lg:gap-x-5">
				<!-- Sidebar -->
				<aside class="px-2 py-6 sm:px-6 lg:col-span-3 lg:px-0 lg:py-0">
					<nav class="space-y-1">
						{#each [{ id: 'profile', label: 'Profile', icon: User }, { id: 'security', label: 'Security', icon: Settings }, { id: 'preferences', label: 'Preferences', icon: Settings }] as tab}
							<button
								class="group flex w-full items-center rounded-md px-3 py-2 text-left text-sm font-medium {activeTab ===
								tab.id
									? 'bg-gray-50 text-blue-700 dark:bg-gray-800 dark:text-blue-400'
									: 'text-gray-900 hover:bg-gray-50 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800'}"
								on:click={() => (activeTab = tab.id)}
							>
								<svelte:component this={tab.icon} class="mr-3 h-5 w-5 text-gray-400" />
								{tab.label}
							</button>
						{/each}
					</nav>
				</aside>

				<!-- Main content -->
				<div class="space-y-6 sm:px-6 lg:col-span-9 lg:px-0">
					{#if activeTab === 'profile'}
						<div class="rounded-lg bg-white shadow dark:bg-gray-800">
							<div class="px-4 py-5 sm:p-6">
								<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
									Profile Information
								</h3>
								<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
									Update your account's profile information.
								</p>

								<form class="mt-6 space-y-6" on:submit|preventDefault={updateProfile}>
									<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
										<div>
											<label
												for="first_name"
												class="block text-sm font-medium text-gray-700 dark:text-gray-300"
											>
												First Name
											</label>
											<input
												type="text"
												name="first_name"
												id="first_name"
												bind:value={profileData.first_name}
												class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
										</div>

										<div>
											<label
												for="last_name"
												class="block text-sm font-medium text-gray-700 dark:text-gray-300"
											>
												Last Name
											</label>
											<input
												type="text"
												name="last_name"
												id="last_name"
												bind:value={profileData.last_name}
												class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
										</div>
									</div>

									<div>
										<label
											for="username"
											class="block text-sm font-medium text-gray-700 dark:text-gray-300"
										>
											Username
										</label>
										<input
											type="text"
											name="username"
											id="username"
											bind:value={profileData.username}
											class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
										/>
									</div>

									<div>
										<label
											for="email"
											class="block text-sm font-medium text-gray-700 dark:text-gray-300"
										>
											Email Address
										</label>
										<input
											type="email"
											name="email"
											id="email"
											bind:value={profileData.email}
											class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
										/>
									</div>

									<div class="flex justify-end">
										<button
											type="submit"
											disabled={saving}
											class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
										>
											{#if saving}
												<div
													class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"
												></div>
											{:else}
												<Save class="mr-2 h-4 w-4" />
											{/if}
											Save Changes
										</button>
									</div>
								</form>
							</div>
						</div>

						<!-- Account Info -->
						{#if $auth.user}
							<div class="rounded-lg bg-white shadow dark:bg-gray-800">
								<div class="px-4 py-5 sm:p-6">
									<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
										Account Information
									</h3>
									<dl class="mt-6 grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
										<div>
											<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
												Member since
											</dt>
											<dd class="mt-1 text-sm text-gray-900 dark:text-white">
												{$auth.user.created_at ? formatDate($auth.user.created_at) : 'Unknown'}
											</dd>
										</div>
										<div>
											<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
												Last login
											</dt>
											<dd class="mt-1 text-sm text-gray-900 dark:text-white">
												{currentUser?.last_login ? formatDate(currentUser.last_login) : 'Never'}
											</dd>
										</div>
										<div>
											<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Role</dt>
											<dd class="mt-1 text-sm text-gray-900 dark:text-white">
												{currentUser?.role}
											</dd>
										</div>
										<div>
											<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
											<dd class="mt-1 text-sm text-gray-900 dark:text-white">
												<span
													class="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800 dark:bg-green-900 dark:text-green-200"
												>
													Active
												</span>
											</dd>
										</div>
									</dl>
								</div>
							</div>
						{/if}
					{:else if activeTab === 'security'}
						<div class="rounded-lg bg-white shadow dark:bg-gray-800">
							<div class="px-4 py-5 sm:p-6">
								<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
									Change Password
								</h3>
								<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
									Update your password to keep your account secure.
								</p>

								<form class="mt-6 space-y-6" on:submit|preventDefault={updatePassword}>
									<div>
										<label
											for="current_password"
											class="block text-sm font-medium text-gray-700 dark:text-gray-300"
										>
											Current Password
										</label>
										<div class="relative mt-1">
											<input
												type={showCurrentPassword ? 'text' : 'password'}
												name="current_password"
												id="current_password"
												bind:value={passwordData.current_password}
												class="block w-full rounded-md border-gray-300 pr-10 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
											<button
												type="button"
												class="absolute inset-y-0 right-0 flex items-center pr-3"
												on:click={() => (showCurrentPassword = !showCurrentPassword)}
											>
												{#if showCurrentPassword}
													<EyeOff class="h-4 w-4 text-gray-400" />
												{:else}
													<Eye class="h-4 w-4 text-gray-400" />
												{/if}
											</button>
										</div>
									</div>

									<div>
										<label
											for="new_password"
											class="block text-sm font-medium text-gray-700 dark:text-gray-300"
										>
											New Password
										</label>
										<div class="relative mt-1">
											<input
												type={showNewPassword ? 'text' : 'password'}
												name="new_password"
												id="new_password"
												bind:value={passwordData.new_password}
												class="block w-full rounded-md border-gray-300 pr-10 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
											<button
												type="button"
												class="absolute inset-y-0 right-0 flex items-center pr-3"
												on:click={() => (showNewPassword = !showNewPassword)}
											>
												{#if showNewPassword}
													<EyeOff class="h-4 w-4 text-gray-400" />
												{:else}
													<Eye class="h-4 w-4 text-gray-400" />
												{/if}
											</button>
										</div>
									</div>

									<div>
										<label
											for="confirm_password"
											class="block text-sm font-medium text-gray-700 dark:text-gray-300"
										>
											Confirm New Password
										</label>
										<div class="relative mt-1">
											<input
												type={showConfirmPassword ? 'text' : 'password'}
												name="confirm_password"
												id="confirm_password"
												bind:value={passwordData.confirm_password}
												class="block w-full rounded-md border-gray-300 pr-10 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
											<button
												type="button"
												class="absolute inset-y-0 right-0 flex items-center pr-3"
												on:click={() => (showConfirmPassword = !showConfirmPassword)}
											>
												{#if showConfirmPassword}
													<EyeOff class="h-4 w-4 text-gray-400" />
												{:else}
													<Eye class="h-4 w-4 text-gray-400" />
												{/if}
											</button>
										</div>
									</div>

									<div class="flex justify-end">
										<button
											type="submit"
											disabled={saving}
											class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
										>
											{#if saving}
												<div
													class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"
												></div>
											{:else}
												<Save class="mr-2 h-4 w-4" />
											{/if}
											Update Password
										</button>
									</div>
								</form>
							</div>
						</div>
					{:else if activeTab === 'preferences'}
						<div class="space-y-6">
							<!-- Theme Settings -->
							<div class="rounded-lg bg-white shadow dark:bg-gray-800">
								<div class="px-4 py-5 sm:p-6">
									<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">Theme</h3>
									<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
										Choose your preferred theme.
									</p>

									<div class="mt-6">
										<select
											bind:value={preferences.theme}
											class="mt-1 block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-base focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
										>
											<option value="light">Light</option>
											<option value="dark">Dark</option>
											<option value="system">System</option>
										</select>
									</div>
								</div>
							</div>

							<!-- Notification Settings -->
							<div class="rounded-lg bg-white shadow dark:bg-gray-800">
								<div class="px-4 py-5 sm:p-6">
									<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
										Notifications
									</h3>
									<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
										Configure your notification preferences.
									</p>

									<div class="mt-6 space-y-4">
										<div class="flex items-center">
											<input
												id="email_alerts"
												type="checkbox"
												bind:checked={preferences.notifications.email_alerts}
												class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
											/>
											<label
												for="email_alerts"
												class="ml-2 block text-sm text-gray-900 dark:text-white"
											>
												Email alerts
											</label>
										</div>

										<div class="flex items-center">
											<input
												id="push_alerts"
												type="checkbox"
												bind:checked={preferences.notifications.push_alerts}
												class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
											/>
											<label
												for="push_alerts"
												class="ml-2 block text-sm text-gray-900 dark:text-white"
											>
												Push notifications
											</label>
										</div>

										<div>
											<label
												for="alert_frequency"
												class="block text-sm font-medium text-gray-700 dark:text-gray-300"
											>
												Alert frequency
											</label>
											<select
												id="alert_frequency"
												bind:value={preferences.notifications.alert_frequency}
												class="mt-1 block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-base focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											>
												<option value="immediate">Immediate</option>
												<option value="hourly">Hourly digest</option>
												<option value="daily">Daily digest</option>
											</select>
										</div>
									</div>
								</div>
							</div>

							<!-- Dashboard Settings -->
							<div class="rounded-lg bg-white shadow dark:bg-gray-800">
								<div class="px-4 py-5 sm:p-6">
									<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
										Dashboard
									</h3>
									<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
										Customize your dashboard preferences.
									</p>

									<div class="mt-6 space-y-4">
										<div>
											<label
												for="refresh_interval"
												class="block text-sm font-medium text-gray-700 dark:text-gray-300"
											>
												Refresh interval (seconds)
											</label>
											<input
												type="number"
												id="refresh_interval"
												min="10"
												max="300"
												bind:value={preferences.dashboard.refresh_interval}
												class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
										</div>

										<div>
											<label
												for="default_time_range"
												class="block text-sm font-medium text-gray-700 dark:text-gray-300"
											>
												Default time range
											</label>
											<select
												id="default_time_range"
												bind:value={preferences.dashboard.default_time_range}
												class="mt-1 block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-base focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											>
												<option value="15m">15 minutes</option>
												<option value="1h">1 hour</option>
												<option value="6h">6 hours</option>
												<option value="24h">24 hours</option>
												<option value="7d">7 days</option>
											</select>
										</div>

										<div class="flex items-center">
											<input
												id="show_system_metrics"
												type="checkbox"
												bind:checked={preferences.dashboard.show_system_metrics}
												class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
											/>
											<label
												for="show_system_metrics"
												class="ml-2 block text-sm text-gray-900 dark:text-white"
											>
												Show system metrics on dashboard
											</label>
										</div>
									</div>
								</div>
							</div>

							<div class="flex justify-end">
								<button
									type="button"
									disabled={saving}
									on:click={updatePreferences}
									class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
								>
									{#if saving}
										<div
											class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"
										></div>
									{:else}
										<Save class="mr-2 h-4 w-4" />
									{/if}
									Save Preferences
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>
