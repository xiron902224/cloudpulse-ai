use sysinfo::{Disk};


pub fn collect(sys: &mut Disk)->f64{
    sys.refresh();
    let usage = sys.usage().total_written_bytes;
    (usage as f64)
}