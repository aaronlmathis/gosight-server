<!-- IAM Layout -->
<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { Users, Shield, Key } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	// Get current path for active navigation
	$: currentPath = $page.url.pathname;

	// IAM navigation items
	const iamNavItems = [
		{
			href: '/iam/users',
			label: 'Users',
			icon: Users,
			description: 'Manage user accounts and assignments'
		},
		{
			href: '/iam/roles',
			label: 'Roles',
			icon: Shield,
			description: 'Define and manage user roles'
		},
		{
			href: '/iam/permissions',
			label: 'Permissions',
			icon: Key,
			description: 'Configure system permissions'
		}
	];

	function isActive(href: string): boolean {
		return currentPath.startsWith(href);
	}

	// Redirect to users by default
	onMount(() => {
		if (currentPath === '/iam' || currentPath === '/iam/') {
			goto('/iam/users');
		}
	});
</script>

<div class="flex flex-1 flex-col overflow-hidden">
	<div class="flex flex-1 overflow-hidden">
		<!-- IAM Sidebar -->
		<nav
			class="w-64 overflow-y-auto border-r border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800"
		>
			<div class="p-4">
				<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
					Identity and Access Management
				</h2>
				<ul class="space-y-2">
					{#each iamNavItems as item}
						<li>
							<a
								href={item.href}
								class="group flex items-start rounded-lg p-3 transition-colors duration-200 {isActive(
									item.href
								)
									? 'border border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-900/20 dark:text-blue-300'
									: 'text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-gray-700'}"
							>
								<svelte:component
									this={item.icon}
									class="mt-0.5 mr-3 h-5 w-5 {isActive(item.href)
										? 'text-blue-600 dark:text-blue-400'
										: 'text-gray-400 group-hover:text-gray-600 dark:text-gray-500 dark:group-hover:text-gray-300'}"
								/>
								<div>
									<div class="text-sm font-medium">{item.label}</div>
									<div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
										{item.description}
									</div>
								</div>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		</nav>

		<!-- Main Content Area -->
		<main class="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-900">
			<slot />
		</main>
	</div>
</div>
