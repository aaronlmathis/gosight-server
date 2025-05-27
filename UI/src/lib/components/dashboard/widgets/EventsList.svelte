<script lang="ts">
	import type { Widget, WidgetData, Event } from '$lib/types/dashboard';
	import { Card, Badge } from 'flowbite-svelte';
	import { Activity, AlertCircle, CheckCircle, Info, Settings, User } from 'lucide-svelte';
	import { onMount, onDestroy } from 'svelte';
	import { dataService } from '$lib/services/dataService';

	export let widget: Widget;

	let events: Event[] = [];
	let loading = true;
	let error = '';
	let unsubscribe: (() => void) | null = null;

	const categoryConfig = {
		system: {
			icon: Settings,
			color: 'blue',
			bgColor: 'bg-blue-50 dark:bg-blue-900/20',
			textColor: 'text-blue-800 dark:text-blue-300',
			iconColor: 'text-blue-600 dark:text-blue-400'
		},
		security: {
			icon: AlertCircle,
			color: 'red',
			bgColor: 'bg-red-50 dark:bg-red-900/20',
			textColor: 'text-red-800 dark:text-red-300',
			iconColor: 'text-red-600 dark:text-red-400'
		},
		deployment: {
			icon: CheckCircle,
			color: 'green',
			bgColor: 'bg-green-50 dark:bg-green-900/20',
			textColor: 'text-green-800 dark:text-green-300',
			iconColor: 'text-green-600 dark:text-green-400'
		},
		user: {
			icon: User,
			color: 'purple',
			bgColor: 'bg-purple-50 dark:bg-purple-900/20',
			textColor: 'text-purple-800 dark:text-purple-300',
			iconColor: 'text-purple-600 dark:text-purple-400'
		},
		application: {
			icon: Activity,
			color: 'yellow',
			bgColor: 'bg-yellow-50 dark:bg-yellow-900/20',
			textColor: 'text-yellow-800 dark:text-yellow-300',
			iconColor: 'text-yellow-600 dark:text-yellow-400'
		},
		default: {
			icon: Info,
			color: 'gray',
			bgColor: 'bg-gray-50 dark:bg-gray-900/20',
			textColor: 'text-gray-800 dark:text-gray-300',
			iconColor: 'text-gray-600 dark:text-gray-400'
		}
	};

	// Load events data
	async function loadEventsData() {
		try {
			loading = true;
			error = '';

			const data = await dataService.getWidgetData(widget);

			if (data.status === 'error') {
				error = data.error || 'Failed to load events';
				return;
			}

			if (data.events && Array.isArray(data.events)) {
				events = data.events;
			}
		} catch (err) {
			console.error('Failed to load events:', err);
			error = err instanceof Error ? err.message : 'Failed to load events';
		} finally {
			loading = false;
		}
	}

	// Setup real-time events subscription
	function setupRealTimeSubscription() {
		// Subscribe to live events store
		unsubscribe = dataService.liveEvents.subscribe((liveEvents) => {
			if (liveEvents.length > 0) {
				// Apply widget filters if configured
				const filteredEvents = applyFilters(liveEvents);
				events = filteredEvents.slice(0, widget.config?.limit || 10);
			}
		});
	}

	// Apply widget-specific filters
	function applyFilters(allEvents: Event[]): Event[] {
		const { category, endpointId, type } = widget.config || {};

		return allEvents.filter((event) => {
			if (category && event.category !== category) return false;
			if (endpointId && event.endpoint_id !== endpointId) return false;
			if (type && event.type !== type) return false;
			return true;
		});
	}

	function formatTimestamp(date: Date | string): string {
		const eventDate = typeof date === 'string' ? new Date(date) : date;
		const now = new Date();
		const diff = now.getTime() - eventDate.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(diff / 3600000);

		if (minutes < 60) {
			return `${minutes}m ago`;
		} else if (hours < 24) {
			return `${hours}h ago`;
		} else {
			return eventDate.toLocaleDateString();
		}
	}

	function getEventConfig(event: Event) {
		return categoryConfig[event.category] || categoryConfig.default;
	}

	onMount(async () => {
		await loadEventsData();
		setupRealTimeSubscription();
	});

	onDestroy(() => {
		if (unsubscribe) {
			unsubscribe();
		}
	});
</script>

<Card class="relative h-full">
	<div class="flex h-full flex-col">
		<!-- Header -->
		{#if widget.config?.showTitle !== false}
			<div
				class="mb-3 flex items-center justify-between text-sm font-medium text-gray-900 dark:text-gray-100"
			>
				<span>{widget.title}</span>
				<div class="flex items-center gap-2">
					{#if loading}
						<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-500"></div>
					{/if}
					<Badge color="gray" class="text-xs">{events.length}</Badge>
				</div>
			</div>
		{/if}

		<!-- Error State -->
		{#if error}
			<div class="flex flex-1 items-center justify-center">
				<div class="text-center">
					<div class="mb-2 text-sm text-red-500 dark:text-red-400">
						<AlertCircle class="mx-auto mb-2 h-8 w-8" />
						Failed to load events
					</div>
					<div class="text-xs text-gray-500 dark:text-gray-400">{error}</div>
				</div>
			</div>
		{:else if loading}
			<!-- Loading State -->
			<div class="flex flex-1 items-center justify-center">
				<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-500"></div>
			</div>
		{:else if events.length === 0}
			<!-- Empty State -->
			<div class="flex flex-1 items-center justify-center">
				<div class="text-center">
					<Activity class="mx-auto mb-2 h-8 w-8 text-gray-400" />
					<div class="text-sm text-gray-500 dark:text-gray-400">No events</div>
				</div>
			</div>
		{:else}
			<!-- Events List -->
			<div class="flex-1 space-y-2 overflow-y-auto">
				{#each events as event}
					{@const config = getEventConfig(event)}
					<div
						class="rounded-lg border border-gray-200 p-2 dark:border-gray-700 {config.bgColor} transition-shadow hover:shadow-sm"
					>
						<div class="flex items-start gap-2">
							<svelte:component
								this={config.icon}
								class="mt-0.5 h-4 w-4 {config.iconColor} flex-shrink-0"
							/>
							<div class="min-w-0 flex-1">
								<div class="flex items-center justify-between gap-2">
									<h4 class="text-sm font-medium {config.textColor} truncate">
										{event.title || event.name || 'Event'}
									</h4>
									<Badge color={config.color} class="flex-shrink-0 text-xs">
										{event.category || 'event'}
									</Badge>
								</div>
								<p class="mt-1 line-clamp-2 text-xs text-gray-600 dark:text-gray-300">
									{event.description || event.message || 'No description available'}
								</p>
								<div
									class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400"
								>
									<span>{event.source || event.endpoint_id || 'Unknown source'}</span>
									<span>{formatTimestamp(event.timestamp || event.created_at || new Date())}</span>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Footer -->
			<div class="mt-3 border-t border-gray-200 pt-2 dark:border-gray-700">
				<button
					class="text-xs font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
				>
					View all events →
				</button>
			</div>
		{/if}
	</div>
</Card>
