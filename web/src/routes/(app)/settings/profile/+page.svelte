<script lang="ts">
	import { onMount } from 'svelte';
	import { superForm } from 'sveltekit-superforms/client';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Save } from 'lucide-svelte';
	import { profileSchema } from '../schema';
	import { api } from '$lib/api/api';
	import AvatarUpload from '$lib/components/AvatarUpload.svelte';
	import { auth } from '$lib/stores/authStore';

	let loading = true;
	let successMessage = '';
	let currentUser: any = null;

	// Initialize superform with empty data
	const profile = superForm(
		{ full_name: '', phone: '' },
		{
			validators: zodClient(profileSchema),
			onUpdated: async ({ form }) => {
				if (form.valid) {
					try {
						const result = await api.updateProfile(form.data);
						if (result.success) {
							successMessage = 'Profile updated successfully!';
							setTimeout(() => (successMessage = ''), 3000);
						}
					} catch (error) {
						console.error('Failed to update profile:', error);
					}
				}
			}
		}
	);

	const { form, enhance, submitting } = profile;

	$: formData = $form;

	// Handle avatar upload
	function handleAvatarUploaded(event: CustomEvent<{ avatar_url: string }>) {
		// Update the current user's avatar in the auth store
		if ($auth.user) {
			const updatedUser = {
				...$auth.user,
				profile: {
					...($auth.user.profile || {}),
					avatar_url: event.detail.avatar_url
				}
			};
			auth.setUser(updatedUser);
			currentUser = updatedUser;
		}
		successMessage = 'Profile picture updated successfully!';
		setTimeout(() => (successMessage = ''), 3000);
	}

	function handleAvatarDeleted() {
		// Remove avatar from the current user in the auth store
		if ($auth.user) {
			const updatedUser = {
				...$auth.user,
				profile: {
					...($auth.user.profile || {}),
					avatar_url: ''
				}
			};
			auth.setUser(updatedUser);
			currentUser = updatedUser;
		}
		successMessage = 'Profile picture removed successfully!';
		setTimeout(() => (successMessage = ''), 3000);
	}

	// Load current user data on mount
	onMount(async () => {
		try {
			const user = await api.getCurrentUser();
			currentUser = user;
			if (user?.profile) {
				form.set({
					full_name: user.profile.full_name || '',
					phone: user.profile.phone || ''
				});
			}
		} catch (error) {
			console.error('Failed to load user profile:', error);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Profile Settings</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center p-8">
		<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-gray-900"></div>
	</div>
{:else}
	<div class="space-y-6">
		<div>
			<h3 class="text-lg font-medium">Profile</h3>
			<p class="text-muted-foreground text-sm">This is how others will see you on the site.</p>
		</div>

		{#if successMessage}
			<div class="rounded-md bg-green-50 p-4">
				<p class="text-sm font-medium text-green-800">{successMessage}</p>
			</div>
		{/if}

		<div class="rounded-lg bg-white shadow dark:bg-gray-800">
			<div class="px-4 py-5 sm:p-6">
				<form use:enhance class="space-y-6">
					<!-- Profile Picture Section -->
					<div class="flex items-center space-x-6">
						<div>
							<AvatarUpload
								currentAvatar={currentUser?.profile?.avatar_url || ''}
								size="lg"
								on:uploaded={handleAvatarUploaded}
								on:deleted={handleAvatarDeleted}
							/>
						</div>
						<div>
							<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Profile Picture</h4>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								Click on your picture to update or remove it.
							</p>
						</div>
					</div>

					<Form.Field form={profile} name="full_name">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Full Name</Form.Label>

								<Input
									{...props}
									bind:value={formData.full_name}
									placeholder="Enter your full name"
								/>
							{/snippet}
						</Form.Control>
						<Form.Description>This is your display name.</Form.Description>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field form={profile} name="phone">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Phone Number</Form.Label>

								<Input {...props} bind:value={formData.phone} placeholder="Enter your phone" />
							{/snippet}
						</Form.Control>
						<Form.Description>Optional</Form.Description>
						<Form.FieldErrors />
					</Form.Field>

					<div class="flex justify-end">
						<Button type="submit" disabled={$submitting}>
							{#if $submitting}
								<div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
							{:else}
								<Save class="mr-2 h-4 w-4" />
							{/if}
							Save Profile
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
