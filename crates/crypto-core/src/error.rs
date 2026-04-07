use std::error::Error;
use std::fmt::{Display, Formatter};

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum CryptoError {
    EmptyCanonicalBytes,
    InvalidKeyId,
    EmptySignature,
    UnsupportedAlgorithm(&'static str),
    SigningUnavailable,
    VerificationUnavailable,
}

impl Display for CryptoError {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::EmptyCanonicalBytes => f.write_str("canonical bytes must not be empty"),
            Self::InvalidKeyId => f.write_str("key id must not be empty"),
            Self::EmptySignature => f.write_str("signature bytes must not be empty"),
            Self::UnsupportedAlgorithm(name) => {
                write!(f, "algorithm is not supported in this context: {name}")
            }
            Self::SigningUnavailable => f.write_str("signing implementation is unavailable"),
            Self::VerificationUnavailable => {
                f.write_str("verification implementation is unavailable")
            }
        }
    }
}

impl Error for CryptoError {}
