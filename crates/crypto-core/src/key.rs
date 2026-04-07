use crate::CryptoError;

#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub enum KeyPurpose {
    Authentication,
    Assertion,
    Encryption,
    Recovery,
}

#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct KeyId(String);

impl KeyId {
    pub fn new(value: impl Into<String>) -> Result<Self, CryptoError> {
        let value = value.into();
        if value.trim().is_empty() {
            return Err(CryptoError::InvalidKeyId);
        }

        Ok(Self(value))
    }

    pub fn as_str(&self) -> &str {
        self.0.as_str()
    }
}

pub trait KeyHandle {
    fn key_id(&self) -> &KeyId;
    fn purpose(&self) -> KeyPurpose;
}

#[cfg(test)]
mod tests {
    use crate::{CryptoError, KeyId};

    #[test]
    fn rejects_blank_key_ids() {
        assert_eq!(KeyId::new("  "), Err(CryptoError::InvalidKeyId));
    }
}
