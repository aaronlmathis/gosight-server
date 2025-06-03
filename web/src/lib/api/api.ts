/**
 * API client for GoSight backend
 */
import type { 
    AlertRule, 
    AlertRulesResponse, 
    EndpointsResponse, 
    Endpoint, 
    LogResponse,
    Role,
    Permission,
    RoleWithPermissions,
    PermissionWithRoles,
    UserWithRoles,
    CreateRoleRequest,
    UpdateRoleRequest,
    CreatePermissionRequest,
    UpdatePermissionRequest,
    AssignRolesRequest,
    AssignPermissionsRequest,
    RolesResponse,
    PermissionsResponse,
    UsersWithRoleResponse,
    RolesWithPermissionResponse
} from './types';

export interface ApiError {
message: string;
status: number;
}

export class GoSightApiError extends Error {
status: number;

constructor(message: string, status: number) {
super(message);
this.name = 'GoSightApiError';
this.status = status;
}
}

// Helper function to safely handle params
function appendSearchParams(searchParams: URLSearchParams, params: Record<string, any>) {
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value));
      }
    });
  }
  return searchParams;
}

// Alert API
export class AlertsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}

async getAll(params?: {
limit?: number;
page?: number;
state?: string;
level?: string;
rule_id?: string;
sort?: string;
order?: string;
endpoint_id?: string;
}) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
const query = searchParams.toString();
return this.api.request(`/alerts${query ? `?${query}` : ''}`);
}

async getActive() {
return this.api.request('/alerts/active');
}

async getRules(): Promise<AlertRule[]> {
    const response = await this.api.request('/alerts/rules');
    // Backend returns array directly, not wrapped in response object
    return Array.isArray(response) ? response : [];
}

async getSummary() {
return this.api.request('/alerts/summary');
}

async create(alert: any) {
return this.api.request('/alerts', {
method: 'POST',
body: JSON.stringify(alert)
});
}

async getContext(alertId: string, window?: string) {
const params = window ? `?window=${window}` : '';
return this.api.request(`/alerts/${alertId}/context${params}`);
}

async acknowledge(alertId: string) {
return this.api.request(`/alerts/${alertId}/acknowledge`, {
method: 'POST'
});
}

async resolve(alertId: string) {
return this.api.request(`/alerts/${alertId}/resolve`, {
method: 'POST'
});
}

async createRule(rule: any) {
return this.api.request('/alerts/rules', {
method: 'POST',
body: JSON.stringify(rule)
});
}

async updateRule(ruleId: string, rule: any) {
return this.api.request(`/alerts/rules/${ruleId}`, {
method: 'PUT',
body: JSON.stringify(rule)
});
}

async deleteRule(ruleId: string) {
return this.api.request(`/alerts/rules/${ruleId}`, {
method: 'DELETE'
});
}
}

// Endpoints API
export class EndpointsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}	async getAll(params?: {
        type?: string;
        status?: string;
        hostname?: string;
        endpointID?: string;
        hostID?: string;
        limit?: number;
    }): Promise<Endpoint[]> {
        const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
        const query = searchParams.toString();
        const response = await this.api.request(`/endpoints${query ? `?${query}` : ''}`);
        // Backend returns array directly, not wrapped in data object
        return Array.isArray(response) ? response : [];
    }	async get(id: string) {
        // Backend doesn't have /endpoints/{id} route, use query parameter instead
        const response = await this.getAll({ endpointID: id });
        return { data: response[0] || null };
    }

async getByType(type: 'hosts' | 'containers', params?: any) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
const query = searchParams.toString();
return this.api.request(`/endpoints/${type}${query ? `?${query}` : ''}`);
}

async getContainers() {
return this.getByType('containers');
}

async getHosts() {
return this.getByType('hosts');
}
}

// Events API
export class EventsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}

async getAll(params?: {
limit?: number;
level?: string;
type?: string;
category?: string;
scope?: string;
target?: string;
source?: string;
contains?: string;
start?: string;
end?: string;
hostID?: string;
endpointID?: string;
endpoint_id?: string;
sort?: string;
}) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
const query = searchParams.toString();
return this.api.request(`/events${query ? `?${query}` : ''}`);
}

async getRecent(limit = 10) {
return this.api.request(`/events/recent?limit=${limit}`);
}
}

// Logs API
export class LogsApi {
    private api: ApiClient;

    constructor(api: ApiClient) {
        this.api = api;
    }

    async getAll(params: {
        limit?: number;
        page?: number;
        level?: string;
        levels?: string[];
        category?: string;
        categories?: string[];
        contains?: string;
        start?: string;
        end?: string;
        hostID?: string;
        endpointID?: string;
        endpoint_id?: string;
        source?: string;
        container?: string;
        container_name?: string;
        app?: string;
        app_name?: string;
        sort?: string;
        order?: string;
        cursor?: string;
        target?: string;
        unit?: string;
        service?: string;
        event_id?: string;
        user?: string;
        container_id?: string;
        platform?: string;
        [key: string]: any; // for dynamic tag_* and field_* and meta_* parameters
    } = {}): Promise<LogResponse> {
        const searchParams = new URLSearchParams();
        
        // Handle array parameters (levels, categories)
        if (params.levels?.length) {
            params.levels.forEach(level => searchParams.append('level', level));
        } else if (params.level) {
            searchParams.append('level', params.level);
        }
        
        if (params.categories?.length) {
            params.categories.forEach(cat => searchParams.append('category', cat));
        } else if (params.category) {
            searchParams.append('category', params.category);
        }
        
        // Handle all other parameters
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '' && 
                !['levels', 'categories'].includes(key)) {
                if (Array.isArray(value)) {
                    value.forEach(v => searchParams.append(key, String(v)));
                } else {
                    searchParams.append(key, String(value));
                }
            }
        });
        
        const query = searchParams.toString();
        return this.api.request(`/logs${query ? `?${query}` : ''}`);
    }

    async getRecent(limit = 50): Promise<LogResponse> {
        return this.api.request(`/logs/latest?limit=${limit}`);
    }
}

// Metrics API
export class MetricsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}

async getAll(params?: {
limit?: number;
endpoint_id?: string;
endpointID?: string;
start?: string;
end?: string;
name?: string;
}) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
const query = searchParams.toString();
return this.api.request(`/metrics${query ? `?${query}` : ''}`);
}

async getSystemOverview() {
return this.api.request('/metrics/system');
    } async getNamespaces() {
        return this.api.request('/metrics');
    }

    async getSubNamespaces(namespace: string) {
        return this.api.request(`/metrics/${namespace}`);
    }

    async getMetricNames(namespace: string, subNamespace: string) {
        return this.api.request(`/metrics/${namespace}/${subNamespace}`);
    }

    async getMetricDimensions(namespace: string, subNamespace: string, metric: string) {
        return this.api.request(`/metrics/${namespace}/${subNamespace}/${metric}/dimensions`);
    }

async getMetricData(namespace: string, subNamespace: string, metric: string, params: any) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
const query = searchParams.toString();
return this.api.request(`/metrics/namespaces/${namespace}/${subNamespace}/${metric}/data${query ? `?${query}` : ''}`);
}

async getMetricLatest(namespace: string, subNamespace: string, metric: string) {
return this.api.request(`/metrics/namespaces/${namespace}/${subNamespace}/${metric}/latest`);
}

async query(params: any) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/metrics/query?${searchParams.toString()}`);
}
}

// Reports API
export class ReportsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}

async getSystemSummary(params: { range: string }) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/reports/summary?${searchParams.toString()}`);
}

async getAlertsReport(params: { range: string; endpoints?: string[] }) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/reports/alerts?${searchParams.toString()}`);
}

async getMetricsReport(params: { range: string; endpoints?: string[] }) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/reports/metrics?${searchParams.toString()}`);
}

async getEventsReport(params: { range: string; endpoints?: string[] }) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/reports/events?${searchParams.toString()}`);
}

async exportReport(params: { type: string; format: string; range: string; endpoints?: string[] }) {
const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
return this.api.request(`/reports/export?${searchParams.toString()}`);
}
}

// Commands API
export class CommandsApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}

async send(endpointId: string, command: { command: string; args: string[] }) {
return this.api.request(`/commands/${endpointId}`, {
method: 'POST',
body: JSON.stringify(command)
});
}
}

// Auth API methods
export class AuthApi {
private api: ApiClient;

constructor(api: ApiClient) {
this.api = api;
}	async login(credentials: { username: string; password: string }) {
        return this.api.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        });
    }

    async getProviders() {
        return this.api.request('/auth/providers');
    }

    async verifyMFA(mfaData: { code: string; remember?: boolean }) {
return this.api.request('/auth/mfa/verify', {
method: 'POST',
body: JSON.stringify(mfaData)
});
}

async register(userData: { username: string; email: string; password: string; confirmPassword: string; first_name?: string; last_name?: string }) {
return this.api.request('/auth/register', {
method: 'POST',
body: JSON.stringify(userData)
});
}

async logout() {
return this.api.request('/auth/logout', {
method: 'POST'
});
}

async getCurrentUser() {
return this.api.request('/auth/me');
}	async getUserProfile() {
        // The profile data is included in the /auth/me response, so we can extract it from there
        const currentUser: any = await this.getCurrentUser();
        return currentUser?.profile || {};
    }

async getUserSettings() {
return this.getUserPreferences();
}

async updateProfile(profileData: any) {
return this.api.request('/users/profile', {
method: 'PUT',
body: JSON.stringify(profileData)
});
}

async updatePassword(passwordData: { current_password: string; new_password: string; confirm_password: string }) {
return this.api.request('/users/password', {
method: 'PUT',
body: JSON.stringify(passwordData)
});
}

async getUserPreferences() {
return this.api.request('/users/preferences');
}

async updateUserPreferences(preferences: any) {
return this.api.request('/users/preferences', {
method: 'PUT',
body: JSON.stringify(preferences)
});
}

async uploadAvatar(file: File) {
const formData = new FormData();
formData.append('avatar', file);

try {
const response = await fetch('/api/v1/users/avatar', {
method: 'POST',
credentials: 'include',
headers: {
'X-API-Version': 'v1'
},
body: formData,
// Don't set Content-Type - let the browser set it with the boundary
});

if (!response.ok) {
const errorData = await response.json().catch(() => ({ message: 'Network error' }));
throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
}

return await response.json();
} catch (error) {
console.error('Avatar upload error:', error);
throw error;
}
}

async cropAvatar(cropData: { x: number; y: number; width: number; height: number }): Promise<{ success: boolean; avatar_url: string; message: string }> {
    try {
        const response = await fetch('/api/v1/users/avatar/crop', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Version': 'v1'
            },
            credentials: 'include',
            body: JSON.stringify(cropData),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'Network error' }));
            throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Avatar crop error:', error);
        throw error;
    }
}

async deleteAvatar() {
return this.api.request('/users/avatar', {
method: 'DELETE'
});
}

async getUploadLimits() {
return this.api.request('/upload/limits');
}
}

// Roles API
export class RolesApi {
    private api: ApiClient;

    constructor(api: ApiClient) {
        this.api = api;
    }

    async getAll(): Promise<Role[]> {
        const response = await this.api.request('/roles');
        return Array.isArray(response) ? response : [];
    }

    async get(id: string): Promise<Role> {
        return this.api.request(`/roles/${id}`);
    }

    async create(role: CreateRoleRequest): Promise<Role> {
        return this.api.request('/roles', {
            method: 'POST',
            body: JSON.stringify(role)
        });
    }

    async update(id: string, role: UpdateRoleRequest): Promise<Role> {
        return this.api.request(`/roles/${id}`, {
            method: 'PUT',
            body: JSON.stringify(role)
        });
    }

    async delete(id: string): Promise<void> {
        return this.api.request(`/roles/${id}`, {
            method: 'DELETE'
        });
    }

    async getPermissions(id: string): Promise<Permission[]> {
        const response = await this.api.request(`/roles/${id}/permissions`);
        return Array.isArray(response) ? response : [];
    }

    async assignPermissions(id: string, request: AssignPermissionsRequest): Promise<void> {
        return this.api.request(`/roles/${id}/permissions`, {
            method: 'POST',
            body: JSON.stringify(request)
        });
    }

    async removePermissions(id: string, request: AssignPermissionsRequest): Promise<void> {
        return this.api.request(`/roles/${id}/permissions`, {
            method: 'DELETE',
            body: JSON.stringify(request)
        });
    }

    async getUsers(id: string): Promise<UserWithRoles[]> {
        const response = await this.api.request(`/roles/${id}/users`);
        return Array.isArray(response) ? response : [];
    }
}

// Permissions API
export class PermissionsApi {
    private api: ApiClient;

    constructor(api: ApiClient) {
        this.api = api;
    }

    async getAll(): Promise<Permission[]> {
        const response = await this.api.request('/permissions');
        return Array.isArray(response) ? response : [];
    }

    async get(id: string): Promise<Permission> {
        return this.api.request(`/permissions/${id}`);
    }

    async create(permission: CreatePermissionRequest): Promise<Permission> {
        return this.api.request('/permissions', {
            method: 'POST',
            body: JSON.stringify(permission)
        });
    }

    async update(id: string, permission: UpdatePermissionRequest): Promise<Permission> {
        return this.api.request(`/permissions/${id}`, {
            method: 'PUT',
            body: JSON.stringify(permission)
        });
    }

    async delete(id: string): Promise<void> {
        return this.api.request(`/permissions/${id}`, {
            method: 'DELETE'
        });
    }

    async getRoles(id: string): Promise<RoleWithPermissions[]> {
        const response = await this.api.request(`/permissions/${id}/roles`);
        return Array.isArray(response) ? response : [];
    }
}

// Users API (IAM-related methods)
export class UsersApi {
    private api: ApiClient;

    constructor(api: ApiClient) {
        this.api = api;
    }

    async getAll(): Promise<UserWithRoles[]> {
        const response = await this.api.request('/users');
        return Array.isArray(response) ? response : [];
    }

    async get(id: string): Promise<UserWithRoles> {
        return this.api.request(`/users/${id}`);
    }

    async create(user: any): Promise<UserWithRoles> {
        return this.api.request('/users', {
            method: 'POST',
            body: JSON.stringify(user)
        });
    }

    async update(id: string, user: any): Promise<UserWithRoles> {
        return this.api.request(`/users/${id}`, {
            method: 'PUT',
            body: JSON.stringify(user)
        });
    }

    async delete(id: string): Promise<void> {
        return this.api.request(`/users/${id}`, {
            method: 'DELETE'
        });
    }

    async getRoles(id: string): Promise<Role[]> {
        const response = await this.api.request(`/users/${id}/roles`);
        return Array.isArray(response) ? response : [];
    }

    async assignRoles(id: string, request: AssignRolesRequest): Promise<void> {
        return this.api.request(`/users/${id}/roles`, {
            method: 'POST',
            body: JSON.stringify(request)
        });
    }

    async removeRoles(id: string, request: AssignRolesRequest): Promise<void> {
        return this.api.request(`/users/${id}/roles`, {
            method: 'DELETE',
            body: JSON.stringify(request)
        });
    }
}

export class ApiClient {
private baseUrl: string;
public alerts: AlertsApi;
public endpoints: EndpointsApi;
public events: EventsApi;
public logs: LogsApi;
public metrics: MetricsApi;
public reports: ReportsApi;
public commands: CommandsApi;
public auth: AuthApi;
public roles: RolesApi;
public permissions: PermissionsApi;
public users: UsersApi;

constructor(baseUrl: string = '/api/v1') {
this.baseUrl = baseUrl;
this.alerts = new AlertsApi(this);
this.endpoints = new EndpointsApi(this);
this.events = new EventsApi(this);
this.logs = new LogsApi(this);
this.metrics = new MetricsApi(this);
this.reports = new ReportsApi(this);
this.commands = new CommandsApi(this);
this.auth = new AuthApi(this);
this.roles = new RolesApi(this);
this.permissions = new PermissionsApi(this);
this.users = new UsersApi(this);
}	async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;

        const response = await fetch(url, {
            ...options,
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Version': 'v1',
                ...options.headers
            }
        });

if (!response.ok) {
const errorMessage = await response.text();
throw new GoSightApiError(
errorMessage || `HTTP ${response.status}`,
response.status
);
}

const contentType = response.headers.get('content-type');
if (contentType?.includes('application/json')) {
return response.json();
}

return response.text() as unknown as T;
}

// Network Devices API
async getNetworkDevices() {
return this.request('/network-devices');
}

async createNetworkDevice(device: any) {
return this.request('/network-devices', {
method: 'POST',
body: JSON.stringify(device)
});
}

async updateNetworkDevice(id: string, device: any) {
return this.request(`/network-devices/${id}`, {
method: 'PUT',
body: JSON.stringify(device)
});
}

async deleteNetworkDevice(id: string) {
return this.request(`/network-devices/${id}`, {
method: 'DELETE'
});
}

async toggleNetworkDeviceStatus(id: string) {
return this.request(`/network-devices/${id}/toggle`, {
method: 'POST'
});
}

// Search API
async globalSearch(query: string) {
return this.request(`/search?q=${encodeURIComponent(query)}`);
}

// Tags API
async getTagKeys() {
return this.request('/tags/keys');
}

async getTagValues() {
return this.request('/tags/values');
}

async getTags(endpointId: string) {
return this.request(`/tags/${endpointId}`);
}

async setTags(endpointId: string, tags: any) {
return this.request(`/tags/${endpointId}`, {
method: 'POST',
body: JSON.stringify(tags)
});
}

async patchTags(endpointId: string, tags: any) {
return this.request(`/tags/${endpointId}`, {
method: 'PATCH',
body: JSON.stringify(tags)
});
}

async deleteTag(endpointId: string, key: string) {
return this.request(`/tags/${endpointId}/${key}`, {
method: 'DELETE'
});
}

// Additional direct methods for backward compatibility
async getEndpoint(id: string) {
return this.endpoints.get(id);
}	async getEndpoints(): Promise<Endpoint[]> {
        return this.endpoints.getAll();
    }

async getAlerts(params?: any) {
return this.alerts.getAll(params);
}

async getMetrics(params?: any) {
return this.metrics.getAll(params);
}

async getLogs(params?: any): Promise<LogResponse> {
return this.logs.getAll(params);
}

async getEvents(params?: any) {
return this.events.getAll(params);
}

async sendCommand(endpointId: string, command: { command: string; args: string[] }) {
return this.commands.send(endpointId, command);
}

async acknowledgeAlert(alertId: string) {
return this.alerts.acknowledge(alertId);
}

async resolveAlert(alertId: string) {
return this.alerts.resolve(alertId);
}

// Auth methods for backward compatibility
async login(credentials: { username: string; password: string }) {
return this.auth.login(credentials);
}

async register(userData: any) {
return this.auth.register(userData);
}

async logout() {
return this.auth.logout();
}

async getCurrentUser() {
return this.auth.getCurrentUser();
}

async updateProfile(data: { full_name: string; phone: string }): Promise<{ success: boolean; message: string }> {
    try {
        const response = await fetch('/api/v1/users/profile', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Version': 'v1'
            },
            credentials: 'include',
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ message: 'Network error' }));
            throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Profile update error:', error);
        throw error;
    }
}

async updatePassword(passwordData: any) {
return this.auth.updatePassword(passwordData);
}

async getUserPreferences() {
return this.auth.getUserPreferences();
}

async getUserSettings() {
return this.auth.getUserSettings();
}

async updateUserPreferences(preferences: any) {
return this.auth.updateUserPreferences(preferences);
}

async cropAvatar(cropData: { x: number; y: number; width: number; height: number }) {
return this.auth.cropAvatar(cropData);
}

async uploadAvatar(file: File) {
return this.auth.uploadAvatar(file);
}

async deleteAvatar() {
return this.auth.deleteAvatar();
}

async getUploadLimits() {
return this.auth.getUploadLimits();
}

// Report methods for backward compatibility
async getSystemSummary(params: any) {
return this.reports.getSystemSummary(params);
}

async getAlertsReport(params: any) {
return this.reports.getAlertsReport(params);
}

async getMetricsReport(params: any) {
return this.reports.getMetricsReport(params);
}

async getEventsReport(params: any) {
return this.reports.getEventsReport(params);
}

async exportReport(params: any) {
return this.reports.exportReport(params);
}

async updateAlertRule(id: string, rule: any) {
return this.alerts.updateRule(id, rule);
}

async deleteAlertRule(id: string) {
return this.alerts.deleteRule(id);
}

async getAlertRules(): Promise<AlertRule[]> {
    return this.alerts.getRules();
}

async getSummary() {
    return this.alerts.getSummary();
}

async createAlertRule(rule: any) {
return this.alerts.createRule(rule);
}
}

// Create and initialize API singleton instance
export const api = new ApiClient();

// Export individual API instances for direct use
export const rolesApi = api.roles;
export const permissionsApi = api.permissions;
export const usersApi = api.users;

// Legacy compatibility function for existing JS code
export function gosightFetch(url: string, options?: RequestInit) {
    return fetch(url, {
        ...options,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Version': 'v1',
            ...options?.headers
        }
    });
}
