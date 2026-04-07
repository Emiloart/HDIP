#![forbid(unsafe_code)]

mod credential;
mod error;
mod identifier;
mod presentation;
mod validation;

pub use credential::{
    CredentialMetadata, CredentialStatusDescriptor, CredentialType, StatusMethod, Timestamp,
};
pub use error::IdentityError;
pub use identifier::{Identifier, IdentifierKind, IdentifierRef};
pub use presentation::{
    ConstraintRule, PresentationConstraint, PresentationRequest, PresentationResponse,
    PresentedCredential,
};
pub use validation::{CredentialState, CredentialValidator, PresentationValidator, StatusChecker};
