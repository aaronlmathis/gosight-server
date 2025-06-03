<script lang="ts">
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';
    import { api } from '$lib/api/api';
    import { auth } from '$lib/stores/authStore';
    import { Button } from "$lib/components/ui/button/index.js";
    import * as Card from "$lib/components/ui/card/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";

    let username = '';
    let password = '';
    let loading = false;
    let error = '';
    let providers: { name: string; display_name: string }[] = [];
    let next = '';
    let shakeError = false;

    onMount(async () => {
        // Get redirect parameter
        next = $page.url.searchParams.get('next') || '/';

        // Check for error parameter from OAuth callback
        const errorParam = $page.url.searchParams.get('error');
        if (errorParam) {
            switch (errorParam) {
                case 'invalid_provider':
                    error = 'Invalid authentication provider selected';
                    break;
                case 'auth_failed':
                    error = 'Authentication failed. Please try again.';
                    break;
                case 'user_load_failed':
                    error = 'Failed to load user information. Please contact support.';
                    break;
                default:
                    error = 'Authentication error occurred';
            }
        }

        // Redirect if already logged in
        auth.subscribe((authState) => {
            if (authState.isAuthenticated && authState.user) {
                goto(next);
            }
        });

        // Load authentication providers
        try {
            const providerData = await api.auth.getProviders();
            providers = providerData.providers || [];
        } catch (err) {
            console.error('Failed to load providers:', err);
        }
    });

    async function handleLogin() {
        if (!username || !password) {
            error = 'Please enter both username and password';
            triggerShake();
            return;
        }

        try {
            loading = true;
            error = '';

            const response = await api.login({ username, password });

            if (response.success) {
                auth.setUser(response.user);
                goto(next);
            } else if (response.mfa_required) {
                goto(`/auth/mfa?next=${encodeURIComponent(next)}`);
            } else {
                error = response.message || 'Login failed';
                triggerShake();
            }
        } catch (err) {
            console.log('Caught error:', err);
            console.log('Error type:', typeof err);
            
            let errorMessage = 'Login failed. Please try again.';
            
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

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === 'Enter') {
            handleLogin();
        }
    }

    function handleSSOLogin(provider: string) {
        window.location.href = `/api/v1/auth/login/start?provider=${provider}&next=${encodeURIComponent(next)}`;
    }

    function getProviderIcon(provider: string): string {
        const icons: Record<string, string> = {
            google: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/google.svg',
            github: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/github.svg',
            azure: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/microsoftazure.svg',
            aws: 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/amazonaws.svg'
        };
        return icons[provider] || `https://cdn.jsdelivr.net/npm/simple-icons@v11/icons/${provider}.svg`;
    }

    function triggerShake() {
        shakeError = true;
        setTimeout(() => {
            shakeError = false;
        }, 800);
    }
</script>
<div class="flex h-screen w-full items-center justify-center px-4 bg-gray-100 dark:bg-gray-900">
<!-- Wrap Card.Root with a div that handles the animation -->
<div class="mx-auto w-full max-w-sm transition-all" class:animate-shake={shakeError}>
    <Card.Root>
        <Card.Header>
            <Card.Title class="text-center text-5xl tracking-tight" style="font-weight: 900;">GoSight</Card.Title>
            <Card.Description class="text-center">Sign in to your account</Card.Description>
        </Card.Header>
        <Card.Content>
            {#if error}
                <div class="mb-4 rounded border border-red-300 bg-red-100 p-2 text-center text-sm text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200">
                    {error}
                </div>
            {/if}

            <form on:submit|preventDefault={handleLogin} class="grid gap-4">
                <div class="grid gap-2">
                    <Label for="username">Username</Label>
                    <Input 
                        id="username" 
                        type="text" 
                        placeholder="Enter your username" 
                        required 
                        bind:value={username}
                        on:keypress={handleKeyPress}
                    />
                </div>
                <div class="grid gap-2">
                    <Label for="password">Password</Label>
                    <Input 
                        id="password" 
                        type="password" 
                        required 
                        bind:value={password}
                        on:keypress={handleKeyPress}
                    />
                </div>
                <Button type="submit" class="w-full" disabled={loading}>
                    {#if loading}
                        <div class="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
                    {/if}
                    Sign in
                </Button>
            </form>

            {#if providers.length > 0}
                <div class="mt-6 text-center text-sm text-muted-foreground">or sign in with</div>
                
                <div class="mt-4 space-y-2">
                    {#each providers as provider}
                        {#if provider.name !== 'local'}
                            <Button 
                                variant="outline" 
                                class="w-full"
                                onclick={() => handleSSOLogin(provider.name)}
                            >
                                <img
                                    src={getProviderIcon(provider.name)}
                                    class="mr-2 h-4 w-4"
                                    alt={provider.display_name}
                                />
                                Sign in with {provider.display_name}
                            </Button>
                        {/if}
                    {/each}
                </div>
            {/if}
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