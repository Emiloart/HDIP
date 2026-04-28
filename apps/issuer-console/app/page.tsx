import Link from "next/link";

export default function IssuerConsolePage() {
  return (
    <main className="console-shell">
      <header className="console-header">
        <p className="eyebrow">HDIP issuer operations</p>
        <h1>Issuer Console</h1>
        <p>Create reusable KYC credentials, look them up by credential ID, and manage terminal status changes.</p>
      </header>
      <nav className="console-nav" aria-label="Issuer console navigation">
        <Link href="/create">Create credential</Link>
        <Link href="/credentials">Credentials</Link>
      </nav>
    </main>
  );
}
