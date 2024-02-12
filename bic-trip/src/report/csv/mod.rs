// A small utility module to encapsulate the logic of writing a Vec<LogEntry> to a CSV file.

use std::error::Error;

use crate::log_parser::LogEntry;

pub fn write_csv(data: &Option<Vec<LogEntry>>, label: &str) -> Result<(), Box<dyn Error>> {
    if data.is_none() {
        Ok(())
    } else {
        let data = data.as_ref().unwrap();
        let mut writer = csv::Writer::from_path(label)?;
        for entry in data {
            writer.serialize(entry)?;
        }
        writer.flush()?;
        Ok(())
    }
}
