package models

import (
    "time"
)

// Constantes para severidad
const (
    SeverityCritical = "critical"
    SeverityWarning  = "warning"
    SeverityInfo     = "info"
)

// Constantes para estado
const (
    StatusActive       = "active"
    StatusResolved     = "resolved"
    StatusAcknowledged = "acknowledged"
)

// Alert representa una alerta del sistema
type Alert struct {
    ID        string     `json:"id"`
    Severity  string     `json:"severity"`
    Message   string     `json:"message"`
    Status    string     `json:"status"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt *time.Time `json:"updated_at,omitempty"`  // ← CAMBIADO a puntero
}