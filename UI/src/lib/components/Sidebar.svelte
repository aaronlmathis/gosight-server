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
			href: '/admin',
			label: 'Admin',
			icon: Settings,
			tooltip: 'Admin'
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
		return item.submenu.some(sub => currentPath.startsWith(sub.href));
	}
</script>

<aside 
	id="sidebar"
	class="fixed top-0 left-0 z-20 flex flex-col w-64 h-full pt-16 transition-transform transform bg-gray-100 dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700 lg:translate-x-0 hidden lg:flex"
	aria-label="Sidebar"
>
	<div class="flex flex-col flex-1 px-3 py-4 overflow-visible">
		<ul class="space-y-2 text-sm font-normal text-gray-700 dark:text-gray-300">
			{#each menuItems as item}
				<li>
					<a
						href={item.href}
						data-tooltip={item.tooltip}
						class="sidebar-link flex items-center p-2 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white transition-all duration-200 relative"
						class:active={isActive(item.href)}
						class:mb-1={item.submenu}
					>
						<svelte:component 
							this={item.icon} 
							class="w-6 h-6 text-gray-800 dark:text-white" 
						/>
						<span class="ml-3">{item.label}</span>
					</a>
					
					{#if item.submenu && (isActive(item.href) || hasActiveSubmenu(item))}
						<ul class="sidebar-submenu ml-6 mt-1 space-y-1 text-sm text-gray-600 dark:text-gray-400 mt-2">
							{#each item.submenu as subItem}
								<li>
									<a
										href={subItem.href}
										class="submenu-link relative block px-3 py-1.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 transition-all duration-150"
										class:active={isActive(subItem.href)}
									>
										<i class="{subItem.icon} text-xs w-3 h-3"></i>
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
	
	<div class="text-gray-500 text-center text-xs mt-6">
		<p class="text-gray-800 dark:text-gray-200">
			Version: <span class="text-black dark:text-white font-semibold">1.0.0</span>
		</p>
	</div>
</aside>
