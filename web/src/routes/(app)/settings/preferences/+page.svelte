<script lang="ts">
	import { onMount } from 'svelte';
	import { superForm } from 'sveltekit-superforms/client';
	import { zodClient } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Save } from 'lucide-svelte';
	import { preferencesSchema } from '../schema';
	import { api } from '$lib/api/api';

	let loading = true;
	let successMessage = '';

	const preferences = superForm(
		{
			theme: 'light',
			notifications: {
				email_alerts: false,
				push_alerts: false,
				alert_frequency: ''
			},
			dashboard: {
				refresh_interval: 30,
				default_time_range: '1h',
				show_system_metrics: true
			}
		},
		{
			validators: zodClient(preferencesSchema),
			dataType: 'json',
			onUpdated: async ({ form }) => {
				if (form.valid) {
					try {
						const result = await api.updateUserPreferences(form.data);
						if (result.success) {
							successMessage = 'Preferences updated successfully!';
							setTimeout(() => (successMessage = ''), 3000);
						}
					} catch (error) {
						console.error('Failed to update preferences:', error);
					}
				}
			}
		}
	);

	const { form, enhance, submitting } = preferences;

	$: formData = $form;

	// Helper function for theme labels
	function getThemeLabel(value: string): string {
		const labels: Record<string, string> = {
			light: 'Light',
			dark: 'Dark',
			system: 'System'
		};
		return labels[value] || 'Select theme';
	}

	// Helper function for time range labels
	function getTimeRangeLabel(value: string): string {
		const labels: Record<string, string> = {
			'1h': '1 Hour',
			'6h': '6 Hours',
			'24h': '24 Hours',
			'7d': '7 Days'
		};
		return labels[value] || 'Select time range';
	}

	// Load current preferences on mount
	onMount(async () => {
		try {
			const userPreferences = await api.getUserPreferences();
			if (userPreferences) {
				form.set({
					theme: userPreferences.theme || 'light',
					notifications: {
						email_alerts: userPreferences.notifications?.email_alerts ?? false,
						push_alerts: userPreferences.notifications?.push_alerts ?? false,
						alert_frequency: userPreferences.notifications?.alert_frequency || ''
					},
					dashboard: {
						refresh_interval: userPreferences.dashboard?.refresh_interval ?? 30,
						default_time_range: userPreferences.dashboard?.default_time_range || '1h',
						show_system_metrics: userPreferences.dashboard?.show_system_metrics ?? true
					}
				});
			}
		} catch (error) {
			console.error('Failed to load preferences:', error);
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Preferences</title>
</svelte:head>

{#if loading}
	<div class="flex items-center justify-center p-8">
		<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-gray-900"></div>
	</div>
{:else}
	<div class="space-y-6">
		<div>
			<h3 class="text-lg font-medium">Preferences</h3>
			<p class="text-muted-foreground text-sm">
				Configure your application preferences and notifications.
			</p>
		</div>

		{#if successMessage}
			<div class="rounded-md bg-green-50 p-4">
				<p class="text-sm font-medium text-green-800">{successMessage}</p>
			</div>
		{/if}

		<div class="">
			<div class=" py-5 sm:p-6">
				<form use:enhance class="space-y-6">
					<!-- Theme Selection -->
					<Form.Field form={preferences} name="theme">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Theme</Form.Label>
								<Select.Root
									type="single"
									value={formData.theme}
									onValueChange={(value: string) => {
										if (value) {
											form.update((data) => ({ ...data, theme: value }));
										}
									}}
								>
									<Select.Trigger {...props} class="w-full">
										{getThemeLabel(formData.theme)}
									</Select.Trigger>
									<Select.Content>
										<Select.Item value="light" label="Light" />
										<Select.Item value="dark" label="Dark" />
										<Select.Item value="system" label="System" />
									</Select.Content>
								</Select.Root>
							{/snippet}
						</Form.Control>
						<Form.Description>Choose your preferred theme.</Form.Description>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Notifications Section -->
					<div class="space-y-4">
						<h4 class="">Notifications</h4>

						<Form.Field form={preferences} name="notifications.email_alerts">
							<Form.Control>
								{#snippet children({ props })}
									<div class="flex items-center space-x-2">
										<input
											{...props}
											type="checkbox"
											bind:checked={formData.notifications.email_alerts}
											class="h-4 w-4 rounded border-gray-300"
										/>
										<Form.Label class="text-sm font-normal">Email alerts</Form.Label>
									</div>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>

						<Form.Field form={preferences} name="notifications.push_alerts">
							<Form.Control>
								{#snippet children({ props })}
									<div class="flex items-center space-x-2">
										<input
											{...props}
											type="checkbox"
											bind:checked={formData.notifications.push_alerts}
											class="h-4 w-4 rounded border-gray-300"
										/>
										<Form.Label class="text-sm font-normal">Push notifications</Form.Label>
									</div>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
					</div>

					<!-- Dashboard Section -->
					<div class="space-y-4">
						<h4 class="text-sm font-medium">Dashboard</h4>

						<Form.Field form={preferences} name="dashboard.refresh_interval">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Refresh Interval (seconds)</Form.Label>
									<Input
										{...props}
										type="number"
										min="10"
										max="300"
										bind:value={formData.dashboard.refresh_interval}
										placeholder="30"
									/>
								{/snippet}
							</Form.Control>
							<Form.Description
								>How often to refresh dashboard data (10-300 seconds).</Form.Description
							>
							<Form.FieldErrors />
						</Form.Field>

						<Form.Field form={preferences} name="dashboard.default_time_range">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Default Time Range</Form.Label>
									<Select.Root
										type="single"
										value={formData.dashboard.default_time_range}
										onValueChange={(value: string) => {
											if (value) {
												form.update((data) => ({
													...data,
													dashboard: {
														...data.dashboard,
														default_time_range: value
													}
												}));
											}
										}}
									>
										<Select.Trigger {...props} class="w-full">
											{getTimeRangeLabel(formData.dashboard.default_time_range)}
										</Select.Trigger>
										<Select.Content>
											<Select.Item value="1h" label="1 Hour" />
											<Select.Item value="6h" label="6 Hours" />
											<Select.Item value="24h" label="24 Hours" />
											<Select.Item value="7d" label="7 Days" />
										</Select.Content>
									</Select.Root>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>

						<Form.Field form={preferences} name="dashboard.show_system_metrics">
							<Form.Control>
								{#snippet children({ props })}
									<div class="flex items-center space-x-2">
										<input
											{...props}
											type="checkbox"
											bind:checked={formData.dashboard.show_system_metrics}
											class="h-4 w-4 rounded border-gray-300"
										/>
										<Form.Label class="text-sm font-normal">Show system metrics</Form.Label>
									</div>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
					</div>

					<div class="flex justify-end">
						<Button type="submit" disabled={$submitting}>
							{#if $submitting}
								<div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
							{:else}
								<Save class="mr-2 h-4 w-4" />
							{/if}
							Save Preferences
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
