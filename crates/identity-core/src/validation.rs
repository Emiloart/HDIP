use crate::{
    CredentialMetadata, CredentialStatusDescriptor, IdentityError, PresentationRequest,
    PresentationResponse,
};

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum CredentialState {
    Valid,
    Revoked,
    Suspended,
    Unknown,
}

pub trait CredentialValidator {
    fn validate_metadata(&self, metadata: &CredentialMetadata) -> Result<(), IdentityError>;
}

pub trait PresentationValidator {
    fn validate_exchange(
        &self,
        request: &PresentationRequest,
        response: &PresentationResponse,
    ) -> Result<(), IdentityError>;
}

pub trait StatusChecker {
    fn check_status(
        &self,
        descriptor: &CredentialStatusDescriptor,
    ) -> Result<CredentialState, IdentityError>;
}
