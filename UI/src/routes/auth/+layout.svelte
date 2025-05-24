<!--
 * Auth Layout
 * This layout provides the centered auth container with dark background
 * matching the original design from web/templates/layouts/layout_auth.html
-->
<script lang="ts">
	import '../../app.html';
</script>

<svelte:head>
	<title>Login – GoSight</title>
	<style>
		body {
			background-color: #111827; /* Tailwind's gray-900 */
		}
	</style>
</svelte:head>

<main class="flex min-h-screen items-center justify-center bg-gray-900 px-4">
	<slot />
</main>

<style>
	:global(html) {
		background-color: #111827 !important;
	}
	:global(body) {
		background-color: #111827 !important;
	}
</style>
