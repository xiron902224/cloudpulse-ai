package handlers

import (
    "encoding/json"
    "log"
    "sync"
    "time"

    "github.com/gofiber/contrib/websocket"
    "github.com/gofiber/fiber/v2"

    "github.com/xiron902224/cloudpulse-ai/api/internal/models"
)

// ============================================
// Hub de WebSocket
// ============================================

// Hub mantiene el registro de clientes conectados
type Hub struct {
    clients   map[*websocket.Conn]bool
    mu        sync.RWMutex
    broadcast chan []byte
}

// Hub global (uno por proceso)
var hub = &Hub{
    clients:   make(map[*websocket.Conn]bool),
    broadcast: make(chan []byte, 256),
}

// ============================================
// Inicialización
// ============================================

// InitWebSocketHub inicia el hub en background
func InitWebSocketHub() {
    go hub.run()
    log.Println("✅ WebSocket Hub iniciado")
}

// run es el loop principal del hub
func (h *Hub) run() {
    for message := range h.broadcast {
        // Copia de clientes para evitar race conditions
        h.mu.RLock()
        clients := make([]*websocket.Conn, 0, len(h.clients))
        for client := range h.clients {
            clients = append(clients, client)
        }
        h.mu.RUnlock()

        // Enviar a cada cliente
        for _, client := range clients {
            if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
                log.Printf("⚠️  Error enviando a cliente: %v", err)
                client.Close()

                h.mu.Lock()
                delete(h.clients, client)
                h.mu.Unlock()
            }
        }
    }
}

// ============================================
// Handlers HTTP
// ============================================

// WebSocketUpgrade verifica que la petición sea WebSocket
func WebSocketUpgrade(c *fiber.Ctx) error {
    if websocket.IsWebSocketUpgrade(c) {
        return c.Next()
    }
    return fiber.ErrUpgradeRequired
}

// WebSocketHandler maneja una conexión WebSocket
func WebSocketHandler(c *websocket.Conn) {
    // Registrar cliente
    hub.mu.Lock()
    hub.clients[c] = true
    clientCount := len(hub.clients)
    hub.mu.Unlock()

    log.Printf("🔌 Cliente conectado (total: %d)", clientCount)

    // Defer: eliminar cliente al desconectar
    defer func() {
        hub.mu.Lock()
        delete(hub.clients, c)
        clientCount := len(hub.clients)
        hub.mu.Unlock()
        c.Close()
        log.Printf("🔌 Cliente desconectado (total: %d)", clientCount)
    }()

    // Enviar mensaje de bienvenida
    welcome := map[string]interface{}{
        "type":    "connected",
        "message": "Conectado a CloudPulse AI",
        "time":    time.Now().Format(time.RFC3339),
    }
    if data, err := json.Marshal(welcome); err == nil {
        c.WriteMessage(websocket.TextMessage, data)
    }

    // Loop de lectura (detecta desconexiones y procesa mensajes)
    for {
        messageType, message, err := c.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("Error de conexión: %v", err)
            }
            break
        }

        handleClientMessage(c, messageType, message)
    }
}

// handleClientMessage procesa mensajes del cliente
func handleClientMessage(c *websocket.Conn, _ int, message []byte) {
    var msg map[string]interface{}
    if err := json.Unmarshal(message, &msg); err != nil {
        return
    }

    msgType, _ := msg["type"].(string)
    switch msgType {
    case "ping":
        pong := map[string]string{"type": "pong"}
        if data, err := json.Marshal(pong); err == nil {
            c.WriteMessage(websocket.TextMessage, data)
        }
    case "subscribe":
        log.Printf("Cliente suscrito a: %v", msg["channel"])
    }
}

// ============================================
// API pública para broadcast
// ============================================

// BroadcastMetric envía una métrica a todos los clientes
func BroadcastMetric(metric *models.Metric) {
    message := map[string]interface{}{
        "type": "metric",
        "data": metric,
    }
    if data, err := json.Marshal(message); err == nil {
        hub.broadcast <- data
    }
}

// BroadcastAlert envía una alerta a todos los clientes
func BroadcastAlert(alert *models.Alert) {
    message := map[string]interface{}{
        "type": "alert",
        "data": alert,
    }
    if data, err := json.Marshal(message); err == nil {
        hub.broadcast <- data
    }
}

// HubClientCount devuelve el número de clientes conectados
func HubClientCount() int {
    hub.mu.RLock()
    defer hub.mu.RUnlock()
    return len(hub.clients)
}