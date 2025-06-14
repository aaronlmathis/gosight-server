<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import { auth } from '$lib/stores/authStore';
	import { goto } from '$app/navigation';
	import BadgeCheckIcon from 'lucide-svelte/icons/badge-check';
	import BellIcon from 'lucide-svelte/icons/bell';
	import ChevronsUpDownIcon from 'lucide-svelte/icons/chevrons-up-down';
	import SettingsIcon from 'lucide-svelte/icons/settings';
	import UserIcon from 'lucide-svelte/icons/user';
	import LogOutIcon from 'lucide-svelte/icons/log-out';

	const sidebar = useSidebar();

	// Add debugging
	$: {
		console.log('Auth debug:', {
			isAuthenticated: $auth.isAuthenticated,
			user: $auth.user,
			avatarData: avatarData
		});
	}

	// Helper function to get user initials
	function getUserInitials(firstName?: string, lastName?: string): string {
		const first = firstName?.charAt(0)?.toUpperCase() || '';
		const last = lastName?.charAt(0)?.toUpperCase() || '';
		return first + last || 'U';
	}

	// Helper function to get display name
	function getDisplayName(user: any): string {
		if (user.firstName && user.lastName) {
			return `${user.firstName} ${user.lastName}`;
		}
		if (user.first_name && user.last_name) {
			return `${user.first_name} ${user.last_name}`;
		}
		return user.username || user.email || 'User';
	}

	// Helper function to get avatar URL or fallback
	function getAvatarData(user: any) {
		console.log('Getting avatar data for user:', user); // Debug log

		const avatarUrl = user.profile?.avatar_url || user.avatar_url;
		const displayName = getDisplayName(user);
		const initials = getUserInitials(
			user.firstName || user.first_name,
			user.lastName || user.last_name
		);

		const result = {
			url: avatarUrl,
			name: displayName,
			initials,
			email: user.email
		};

		console.log('Avatar data result:', result); // Debug log
		return result;
	}

	// Handle logout
	function handleLogout() {
		auth.logout();
		goto('/auth/login');
	}

	// Handle navigation
	function handleNavigation(path: string) {
		goto(path);
	}

	// Reactive avatar data
	$: avatarData = $auth.user ? getAvatarData($auth.user) : null;
</script>

{#if $auth.isAuthenticated && $auth.user && avatarData}
	<Sidebar.Menu>
		<Sidebar.MenuItem>
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Sidebar.MenuButton
							size="lg"
							class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
							{...props}
						>
							<Avatar.Root class="size-8 rounded-lg">
								{#if avatarData.url}
									<Avatar.Image src={avatarData.url} alt={avatarData.name} />
								{/if}
								<Avatar.Fallback class="bg-primary/10 text-primary rounded-lg font-medium">
									{avatarData.initials}
								</Avatar.Fallback>
							</Avatar.Root>
							<div class="grid flex-1 text-left text-sm leading-tight">
								<span class="truncate font-medium">{avatarData.name}</span>
								<span class="text-muted-foreground truncate text-xs">{avatarData.email}</span>
							</div>
							<ChevronsUpDownIcon class="ml-auto size-4" />
						</Sidebar.MenuButton>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content
					class="w-[--bits-dropdown-menu-anchor-width] min-w-56 rounded-lg"
					side={sidebar.isMobile ? 'bottom' : 'right'}
					align="end"
					sideOffset={4}
				>
					<DropdownMenu.Label class="p-0 font-normal">
						<div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
							<Avatar.Root class="size-8 rounded-lg">
								{#if avatarData.url}
									<Avatar.Image src={avatarData.url} alt={avatarData.name} />
								{/if}
								<Avatar.Fallback class="bg-primary/10 text-primary rounded-lg font-medium">
									{avatarData.initials}
								</Avatar.Fallback>
							</Avatar.Root>
							<div class="grid flex-1 text-left text-sm leading-tight">
								<span class="truncate font-medium">{avatarData.name}</span>
								<span class="text-muted-foreground truncate text-xs">{avatarData.email}</span>
							</div>
						</div>
					</DropdownMenu.Label>
					<DropdownMenu.Separator />
					<DropdownMenu.Group>
						<DropdownMenu.Item onclick={() => handleNavigation('/settings/profile')}>
							<UserIcon class="size-4" />
							Profile
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleNavigation('/settings/security')}>
							<SettingsIcon class="size-4" />
							Security
						</DropdownMenu.Item>

						<DropdownMenu.Item onclick={() => handleNavigation('/settings/preferences')}>
							<BellIcon class="size-4" />
							Preferences
						</DropdownMenu.Item>
					</DropdownMenu.Group>
					<DropdownMenu.Separator />
					<DropdownMenu.Item onclick={handleLogout} class="text-destructive focus:text-destructive">
						<LogOutIcon class="size-4" />
						Log out
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</Sidebar.MenuItem>
	</Sidebar.Menu>
{/if}
