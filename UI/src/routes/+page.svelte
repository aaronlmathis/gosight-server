<script lang="ts">
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import CustomizableDashboard from '$lib/components/dashboard/CustomizableDashboard.svelte';
</script>

<svelte:head>
	<title>Dashboard - GoSight</title>
</svelte:head>

<PermissionGuard requiredPermission="gosight:dashboard:view">
	<CustomizableDashboard />
</PermissionGuard>
