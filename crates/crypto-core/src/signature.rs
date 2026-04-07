use std::fmt::{Debug, Formatter};

use crate::{CanonicalBytes, CryptoError, KeyId};

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum SignatureAlgorithm {
    Ed25519,
    Es256,
}

#[derive(Clone, Eq, PartialEq)]
pub struct SignatureBytes(Vec<u8>);

impl SignatureBytes {
    pub fn new(bytes: Vec<u8>) -> Result<Self, CryptoError> {
        if bytes.is_empty() {
            return Err(CryptoError::EmptySignature);
        }

        Ok(Self(bytes))
    }

    pub fn as_slice(&self) -> &[u8] {
        self.0.as_slice()
    }
}

impl Debug for SignatureBytes {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("SignatureBytes")
            .field("len", &self.0.len())
            .finish()
    }
}

pub trait Signer {
    fn key_id(&self) -> &KeyId;
    fn algorithm(&self) -> SignatureAlgorithm;
    fn sign(&self, payload: &CanonicalBytes) -> Result<SignatureBytes, CryptoError>;
}

pub trait Verifier {
    fn key_id(&self) -> &KeyId;
    fn algorithm(&self) -> SignatureAlgorithm;
    fn verify(
        &self,
        payload: &CanonicalBytes,
        signature: &SignatureBytes,
    ) -> Result<(), CryptoError>;
}

#[cfg(test)]
mod tests {
    use crate::{CryptoError, SignatureBytes};

    #[test]
    fn rejects_empty_signatures() {
        assert_eq!(
            SignatureBytes::new(Vec::new()),
            Err(CryptoError::EmptySignature)
        );
    }
}
