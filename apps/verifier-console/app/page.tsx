import { AppShell, SurfaceCard } from "@hdip/ui";

import { trustRegistryApi, verifierApi } from "../lib/api";

export default function VerifierConsolePage() {
  return (
    <AppShell
      eyebrow="HDIP foundation"
      title="Verifier Console"
      description="This shell preserves clear operator-surface boundaries while contracts, health endpoints, and typed client edges settle before verifier policy logic begins."
    >
      <SurfaceCard
        title="Verifier API boundary"
        body={`Configured endpoint: ${verifierApi.baseUrl}. The shell is limited to typed client setup and presentation-safe UI framing.`}
        accent="positive"
      />
      <SurfaceCard
        title="Trust registry boundary"
        body={`Configured endpoint: ${trustRegistryApi.baseUrl}. Trust metadata flows are not implemented in this slice.`}
      />
      <SurfaceCard
        title="Auth placeholder"
        body="Real auth and role controls are intentionally deferred. The app establishes the place where they will later attach without hard-coding business logic now."
      />
    </AppShell>
  );
}
