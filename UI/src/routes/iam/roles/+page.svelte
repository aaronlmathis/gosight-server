<script lang="ts">
	import { onMount } from 'svelte';
	import { rolesApi, permissionsApi } from '$lib/api';
	import type {
		RoleWithPermissions,
		Permission,
		CreateRoleRequest,
		UpdateRoleRequest
	} from '$lib/types';

	let roles: RoleWithPermissions[] = [];
	let allPermissions: Permission[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';

	// Modal states
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let showPermissionsModal = false;
	let selectedRole: RoleWithPermissions | null = null;

	// Form data
	let roleForm = {
		name: '',
		description: ''
	};

	// Permission assignment
	let selectedPermissions: string[] = [];

	// Filtered roles
	$: filteredRoles = roles.filter(
		(role) =>
			role.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
			(role.description && role.description.toLowerCase().includes(searchTerm.toLowerCase()))
	);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		try {
			loading = true;
			error = '';
			const [rolesResponse, permissionsResponse] = await Promise.all([
				rolesApi.getAll(),
				permissionsApi.getAll()
			]);
			// Convert Role[] to RoleWithPermissions[] by adding empty permissions array
			roles = rolesResponse.map((role) => ({
				...role,
				permissions: []
			}));
			allPermissions = permissionsResponse;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		roleForm = { name: '', description: '' };
		showCreateModal = true;
	}

	function openEditModal(role: RoleWithPermissions) {
		selectedRole = role;
		roleForm = {
			name: role.name,
			description: role.description || ''
		};
		showEditModal = true;
	}

	function openDeleteModal(role: RoleWithPermissions) {
		selectedRole = role;
		showDeleteModal = true;
	}

	function openPermissionsModal(role: RoleWithPermissions) {
		selectedRole = role;
		selectedPermissions = role.permissions?.map((p) => p.id) || [];
		showPermissionsModal = true;
	}

	async function createRole() {
		try {
			const request: CreateRoleRequest = {
				name: roleForm.name,
				description: roleForm.description
			};
			await rolesApi.create(request);
			await loadData();
			showCreateModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create role';
		}
	}

	async function updateRole() {
		if (!selectedRole) return;

		try {
			const request: UpdateRoleRequest = {
				name: roleForm.name,
				description: roleForm.description
			};
			await rolesApi.update(selectedRole.id, request);
			await loadData();
			showEditModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update role';
		}
	}

	async function deleteRole() {
		if (!selectedRole) return;

		try {
			await rolesApi.delete(selectedRole.id);
			await loadData();
			showDeleteModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete role';
		}
	}

	async function updateRolePermissions() {
		if (!selectedRole) return;

		try {
			await rolesApi.assignPermissions(selectedRole.id, { permission_ids: selectedPermissions });
			await loadData();
			showPermissionsModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update role permissions';
		}
	}
</script>

<svelte:head>
	<title>Roles Management - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Roles Management</h1>
			<p class="text-gray-600">Manage system roles and their permissions</p>
		</div>
		<button
			on:click={openCreateModal}
			class="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
		>
			<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 6v6m0 0v6m0-6h6m-6 0H6"
				/>
			</svg>
			Create Role
		</button>
	</div>

	{#if error}
		<div class="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
			{error}
		</div>
	{/if}

	<!-- Search -->
	<div class="mb-6">
		<input
			type="text"
			bind:value={searchTerm}
			placeholder="Search roles..."
			class="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-transparent focus:ring-2 focus:ring-blue-500 md:w-96"
		/>
	</div>

	{#if loading}
		<div class="flex h-64 items-center justify-center">
			<div class="h-12 w-12 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else}
		<!-- Roles Table -->
		<div class="overflow-hidden rounded-lg bg-white shadow">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Role
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Description
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Permissions
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Created
						</th>
						<th
							class="px-6 py-3 text-right text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Actions
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white">
					{#each filteredRoles as role}
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-gray-900">{role.name}</div>
							</td>
							<td class="px-6 py-4">
								<div class="text-sm text-gray-900">{role.description}</div>
							</td>
							<td class="px-6 py-4">
								<div class="flex flex-wrap gap-1">
									{#each (role.permissions || []).slice(0, 3) as permission}
										<span
											class="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800"
										>
											{permission.name}
										</span>
									{/each}
									{#if (role.permissions || []).length > 3}
										<span
											class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800"
										>
											+{(role.permissions || []).length - 3} more
										</span>
									{/if}
									{#if (role.permissions || []).length === 0}
										<span class="text-sm text-gray-500">No permissions</span>
									{/if}
								</div>
							</td>
							<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-500">
								{new Date(role.created_at).toLocaleDateString()}
							</td>
							<td class="px-6 py-4 text-right text-sm font-medium whitespace-nowrap">
								<div class="flex justify-end gap-2">
									<button
										on:click={() => openPermissionsModal(role)}
										class="text-blue-600 hover:text-blue-900"
										title="Manage Permissions"
									>
										<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
											/>
										</svg>
									</button>
									<button
										on:click={() => openEditModal(role)}
										class="text-indigo-600 hover:text-indigo-900"
										title="Edit Role"
									>
										<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
											/>
										</svg>
									</button>
									<button
										on:click={() => openDeleteModal(role)}
										class="text-red-600 hover:text-red-900"
										title="Delete Role"
									>
										<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
											/>
										</svg>
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			{#if filteredRoles.length === 0}
				<div class="py-12 text-center">
					<svg
						class="mx-auto h-12 w-12 text-gray-400"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
						/>
					</svg>
					<h3 class="mt-2 text-sm font-medium text-gray-900">No roles found</h3>
					<p class="mt-1 text-sm text-gray-500">
						{searchTerm
							? 'Try adjusting your search criteria.'
							: 'Get started by creating a new role.'}
					</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create Role Modal -->
{#if showCreateModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div class="relative top-20 mx-auto w-96 rounded-md border bg-white p-5 shadow-lg">
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Create New Role</h3>
				<form on:submit|preventDefault={createRole}>
					<div class="mb-4">
						<label for="role-name" class="mb-2 block text-sm font-medium text-gray-700">Name</label>
						<input
							id="role-name"
							type="text"
							bind:value={roleForm.name}
							required
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label for="role-description" class="mb-2 block text-sm font-medium text-gray-700"
							>Description</label
						>
						<textarea
							id="role-description"
							bind:value={roleForm.description}
							rows="3"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						></textarea>
					</div>
					<div class="flex justify-end gap-3">
						<button
							type="button"
							on:click={() => (showCreateModal = false)}
							class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
						>
							Create Role
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Edit Role Modal -->
{#if showEditModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div class="relative top-20 mx-auto w-96 rounded-md border bg-white p-5 shadow-lg">
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Edit Role</h3>
				<form on:submit|preventDefault={updateRole}>
					<div class="mb-4">
						<label for="edit-role-name" class="mb-2 block text-sm font-medium text-gray-700"
							>Name</label
						>
						<input
							id="edit-role-name"
							type="text"
							bind:value={roleForm.name}
							required
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label for="edit-role-description" class="mb-2 block text-sm font-medium text-gray-700"
							>Description</label
						>
						<textarea
							id="edit-role-description"
							bind:value={roleForm.description}
							rows="3"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						></textarea>
					</div>
					<div class="flex justify-end gap-3">
						<button
							type="button"
							on:click={() => (showEditModal = false)}
							class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
						>
							Update Role
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Role Modal -->
{#if showDeleteModal && selectedRole}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div class="relative top-20 mx-auto w-96 rounded-md border bg-white p-5 shadow-lg">
			<div class="mt-3 text-center">
				<div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
					<svg class="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
						/>
					</svg>
				</div>
				<h3 class="mt-2 text-lg font-medium text-gray-900">Delete Role</h3>
				<div class="mt-2 px-7 py-3">
					<p class="text-sm text-gray-500">
						Are you sure you want to delete the role "{selectedRole.name}"? This action cannot be
						undone.
					</p>
				</div>
				<div class="mt-4 flex justify-center gap-3">
					<button
						on:click={() => (showDeleteModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
					>
						Cancel
					</button>
					<button
						on:click={deleteRole}
						class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
					>
						Delete
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Permissions Modal -->
{#if showPermissionsModal && selectedRole}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div
			class="relative top-20 mx-auto max-h-96 w-96 overflow-y-auto rounded-md border bg-white p-5 shadow-lg"
		>
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">
					Manage Permissions for "{selectedRole.name}"
				</h3>
				<div class="mb-4 space-y-2">
					{#each allPermissions as permission}
						<label class="flex items-center">
							<input
								type="checkbox"
								bind:group={selectedPermissions}
								value={permission.id}
								class="focus:ring-opacity-50 rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200"
							/>
							<span class="ml-2 text-sm text-gray-900">{permission.name}</span>
							{#if permission.description}
								<span class="ml-2 text-xs text-gray-500">- {permission.description}</span>
							{/if}
						</label>
					{/each}
				</div>
				<div class="flex justify-end gap-3">
					<button
						on:click={() => (showPermissionsModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
					>
						Cancel
					</button>
					<button
						on:click={updateRolePermissions}
						class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					>
						Update Permissions
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
