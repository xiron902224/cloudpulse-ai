import { useEffect, useRef, useState, useCallback } from 'react';

import {
  WebSocketClient,
  type ConnectionStatus,
  type WebSocketMessage,
} from '../services/websocket';

// ============================================
// Tipos
// ============================================

interface UseWebSocketOptions {
  /** URL del WebSocket. Si no se pasa, usa VITE_WS_URL o localhost */
  url?: string;

  /** Callback cuando llega un mensaje */
  onMessage?: (message: WebSocketMessage) => void;

  /** Callback cuando se conecta */
  onOpen?: () => void;

  /** Callback cuando se desconecta */
  onClose?: (event: CloseEvent) => void;

  /** Callback cuando hay error */
  onError?: (event: Event) => void;

  /** Intervalo de reconexión base en ms */
  reconnectInterval?: number;

  /** Máximo de intentos de reconexión */
  maxReconnectAttempts?: number;

  /** Intervalo del heartbeat */
  heartbeatInterval?: number;

  /** Si es false, no conecta automáticamente */
  enabled?: boolean;
}

interface UseWebSocketReturn {
  /** Estado de la conexión */
  status: ConnectionStatus;

  /** Último mensaje recibido */
  lastMessage: WebSocketMessage | null;

  /** Historial de mensajes (últimos 50) */
  messages: WebSocketMessage[];

  /** Enviar un mensaje */
  send: (message: WebSocketMessage) => void;

  /** Reconectar manualmente */
  reconnect: () => void;

  /** Desconectar */
  disconnect: () => void;

  /** ¿Está conectado? */
  isConnected: boolean;
}

// ============================================
// Hook principal
// ============================================

/**
 * Hook para manejar una conexión WebSocket en React.
 *
 * @example
 * ```tsx
 * const { status, lastMessage, send } = useWebSocket({
 *   url: 'ws://localhost:8080/ws',
 *   onMessage: (msg) => console.log(msg),
 * });
 * ```
 */
export function useWebSocket(
  options: UseWebSocketOptions = {}
): UseWebSocketReturn {
  const {
    url,
    onMessage,
    onOpen,
    onClose,
    onError,
    reconnectInterval,
    maxReconnectAttempts,
    heartbeatInterval,
    enabled = true,
  } = options;

  // ---------- Estado reactivo ----------
  const [status, setStatus] = useState<ConnectionStatus>('connecting');
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const [messages, setMessages] = useState<WebSocketMessage[]>([]);

  // ---------- Refs ----------
  const clientRef = useRef<WebSocketClient | null>(null);

  // Callbacks en refs para no causar reconexiones
  const onMessageRef = useRef(onMessage);
  const onOpenRef = useRef(onOpen);
  const onCloseRef = useRef(onClose);
  const onErrorRef = useRef(onError);

  // Actualizar refs cuando cambien los callbacks
  useEffect(() => {
    onMessageRef.current = onMessage;
    onOpenRef.current = onOpen;
    onCloseRef.current = onClose;
    onErrorRef.current = onError;
  }, [onMessage, onOpen, onClose, onError]);

  // ---------- Crear cliente WebSocket ----------
  useEffect(() => {
    if (!enabled) {
      setStatus('disconnected');
      return;
    }

    // Crear cliente
    const client = new WebSocketClient({
      url:
        url ||
        import.meta.env.VITE_WS_URL ||
        'ws://localhost:8080/ws',
      reconnectInterval,
      maxReconnectAttempts,
      heartbeatInterval,

      onStatusChange: (newStatus) => {
        setStatus(newStatus);
      },

      onMessage: (message) => {
        // Actualizar último mensaje
        setLastMessage(message);

        // Actualizar historial (máximo 50)
        setMessages((prev) => {
          const next = [...prev, message];
          return next.length > 50 ? next.slice(-50) : next;
        });

        // Llamar callback del usuario
        onMessageRef.current?.(message);
      },

      onOpen: () => {
        onOpenRef.current?.();
      },

      onClose: (event) => {
        onCloseRef.current?.(event);
      },

      onError: (event) => {
        onErrorRef.current?.(event);
      },
    });

    clientRef.current = client;
    client.connect();

    // Cleanup
    return () => {
      client.destroy();
      clientRef.current = null;
    };
  }, [
    url,
    enabled,
    reconnectInterval,
    maxReconnectAttempts,
    heartbeatInterval,
  ]);

  // ---------- Enviar mensaje ----------
  const send = useCallback((message: WebSocketMessage) => {
    clientRef.current?.send(message);
  }, []);

  // ---------- Reconectar ----------
  const reconnect = useCallback(() => {
    clientRef.current?.forceReconnect();
  }, []);

  // ---------- Desconectar ----------
  const disconnect = useCallback(() => {
    clientRef.current?.disconnect();
  }, []);

  // ---------- ¿Conectado? ----------
  const isConnected = status === 'connected';

  return {
    status,
    lastMessage,
    messages,
    send,
    reconnect,
    disconnect,
    isConnected,
  };
}