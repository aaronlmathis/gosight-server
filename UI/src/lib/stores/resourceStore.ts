import { writable, derived } from 'svelte/store';
import type { Resource, ResourceFilter } from '$lib/types/resource';
import { api } from '$lib/api';

export const resources = writable<Resource[]>([]);
export const resourceFilter = writable<ResourceFilter>({});
export const selectedResources = writable<Set<string>>(new Set());

export const resourcesByKind = derived(resources, ($resources) => {
    const grouped = new Map<string, Resource[]>();
    $resources.forEach(resource => {
        if (!grouped.has(resource.kind)) {
            grouped.set(resource.kind, []);
        }
        grouped.get(resource.kind)!.push(resource);
    });
    return grouped;
});

export const resourceCounts = derived(resourcesByKind, ($resourcesByKind) => {
    const counts = new Map<string, number>();
    $resourcesByKind.forEach((resources, kind) => {
        counts.set(kind, resources.length);
    });
    return counts;
});

export async function loadResources(filter?: ResourceFilter) {
    try {
        const response = await api.request('/resources');
        resources.set(Array.isArray(response) ? response : []);
    } catch (error) {
        console.error('Failed to load resources:', error);
    }
}

export async function updateResourceTags(resourceId: string, tags: Record<string, string>) {
    try {
        await api.request(`/resources/${resourceId}/tags`, {
            method: 'PUT',
            body: JSON.stringify(tags)
        });
        // Refresh resources
        await loadResources();
    } catch (error) {
        console.error('Failed to update resource tags:', error);
    }
}