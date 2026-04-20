import { AppShell, SurfaceCard } from "@hdip/ui";

import { issuerApi } from "../lib/api";

export default async function IssuerConsolePage() {
  const [profile, template] = await Promise.all([
    issuerApi.profile(),
    issuerApi.template("hdip-passport-basic"),
  ]);

  return (
    <AppShell
      eyebrow="HDIP foundation"
      title="Issuer Console"
      description="This shell now exercises the first stubbed issuer flow: profile metadata and a credential template are fetched through typed client boundaries while real issuance remains deferred."
    >
      <SurfaceCard
        title="Issuer profile"
        body={`${profile.displayName} (${profile.issuerId}) publishes its stub issuer endpoint at ${profile.endpoint}.`}
        accent="positive"
      />
      <SurfaceCard
        title="Credential template"
        body={`${template.displayName} v${template.version} advertises credential types ${template.credentialTypes.join(", ")}.`}
      />
      <SurfaceCard
        title="Stub flow handoff"
        body={`Supported template IDs: ${profile.supportedCredentialTemplates.join(", ")}. The verifier console consumes the matching stub policy request and decision path next.`}
      />
    </AppShell>
  );
}
