declare module 'svelte-apexcharts' {
	export function chart(node: HTMLElement, options: any): {
		update(options: any): void;
	};
}
