use crate::air::FieldRef;

impl FieldRef {
    pub(crate) fn root_parent(&self) -> String {
        match &self.parent {
            Some(parent) => parent.root_parent(),
            None => self.name.clone(),
        }
    }
}
