/**
 * Utility functions for formatting, dates, etc.
 */

// Date formatting utilities
export function formatDate(date: string | Date, format: 'short' | 'long' | 'time' = 'short'): string {
	if (!date) return '—';
	
	const d = typeof date === 'string' ? new Date(date) : date;
	if (isNaN(d.getTime())) return '—';

	switch (format) {
		case 'time':
			return d.toLocaleTimeString();
		case 'long':
			return d.toLocaleString();
		case 'short':
		default:
			return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}
}

export function timeAgo(date: string | Date): string {
	if (!date) return '—';
	
	const d = typeof date === 'string' ? new Date(date) : date;
	if (isNaN(d.getTime())) return '—';

	const now = new Date();
	const diffMs = now.getTime() - d.getTime();
	const diffSecs = Math.floor(diffMs / 1000);
	const diffMins = Math.floor(diffSecs / 60);
	const diffHours = Math.floor(diffMins / 60);
	const diffDays = Math.floor(diffHours / 24);

	if (diffSecs < 60) return 'just now';
	if (diffMins < 60) return `${diffMins}m ago`;
	if (diffHours < 24) return `${diffHours}h ago`;
	if (diffDays < 7) return `${diffDays}d ago`;
	
	return formatDate(d, 'short');
}

// String utilities
export function capitalize(str: string): string {
	if (!str) return '';
	return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

export function getUserInitials(firstName: string | undefined, lastName: string | undefined): string {
	const first = (firstName || '').trim();
	const last = (lastName || '').trim();
	
	if (!first && !last) return '?';
	if (!first) return last.charAt(0).toUpperCase();
	if (!last) return first.charAt(0).toUpperCase();
	
	return (first.charAt(0) + last.charAt(0)).toUpperCase();
}

// Duration formatting
export function formatDuration(seconds: number): string {
  if (seconds === undefined || seconds === null) return '—';
  
  if (seconds < 60) {
    return `${seconds.toFixed(1)}s`;
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}m ${remainingSeconds}s`;
  } else if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  } else {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    return `${days}d ${hours}h`;
  }
}

export function escapeHTML(str: string): string {
	if (!str) return '';
	const div = document.createElement('div');
	div.textContent = str;
	return div.innerHTML;
}

export function truncate(str: string, length: number = 50): string {
	if (!str) return '';
	if (str.length <= length) return str;
	return str.slice(0, length) + '...';
}

// Status badge utilities
export function getStatusBadgeClass(status: string): string {
	const statusLower = status?.toLowerCase() || 'unknown';
	
	switch (statusLower) {
		case 'online':
		case 'running':
		case 'active':
		case 'ok':
		case 'healthy':
		case 'resolved':
			return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
		
		case 'offline':
		case 'stopped':
		case 'inactive':
		case 'unhealthy':
			return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
		
		case 'warning':
		case 'degraded':
		case 'pending':
		case 'acknowledged':
			return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
		
		case 'firing':
		case 'critical':
		case 'error':
		case 'high':
			return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
		
		case 'medium':
			return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200';
		
		case 'low':
			return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
			
		case 'info':
		case 'unknown':
		default:
			return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
	}
}

// Alias for status badge to maintain backward compatibility
export const getBadgeClass = getStatusBadgeClass;

export function getLevelBadgeClass(level: string): string {
	const levelLower = level.toLowerCase();
	
	switch (levelLower) {
		case 'critical':
		case 'error':
			return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
		
		case 'warning':
		case 'warn':
			return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200';
		
		case 'info':
		case 'information':
			return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
		
		case 'debug':
		case 'trace':
			return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
		
		default:
			return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
	}
}

export function getScopeBadgeClass(scope: string): string {
	const scopeLower = scope.toLowerCase();
	
	switch (scopeLower) {
		case 'agent':
		case 'host':
			return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
		
		case 'container':
			return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200';
		
		case 'endpoint':
			return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
		
		case 'user':
			return 'bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200';
		
		case 'rule':
		case 'system':
			return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
		
		default:
			return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
	}
}

// Number formatting
export function formatNumber(num: number | string): string {
	if (typeof num === 'string') {
		const parsed = parseFloat(num);
		if (isNaN(parsed)) return num;
		num = parsed;
	}
	
	if (num >= 1e9) return (num / 1e9).toFixed(1) + 'B';
	if (num >= 1e6) return (num / 1e6).toFixed(1) + 'M';
	if (num >= 1e3) return (num / 1e3).toFixed(1) + 'K';
	
	return num.toLocaleString();
}

// Status badge function
export function getStatusBadge(status: string): string {
	const statusLower = status?.toLowerCase() || 'unknown';
	const badgeClass = getStatusBadgeClass(statusLower);
	return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${badgeClass}">${status}</span>`;
}

export function formatBytes(bytes: number | string): string {
	if (typeof bytes === 'string') {
		const parsed = parseFloat(bytes);
		if (isNaN(parsed)) return bytes;
		bytes = parsed;
	}
	
	if (bytes === 0) return '0 B';
	
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	
	return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export function formatPercent(value: number | string, decimals: number = 1): string {
	if (typeof value === 'string') {
		const parsed = parseFloat(value);
		if (isNaN(parsed)) return value;
		value = parsed;
	}
	
	return value.toFixed(decimals) + '%';
}

// URL utilities
export function buildQueryString(params: Record<string, any>): string {
	const searchParams = new URLSearchParams();
	
	Object.entries(params).forEach(([key, value]) => {
		if (value !== null && value !== undefined && value !== '') {
			if (Array.isArray(value)) {
				value.forEach(v => searchParams.append(key, v.toString()));
			} else {
				searchParams.append(key, value.toString());
			}
		}
	});
	
	return searchParams.toString();
}

// Validation utilities
export function isValidEmail(email: string): boolean {
	const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	return emailRegex.test(email);
}

export function isValidUrl(url: string): boolean {
	try {
		new URL(url);
		return true;
	} catch {
		return false;
	}
}

// Debounce utility
export function debounce<T extends (...args: any[]) => any>(
	func: T,
	delay: number
): (...args: Parameters<T>) => void {
	let timeoutId: ReturnType<typeof setTimeout>;
	
	return (...args: Parameters<T>) => {
		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => func(...args), delay);
	};
}

// Local storage utilities
export function saveToLocalStorage(key: string, value: any): void {
	try {
		localStorage.setItem(key, JSON.stringify(value));
	} catch (error) {
		console.error('Failed to save to localStorage:', error);
	}
}

export function loadFromLocalStorage<T>(key: string, defaultValue: T): T {
	try {
		const item = localStorage.getItem(key);
		return item ? JSON.parse(item) : defaultValue;
	} catch (error) {
		console.error('Failed to load from localStorage:', error);
		return defaultValue;
	}
}

// Theme utilities
export function toggleTheme(): void {
	const isDark = document.documentElement.classList.contains('dark');
	if (isDark) {
		document.documentElement.classList.remove('dark');
		saveToLocalStorage('theme', 'light');
	} else {
		document.documentElement.classList.add('dark');
		saveToLocalStorage('theme', 'dark');
	}
}

export function initTheme(): void {
	const savedTheme = loadFromLocalStorage('theme', 'light') as 'light' | 'dark' | 'auto';
	const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
	
	if (savedTheme === 'dark' || (savedTheme === 'auto' && prefersDark)) {
		document.documentElement.classList.add('dark');
	}
}
