import Link from "next/link";

import { CredentialDetailWorkflow } from "../../components";

export default async function CredentialDetailPage(props: {
  params: Promise<{ credentialId: string }>;
}) {
  const params = await props.params;

  return (
    <main className="console-shell">
      <nav className="top-nav" aria-label="Issuer console navigation">
        <Link href="/">Home</Link>
        <Link href="/credentials">Credentials</Link>
        <Link href="/create">Create credential</Link>
      </nav>
      <CredentialDetailWorkflow credentialId={params.credentialId} />
    </main>
  );
}
