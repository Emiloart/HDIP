use crate::{IdentifierRef, IdentityError};

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct Timestamp(String);

impl Timestamp {
    pub fn new(value: impl Into<String>) -> Result<Self, IdentityError> {
        let value = value.into();
        if value.trim().is_empty() {
            return Err(IdentityError::EmptyValue("timestamp"));
        }

        Ok(Self(value))
    }

    pub fn as_str(&self) -> &str {
        self.0.as_str()
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CredentialType(String);

impl CredentialType {
    pub fn new(value: impl Into<String>) -> Result<Self, IdentityError> {
        let value = value.into();
        if value.trim().is_empty() {
            return Err(IdentityError::EmptyValue("credential type"));
        }

        Ok(Self(value))
    }

    pub fn as_str(&self) -> &str {
        self.0.as_str()
    }
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum StatusMethod {
    BitstringStatusList,
    Custom,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CredentialStatusDescriptor {
    pub method: StatusMethod,
    pub endpoint: String,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CredentialMetadata {
    pub id: IdentifierRef,
    pub issuer: IdentifierRef,
    pub credential_types: Vec<CredentialType>,
    pub issued_at: Timestamp,
    pub expires_at: Option<Timestamp>,
    pub status: Option<CredentialStatusDescriptor>,
}

impl CredentialMetadata {
    pub fn validate(&self) -> Result<(), IdentityError> {
        if self.credential_types.is_empty() {
            return Err(IdentityError::MissingCredentialTypes);
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::{
        CredentialMetadata, CredentialType, Identifier, IdentifierKind, IdentifierRef, Timestamp,
    };

    #[test]
    fn validates_presence_of_credential_types() {
        let metadata = CredentialMetadata {
            id: IdentifierRef::new(
                IdentifierKind::Opaque,
                Identifier::new("cred-1").expect("id"),
            ),
            issuer: IdentifierRef::new(
                IdentifierKind::DidWeb,
                Identifier::new("did:web:issuer.example").expect("issuer"),
            ),
            credential_types: vec![CredentialType::new("KycCredential").expect("type")],
            issued_at: Timestamp::new("2026-04-06T21:00:00Z").expect("timestamp"),
            expires_at: None,
            status: None,
        };

        assert!(metadata.validate().is_ok());
    }
}
