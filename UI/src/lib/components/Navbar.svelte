<script lang="ts">
	import { darkMode, activeAlertsCount } from '$lib/stores';
	import { auth } from '$lib/stores/auth';
	import { getUserInitials } from '$lib/utils';
	import { Moon, Sun, Bell, Menu } from 'lucide-svelte';

	let dropdownOpen = false;
	let alertDropdownOpen = false;

	function toggleTheme() {
		darkMode.update((isDark) => !isDark);
	}

	function toggleSidebar() {
		const sidebar = document.getElementById('sidebar');
		const backdrop = document.getElementById('sidebarBackdrop');
		const isMobile = window.innerWidth < 1024;
		const submenus = document.querySelectorAll('.sidebar-submenu');

		if (isMobile) {
			const isNowHidden = sidebar?.classList.toggle('hidden');
			backdrop?.classList.toggle('hidden');
			submenus.forEach((ul) => {
				if (ul instanceof HTMLElement) {
					ul.style.display = isNowHidden ? 'none' : '';
				}
			});
		} else {
			const collapsed = document.body.classList.toggle('sidebar-collapsed');
			localStorage.setItem('sidebarCollapsed', collapsed.toString());
		}
	}

	function toggleDropdown() {
		dropdownOpen = !dropdownOpen;
	}

	function toggleAlertDropdown() {
		alertDropdownOpen = !alertDropdownOpen;
	}

	// Close dropdowns when clicking outside
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as Element;
		if (!target.closest('#dropdownUserAvatarButton') && !target.closest('#dropdownUserAvatar')) {
			dropdownOpen = false;
		}
		if (!target.closest('#alert-bell') && !target.closest('#alert-dropdown')) {
			alertDropdownOpen = false;
		}
	}
</script>

<svelte:window on:click={handleClickOutside} />

<nav class="fixed z-30 w-full border-b border-gray-700 bg-gray-800 text-white">
	<div class="px-3 py-3 lg:px-5 lg:pl-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center justify-start">
				<!-- Mobile sidebar toggle -->
				<button
					on:click={toggleSidebar}
					class="rounded p-2 text-gray-400 hover:bg-gray-700 hover:text-white focus:ring-2 focus:ring-gray-600"
					aria-label="Toggle sidebar"
				>
					<Menu class="h-6 w-6" />
				</button>

				<a href="/" class="ml-2 flex md:mr-24">
					<span
						class="self-center text-xl font-bold tracking-wide whitespace-nowrap text-blue-400 sm:text-2xl"
					>
						GoSight
					</span>
				</a>
			</div>

			<div class="flex items-center space-x-4">
				<!-- Theme toggle -->
				<button
					on:click={toggleTheme}
					class="text-gray-400 hover:text-white focus:outline-none"
					aria-label="Toggle theme"
				>
					{#if !$darkMode}
						<Moon class="h-5 w-5" />
					{:else}
						<Sun class="h-5 w-5" />
					{/if}
				</button>

				<!-- Notification bell -->
				<div class="relative">
					<button
						id="alert-bell"
						on:click={toggleAlertDropdown}
						class="relative text-gray-400 hover:text-white focus:outline-none"
						aria-label="Notifications"
					>
						<Bell class="h-5 w-5" />
						{#if $activeAlertsCount > 0}
							<span class="absolute -top-1 -right-1 h-2 w-2 rounded-full bg-red-600"></span>
						{/if}
					</button>

					{#if alertDropdownOpen}
						<div
							id="alert-dropdown"
							class="absolute right-0 z-50 mt-2 w-80 rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="border-b border-gray-200 p-4 dark:border-gray-700">
								<h3 class="font-semibold text-gray-900 dark:text-white">Notifications</h3>
							</div>
							<div class="max-h-64 overflow-y-auto">
								{#if $activeAlertsCount === 0}
									<div class="p-4 text-center text-gray-500 dark:text-gray-400">
										No new notifications
									</div>
								{:else}
									<!-- Alert notifications will be populated here -->
									<div class="p-4 text-center text-gray-500 dark:text-gray-400">
										{$activeAlertsCount} new alerts
									</div>
								{/if}
							</div>
						</div>
					{/if}
				</div>

				<!-- User dropdown -->
				{#if $auth.isAuthenticated && $auth.user}
					<div class="relative">
						<button
							id="dropdownUserAvatarButton"
							on:click={toggleDropdown}
							class="h-11 w-11 rounded-full object-cover shadow-md focus:ring-2 focus:ring-gray-600"
						>
							{#if $auth.user.profile?.avatar_url}
								<img
									class="h-11 w-11 rounded-full object-cover"
									src={$auth.user.profile.avatar_url}
									alt="User avatar"
								/>
							{:else}
								<div
									class="relative inline-flex h-11 w-11 items-center justify-center overflow-hidden rounded-full bg-gray-100 dark:bg-gray-600"
								>
									<span class="font-medium text-gray-600 dark:text-gray-300">
										{getUserInitials(
											$auth.user.firstName || $auth.user.first_name,
											$auth.user.lastName || $auth.user.last_name
										)}
									</span>
								</div>
							{/if}
						</button>

						{#if dropdownOpen}
							<div
								id="dropdownUserAvatar"
								class="absolute right-0 z-50 mt-2 w-48 rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
							>
								<div class="border-b border-gray-200 px-4 py-3 dark:border-gray-700">
									<span class="block text-sm text-gray-900 dark:text-white">
										{$auth.user.firstName}
										{$auth.user.lastName}
									</span>
									<span class="block truncate text-sm text-gray-500 dark:text-gray-400">
										{$auth.user.email}
									</span>
								</div>
								<ul class="py-2">
									<li>
										<a
											href="/profile"
											class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
										>
											Profile
										</a>
									</li>
									<li>
										<a
											href="/settings"
											class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-gray-700"
										>
											Settings
										</a>
									</li>
									<li>
										<button
											on:click={() => auth.logout()}
											class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-100 dark:text-red-400 dark:hover:bg-gray-700"
										>
											Sign out
										</button>
									</li>
								</ul>
							</div>
						{/if}
					</div>
				{:else}
					<a href="/auth/login" class="font-medium text-blue-400 hover:text-blue-300"> Sign in </a>
				{/if}
			</div>
		</div>
	</div>
</nav>
