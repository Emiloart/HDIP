import Link from "next/link";

import { CredentialLookupWorkflow } from "../components";

export default function CredentialsPage() {
  return (
    <main className="console-shell">
      <nav className="top-nav" aria-label="Issuer console navigation">
        <Link href="/">Home</Link>
        <Link href="/create">Create credential</Link>
      </nav>
      <CredentialLookupWorkflow />
    </main>
  );
}
