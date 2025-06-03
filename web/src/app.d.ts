// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	interface Window {
		ApexCharts: any;
	}

	// Global ApexCharts constructor
	declare class ApexCharts {
		constructor(element: Element | string, options: any);
		render(): Promise<void>;
		updateSeries(series: any, animate?: boolean): Promise<void>;
		updateOptions(options: any, redrawPaths?: boolean, animate?: boolean): Promise<void>;
		destroy(): void;
	}
}

declare module 'svelte-apexcharts' {
	import { SvelteComponentTyped } from 'svelte';

	export interface ApexChartProps {
		options?: any;
		series?: any;
		type?: string;
		width?: string | number;
		height?: string | number;
	}

	export default class ApexChart extends SvelteComponentTyped<ApexChartProps> { }
}

export {};
