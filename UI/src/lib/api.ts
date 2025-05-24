/**
 * API client for GoSight backend
 */
import type { AlertRule, AlertRulesResponse, EndpointsResponse, Endpoint } from './types';

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
	}): Promise<Endpoint[]> {
		const searchParams = appendSearchParams(new URLSearchParams(), params ?? {});
		const query = searchParams.toString();
		const response = await this.api.request(`/endpoints${query ? `?${query}` : ''}`);
		// Backend returns { data: Endpoint[] } structure
		return (response as any).data || [];
	}

async get(id: string) {
return this.api.request(`/endpoints/${id}`);
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
return this.api.request(`/logs${query ? `?${query}` : ''}`);
}

async getRecent(limit = 50) {
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
}

async getNamespaces() {
return this.api.request('/metrics/namespaces');
}

async getSubNamespaces(namespace: string) {
return this.api.request(`/metrics/namespaces/${namespace}`);
}

async getMetricNames(namespace: string, subNamespace: string) {
return this.api.request(`/metrics/namespaces/${namespace}/${subNamespace}`);
}

async getMetricDimensions(namespace: string, subNamespace: string, metric: string) {
return this.api.request(`/metrics/namespaces/${namespace}/${subNamespace}/${metric}/dimensions`);
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
}

async login(credentials: { username: string; password: string }) {
return this.api.request('/auth/login', {
method: 'POST',
body: JSON.stringify(credentials)
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
}

async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
const url = `${this.baseUrl}${endpoint}`;

const response = await fetch(url, {
...options,
headers: {
'Content-Type': 'application/json',
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

async getLogs(params?: any) {
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

async updateProfile(profileData: any) {
return this.auth.updateProfile(profileData);
}

async updatePassword(passwordData: any) {
return this.auth.updatePassword(passwordData);
}

async getUserPreferences() {
return this.auth.getUserPreferences();
}

async updateUserPreferences(preferences: any) {
return this.auth.updateUserPreferences(preferences);
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

async createAlertRule(rule: any) {
return this.alerts.createRule(rule);
}
}

// Create and initialize API singleton instance
export const api = new ApiClient();

// Legacy compatibility function for existing JS code
export function gosightFetch(url: string, options?: RequestInit) {
return fetch(url, {
...options,
headers: {
'Content-Type': 'application/json',
...options?.headers
}
});
}
