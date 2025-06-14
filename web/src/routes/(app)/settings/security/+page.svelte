<script lang="ts">
	import { superForm } from 'sveltekit-superforms/client';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Eye, EyeOff, Save } from 'lucide-svelte';
	import { passwordSchema } from '../schema';
	import { api } from '$lib/api/api';

	let successMessage = '';
	let showCurrentPassword = false;
	let showNewPassword = false;
	let showConfirmPassword = false;

	// Initialize superform
	const security = superForm(
		{ current_password: '', new_password: '', confirm_password: '' },
		{
			validators: zodClient(passwordSchema),
			onUpdated: async ({ form }) => {
				if (form.valid) {
					try {
						const result = await api.updatePassword({
							current_password: form.data.current_password,
							new_password: form.data.new_password
						});
						if (result.success) {
							successMessage = 'Password updated successfully!';

							form.set({
								current_password: '',
								new_password: '',
								confirm_password: ''
							});
							setTimeout(() => (successMessage = ''), 3000);
						}
					} catch (error) {
						console.error('Failed to update password:', error);
					}
				}
			}
		}
	);

	const { form, enhance, submitting } = security;

	$: formData = $form;
</script>

<svelte:head>
	<title>Security Settings</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h3 class="text-lg font-medium">Security</h3>
		<p class="text-muted-foreground text-sm">Update your password and security preferences.</p>
	</div>

	{#if successMessage}
		<div class="rounded-md bg-green-50 p-4">
			<p class="text-sm font-medium text-green-800">{successMessage}</p>
		</div>
	{/if}

	<div class="rounded-lg bg-white shadow dark:bg-gray-800">
		<div class="px-4 py-5 sm:p-6">
			<form use:enhance class="space-y-6">
				<Form.Field form={security} name="current_password">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Current Password</Form.Label>
							<div class="relative">
								<Input
									{...props}
									type={showCurrentPassword ? 'text' : 'password'}
									bind:value={formData.current_password}
									placeholder="Enter current password"
									class="pr-10"
								/>
								<button
									type="button"
									class="absolute inset-y-0 right-0 flex items-center pr-3"
									on:click={() => (showCurrentPassword = !showCurrentPassword)}
								>
									{#if showCurrentPassword}
										<EyeOff class="h-4 w-4 text-gray-400" />
									{:else}
										<Eye class="h-4 w-4 text-gray-400" />
									{/if}
								</button>
							</div>
						{/snippet}
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>

				<Form.Field form={security} name="new_password">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>New Password</Form.Label>
							<div class="relative">
								<Input
									{...props}
									type={showNewPassword ? 'text' : 'password'}
									bind:value={formData.new_password}
									placeholder="Enter new password"
									class="pr-10"
								/>
								<button
									type="button"
									class="absolute inset-y-0 right-0 flex items-center pr-3"
									on:click={() => (showNewPassword = !showNewPassword)}
								>
									{#if showNewPassword}
										<EyeOff class="h-4 w-4 text-gray-400" />
									{:else}
										<Eye class="h-4 w-4 text-gray-400" />
									{/if}
								</button>
							</div>
						{/snippet}
					</Form.Control>
					<Form.Description>Must be at least 8 characters long.</Form.Description>
					<Form.FieldErrors />
				</Form.Field>

				<Form.Field form={security} name="confirm_password">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Confirm New Password</Form.Label>
							<div class="relative">
								<Input
									{...props}
									type={showConfirmPassword ? 'text' : 'password'}
									bind:value={formData.confirm_password}
									placeholder="Confirm new password"
									class="pr-10"
								/>
								<button
									type="button"
									class="absolute inset-y-0 right-0 flex items-center pr-3"
									on:click={() => (showConfirmPassword = !showConfirmPassword)}
								>
									{#if showConfirmPassword}
										<EyeOff class="h-4 w-4 text-gray-400" />
									{:else}
										<Eye class="h-4 w-4 text-gray-400" />
									{/if}
								</button>
							</div>
						{/snippet}
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>

				<div class="flex justify-end">
					<Button type="submit" disabled={$submitting}>
						{#if $submitting}
							<div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
						{:else}
							<Save class="mr-2 h-4 w-4" />
						{/if}
						Update Password
					</Button>
				</div>
			</form>
		</div>
	</div>
</div>
