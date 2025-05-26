<script lang="ts">
	import type { Event } from '$lib/types';
	export let events: Event[];
</script>

<div class="p-4" id="activity" role="tabpanel" aria-labelledby="activity-tab">
	<!-- Activity Tab: events list -->
	<h3 class="mb-4 text-lg font-medium">Activity</h3>
	<ul class="space-y-2 text-sm text-gray-700 dark:text-gray-300">
		{#each events as e}
			<li class="border-b border-gray-200 py-2 dark:border-gray-700">
				<div><strong>{e.type}</strong> at {new Date(e.timestamp).toLocaleString()}</div>
				<div class="text-gray-500 dark:text-gray-400">{e.message}</div>
			</li>
		{/each}
	</ul>
</div>
