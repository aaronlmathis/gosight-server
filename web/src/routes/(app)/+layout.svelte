<script lang="ts">
	import '../../app.css';
	import TopNavbar from '$lib/components/TopNavbar.svelte';
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Toaster } from '$lib/components/ui/sonner';

	let { children } = $props();
</script>

<TopNavbar />

<div class="flex h-screen">
	<Sidebar.Provider class="pt-16">
		<AppSidebar />
		<Sidebar.Inset class="flex-1 flex flex-col">
			{@render children()}
		</Sidebar.Inset>
	</Sidebar.Provider>
</div>

<Toaster position="top-right" />
