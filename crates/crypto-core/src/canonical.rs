use std::fmt::{Debug, Formatter};

use crate::CryptoError;

#[derive(Clone, Eq, PartialEq, Ord, PartialOrd, Hash)]
pub struct CanonicalBytes(Vec<u8>);

impl CanonicalBytes {
    pub fn from_vec(bytes: Vec<u8>) -> Result<Self, CryptoError> {
        if bytes.is_empty() {
            return Err(CryptoError::EmptyCanonicalBytes);
        }

        Ok(Self(bytes))
    }

    pub fn from_slice(bytes: &[u8]) -> Result<Self, CryptoError> {
        Self::from_vec(bytes.to_vec())
    }

    pub fn as_slice(&self) -> &[u8] {
        self.0.as_slice()
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }
}

impl AsRef<[u8]> for CanonicalBytes {
    fn as_ref(&self) -> &[u8] {
        self.as_slice()
    }
}

impl Debug for CanonicalBytes {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("CanonicalBytes")
            .field("len", &self.0.len())
            .finish()
    }
}

#[cfg(test)]
mod tests {
    use super::CanonicalBytes;
    use crate::CryptoError;

    #[test]
    fn rejects_empty_payloads() {
        assert_eq!(
            CanonicalBytes::from_vec(Vec::new()),
            Err(CryptoError::EmptyCanonicalBytes)
        );
    }

    #[test]
    fn tracks_length_without_exposing_content() {
        let value = CanonicalBytes::from_slice(b"hdip").expect("canonical bytes");
        assert_eq!(value.len(), 4);
        assert_eq!(format!("{value:?}"), "CanonicalBytes { len: 4 }");
    }
}
