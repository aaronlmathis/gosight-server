<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { page } from '$app/stores';

	interface NavItem {
		href: string;
		label: string;
		requiredPermission?: string;
		icon?: string;
	}

	const navItems: NavItem[] = [
		{
			href: '/',
			label: 'Dashboard',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
		},
		{
			href: '/alerts/active',
			label: 'Active Alerts',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.98-.833-2.75 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z'
		},
		{
			href: '/alerts/history',
			label: 'Alert History',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z'
		},
		{
			href: '/alerts/rules',
			label: 'Alert Rules',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z'
		},
		{
			href: '/metrics',
			label: 'Metrics',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'
		},
		{
			href: '/logs',
			label: 'Logs',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z'
		},
		{
			href: '/activity',
			label: 'Activity',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6'
		},
		{
			href: '/endpoints',
			label: 'Endpoints',
			requiredPermission: 'gosight:dashboard:view',
			icon: 'M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z'
		}
	];

	function hasPermission(permission?: string): boolean {
		if (!permission) return true;
		return auth.hasPermission(permission);
	}

	function isCurrentPage(href: string): boolean {
		return $page.url.pathname === href;
	}
</script>

<nav class="border-b border-gray-200 bg-white shadow-sm">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="flex h-16 justify-between">
			<div class="flex">
				<div class="flex flex-shrink-0 items-center">
					<span class="text-2xl font-bold text-blue-600">GoSight</span>
				</div>
				<div class="hidden sm:ml-6 sm:flex sm:space-x-8">
					{#each navItems as item}
						{#if hasPermission(item.requiredPermission)}
							<a
								href={item.href}
								class="inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium {isCurrentPage(
									item.href
								)
									? 'border-blue-500 text-gray-900'
									: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
							>
								{#if item.icon}
									<svg class="mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d={item.icon}
										/>
									</svg>
								{/if}
								{item.label}
							</a>
						{/if}
					{/each}
				</div>
			</div>
			<div class="hidden sm:ml-6 sm:flex sm:items-center">
				{#if $auth.isAuthenticated && $auth.user}
					<div class="relative ml-3">
						<div class="flex items-center space-x-4">
							<span class="text-sm text-gray-700">
								Welcome, {$auth.user.username}
							</span>
							<button
								on:click={() => auth.logout()}
								class="rounded-full bg-white p-1 text-gray-400 hover:text-gray-500 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
							>
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
									/>
								</svg>
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Mobile menu -->
	<div class="sm:hidden">
		<div class="space-y-1 pt-2 pb-3">
			{#each navItems as item}
				{#if hasPermission(item.requiredPermission)}
					<a
						href={item.href}
						class="block border-l-4 py-2 pr-4 pl-3 text-base font-medium {isCurrentPage(item.href)
							? 'border-blue-500 bg-blue-50 text-blue-700'
							: 'border-transparent text-gray-500 hover:border-gray-300 hover:bg-gray-50 hover:text-gray-700'}"
					>
						<div class="flex items-center">
							{#if item.icon}
								<svg class="mr-3 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d={item.icon}
									/>
								</svg>
							{/if}
							{item.label}
						</div>
					</a>
				{/if}
			{/each}
		</div>
		{#if $auth.isAuthenticated && $auth.user}
			<div class="border-t border-gray-200 pt-4 pb-3">
				<div class="flex items-center px-4">
					<div class="text-base font-medium text-gray-800">{$auth.user.username}</div>
				</div>
				<div class="mt-3 space-y-1">
					<button
						on:click={() => auth.logout()}
						class="block w-full px-4 py-2 text-left text-base font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-800"
					>
						Sign out
					</button>
				</div>
			</div>
		{/if}
	</div>
</nav>
