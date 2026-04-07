use crate::{IdentifierRef, IdentityError, Timestamp};

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum ConstraintRule {
    Predicate(String),
    AllOf(Vec<String>),
    AnyOf(Vec<String>),
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct PresentationConstraint {
    pub id: String,
    pub description: String,
    pub rule: ConstraintRule,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct PresentationRequest {
    pub id: String,
    pub verifier: IdentifierRef,
    pub purpose: String,
    pub nonce: String,
    pub constraints: Vec<PresentationConstraint>,
}

impl PresentationRequest {
    pub fn validate(&self) -> Result<(), IdentityError> {
        if self.constraints.is_empty() {
            return Err(IdentityError::MissingConstraints);
        }

        Ok(())
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct PresentedCredential {
    pub credential_id: IdentifierRef,
    pub disclosed_fields: Vec<String>,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct PresentationResponse {
    pub request_id: String,
    pub submissions: Vec<PresentedCredential>,
    pub fulfilled_at: Timestamp,
}

#[cfg(test)]
mod tests {
    use crate::{
        ConstraintRule, Identifier, IdentifierKind, IdentifierRef, PresentationConstraint,
        PresentationRequest,
    };

    #[test]
    fn validates_non_empty_constraints() {
        let request = PresentationRequest {
            id: String::from("request-1"),
            verifier: IdentifierRef::new(
                IdentifierKind::DidWeb,
                Identifier::new("did:web:verifier.example").expect("verifier"),
            ),
            purpose: String::from("KYC reuse"),
            nonce: String::from("nonce"),
            constraints: vec![PresentationConstraint {
                id: String::from("constraint-1"),
                description: String::from("must prove KYC status"),
                rule: ConstraintRule::Predicate(String::from("kyc_verified == true")),
            }],
        };

        assert!(request.validate().is_ok());
    }
}
