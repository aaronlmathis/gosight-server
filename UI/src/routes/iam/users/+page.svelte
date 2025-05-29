<!-- Users Management Page -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { UserWithRoles, Role } from '$lib/types';
	import { Plus, Search, Edit, Trash2, UserPlus, Shield } from 'lucide-svelte';

	let users: UserWithRoles[] = [];
	let allRoles: Role[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let showCreateModal = false;
	let showEditModal = false;
	let showRoleModal = false;
	let selectedUser: UserWithRoles | null = null;

	// Filter users based on search term
	$: filteredUsers = users.filter(
		(user) =>
			user.email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
			user.username?.toLowerCase().includes(searchTerm.toLowerCase()) ||
			user.first_name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
			user.last_name?.toLowerCase().includes(searchTerm.toLowerCase())
	);

	// Form data
	let newUser = {
		email: '',
		username: '',
		first_name: '',
		last_name: '',
		password: ''
	};

	let editUser = {
		id: '',
		email: '',
		username: '',
		first_name: '',
		last_name: ''
	};

	let selectedRoleIds: string[] = [];

	async function loadUsers() {
		try {
			loading = true;
			users = await api.users.getAll();
		} catch (err) {
			error = `Failed to load users: ${err instanceof Error ? err.message : 'Unknown error'}`;
		} finally {
			loading = false;
		}
	}

	async function loadRoles() {
		try {
			allRoles = await api.roles.getAll();
		} catch (err) {
			console.error('Failed to load roles:', err);
		}
	}

	async function createUser() {
		try {
			await api.users.create(newUser);
			await loadUsers();
			showCreateModal = false;
			resetNewUser();
		} catch (err) {
			error = `Failed to create user: ${err instanceof Error ? err.message : 'Unknown error'}`;
		}
	}

	async function updateUser() {
		try {
			await api.users.update(editUser.id, editUser);
			await loadUsers();
			showEditModal = false;
		} catch (err) {
			error = `Failed to update user: ${err instanceof Error ? err.message : 'Unknown error'}`;
		}
	}

	async function deleteUser(userId: string) {
		if (!confirm('Are you sure you want to delete this user?')) return;

		try {
			await api.users.delete(userId);
			await loadUsers();
		} catch (err) {
			error = `Failed to delete user: ${err instanceof Error ? err.message : 'Unknown error'}`;
		}
	}

	async function assignRoles() {
		if (!selectedUser) return;

		try {
			await api.users.assignRoles(selectedUser.id, { role_ids: selectedRoleIds });
			await loadUsers();
			showRoleModal = false;
		} catch (err) {
			error = `Failed to assign roles: ${err instanceof Error ? err.message : 'Unknown error'}`;
		}
	}

	function openEditModal(user: UserWithRoles) {
		editUser = {
			id: user.id,
			email: user.email || '',
			username: user.username || '',
			first_name: user.first_name || user.firstName || '',
			last_name: user.last_name || user.lastName || ''
		};
		showEditModal = true;
	}

	function openRoleModal(user: UserWithRoles) {
		selectedUser = user;
		selectedRoleIds = user.roles?.map((r) => r.id) || [];
		showRoleModal = true;
	}

	function resetNewUser() {
		newUser = {
			email: '',
			username: '',
			first_name: '',
			last_name: '',
			password: ''
		};
	}

	onMount(() => {
		loadUsers();
		loadRoles();
	});
</script>

<svelte:head>
	<title>Users - GoSight Admin</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6">
		<div class="mb-4 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Users</h1>
				<p class="text-gray-600 dark:text-gray-400">Manage user accounts and role assignments</p>
			</div>
			<button
				on:click={() => (showCreateModal = true)}
				class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
			>
				<Plus class="h-4 w-4" />
				Create User
			</button>
		</div>

		<!-- Search Bar -->
		<div class="relative">
			<Search class="absolute top-1/2 left-3 h-5 w-5 -translate-y-1/2 transform text-gray-400" />
			<input
				bind:value={searchTerm}
				type="text"
				placeholder="Search users..."
				class="w-full rounded-lg border border-gray-300 bg-white py-2 pr-4 pl-10 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
			/>
		</div>
	</div>

	{#if error}
		<div
			class="mb-6 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400"
		>
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else}
		<!-- Users Table -->
		<div class="overflow-hidden rounded-lg bg-white shadow dark:bg-gray-800">
			<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
				<thead class="bg-gray-50 dark:bg-gray-700">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
						>
							User
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
						>
							Email
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
						>
							Roles
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
						>
							Created
						</th>
						<th
							class="px-6 py-3 text-right text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
						>
							Actions
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-800">
					{#each filteredUsers as user}
						<tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex items-center">
									<div class="h-10 w-10 flex-shrink-0">
										<div
											class="flex h-10 w-10 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900"
										>
											<span class="text-sm font-medium text-blue-600 dark:text-blue-400">
												{(user.first_name || user.firstName || user.username || user.email || '')
													.charAt(0)
													.toUpperCase()}
											</span>
										</div>
									</div>
									<div class="ml-4">
										<div class="text-sm font-medium text-gray-900 dark:text-white">
											{user.first_name || user.firstName || ''}
											{user.last_name || user.lastName || ''}
										</div>
										<div class="text-sm text-gray-500 dark:text-gray-400">
											@{user.username || user.email}
										</div>
									</div>
								</div>
							</td>
							<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-900 dark:text-white">
								{user.email || ''}
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="flex flex-wrap gap-1">
									{#each user.roles || [] as role}
										<span
											class="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-200"
										>
											{role.name}
										</span>
									{/each}
								</div>
							</td>
							<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-500 dark:text-gray-400">
								{user.created_at ? new Date(user.created_at).toLocaleDateString() : ''}
							</td>
							<td class="px-6 py-4 text-right text-sm font-medium whitespace-nowrap">
								<div class="flex justify-end gap-2">
									<button
										on:click={() => openRoleModal(user)}
										class="p-1 text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
										title="Manage Roles"
									>
										<Shield class="h-4 w-4" />
									</button>
									<button
										on:click={() => openEditModal(user)}
										class="p-1 text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-300"
										title="Edit User"
									>
										<Edit class="h-4 w-4" />
									</button>
									<button
										on:click={() => deleteUser(user.id)}
										class="p-1 text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
										title="Delete User"
									>
										<Trash2 class="h-4 w-4" />
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			{#if filteredUsers.length === 0}
				<div class="py-12 text-center">
					<UserPlus class="mx-auto mb-4 h-12 w-12 text-gray-400" />
					<h3 class="text-sm font-medium text-gray-900 dark:text-white">No users found</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">
						{searchTerm
							? 'Try adjusting your search terms.'
							: 'Get started by creating a new user.'}
					</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create User Modal -->
{#if showCreateModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 flex items-center justify-center bg-black p-4">
		<div class="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-gray-800">
			<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">Create New User</h3>
			</div>
			<form on:submit|preventDefault={createUser} class="px-6 py-4">
				<div class="space-y-4">
					<div>
						<label
							for="email"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Email *
						</label>
						<input
							id="email"
							bind:value={newUser.email}
							type="email"
							required
							class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>
					<div>
						<label
							for="username"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Username *
						</label>
						<input
							id="username"
							bind:value={newUser.username}
							type="text"
							required
							class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label
								for="first_name"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								First Name
							</label>
							<input
								id="first_name"
								bind:value={newUser.first_name}
								type="text"
								class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
						<div>
							<label
								for="last_name"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								Last Name
							</label>
							<input
								id="last_name"
								bind:value={newUser.last_name}
								type="text"
								class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
					</div>
					<div>
						<label
							for="password"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Password *
						</label>
						<input
							id="password"
							bind:value={newUser.password}
							type="password"
							required
							class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>
				</div>
				<div class="mt-6 flex justify-end gap-3">
					<button
						type="button"
						on:click={() => (showCreateModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
					>
						Create User
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Edit User Modal -->
{#if showEditModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 flex items-center justify-center bg-black p-4">
		<div class="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-gray-800">
			<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">Edit User</h3>
			</div>
			<form on:submit|preventDefault={updateUser} class="px-6 py-4">
				<div class="space-y-4">
					<div>
						<label
							for="edit_email"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Email
						</label>
						<input
							id="edit_email"
							bind:value={editUser.email}
							type="email"
							class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>
					<div>
						<label
							for="edit_username"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Username
						</label>
						<input
							id="edit_username"
							bind:value={editUser.username}
							type="text"
							class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label
								for="edit_first_name"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								First Name
							</label>
							<input
								id="edit_first_name"
								bind:value={editUser.first_name}
								type="text"
								class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
						<div>
							<label
								for="edit_last_name"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								Last Name
							</label>
							<input
								id="edit_last_name"
								bind:value={editUser.last_name}
								type="text"
								class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
					</div>
				</div>
				<div class="mt-6 flex justify-end gap-3">
					<button
						type="button"
						on:click={() => (showEditModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
					>
						Update User
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Role Assignment Modal -->
{#if showRoleModal && selectedUser}
	<div class="bg-opacity-50 fixed inset-0 z-50 flex items-center justify-center bg-black p-4">
		<div class="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-gray-800">
			<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">
					Manage Roles - {selectedUser.username || selectedUser.email}
				</h3>
			</div>
			<form on:submit|preventDefault={assignRoles} class="px-6 py-4">
				<div class="space-y-3">
					{#each allRoles as role}
						<label class="flex items-center">
							<input
								type="checkbox"
								bind:group={selectedRoleIds}
								value={role.id}
								class="rounded border-gray-300 text-blue-600 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700"
							/>
							<span class="ml-3 text-sm text-gray-900 dark:text-white">
								{role.name}
								{#if role.description}
									<span class="block text-xs text-gray-500 dark:text-gray-400">
										{role.description}
									</span>
								{/if}
							</span>
						</label>
					{/each}
				</div>
				<div class="mt-6 flex justify-end gap-3">
					<button
						type="button"
						on:click={() => (showRoleModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
					>
						Update Roles
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
