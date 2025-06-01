<script lang="ts">
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';

	let {
		groups
	}: {
		groups: {
			groupTitle: string;
			items: {
				title: string;
				url: string;
				// this should be `Component` after @lucide/svelte updates types
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				icon?: any;
				isActive?: boolean;
				items?: {
					title: string;
					url: string;
				}[];
			}[];
		}[];
	} = $props();
</script>

{#each groups as group (group.groupTitle)}
	<Sidebar.Group>
		<Sidebar.GroupLabel>{group.groupTitle}</Sidebar.GroupLabel>
		<Sidebar.Menu>
			{#each group.items as item (item.title)}
				{#if item.items && item.items.length > 0}
					<!-- Collapsible item with sub-menu -->
					<Collapsible.Root open={item.isActive} class="group/collapsible">
						<Sidebar.MenuItem>
							<Collapsible.Trigger>
								{#snippet child({ props })}
									<Sidebar.MenuButton {...props} tooltipContent={item.title}>
										{#if item.icon}
											<item.icon />
										{/if}
										<span>{item.title}</span>
										<ChevronRightIcon
											class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
										/>
									</Sidebar.MenuButton>
								{/snippet}
							</Collapsible.Trigger>
						</Sidebar.MenuItem>
						<Collapsible.Content>
							<Sidebar.MenuSub>
								{#each item.items as subItem (subItem.title)}
									<Sidebar.MenuSubItem>
										<Sidebar.MenuSubButton>
											{#snippet child({ props })}
												<a href={subItem.url} {...props}>
													<span>{subItem.title}</span>
												</a>
											{/snippet}
										</Sidebar.MenuSubButton>
									</Sidebar.MenuSubItem>
								{/each}
							</Sidebar.MenuSub>
						</Collapsible.Content>
					</Collapsible.Root>
				{:else}
					<!-- Simple menu item -->
					<Sidebar.MenuItem>
						<Sidebar.MenuButton tooltipContent={item.title} isActive={item.isActive}>
							{#if item.icon}
								<item.icon />
							{/if}
							<span>{item.title}</span>
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/if}
			{/each}
		</Sidebar.Menu>
	</Sidebar.Group>
{/each}
