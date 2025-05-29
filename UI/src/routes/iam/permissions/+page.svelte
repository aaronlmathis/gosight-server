<script lang="ts">
	import { onMount } from 'svelte';
	import { permissionsApi, rolesApi } from '$lib/api';
	import type {
		PermissionWithRoles,
		Role,
		CreatePermissionRequest,
		UpdatePermissionRequest
	} from '$lib/types';

	let permissions: PermissionWithRoles[] = [];
	let allRoles: Role[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';

	// Modal states
	let showCreateModal = false;
	let showEditModal = false;
	let showDeleteModal = false;
	let showRolesModal = false;
	let selectedPermission: PermissionWithRoles | null = null;

	// Form data
	let permissionForm = {
		name: '',
		description: '',
		resource: '',
		action: ''
	};

	// Role assignment
	let selectedRoles: string[] = [];

	// Filtered permissions
	$: filteredPermissions = permissions.filter(
		(permission) =>
			permission.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
			(permission.description &&
				permission.description.toLowerCase().includes(searchTerm.toLowerCase())) ||
			permission.resource.toLowerCase().includes(searchTerm.toLowerCase()) ||
			permission.action.toLowerCase().includes(searchTerm.toLowerCase())
	);

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		try {
			loading = true;
			error = '';
			const [permissionsResponse, rolesResponse] = await Promise.all([
				permissionsApi.getAll(),
				rolesApi.getAll()
			]);
			// Convert Permission[] to PermissionWithRoles[] by adding empty roles array
			permissions = permissionsResponse.map((permission) => ({
				...permission,
				roles: []
			}));
			allRoles = rolesResponse;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		permissionForm = { name: '', description: '', resource: '', action: '' };
		showCreateModal = true;
	}

	function openEditModal(permission: PermissionWithRoles) {
		selectedPermission = permission;
		permissionForm = {
			name: permission.name,
			description: permission.description || '',
			resource: permission.resource,
			action: permission.action
		};
		showEditModal = true;
	}

	function openDeleteModal(permission: PermissionWithRoles) {
		selectedPermission = permission;
		showDeleteModal = true;
	}

	function openRolesModal(permission: PermissionWithRoles) {
		selectedPermission = permission;
		selectedRoles = permission.roles?.map((r) => r.id) || [];
		showRolesModal = true;
	}

	async function createPermission() {
		try {
			const request: CreatePermissionRequest = {
				name: permissionForm.name,
				description: permissionForm.description,
				resource: permissionForm.resource,
				action: permissionForm.action
			};
			await permissionsApi.create(request);
			await loadData();
			showCreateModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create permission';
		}
	}

	async function updatePermission() {
		if (!selectedPermission) return;

		try {
			const request: UpdatePermissionRequest = {
				name: permissionForm.name,
				description: permissionForm.description,
				resource: permissionForm.resource,
				action: permissionForm.action
			};
			await permissionsApi.update(selectedPermission.id, request);
			await loadData();
			showEditModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update permission';
		}
	}

	async function deletePermission() {
		if (!selectedPermission) return;

		try {
			await permissionsApi.delete(selectedPermission.id);
			await loadData();
			showDeleteModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete permission';
		}
	}

	async function updatePermissionRoles() {
		if (!selectedPermission) return;

		try {
			// Since we can't directly assign roles to permissions, we need to update roles
			// to have or remove this permission. This is a simplified approach.
			// In a real implementation, you might want to show the current state
			// and let users manage it through the roles page instead.

			// For now, we'll just reload data and close the modal
			// The actual role-permission assignment should be done from the roles page
			await loadData();
			showRolesModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update permission roles';
		}
	}
</script>

<svelte:head>
	<title>Permissions Management - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Permissions Management</h1>
			<p class="text-gray-600">Manage system permissions and their role assignments</p>
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
			Create Permission
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
			placeholder="Search permissions..."
			class="w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-transparent focus:ring-2 focus:ring-blue-500 md:w-96"
		/>
	</div>

	{#if loading}
		<div class="flex h-64 items-center justify-center">
			<div class="h-12 w-12 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else}
		<!-- Permissions Table -->
		<div class="overflow-hidden rounded-lg bg-white shadow">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Permission
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Description
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Resource
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Action
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>
							Roles
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
					{#each filteredPermissions as permission}
						<tr class="hover:bg-gray-50">
							<td class="px-6 py-4 whitespace-nowrap">
								<div class="text-sm font-medium text-gray-900">{permission.name}</div>
							</td>
							<td class="px-6 py-4">
								<div class="text-sm text-gray-900">{permission.description}</div>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span
									class="inline-flex items-center rounded-full bg-purple-100 px-2.5 py-0.5 text-xs font-medium text-purple-800"
								>
									{permission.resource}
								</span>
							</td>
							<td class="px-6 py-4 whitespace-nowrap">
								<span
									class="inline-flex items-center rounded-full bg-orange-100 px-2.5 py-0.5 text-xs font-medium text-orange-800"
								>
									{permission.action}
								</span>
							</td>
							<td class="px-6 py-4">
								<div class="flex flex-wrap gap-1">
									{#each (permission.roles || []).slice(0, 3) as role}
										<span
											class="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800"
										>
											{role.name}
										</span>
									{/each}
									{#if (permission.roles || []).length > 3}
										<span
											class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800"
										>
											+{(permission.roles || []).length - 3} more
										</span>
									{/if}
									{#if (permission.roles || []).length === 0}
										<span class="text-sm text-gray-500">No roles</span>
									{/if}
								</div>
							</td>
							<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-500">
								{new Date(permission.created_at).toLocaleDateString()}
							</td>
							<td class="px-6 py-4 text-right text-sm font-medium whitespace-nowrap">
								<div class="flex justify-end gap-2">
									<button
										on:click={() => openRolesModal(permission)}
										class="text-blue-600 hover:text-blue-900"
										title="Manage Roles"
									>
										<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
											/>
										</svg>
									</button>
									<button
										on:click={() => openEditModal(permission)}
										class="text-indigo-600 hover:text-indigo-900"
										title="Edit Permission"
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
										on:click={() => openDeleteModal(permission)}
										class="text-red-600 hover:text-red-900"
										title="Delete Permission"
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

			{#if filteredPermissions.length === 0}
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
					<h3 class="mt-2 text-sm font-medium text-gray-900">No permissions found</h3>
					<p class="mt-1 text-sm text-gray-500">
						{searchTerm
							? 'Try adjusting your search criteria.'
							: 'Get started by creating a new permission.'}
					</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create Permission Modal -->
{#if showCreateModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div class="relative top-20 mx-auto w-96 rounded-md border bg-white p-5 shadow-lg">
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Create New Permission</h3>
				<form on:submit|preventDefault={createPermission}>
					<div class="mb-4">
						<label for="create-permission-name" class="mb-2 block text-sm font-medium text-gray-700"
							>Name</label
						>
						<input
							id="create-permission-name"
							type="text"
							bind:value={permissionForm.name}
							required
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label
							for="create-permission-description"
							class="mb-2 block text-sm font-medium text-gray-700">Description</label
						>
						<textarea
							id="create-permission-description"
							bind:value={permissionForm.description}
							rows="3"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						></textarea>
					</div>
					<div class="mb-4">
						<label
							for="create-permission-resource"
							class="mb-2 block text-sm font-medium text-gray-700">Resource</label
						>
						<input
							id="create-permission-resource"
							type="text"
							bind:value={permissionForm.resource}
							required
							placeholder="e.g., users, reports, settings"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label
							for="create-permission-action"
							class="mb-2 block text-sm font-medium text-gray-700">Action</label
						>
						<input
							id="create-permission-action"
							type="text"
							bind:value={permissionForm.action}
							required
							placeholder="e.g., read, write, delete, manage"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
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
							Create Permission
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Edit Permission Modal -->
{#if showEditModal}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div class="relative top-20 mx-auto w-96 rounded-md border bg-white p-5 shadow-lg">
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Edit Permission</h3>
				<form on:submit|preventDefault={updatePermission}>
					<div class="mb-4">
						<label for="edit-permission-name" class="mb-2 block text-sm font-medium text-gray-700"
							>Name</label
						>
						<input
							id="edit-permission-name"
							type="text"
							bind:value={permissionForm.name}
							required
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label
							for="edit-permission-description"
							class="mb-2 block text-sm font-medium text-gray-700">Description</label
						>
						<textarea
							id="edit-permission-description"
							bind:value={permissionForm.description}
							rows="3"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						></textarea>
					</div>
					<div class="mb-4">
						<label
							for="edit-permission-resource"
							class="mb-2 block text-sm font-medium text-gray-700">Resource</label
						>
						<input
							id="edit-permission-resource"
							type="text"
							bind:value={permissionForm.resource}
							required
							placeholder="e.g., users, reports, settings"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
					</div>
					<div class="mb-4">
						<label for="edit-permission-action" class="mb-2 block text-sm font-medium text-gray-700"
							>Action</label
						>
						<input
							id="edit-permission-action"
							type="text"
							bind:value={permissionForm.action}
							required
							placeholder="e.g., read, write, delete, manage"
							class="w-full rounded-md border border-gray-300 px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:outline-none"
						/>
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
							Update Permission
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Permission Modal -->
{#if showDeleteModal && selectedPermission}
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
				<h3 class="mt-2 text-lg font-medium text-gray-900">Delete Permission</h3>
				<div class="mt-2 px-7 py-3">
					<p class="text-sm text-gray-500">
						Are you sure you want to delete the permission "{selectedPermission.name}"? This action
						cannot be undone.
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
						on:click={deletePermission}
						class="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
					>
						Delete
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Roles Modal -->
{#if showRolesModal && selectedPermission}
	<div class="bg-opacity-50 fixed inset-0 z-50 h-full w-full overflow-y-auto bg-gray-600">
		<div
			class="relative top-20 mx-auto max-h-96 w-96 overflow-y-auto rounded-md border bg-white p-5 shadow-lg"
		>
			<div class="mt-3">
				<h3 class="mb-4 text-lg font-medium text-gray-900">
					Manage Roles for "{selectedPermission.name}"
				</h3>
				<div class="mb-4 space-y-2">
					{#each allRoles as role}
						<label class="flex items-center">
							<input
								type="checkbox"
								bind:group={selectedRoles}
								value={role.id}
								class="focus:ring-opacity-50 rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200"
							/>
							<span class="ml-2 text-sm text-gray-900">{role.name}</span>
							{#if role.description}
								<span class="ml-2 text-xs text-gray-500">- {role.description}</span>
							{/if}
						</label>
					{/each}
				</div>
				<div class="flex justify-end gap-3">
					<button
						on:click={() => (showRolesModal = false)}
						class="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
					>
						Cancel
					</button>
					<button
						on:click={updatePermissionRoles}
						class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					>
						Update Roles
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
