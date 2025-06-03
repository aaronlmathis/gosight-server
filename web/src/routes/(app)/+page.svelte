<script lang="ts">
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';

	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import TopNavbar from '$lib/components/TopNavbar.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import DashboardGrid from '$lib/components/dashboard/EnhancedDashboardGrid.svelte';
	import DashboardTabs from '$lib/components/dashboard/DashboardTabs.svelte';
	import EditModeToolbar from '$lib/components/dashboard/EditModeToolbar.svelte';
	import { dashboardStore, isEditMode } from '$lib/stores/dashboardStore';
	import { onMount } from 'svelte';
	import { Edit3, Save, RotateCcw } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';

	// Import debug utilities in development
	import '$lib/debug';

	onMount(async () => {
		// Load dashboard
		await dashboardStore.load();

		// Dashboard is now loaded - widgets will persist if saved
		const currentDashboard = $dashboardStore;
		console.log('Current dashboard state:', currentDashboard);
	});

	// Edit mode functions
	function toggleEditMode() {
		isEditMode.update((mode: boolean) => !mode);
		const message = $isEditMode ? 'Edit mode enabled' : 'Edit mode disabled';
		console.log('🔧 Edit mode toggled:', !$isEditMode, '->', $isEditMode);
		toast.success(message);
	}

	function saveDashboard() {
		toast.success('Dashboard saved', {
			description: 'All changes have been saved successfully'
		});
	}

	function resetDashboard() {
		if (confirm('Are you sure you want to reset the dashboard? This will remove all widgets.')) {
			dashboardStore.reset();
			toast.success('Dashboard reset');
		}
	}
</script>

<svelte:head>
	<title>Dashboard - GoSight</title>
</svelte:head>
<PermissionGuard requiredPermission="gosight:dashboard:view">
	<!-- Top Navbar (Fixed at top) -->
	<TopNavbar />

	<!-- Main Layout with sidebar below header -->
	<div class="flex h-screen">
		<Sidebar.Provider class="pt-16">
			<AppSidebar />
			<Sidebar.Inset class="flex flex-1 flex-col">
				<!-- Breadcrumb Header -->
				<header class="bg-background flex h-16 shrink-0 items-center gap-2 border-b">
					<div class="flex items-center gap-2 px-4">
						<Sidebar.Trigger class="-ml-1" />
						<Separator orientation="vertical" class="mr-2 data-[orientation=vertical]:h-4" />
						<Breadcrumb.Root>
							<Breadcrumb.List>
								<Breadcrumb.Item class="hidden md:block">
									<Breadcrumb.Link href="#" class="">GoSight</Breadcrumb.Link>
								</Breadcrumb.Item>
								<Breadcrumb.Separator class="hidden md:block" />
								<Breadcrumb.Item>
									<Breadcrumb.Page>Dashboard</Breadcrumb.Page>
								</Breadcrumb.Item>
							</Breadcrumb.List>
						</Breadcrumb.Root>
					</div>

					<!-- Edit Mode Controls -->
					<div class="ml-auto flex items-center gap-2 pr-4">
						<!-- Edit Mode Toggle -->
						<Button
							variant={$isEditMode ? 'default' : 'outline'}
							size="sm"
							onclick={toggleEditMode}
							class="gap-2"
						>
							<Edit3 class="h-4 w-4" />
							{$isEditMode ? 'Exit Edit' : 'Edit Mode'}
						</Button>

						<!-- Save and Reset buttons (only visible in edit mode) -->
						{#if $isEditMode}
							<Button variant="outline" size="sm" onclick={saveDashboard} class="gap-2">
								<Save class="h-4 w-4" />
								Save
							</Button>

							<Button variant="outline" size="sm" onclick={resetDashboard} class="gap-2">
								<RotateCcw class="h-4 w-4" />
								Reset
							</Button>
						{/if}
					</div>
				</header>

				<!-- Dashboard Content with Tabs -->
				<div class="flex flex-1 flex-col">
					<DashboardTabs>
						{#snippet children({ dashboard }: { dashboard: any })}
							<div class="flex flex-1 flex-col">
								<!-- Edit Mode Toolbar (appears below tabs when in edit mode) -->
								{#if $isEditMode}
									<div class="px-4 pt-4">
										<EditModeToolbar />
									</div>
								{/if}

								<!-- Dashboard Grid -->
								<div class="flex-1 overflow-auto">
									<DashboardGrid />
								</div>
							</div>
						{/snippet}
					</DashboardTabs>
				</div>
			</Sidebar.Inset>
		</Sidebar.Provider>
	</div>
</PermissionGuard>
