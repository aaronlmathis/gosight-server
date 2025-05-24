/**
 * WebSocket utilities for real-time updates
 */
import { writable, type Writable } from 'svelte/store';

interface WebSocketState {
	connected: boolean;
	error?: string;
}

export class GoSightWebSocket {
	private ws: WebSocket | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000;
	private endpoint: string;

	public state: Writable<WebSocketState> = writable({ connected: false });
	public messages: Writable<any[]> = writable([]);

	constructor(endpoint: string) {
		this.endpoint = endpoint;
		// Don't auto-connect, let the manager handle connection
	}

	public connect() {
		try {
			const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
			this.ws = new WebSocket(`${protocol}//${location.host}/ws/${this.endpoint}`);

			this.ws.addEventListener('open', () => {
				this.reconnectAttempts = 0;
				this.state.set({ connected: true });
			});

			this.ws.addEventListener('close', () => {
				this.state.set({ connected: false });
				this.scheduleReconnect();
			});

			this.ws.addEventListener('error', (error) => {
				console.error(`WebSocket error on ${this.endpoint}:`, error);
				this.state.set({ connected: false, error: 'Connection error' });
			});

			this.ws.addEventListener('message', (event) => {
				if (event.data === 'ping') return;

				try {
					const data = JSON.parse(event.data);
					this.messages.update(messages => [data, ...messages.slice(0, 99)]); // Keep last 100 messages
				} catch (err) {
					console.error('Failed to parse WebSocket message:', err);
				}
			});
		} catch (err) {
			console.error(`Failed to create WebSocket for ${this.endpoint}:`, err);
			this.scheduleReconnect();
		}
	}

	private scheduleReconnect() {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			this.state.set({ connected: false, error: 'Max reconnection attempts reached' });
			return;
		}

		setTimeout(() => {
			this.reconnectAttempts++;
			this.connect();
		}, this.reconnectDelay * Math.pow(2, this.reconnectAttempts));
	}

	public disconnect() {
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
	}

	public send(data: any) {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(data));
		}
	}
}

// WebSocket instances for different data types
export const alertsWS = new GoSightWebSocket('alerts');
export const eventsWS = new GoSightWebSocket('events');
export const logsWS = new GoSightWebSocket('logs');
export const metricsWS = new GoSightWebSocket('metrics');
export const commandsWS = new GoSightWebSocket('command');
export const processesWS = new GoSightWebSocket('process');

// WebSocket manager for convenience access to all websockets
export const websocketManager = {
  alerts: alertsWS,
  events: eventsWS,
  logs: logsWS,
  metrics: metricsWS,
  commands: commandsWS,
  processes: processesWS,

  connect() {
    alertsWS.connect();
    eventsWS.connect();
    logsWS.connect();
    metricsWS.connect();
    commandsWS.connect();
    processesWS.connect();
  },
  
  disconnect() {
    alertsWS.disconnect();
    eventsWS.disconnect();
    logsWS.disconnect();
    metricsWS.disconnect();
    commandsWS.disconnect();
    processesWS.disconnect();
  },
  
  subscribeToAlerts(callback: (data: any) => void) {
    const unsubscribe = alertsWS.messages.subscribe((messages) => {
      if (messages.length > 0) {
        callback(messages[0]);
      }
    });
    return unsubscribe;
  },
  
  subscribeToEvents(callback: (data: any) => void) {
    const unsubscribe = eventsWS.messages.subscribe((messages) => {
      if (messages.length > 0) {
        callback(messages[0]);
      }
    });
    return unsubscribe;
  },
  
  subscribeToLogs(callback: (data: any) => void) {
    const unsubscribe = logsWS.messages.subscribe((messages) => {
      if (messages.length > 0) {
        callback(messages[0]);
      }
    });
    return unsubscribe;
  },
  
  subscribeToMetrics(callback: (data: any) => void) {
    const unsubscribe = metricsWS.messages.subscribe((messages) => {
      if (messages.length > 0) {
        callback(messages[0]);
      }
    });
    return unsubscribe;
  }
};
