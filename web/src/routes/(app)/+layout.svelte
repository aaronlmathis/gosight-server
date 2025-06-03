<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/authStore';
	import '../../app.css';
	import TopNavbar from '$lib/components/TopNavbar.svelte';
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Toaster } from '$lib/components/ui/sonner';

	let { children } = $props();

	onMount(async () => {
        // Initialize auth store
        console.log('Initializing auth store...');
        try {
            await auth.init();
            console.log('Auth initialized:', {
                isAuthenticated: $auth.isAuthenticated,
                user: $auth.user
            });
        } catch (error) {
            console.error('Auth initialization failed:', error);
        }
    });
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

<Toaster position="bottom-right" />
