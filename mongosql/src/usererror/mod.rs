pub use usererrordisplay_impl::UserErrorDisplay;
pub trait UserError {
    fn code(&self) -> u32;
    fn user_message(&self) -> Option<String>;
    fn technical_message(&self) -> String;
}

#[cfg(test)]
mod test;
