<script context="module">
	// Turn off SSR and built-in data fetching for settings routes
	export const ssr = false;
</script>

<script lang="ts">
	import SettingsSidebarNav from '$lib/components/settings-sidebar-nav.svelte';
	import { Separator } from '$lib/components/ui/separator/index.js';

	let { children } = $props();

	const sidebarNavItems = [
		{
			title: 'Profile',
			href: '/settings/profile'
		},
		{
			title: 'Security',
			href: '/settings/security'
		},
		{
			title: 'Preferences',
			href: '/settings/preferences'
		}
	];
</script>

<div class="md:hidden">
	<img
		src="/img/examples/forms-light.png"
		width={1280}
		height={791}
		alt="Forms"
		class="block dark:hidden"
	/>
	<img
		src="/img/examples/forms-dark.png"
		width={1280}
		height={791}
		alt="Forms"
		class="hidden dark:block"
	/>
</div>
<div class="hidden space-y-6 p-10 pb-16 md:block">
	<div class="space-y-0.5">
		<h2 class="text-2xl font-bold tracking-tight">Settings</h2>
		<p class="text-muted-foreground">Manage your account settings and set e-mail preferences.</p>
	</div>
	<Separator class="my-6" />
	<div class="flex flex-col space-y-8 lg:flex-row lg:space-y-0 lg:space-x-12">
		<aside class="-mx-4 lg:w-1/5">
			<SettingsSidebarNav items={sidebarNavItems} />
		</aside>
		<div class="flex-1 lg:ml-12 lg:max-w-2xl">{@render children()}</div>
	</div>
</div>
