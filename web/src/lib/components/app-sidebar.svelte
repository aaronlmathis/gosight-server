<script lang="ts" module>
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
	import DashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import SearchIcon from '@lucide/svelte/icons/search';
	import ServerIcon from '@lucide/svelte/icons/server';
	import ContainerIcon from '@lucide/svelte/icons/box';
	import ServiceIcon from '@lucide/svelte/icons/cpu';
	import NetworkIcon from '@lucide/svelte/icons/network';
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import ZapIcon from '@lucide/svelte/icons/zap';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import UsersIcon from '@lucide/svelte/icons/users';
	import GalleryVerticalEndIcon from '@lucide/svelte/icons/gallery-vertical-end';
	import EyeIcon from '@lucide/svelte/icons/eye';

	// This is sample data.
	const data = {
		user: {
			name: 'Aaron Mathis',
			email: 'aaron@gosight.io',
			avatar: '/dev/avatar.jpg'
		},
		teams: [
			{
				name: 'GoSight',
				logo: GalleryVerticalEndIcon,
				plan: 'Enterprise'
			}
		],
		navGroups: [
			{
				groupTitle: 'Platform',
				items: [
					{
						title: 'Dashboards',
						url: '#',
						icon: DashboardIcon,
						isActive: true
					},
					{
						title: 'Explore',
						url: '#',
						icon: SearchIcon
					}
				]
			},
			{
				groupTitle: 'Resources',
				items: [
					{
						title: 'Hosts',
						url: '#',
						icon: ServerIcon
					},
					{
						title: 'Containers',
						url: '#',
						icon: ContainerIcon
					},
					{
						title: 'Services',
						url: '#',
						icon: ServiceIcon
					},
					{
						title: 'Network Devices',
						url: '#',
						icon: NetworkIcon
					}
				]
			},
			{
				groupTitle: 'Telemetry',
				items: [
					{
						title: 'Metrics',
						url: '#',
						icon: ActivityIcon,
						items: [
							{
								title: 'Metric Explorer',
								url: '#'
							}
						]
					},
					{
						title: 'Logs',
						url: '#',
						icon: FileTextIcon,
						items: [
							{
								title: 'Log Explorer',
								url: '#'
							},
							{
								title: 'Log Analyzer',
								url: '#'
							}
						]
					},
					{
						title: 'Traces',
						url: '#',
						icon: ZapIcon
					}
				]
			},
			{
				groupTitle: 'Alerts & Events',
				items: [
					{
						title: 'Alerts',
						url: '#',
						icon: AlertTriangleIcon,
						items: [
							{
								title: 'Rule Builder',
								url: '#'
							},
							{
								title: 'Alert History',
								url: '#'
							},
							{
								title: 'Active Alerts',
								url: '#'
							}
						]
					},
					{
						title: 'Events',
						url: '#',
						icon: CalendarIcon
					}
				]
			},
			{
				groupTitle: 'Administration',
				items: [
					{
						title: 'Settings',
						url: '#',
						icon: SettingsIcon
					},
					{
						title: 'Teams & Access',
						url: '#',
						icon: UsersIcon,
						items: [
							{
								title: 'Users',
								url: '#'
							},
							{
								title: 'Roles',
								url: '#'
							},
							{
								title: 'Permissions',
								url: '#'
							}
						]
					}
				]
			}
		]
	};
</script>

<script lang="ts">
	import NavGroups from './nav-groups.svelte';
	import NavUser from './nav-user.svelte';
	import TeamSwitcher from './team-switcher.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import type { ComponentProps } from 'svelte';

	let {
		ref = $bindable(null),
		collapsible = 'icon',
		...restProps
	}: ComponentProps<typeof Sidebar.Root> = $props();
</script>

<Sidebar.Root {collapsible} {...restProps} class="top-16 h-[calc(100vh-4rem)]">
	<Sidebar.Content>
		<NavGroups groups={data.navGroups} />
	</Sidebar.Content>
	<Sidebar.Footer>
		<NavUser user={data.user} />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
