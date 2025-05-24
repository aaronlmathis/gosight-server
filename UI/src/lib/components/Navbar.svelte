<script lang="ts">
	import { darkMode, activeAlertsCount } from '$lib/stores';
	import { Moon, Sun, Bell, Menu } from 'lucide-svelte';
	import type { User } from '$lib/types';

	export let user: User | null;
	export let currentPath: string;

	let dropdownOpen = false;
	let alertDropdownOpen = false;

	function toggleTheme() {
		darkMode.update(isDark => !isDark);
	}

	function toggleSidebar() {
		const sidebar = document.getElementById('sidebar');
		const backdrop = document.getElementById('sidebarBackdrop');
		const isMobile = window.innerWidth < 1024;
		const submenus = document.querySelectorAll('.sidebar-submenu');

		if (isMobile) {
			const isNowHidden = sidebar?.classList.toggle('hidden');
			backdrop?.classList.toggle('hidden');
			submenus.forEach(ul => {
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

<nav class="fixed z-30 w-full bg-gray-800 border-b border-gray-700 text-white">
	<div class="px-3 py-3 lg:px-5 lg:pl-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center justify-start">
				<!-- Mobile sidebar toggle -->
				<button
					on:click={toggleSidebar}
					class="p-2 text-gray-400 rounded hover:text-white hover:bg-gray-700 focus:ring-2 focus:ring-gray-600"
					aria-label="Toggle sidebar"
				>
					<Menu class="w-6 h-6" />
				</button>
				
				<a href="/" class="flex ml-2 md:mr-24">
					<span class="self-center text-xl sm:text-2xl whitespace-nowrap font-bold text-blue-400 tracking-wide">
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
						<Moon class="w-5 h-5" />
					{:else}
						<Sun class="w-5 h-5" />
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
						<Bell class="w-5 h-5" />
						{#if $activeAlertsCount > 0}
							<span class="absolute -top-1 -right-1 w-2 h-2 bg-red-600 rounded-full"></span>
						{/if}
					</button>
					
					{#if alertDropdownOpen}
						<div
							id="alert-dropdown"
							class="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50"
						>
							<div class="p-4 border-b border-gray-200 dark:border-gray-700">
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
				{#if user}
					<div class="relative">
						<button
							id="dropdownUserAvatarButton"
							on:click={toggleDropdown}
							class="w-11 h-11 rounded-full shadow-md object-cover focus:ring-2 focus:ring-gray-600"
						>
							<img
								class="w-11 h-11 rounded-full object-cover"
								src={user.avatar || '/default-avatar.png'}
								alt="User avatar"
							/>
						</button>
						
						{#if dropdownOpen}
							<div
								id="dropdownUserAvatar"
								class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 z-50"
							>
								<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
									<span class="block text-sm text-gray-900 dark:text-white">
										{user.firstName} {user.lastName}
									</span>
									<span class="block text-sm text-gray-500 dark:text-gray-400 truncate">
										{user.email}
									</span>
								</div>
								<ul class="py-2">
									<li>
										<a
											href="/profile"
											class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700"
										>
											Profile
										</a>
									</li>
									<li>
										<a
											href="/settings"
											class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700"
										>
											Settings
										</a>
									</li>
									<li>
										<a
											href="/logout"
											class="block px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-gray-100 dark:hover:bg-gray-700"
										>
											Sign out
										</a>
									</li>
								</ul>
							</div>
						{/if}
					</div>
				{:else}
					<a
						href="/login"
						class="text-blue-400 hover:text-blue-300 font-medium"
					>
						Sign in
					</a>
				{/if}
			</div>
		</div>
	</div>
</nav>
