
use sysinfo::System;
pub fn collect(sys: &mut System)->f64{
    sys.refresh_memory();
    let total = sys.total_memory();
    let usage = sys.used_memory();
    if total == 0 {
        return 0.0;
    }
    let percentage =(usage as f64/total as f64)*100.0;
    (percentage *100.0).round()/100.0
    
}
pub fn collect_disk() -> f64 {
    // TODO: Implementar con sysinfo::Disks
    0.0
}
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_collect_memory() {
        let mut sys = System::new_all();
        let memory = collect(&mut sys);

        assert!(
            memory >= 0.0,
            "Memoria no puede ser negativa, obtenido: {}",
            memory
        );
        assert!(
            memory <= 100.0,
            "Memoria no puede ser mayor a 100, obtenido: {}",
            memory
        );
    }

    #[test]
    fn test_collect_disk() {
        let disk = collect_disk();
        assert!(disk >= 0.0);
        assert!(disk <= 100.0);
    }

    #[test]
    fn test_memory_is_realistic() {
        let mut sys = System::new_all();
        let memory = collect(&mut sys);

        // En un sistema real, el uso de memoria nunca es exactamente 0
        // a menos que estés en un sistema muy peculiar
        assert!(
            memory > 0.0,
            "Uso de memoria debería ser mayor a 0 en un sistema real"
        );
    }
}