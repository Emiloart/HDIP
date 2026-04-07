use std::fmt::{Debug, Formatter};

#[derive(Clone, Eq, PartialEq)]
pub struct Sensitive<T>(T);

impl<T> Sensitive<T> {
    pub fn new(value: T) -> Self {
        Self(value)
    }

    pub fn expose_ref(&self) -> &T {
        &self.0
    }

    pub fn into_inner(self) -> T {
        self.0
    }
}

impl<T> Debug for Sensitive<T> {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        f.write_str("Sensitive([REDACTED])")
    }
}

#[cfg(test)]
mod tests {
    use crate::Sensitive;

    #[test]
    fn redacts_debug_output() {
        let value = Sensitive::new(String::from("secret"));
        assert_eq!(format!("{value:?}"), "Sensitive([REDACTED])");
    }
}
