package models

import (
    "encoding/json"
    "testing"
)

func TestMetricJSONSerialization(t *testing.T) {
    // Crear métrica de prueba
    metric := Metric{
        ID:          "abc123",
        CPUUsage:    45.5,
        MemoryUsage: 62.3,
        DiskUsage:   71.0,
        NetworkIn:   250.5,
        NetworkOut:  120.8,
        Timestamp:   1705310400,
    }

    // Convertir a JSON
    jsonData, err := json.Marshal(metric)
    if err != nil {
        t.Fatalf("Error serializando: %v", err)
    }

    // Verificar JSON
    expected := `{"id":"abc123","cpu_usage":45.5,"memory_usage":62.3,"disk_usage":71,"network_in":250.5,"network_out":120.8,"timestamp":1705310400,"created_at":"0001-01-01T00:00:00Z"}`
    
    if string(jsonData) != expected {
        t.Errorf("JSON esperado: %s\nJSON obtenido: %s", expected, string(jsonData))
    }
}

func TestMetricDeserialization(t *testing.T) {
    // JSON de entrada
    jsonInput := `{"id":"abc123","cpu_usage":45.5,"memory_usage":62.3,"disk_usage":71.0,"network_in":250.5,"network_out":120.8,"timestamp":1705310400}`

    // Convertir a struct
    var metric Metric
    if err := json.Unmarshal([]byte(jsonInput), &metric); err != nil {
        t.Fatalf("Error deserializando: %v", err)
    }

    // Verificar campos
    if metric.ID != "abc123" {
        t.Errorf("ID esperado: abc123, obtenido: %s", metric.ID)
    }

    if metric.CPUUsage != 45.5 {
        t.Errorf("CPUUsage esperado: 45.5, obtenido: %f", metric.CPUUsage)
    }

    if metric.MemoryUsage != 62.3 {
        t.Errorf("MemoryUsage esperado: 62.3, obtenido: %f", metric.MemoryUsage)
    }
}

func TestAlertConstants(t *testing.T) {
    // Verificar que las constantes tienen los valores correctos
    if SeverityCritical != "critical" {
        t.Errorf("SeverityCritical debería ser 'critical'")
    }
    
    if SeverityWarning != "warning" {
        t.Errorf("SeverityWarning debería ser 'warning'")
    }
    
    if SeverityInfo != "info" {
        t.Errorf("SeverityInfo debería ser 'info'")
    }
    
    if StatusActive != "active" {
        t.Errorf("StatusActive debería ser 'active'")
    }
    
    if StatusResolved != "resolved" {
        t.Errorf("StatusResolved debería ser 'resolved'")
    }
    
    if StatusAcknowledged != "acknowledged" {
        t.Errorf("StatusAcknowledged debería ser 'acknowledged'")
    }
}