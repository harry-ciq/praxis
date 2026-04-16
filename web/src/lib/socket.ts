const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";

type MessageHandler = (data: unknown) => void;

class SocketClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<MessageHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private currentToken: string | null = null;

  connect(token: string) {
    // Store the token so reconnection always uses the latest one
    this.currentToken = token;

    // Skip if already open or currently connecting
    if (
      this.ws &&
      (this.ws.readyState === WebSocket.OPEN ||
        this.ws.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    // Always reset reconnect counter on an explicit connect() call
    this.reconnectAttempts = 0;
    this._doConnect();
  }

  private _doConnect() {
    const token = this.currentToken;
    if (!token) return;

    // Clean up any existing socket
    if (this.ws) {
      this.ws.onopen = null;
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      if (
        this.ws.readyState === WebSocket.OPEN ||
        this.ws.readyState === WebSocket.CONNECTING
      ) {
        this.ws.close();
      }
      this.ws = null;
    }

    console.log("[WS] connecting...");
    this.ws = new WebSocket(`${WS_URL}?token=${token}`);

    this.ws.onopen = () => {
      console.log("[WS] connected");
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        const handlers = this.handlers.get(msg.type);
        if (handlers) {
          handlers.forEach((handler) => handler(msg.data));
        }
      } catch (err) {
        console.error("[WS] failed to parse message:", err);
      }
    };

    this.ws.onerror = (event) => {
      console.error("[WS] error:", event);
    };

    this.ws.onclose = (event) => {
      console.log(
        `[WS] closed (code=${event.code}, reason=${event.reason || "none"})`,
      );
      this.ws = null;

      // Don't reconnect if we were intentionally disconnected
      if (!this.currentToken) return;

      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay =
          this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
        console.log(
          `[WS] reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`,
        );
        this.reconnectAttempts++;
        setTimeout(() => {
          // Re-read token from localStorage in case it was refreshed
          const freshToken = localStorage.getItem("praxis_access_token");
          if (freshToken) {
            this.currentToken = freshToken;
          }
          if (this.currentToken) {
            this._doConnect();
          }
        }, delay);
      } else {
        console.warn("[WS] max reconnection attempts reached");
      }
    };
  }

  disconnect() {
    console.log("[WS] disconnect requested");
    this.currentToken = null; // Prevents reconnection in onclose
    if (this.ws) {
      this.ws.onclose = null; // Prevent the reconnect handler from firing
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = 0;
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  send(type: string, data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }

  /** Check current connection state (for debugging) */
  get state(): string {
    if (!this.ws) return "NONE";
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return "CONNECTING";
      case WebSocket.OPEN:
        return "OPEN";
      case WebSocket.CLOSING:
        return "CLOSING";
      case WebSocket.CLOSED:
        return "CLOSED";
      default:
        return "UNKNOWN";
    }
  }
}

export const socket = new SocketClient();
