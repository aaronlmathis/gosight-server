<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth';
	import Navigation from '$lib/components/Navigation.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { darkMode, user, activeAlertsCount } from '$lib/stores';
	import { alertsWS, eventsWS, logsWS, metricsWS } from '$lib/websocket';
	import type { LayoutData } from './$types';

	export let data: LayoutData;

	// Initialize stores with server data
	$: if (data.user) {
		user.set(data.user);
		// Also set user in auth store if we have user data from server
		auth.setUser(data.user);
	}

	// Initialize theme
	onMount(async () => {
		// Initialize auth store to check for existing session
		await auth.init();

		// Initialize theme
		const savedTheme = localStorage.getItem('color-theme');
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

		if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
			document.documentElement.classList.add('dark');
			darkMode.set(true);
		} else {
			document.documentElement.classList.remove('dark');
			darkMode.set(false);
		}

		// Subscribe to theme changes
		darkMode.subscribe((isDark) => {
			if (isDark) {
				document.documentElement.classList.add('dark');
				localStorage.setItem('color-theme', 'dark');
			} else {
				document.documentElement.classList.remove('dark');
				localStorage.setItem('color-theme', 'light');
			}
		});

		// Initialize WebSocket connections for real-time updates
		if (browser && (data.user || $auth.isAuthenticated)) {
			// WebSocket connections are already initialized via imports
			// Individual pages will connect to their specific WebSocket endpoints as needed
		}

		// Restore sidebar collapsed state
		if (window.innerWidth >= 1024 && localStorage.getItem('sidebarCollapsed') === 'true') {
			document.body.classList.add('sidebar-collapsed');
		}
	});

	$: currentPath = $page.url.pathname;
	$: isAuthPage = currentPath.startsWith('/auth');
</script>

<svelte:head>
	<title>{data.title || 'GoSight'}</title>
</svelte:head>

{#if isAuthPage}
	<!-- Auth pages use their own layout -->
	<slot />
{:else}
	<div class="min-h-screen bg-white dark:bg-gray-900">
		<!-- Alert spacer for notifications -->
		<div id="alert-spacer" class="h-0 transition-all duration-300"></div>

		<!-- Top Navbar -->
		<Navbar />

		<!-- Sidebar -->
		<Sidebar {currentPath} />

		<!-- Sidebar backdrop for mobile -->
		<div
			class="fixed inset-0 z-10 hidden bg-gray-900/50 dark:bg-gray-900/90"
			id="sidebarBackdrop"
		></div>

		<!-- Main content -->
		<main class="relative pt-16 lg:pl-64">
			<div class="p-4">
				<slot />
			</div>
		</main>
	</div>
{/if}

<style>
	:global(.sidebar-link::before) {
		content: '';
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 4px;
		background-color: transparent;
		transition: background-color 0.2s ease;
	}

	:global(.sidebar-link:hover::before) {
		background-color: #9ca3af; /* gray-400 */
	}

	:global(.sidebar-link.active::before) {
		background-color: #3b82f6; /* blue-500 */
	}

	:global(.sidebar-link.active:hover::before) {
		background-color: #3b82f6; /* STAY blue on hover */
	}

	:global(.sidebar-link.active) {
		background-color: #ffffff; /* bg-white */
		color: #1f2937; /* text-gray-900 */
	}

	:global(body.sidebar-collapsed #sidebar) {
		width: 4rem;
		overflow: visible;
	}

	:global(body.sidebar-collapsed #sidebar span) {
		display: none;
	}

	:global(body.sidebar-collapsed #sidebar a) {
		justify-content: center;
	}

	:global(body.sidebar-collapsed #sidebar svg) {
		margin-right: 0;
		display: block;
	}

	@media (min-width: 1024px) {
		:global(body.sidebar-collapsed main) {
			padding-left: 4rem !important;
		}
	}

	/* Tooltip setup for collapsed sidebar */
	:global(body.sidebar-collapsed #sidebar a[data-tooltip]) {
		position: relative;
		overflow: visible;
	}

	:global(body.sidebar-collapsed #sidebar a[data-tooltip]::after) {
		content: attr(data-tooltip);
		position: absolute;
		left: 4.5rem;
		top: 50%;
		transform: translateY(-50%);
		background-color: #1f2937;
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		white-space: nowrap;
		z-index: 9999;
		opacity: 0;
		transition: opacity 0.15s ease-in-out;
		pointer-events: none;
	}

	:global(body.sidebar-collapsed #sidebar a[data-tooltip]:hover::after) {
		opacity: 1;
	}

	/* Submenu styling */
	:global(.sidebar-submenu) {
		border-left: 1px solid #e5e7eb; /* gray-200 */
		margin-left: 1.25rem; /* match icon space */
		padding-left: 0.75rem;
	}

	:global(.submenu-link::before) {
		content: '';
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 3px;
		background-color: transparent;
		transition: background-color 0.2s ease;
	}

	:global(.submenu-link:hover::before) {
		background-color: transparent;
	}

	:global(.submenu-link.active::before) {
		background-color: #60a5fa; /* blue-400 */
	}

	:global(.submenu-link:hover) {
		color: #3b82f6;
	}

	:global(.submenu-link.active) {
		background-color: #f1f5f9; /* gray-100 */
		color: #3b82f6;
	}

	/* Hide submenu entirely when collapsed */
	:global(body.sidebar-collapsed .sidebar-submenu) {
		display: none !important;
	}

	/* Scrollbar styling */
	:global(.scrollbar-thin) {
		scrollbar-width: thin;
	}

	:global(.scrollbar-dark) {
		scrollbar-color: #4b5563 #111827; /* thumb / track */
	}

	/* Chrome & Edge */
	:global(.scrollbar-dark::-webkit-scrollbar) {
		width: 6px;
	}

	:global(.scrollbar-dark::-webkit-scrollbar-track) {
		background: #111827; /* dark background */
	}

	:global(.scrollbar-dark::-webkit-scrollbar-thumb) {
		background-color: #4b5563; /* gray-600 */
		border-radius: 10px;
		border: 1px solid #1f2937; /* dark border */
	}

	:global(.scrollbar-dark::-webkit-scrollbar-thumb:hover) {
		background-color: #6b7280; /* gray-500 */
	}

	/* Blinking cursor animation */
	:global(.blink-cursor::after) {
		content: '';
		display: inline-block;
		width: 0.6ch;
		height: 1em;
		margin-left: 2px;
		background-color: #f3f4f6;
		animation: blink 1s steps(2, start) infinite;
		vertical-align: middle;
	}

	@keyframes blink {
		0%,
		100% {
			opacity: 0;
		}
		50% {
			opacity: 1;
		}
	}

	/* Mobile submenu hiding */
	@media (max-width: 1024px) {
		:global(#sidebar.hidden .sidebar-submenu) {
			display: none !important;
		}
	}
</style>
