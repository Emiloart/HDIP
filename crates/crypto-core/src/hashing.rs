use std::fmt::{Debug, Formatter};

use sha2::{Digest, Sha256};

use crate::CanonicalBytes;

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum HashAlgorithm {
    Sha256,
}

#[derive(Clone, Copy, Eq, PartialEq)]
pub struct DigestBytes<const N: usize>([u8; N]);

impl<const N: usize> DigestBytes<N> {
    pub fn as_bytes(&self) -> &[u8; N] {
        &self.0
    }

    pub fn to_hex(self) -> String {
        self.0.iter().map(|byte| format!("{byte:02x}")).collect()
    }
}

impl<const N: usize> Debug for DigestBytes<N> {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        f.debug_tuple("DigestBytes").field(&self.to_hex()).finish()
    }
}

pub fn sha256_digest(bytes: &CanonicalBytes) -> DigestBytes<32> {
    let mut hasher = Sha256::new();
    hasher.update(bytes.as_slice());
    let digest = hasher.finalize();
    let mut value = [0_u8; 32];
    value.copy_from_slice(digest.as_slice());
    DigestBytes(value)
}

#[cfg(test)]
mod tests {
    use super::sha256_digest;
    use crate::CanonicalBytes;

    #[test]
    fn computes_known_digest() {
        let digest = sha256_digest(&CanonicalBytes::from_slice(b"hdip").expect("canonical"));
        assert_eq!(
            digest.to_hex(),
            "fd89fd7f11a61aa8c5cdf510a224150fe47116b2baaa8b11dd45f20a909b5a08"
        );
    }
}
