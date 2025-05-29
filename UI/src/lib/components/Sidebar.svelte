<script lang="ts">
	import {
		Home,
		Network,
		BarChart3,
		Activity,
		FileText,
		AlertTriangle,
		Settings,
		TrendingUp,
		History,
		Bell
	} from 'lucide-svelte';

	export let currentPath: string;

	interface MenuItem {
		href: string;
		label: string;
		icon: any;
		tooltip: string;
		submenu?: SubMenuItem[];
	}

	interface SubMenuItem {
		href: string;
		label: string;
		icon: string;
	}

	const menuItems: MenuItem[] = [
		{
			href: '/',
			label: 'Overview',
			icon: Home,
			tooltip: 'Dashboard'
		},
		{
			href: '/endpoints',
			label: 'Endpoints',
			icon: Network,
			tooltip: 'Endpoints'
		},
		{
			href: '/metrics',
			label: 'Metric Explorer',
			icon: BarChart3,
			tooltip: 'Metrics'
		},
		{
			href: '/activity',
			label: 'Events',
			icon: Activity,
			tooltip: 'Unified Activity Stream'
		},
		{
			href: '/logs',
			label: 'Logs',
			icon: FileText,
			tooltip: 'Logs'
		},
		{
			href: '/alerts',
			label: 'Alerts',
			icon: AlertTriangle,
			tooltip: 'Alerts',
			submenu: [
				{
					href: '/alerts/rules',
					label: 'Rule Builder',
					icon: 'fas fa-cog'
				},
				{
					href: '/alerts/history',
					label: 'Alert History',
					icon: 'fas fa-history'
				},
				{
					href: '/alerts/active',
					label: 'Active Alerts',
					icon: 'fas fa-bell'
				}
			]
		},
		{
			href: '/reports',
			label: 'Reports',
			icon: TrendingUp,
			tooltip: 'Reports'
		},
		{
			href: '/iam',
			label: 'IAM',
			icon: Settings,
			tooltip: 'Identity and Access Management',
			submenu: [
				{
					href: '/iam/users',
					label: 'Users',
					icon: 'fas fa-users'
				},
				{
					href: '/iam/roles',
					label: 'Roles',
					icon: 'fas fa-user-tag'
				},
				{
					href: '/iam/permissions',
					label: 'Permissions',
					icon: 'fas fa-key'
				}
			]
		}
	];

	function isActive(href: string): boolean {
		if (href === '/') {
			return currentPath === '/';
		}
		return currentPath.startsWith(href);
	}

	function hasActiveSubmenu(item: MenuItem): boolean {
		if (!item.submenu) return false;
		return item.submenu.some((sub) => currentPath.startsWith(sub.href));
	}
</script>

<aside
	id="sidebar"
	class="fixed top-0 left-0 z-20 flex hidden h-full w-64 transform flex-col border-r border-gray-200 bg-gray-100 pt-20 transition-transform lg:flex lg:translate-x-0 dark:border-gray-700 dark:bg-gray-900"
	aria-label="Sidebar"
>
	<div class="flex flex-1 flex-col overflow-visible px-3 py-6">
		<ul class="space-y-2 text-sm font-normal text-gray-700 dark:text-gray-300">
			{#each menuItems as item}
				<li>
					<a
						href={item.href}
						data-tooltip={item.tooltip}
						class="sidebar-link relative flex items-center p-2 transition-all duration-200 hover:bg-gray-200 hover:text-gray-900 dark:hover:bg-gray-700 dark:hover:text-white"
						class:active={isActive(item.href)}
						class:mb-1={item.submenu}
					>
						<svelte:component this={item.icon} class="h-6 w-6 text-gray-800 dark:text-white" />
						<span class="ml-3">{item.label}</span>
					</a>

					{#if item.submenu && (isActive(item.href) || hasActiveSubmenu(item))}
						<ul
							class="sidebar-submenu mt-1 mt-2 ml-6 space-y-1 text-sm text-gray-600 dark:text-gray-400"
						>
							{#each item.submenu as subItem}
								<li>
									<a
										href={subItem.href}
										class="submenu-link relative block rounded-md px-3 py-1.5 transition-all duration-150 hover:bg-gray-100 dark:hover:bg-gray-800"
										class:active={isActive(subItem.href)}
									>
										<i class="{subItem.icon} h-3 w-3 text-xs"></i>
										{subItem.label}
									</a>
								</li>
							{/each}
						</ul>
					{/if}
				</li>
			{/each}
		</ul>
	</div>

	<div class="mt-6 text-center text-xs text-gray-500">
		<p class="text-gray-800 dark:text-gray-200">
			Version: <span class="font-semibold text-black dark:text-white">1.0.0</span>
		</p>
	</div>
</aside>
