mod client;
mod collector;

use anyhow::Result;
use axum::{routing::get, Json, Router};
use serde::Serialize;
use std::env;
use std::net::SocketAddr;
use std::time::Duration;
use sysinfo::System;
use tokio::time::interval;
use tracing::{info, warn, Level};
use tracing_subscriber;

use crate::client::ApiClient;
use crate::collector::Metric;

/// Respuesta del endpoint /health
#[derive(Serialize)]
struct HealthResponse {
    status: &'static str,
    service: &'static str,
    version: &'static str,
}

/// Handler para /health
async fn health_check() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy",
        service: "cloudpulse-collector",
        version: "0.1.0",
    })
}

/// Punto de entrada principal
#[tokio::main]
async fn main() -> Result<()> {
    // 1. Configurar logging estructurado
    tracing_subscriber::fmt()
        .with_max_level(Level::INFO)
        .with_target(false)
        .init();

    info!("🚀 Iniciando CloudPulse Collector");

    // 2. Leer configuración desde variables de entorno
    let api_url = env::var("API_URL")
        .unwrap_or_else(|_| "http://localhost:8080".to_string());

    let collector_port = env::var("COLLECTOR_PORT")
        .unwrap_or_else(|_| "9090".to_string());

    let collect_interval: u64 = env::var("COLLECT_INTERVAL")
        .unwrap_or_else(|_| "10".to_string())
        .parse()
        .unwrap_or(10);

    info!("📡 API URL: {}", api_url);
    info!("🔌 Puerto HTTP: {}", collector_port);
    info!("⏱️  Intervalo de recolección: {}s", collect_interval);

    // 3. Crear cliente API
    let api_client = ApiClient::new(api_url.clone())?;

    // 4. Verificar disponibilidad de la API
    if api_client.health_check().await {
        info!("✅ API disponible");
    } else {
        warn!("⚠️  API no disponible, se reintentará al enviar métricas");
    }

    // 5. Configurar servidor HTTP
    let port: u16 = collector_port.parse().unwrap_or(9090);
    let addr = SocketAddr::from(([0, 0, 0, 0], port));

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/metrics", get(health_check)); // TODO: métricas Prometheus

    // 6. Iniciar tarea de recolección
    let collector_task = tokio::spawn(async move {
        let mut sys = System::new_all();
        let mut ticker = interval(Duration::from_secs(collect_interval));

        // Primera lectura (inicializar contadores de CPU)
        let _ = collector::cpu::collect(&mut sys);

        loop {
            ticker.tick().await;

            // Recolectar métrica
            let metric = Metric::collect(&mut sys);

            info!(
                "📊 CPU: {:.2}% | Mem: {:.2}% | Disco: {:.2}%",
                metric.cpu_usage, metric.memory_usage, metric.disk_usage
            );

            // Enviar a la API
            if let Err(e) = api_client.send_metric(&metric).await {
                warn!("⚠️  Error enviando métrica: {}", e);
            }
        }
    });

    // 7. Iniciar servidor HTTP
    info!("🌐 Servidor HTTP escuchando en http://{}", addr);
    info!("   Health: http://{}/health", addr);
    info!("   Metrics: http://{}/metrics", addr);

    let server = axum::serve(
        tokio::net::TcpListener::bind(addr).await?,
        app.into_make_service(),
    );

    // 8. Ejecutar ambos en paralelo
    tokio::select! {
        result = server => {
            if let Err(e) = result {
                warn!("⚠️  Servidor HTTP detenido con error: {}", e);
            }
            info!("Servidor HTTP detenido");
        }
        _ = collector_task => {
            info!("Tarea de recolección detenida");
        }
        _ = tokio::signal::ctrl_c() => {
            info!("🛑 Señal Ctrl+C recibida, cerrando...");
        }
    }

    info!("👋 CloudPulse Collector finalizado");
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_health_check() {
        let response = health_check().await;
        let json = response.0;

        assert_eq!(json.status, "healthy");
        assert_eq!(json.service, "cloudpulse-collector");
        assert_eq!(json.version, "0.1.0");
    }
}