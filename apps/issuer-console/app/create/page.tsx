import Link from "next/link";

import { CreateCredentialWorkflow } from "../components";

export default function CreateCredentialPage() {
  return (
    <main className="console-shell">
      <nav className="top-nav" aria-label="Issuer console navigation">
        <Link href="/">Home</Link>
        <Link href="/credentials">Credentials</Link>
      </nav>
      <CreateCredentialWorkflow />
    </main>
  );
}
