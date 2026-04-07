import { AppShell, SurfaceCard } from "@hdip/ui";

import { issuerApi, trustRegistryApi } from "../lib/api";

export default function IssuerConsolePage() {
  return (
    <AppShell
      eyebrow="HDIP foundation"
      title="Issuer Console"
      description="This shell establishes typed API boundaries, environment loading, and shared layout discipline before any real credential issuance logic lands."
    >
      <SurfaceCard
        title="Issuer API boundary"
        body={`Configured endpoint: ${issuerApi.baseUrl}. Health and readiness are the only live contracts in this slice.`}
        accent="positive"
      />
      <SurfaceCard
        title="Trust registry boundary"
        body={`Configured endpoint: ${trustRegistryApi.baseUrl}. Registry business logic is intentionally deferred.`}
      />
      <SurfaceCard
        title="Auth placeholder"
        body="Authentication is intentionally absent in this slice. Real operator auth and authorization land after the service and contract baseline is stable."
      />
    </AppShell>
  );
}
