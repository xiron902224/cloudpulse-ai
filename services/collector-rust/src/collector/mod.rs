pub mod cpu;
pub mod memory;

use serde::{Deserialize, Serialize};
use sysinfo::System;

/// Métrica del sistema
///
/// Representa una lectura de todos los recursos del sistema
/// en un momento específico.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Metric {
    /// Uso de CPU en porcentaje (0-100)
    pub cpu_usage: f64,

    /// Uso de memoria en porcentaje (0-100)
    pub memory_usage: f64,

    /// Uso de disco en porcentaje (0-100)
    pub disk_usage: f64,

    /// Tráfico de red entrante (KB/s)
    pub network_in: f64,

    /// Tráfico de red saliente (KB/s)
    pub network_out: f64,

    /// Timestamp Unix (segundos desde 1970)
    pub timestamp: i64,
}

impl Metric {
    /// Crea una nueva métrica recolectando datos del sistema
    ///
    /// # Argumentos
    /// * `sys` - Referencia mutable al sistema
    ///
    /// # Retorna
    /// Una instancia de `Metric` con los valores actuales
    pub fn collect(sys: &mut System) -> Self {
        Self {
            cpu_usage: cpu::collect(sys),
            memory_usage: memory::collect(sys),
            disk_usage: memory::collect_disk(),
            network_in: 0.0,
            network_out: 0.0,
            timestamp: chrono::Utc::now().timestamp(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_metric_collect() {
        let mut sys = System::new_all();

        // Esperar para tener datos reales de CPU
        std::thread::sleep(std::time::Duration::from_millis(200));

        let metric = Metric::collect(&mut sys);

        assert!(metric.cpu_usage >= 0.0);
        assert!(metric.cpu_usage <= 100.0);
        assert!(metric.memory_usage >= 0.0);
        assert!(metric.memory_usage <= 100.0);
        assert!(metric.timestamp > 0);
    }

    #[test]
    fn test_metric_serialization() {
        let metric = Metric {
            cpu_usage: 45.5,
            memory_usage: 62.3,
            disk_usage: 71.0,
            network_in: 100.0,
            network_out: 50.0,
            timestamp: 1705310400,
        };

        // Serializar a JSON
        let json = serde_json::to_string(&metric).unwrap();

        // Verificar que contiene los campos
        assert!(json.contains("cpu_usage"));
        assert!(json.contains("45.5"));

        // Deserializar
        let parsed: Metric = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed.cpu_usage, 45.5);
    }

    #[test]
    fn test_metric_deserialization_from_json() {
        let json = r#"{
            "cpu_usage": 25.5,
            "memory_usage": 60.0,
            "disk_usage": 45.0,
            "network_in": 100.0,
            "network_out": 50.0,
            "timestamp": 1705310400
        }"#;

        let metric: Metric = serde_json::from_str(json).unwrap();

        assert_eq!(metric.cpu_usage, 25.5);
        assert_eq!(metric.memory_usage, 60.0);
        assert_eq!(metric.timestamp, 1705310400);
    }
}