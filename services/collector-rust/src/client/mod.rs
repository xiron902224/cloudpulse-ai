use anyhow::{Context, Result};
use reqwest::Client;
use std::time::Duration;
use tracing::{error, info, warn};

use crate::collector::Metric;

/// Cliente para comunicarse con la API
pub struct ApiClient {
    client: Client,
    base_url: String,
}

impl ApiClient {
    /// Crea un nuevo cliente API
    ///
    /// # Argumentos
    /// * `base_url` - URL base de la API (ej: "http://localhost:8080")
    ///
    /// # Retorna
    /// Un `ApiClient` configurado o un error
    pub fn new(base_url: String) -> Result<Self> {
        let client = Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .context("Error creando cliente HTTP")?;

        Ok(Self { client, base_url })
    }

    /// Envía una métrica a la API
    ///
    /// # Argumentos
    /// * `metric` - Métrica a enviar
    ///
    /// # Retorna
    /// `Ok(())` si la métrica se envió correctamente
    pub async fn send_metric(&self, metric: &Metric) -> Result<()> {
        let url = format!("{}/api/v1/metrics", self.base_url);

        let response = self
            .client
            .post(&url)
            .json(metric)
            .send()
            .await
            .context("Error enviando métrica a la API")?;

        // Verificar que la respuesta sea exitosa
        if !response.status().is_success() {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            error!("API respondió con error: {} - {}", status, body);
            anyhow::bail!("API error: {}", status);
        }

        info!(
            "✅ Métrica enviada: CPU {:.2}% | Mem {:.2}%",
            metric.cpu_usage, metric.memory_usage
        );

        Ok(())
    }

    /// Verifica que la API esté disponible
    ///
    /// # Retorna
    /// `true` si la API responde a `/health`
    pub async fn health_check(&self) -> bool {
        let url = format!("{}/health", self.base_url);

        match self.client.get(&url).send().await {
            Ok(response) => {
                if response.status().is_success() {
                    info!("✅ API disponible");
                    true
                } else {
                    warn!("⚠️  API respondió con error: {}", response.status());
                    false
                }
            }
            Err(e) => {
                warn!("⚠️  No se pudo conectar a la API: {}", e);
                false
            }
        }
    }

    /// Obtiene la URL base configurada
    pub fn base_url(&self) -> &str {
        &self.base_url
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_client_creation() {
        let client = ApiClient::new("http://localhost:8080".to_string());
        assert!(client.is_ok());

        let client = client.unwrap();
        assert_eq!(client.base_url(), "http://localhost:8080");
    }

    #[test]
    fn test_client_with_invalid_url() {
        // No valida la URL al construir, solo al hacer la petición
        let client = ApiClient::new("not-a-valid-url".to_string());
        assert!(client.is_ok());
    }

    #[tokio::test]
    async fn test_health_check_unreachable() {
        // Puerto improbable que no tenga nada corriendo
        let client = ApiClient::new("http://localhost:59999".to_string()).unwrap();
        let available = client.health_check().await;
        assert!(!available);
    }
}