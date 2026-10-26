use sysinfo::System;

/// Recolecta el uso de CPU del sistema
///
/// # Argumentos
/// * `sys` - Referencia mutable al sistema
///
/// # Retorna
/// Uso de CPU en porcentaje (0-100)
pub fn collect(sys: &mut System) -> f64 {
    // Actualizar información de CPU
    sys.refresh_cpu_usage();

    // Obtener uso promedio de todos los núcleos
    let usage = sys.global_cpu_usage();
    // Convertir a f64 y redondear a 2 decimales
    (usage as f64 * 100.0).round() / 100.0
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_collect_cpu() {
        let mut sys = System::new_all();

        // Primera llamada (inicializa contadores)
        let _ = collect(&mut sys);

        // Esperar para tener una segunda lectura
        std::thread::sleep(std::time::Duration::from_millis(200));

        // Segunda llamada (datos reales)
        let cpu = collect(&mut sys);

        // Verificar rango válido
        assert!(cpu >= 0.0, "CPU no puede ser negativo");
        assert!(cpu <= 100.0, "CPU no puede ser mayor a 100");
    }
}