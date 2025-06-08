<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';
    import { api } from '$lib/api/api';
    import { auth } from '$lib/stores/authStore';
    import { Shield, ArrowLeft } from 'lucide-svelte';
    import { Button } from "$lib/components/ui/button/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import { Label } from "$lib/components/ui/label/index.js";
    import { Checkbox } from "$lib/components/ui/checkbox/index.js";
    import * as InputOTP from "$lib/components/ui/input-otp/index.js";

    let code = '';
    let remember = false;
    let loading = false;
    let error = '';
    let next = '';
    let shakeError = false;

    // Reactive statement to handle code changes and auto-submit
    $: if (code && code.length === 6) {
        console.log('Auto-submitting with code:', code);
        handleMFAVerify();
    }

    onMount(() => {
        // Get redirect parameter
        next = $page.url.searchParams.get('next') || '/';

        // Redirect if already logged in
        auth.subscribe((authState) => {
            if (authState.isAuthenticated && authState.user) {
                goto(next);
            }
        });
    });

    async function handleMFAVerify() {
        if (!code) {
            error = 'Please enter the 6-digit verification code';
            triggerShake();
            return;
        }

        if (code.length !== 6 || !/^\d+$/.test(code)) {
            error = 'Please enter a valid 6-digit code';
            triggerShake();
            return;
        }

        try {
            loading = true;
            error = '';

            const response = await api.auth.verifyMFA({ code, remember });

            if (response.success) {
                // Set user in auth store
                auth.setUser(response.user);

                // Redirect to original page or dashboard
                window.location.href = next;
            } else {
                error = response.message || 'MFA verification failed';
                triggerShake();
            }
        } catch (err) {
            console.log('MFA error:', err);
            
            let errorMessage = 'MFA verification failed. Please try again.';
            
            if (err && typeof err === 'object' && 'message' in err) {
                try {
                    // Try to parse the message as JSON first
                    const parsed = JSON.parse(err.message);
                    errorMessage = parsed.message || err.message;
                } catch {
                    // If it's not JSON, use the message as-is
                    errorMessage = err.message;
                }
            } else if (typeof err === 'string') {
                errorMessage = err;
            }
            
            error = errorMessage;
            triggerShake();
        } finally {
            loading = false;
        }
    }

    function handleBackToLogin() {
        goto(`/auth/login?next=${encodeURIComponent(next)}`);
    }

    function triggerShake() {
        shakeError = true;
        setTimeout(() => {
            shakeError = false;
        }, 800);
    }
</script>

<svelte:head>
    <title>Two-Factor Authentication - GoSight</title>
</svelte:head>

<div class="flex h-screen w-full items-center justify-center px-4 bg-gray-100 dark:bg-gray-900">
    <div class="mx-auto w-full max-w-md transition-all" class:animate-shake={shakeError}>
        <Card.Root>
            <Card.Header class="text-center">
                <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 mb-4">
                    <Shield class="h-6 w-6 text-foreground" />
                </div>
                <Card.Title class="text-center text-2xl font-bold tracking-tight text-foreground">Two-Factor Authentication</Card.Title>
                <Card.Description>
                    Enter the 6-digit code from your authenticator app
                </Card.Description>
            </Card.Header>

            <Card.Content>
                {#if error}
                    <div class="mb-4 rounded border border-destructive/20 bg-destructive/10 p-3 text-center text-sm text-destructive">
                        {error}
                    </div>
                {/if}

                <form on:submit|preventDefault={handleMFAVerify} class="space-y-6">
                    <div class="space-y-2">
                        <Label class="text-center block">Verification Code</Label>
                        <div class="flex justify-center">
                            <InputOTP.Root 
                                maxlength={6} 
                                bind:value={code}
                                disabled={loading}
                            >
                                {#snippet children({ cells })}
                                    <InputOTP.Group>
                                        {#each cells.slice(0, 3) as cell (cell)}
                                            <InputOTP.Slot {cell} />
                                        {/each}
                                    </InputOTP.Group>
                                    <InputOTP.Separator />
                                    <InputOTP.Group>
                                        {#each cells.slice(3, 6) as cell (cell)}
                                            <InputOTP.Slot {cell} />
                                        {/each}
                                    </InputOTP.Group>
                                {/snippet}
                            </InputOTP.Root>
                        </div>
                        <p class="text-xs text-center text-muted-foreground">Current code: {code}</p> <!-- Debug info -->
                    </div>

                    <div class="flex items-center justify-center space-x-2">
                        <Checkbox 
                            id="remember" 
                            bind:checked={remember} 
                            disabled={loading}
                        />
                        <Label 
                            for="remember" 
                            class="text-sm font-normal leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                            Remember this device for 30 days
                        </Label>
                    </div>

                    <Button
                        type="submit"
                        class="w-full"
                        disabled={loading}
                        onclick={() => {
                            console.log('Manual submit with code:', code); // Add debugging
                            handleMFAVerify();
                        }}
                    >
                        {#if loading}
                            <div class="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
                        {/if}
                        {loading ? 'Verifying...' : 'Verify Code'}
                    </Button>

                    <Button
                        type="button"
                        variant="outline"
                        class="w-full"
                        onclick={handleBackToLogin}
                        disabled={loading}
                    >
                        <ArrowLeft class="mr-2 h-4 w-4" />
                        Back to Login
                    </Button>
                </form>

                <div class="mt-6 text-center">
                    <p class="text-xs text-muted-foreground">
                        Having trouble? Contact your administrator for assistance.
                    </p>
                </div>
            </Card.Content>
        </Card.Root>
    </div>
</div>

<style>
    @keyframes shake {
        10%, 90% { transform: translateX(-1px); }
        20%, 80% { transform: translateX(2px); }
        30%, 50%, 70% { transform: translateX(-4px); }
        40%, 60% { transform: translateX(4px); }
    }

    :global(.animate-shake) {
        animation: shake 0.8s ease-in-out;
    }
</style>