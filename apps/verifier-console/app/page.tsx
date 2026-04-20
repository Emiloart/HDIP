import { AppShell, SurfaceCard } from "@hdip/ui";

import { verifierApi } from "../lib/api";

export default async function VerifierConsolePage() {
  const [policyRequest, result] = await Promise.all([
    verifierApi.policyRequest("kyc-passport-basic"),
    verifierApi.stubResult("kyc-passport-basic-review"),
  ]);

  return (
    <AppShell
      eyebrow="HDIP foundation"
      title="Verifier Console"
      description="This shell now exercises the first stubbed verifier flow: a policy request and a deterministic result are fetched through typed client boundaries while real proof submission remains deferred."
    >
      <SurfaceCard
        title="Policy request"
        body={`${policyRequest.requestId}: ${policyRequest.purpose}. Required predicates: ${policyRequest.requiredPredicates.join("; ")}.`}
        accent="positive"
      />
      <SurfaceCard
        title="Stub verifier result"
        body={`Decision: ${result.decision}. Reasons: ${result.reasons.join("; ")}.`}
      />
      <SurfaceCard
        title="Flow boundary"
        body="This is still a read-only foundation flow. No credential proof, holder submission, or live trust-registry composition happens in this slice."
      />
    </AppShell>
  );
}
