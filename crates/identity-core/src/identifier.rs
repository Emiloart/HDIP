use crate::IdentityError;

#[derive(Clone, Copy, Debug, Eq, PartialEq, Hash)]
pub enum IdentifierKind {
    DidWeb,
    DidKey,
    HttpsUrl,
    Opaque,
}

#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct Identifier(String);

impl Identifier {
    pub fn new(value: impl Into<String>) -> Result<Self, IdentityError> {
        let value = value.into();
        if value.trim().is_empty() {
            return Err(IdentityError::EmptyValue("identifier"));
        }

        Ok(Self(value))
    }

    pub fn as_str(&self) -> &str {
        self.0.as_str()
    }
}

#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct IdentifierRef {
    pub kind: IdentifierKind,
    pub value: Identifier,
}

impl IdentifierRef {
    pub fn new(kind: IdentifierKind, value: Identifier) -> Self {
        Self { kind, value }
    }
}

#[cfg(test)]
mod tests {
    use crate::{Identifier, IdentityError};

    #[test]
    fn rejects_blank_identifier() {
        assert_eq!(
            Identifier::new("  "),
            Err(IdentityError::EmptyValue("identifier"))
        );
    }
}
