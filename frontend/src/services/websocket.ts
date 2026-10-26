// ============================================
// Tipos
// ============================================

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export type MessageHandler = (message: WebSocketMessage) => void;

export interface WebSocketMessage {
  type: string;
  data?: unknown;
  message?: string;
  time?: string;
}

export interface WebSocketClientOptions {
  /** URL del WebSocket (ej: ws://localhost:8080/ws) */
  url: string;

  /** Intervalo de reconexión base en ms (default: 3000) */
  reconnectInterval?: number;

  /** Máximo de intentos de reconexión (default: 10) */
  maxReconnectAttempts?: number;

  /** Intervalo del heartbeat en ms (default: 30000) */
  heartbeatInterval?: number;

  /** Callback cuando se conecta */
  onOpen?: () => void;

  /** Callback cuando se desconecta */
  onClose?: (event: CloseEvent) => void;

  /** Callback cuando hay error */
  onError?: (event: Event) => void;

  /** Callback cuando llega un mensaje */
  onMessage?: MessageHandler;

  /** Callback cuando cambia el estado */
  onStatusChange?: (status: ConnectionStatus) => void;
}

// ============================================
// Cliente WebSocket
// ============================================

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private status: ConnectionStatus = 'connecting';

  // Configuración
  private reconnectInterval: number;
  private maxReconnectAttempts: number;
  private heartbeatInterval: number;

  // Estado de reconexión
  private reconnectAttempts = 0;
  private reconnectTimer: number | null = null;
  private heartbeatTimer: number | null = null;
  private shouldReconnect = true;

  // Cola de mensajes pendientes (cuando está desconectado)
  private messageQueue: WebSocketMessage[] = [];

  // Listeners
  private messageListeners = new Set<MessageHandler>();

  // Callbacks
  private onOpenCallback?: () => void;
  private onCloseCallback?: (event: CloseEvent) => void;
  private onErrorCallback?: (event: Event) => void;
  private onStatusChangeCallback?: (status: ConnectionStatus) => void;

  constructor(options: WebSocketClientOptions) {
    this.url = options.url;
    this.reconnectInterval = options.reconnectInterval ?? 3000;
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 10;
    this.heartbeatInterval = options.heartbeatInterval ?? 30_000;

    this.onOpenCallback = options.onOpen;
    this.onCloseCallback = options.onClose;
    this.onErrorCallback = options.onError;
    this.onStatusChangeCallback = options.onStatusChange;

    if (options.onMessage) {
      this.messageListeners.add(options.onMessage);
    }
  }

  // ============================================
  // Estado
  // ============================================

  public getStatus(): ConnectionStatus {
    return this.status;
  }

  private setStatus(status: ConnectionStatus): void {
    if (this.status === status) return;
    this.status = status;
    this.onStatusChangeCallback?.(status);
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // ============================================
  // Conexión
  // ============================================

  public connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('🔌 WebSocket ya está conectado');
      return;
    }

    if (this.ws?.readyState === WebSocket.CONNECTING) {
      console.log('🔌 WebSocket ya está conectando');
      return;
    }

    this.shouldReconnect = true;
    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = this.handleOpen.bind(this);
      this.ws.onclose = this.handleClose.bind(this);
      this.ws.onerror = this.handleError.bind(this);
      this.ws.onmessage = this.handleMessage.bind(this);
    } catch (error) {
      console.error('❌ Error creando WebSocket:', error);
      this.setStatus('error');
      this.scheduleReconnect();
    }
  }

  public disconnect(): void {
    this.shouldReconnect = false;
    this.clearTimers();

    if (this.ws) {
      this.ws.close(1000, 'Cliente desconectado');
      this.ws = null;
    }

    this.setStatus('disconnected');
  }

  // ============================================
  // Handlers de eventos
  // ============================================

  private handleOpen(): void {
    console.log('🔌 WebSocket conectado');
    this.reconnectAttempts = 0;
    this.setStatus('connected');

    this.onOpenCallback?.();

    // Iniciar heartbeat
    this.startHeartbeat();

    // Enviar mensajes en cola
    this.flushMessageQueue();
  }

  private handleClose(event: CloseEvent): void {
    console.log('🔌 WebSocket desconectado', event.code, event.reason);
    this.setStatus('disconnected');
    this.stopHeartbeat();

    this.onCloseCallback?.(event);

    // Reconectar si no fue cierre intencional
    if (this.shouldReconnect && event.code !== 1000) {
      this.scheduleReconnect();
    }
  }

  private handleError(event: Event): void {
    console.error('❌ WebSocket error:', event);
    this.setStatus('error');
    this.onErrorCallback?.(event);
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);

      // Responder a pings automáticamente
      if (message.type === 'ping') {
        this.send({ type: 'pong' });
        return;
      }

      // Ignorar pongs
      if (message.type === 'pong') {
        return;
      }

      // Notificar a todos los listeners
      this.messageListeners.forEach((handler) => {
        try {
          handler(message);
        } catch (err) {
          console.error('Error en listener:', err);
        }
      });
    } catch (err) {
      console.error('Error parseando mensaje WebSocket:', err);
    }
  }

  // ============================================
  // Envío de mensajes
  // ============================================

  public send(message: WebSocketMessage): void {
    if (this.isConnected()) {
      try {
        this.ws!.send(JSON.stringify(message));
      } catch (err) {
        console.error('Error enviando mensaje:', err);
        this.messageQueue.push(message);
      }
    } else {
      // Guardar en cola para enviar al reconectar
      this.messageQueue.push(message);

      // Si la cola crece mucho, vaciar la mitad (los más viejos)
      if (this.messageQueue.length > 100) {
        this.messageQueue = this.messageQueue.slice(-50);
        console.warn('⚠️  Cola de mensajes llena, descartando antiguos');
      }
    }
  }

  private flushMessageQueue(): void {
    if (this.messageQueue.length === 0) return;

    console.log(`📤 Enviando ${this.messageQueue.length} mensajes en cola`);
    const queue = [...this.messageQueue];
    this.messageQueue = [];

    queue.forEach((msg) => {
      if (this.isConnected()) {
        this.ws!.send(JSON.stringify(msg));
      }
    });
  }

  // ============================================
  // Listeners
  // ============================================

  public onMessage(handler: MessageHandler): () => void {
    this.messageListeners.add(handler);
    // Devolver función de cleanup
    return () => {
      this.messageListeners.delete(handler);
    };
  }

  // ============================================
  // Reconexión
  // ============================================

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('❌ Máximo de intentos de reconexión alcanzado');
      this.setStatus('error');
      return;
    }

    // Backoff exponencial con tope
    const delay = Math.min(
      this.reconnectInterval * Math.pow(1.5, this.reconnectAttempts),
      30_000
    );

    this.reconnectAttempts += 1;

    console.log(
      `🔄 Reintentando conexión en ${delay}ms ` +
        `(intento ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.reconnectTimer = window.setTimeout(() => {
      this.connect();
    }, delay);
  }

  public forceReconnect(): void {
    console.log('🔄 Forzando reconexión manual');
    this.reconnectAttempts = 0;
    this.clearTimers();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.connect();
  }

  // ============================================
  // Heartbeat
  // ============================================

  private startHeartbeat(): void {
    this.stopHeartbeat();

    this.heartbeatTimer = window.setInterval(() => {
      if (this.isConnected()) {
        this.send({ type: 'ping' });
      }
    }, this.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer !== null) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  // ============================================
  // Cleanup
  // ============================================

  private clearTimers(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.stopHeartbeat();
  }

  public destroy(): void {
    this.shouldReconnect = false;
    this.clearTimers();
    this.messageListeners.clear();
    this.messageQueue = [];

    if (this.ws) {
      this.ws.close(1000, 'Cliente destruido');
      this.ws = null;
    }
  }
}

// ============================================
// Helper: crear instancia
// ============================================

/**
 * Crea una instancia de WebSocketClient con la URL por defecto.
 */
export function createWebSocketClient(
  options: Partial<WebSocketClientOptions> & { url?: string }
): WebSocketClient {
  const url =
    options.url ||
    import.meta.env.VITE_WS_URL ||
    'ws://localhost:8080/ws';

  return new WebSocketClient({
    url,
    ...options,
  });
}