use std::error::Error;
use std::fmt::{Display, Formatter};

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum IdentityError {
    EmptyValue(&'static str),
    MissingCredentialTypes,
    MissingConstraints,
    UnsupportedStatusMethod,
    ValidationUnavailable,
    StatusUnavailable,
}

impl Display for IdentityError {
    fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::EmptyValue(field) => write!(f, "{field} must not be empty"),
            Self::MissingCredentialTypes => f.write_str("credential metadata must include types"),
            Self::MissingConstraints => {
                f.write_str("presentation request must include at least one constraint")
            }
            Self::UnsupportedStatusMethod => {
                f.write_str("status method is not supported in this context")
            }
            Self::ValidationUnavailable => f.write_str("validation implementation is unavailable"),
            Self::StatusUnavailable => f.write_str("status implementation is unavailable"),
        }
    }
}

impl Error for IdentityError {}
