package httpapi

const defaultPolicyID = "kyc-passport-basic"
const defaultRequestID = "kyc-passport-basic-review"

type verifierPolicyRequestResponse struct {
	RequestID          string   `json:"requestId"`
	Purpose            string   `json:"purpose"`
	RequiredPredicates []string `json:"requiredPredicates"`
}

type verifierResultResponse struct {
	RequestID string   `json:"requestId"`
	Decision  string   `json:"decision"`
	Reasons   []string `json:"reasons"`
}

func stubPolicyRequest(policyID string) (verifierPolicyRequestResponse, bool) {
	if policyID != defaultPolicyID {
		return verifierPolicyRequestResponse{}, false
	}

	return verifierPolicyRequestResponse{
		RequestID: defaultRequestID,
		Purpose:   "Review a reusable HDIP passport credential for marketplace onboarding",
		RequiredPredicates: []string{
			"credentialType == HDIPPassportCredential",
			"credentialType == KycCredential",
			"issuerId == did:web:issuer.hdip.dev",
		},
	}, true
}

func stubVerifierResult(requestID string) (verifierResultResponse, bool) {
	if requestID != defaultRequestID {
		return verifierResultResponse{}, false
	}

	return verifierResultResponse{
		RequestID: defaultRequestID,
		Decision:  "allow",
		Reasons: []string{
			"stub flow matched the expected issuer profile",
			"stub flow matched the HDIP passport template contract",
			"real credential verification is deferred in this slice",
		},
	}, true
}
