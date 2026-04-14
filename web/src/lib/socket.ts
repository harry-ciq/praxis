const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";

type MessageHandler = (data: unknown) => void;

class SocketClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  connect(token: string) {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(`${WS_URL}?token=${token}`);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      const handlers = this.handlers.get(msg.type);
      if (handlers) {
        handlers.forEach((handler) => handler(msg.data));
      }
    };

    this.ws.onclose = () => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => {
          this.reconnectAttempts++;
          this.connect(token);
        }, this.reconnectDelay * Math.pow(2, this.reconnectAttempts));
      }
    };
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
    this.reconnectAttempts = this.maxReconnectAttempts;
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => this.handlers.get(type)?.delete(handler);
  }

  send(type: string, data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }
}

export const socket = new SocketClient();
