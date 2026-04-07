"use client";

export default function ErrorBoundary(props: { error: Error; reset: () => void }) {
  return (
    <main style={{ padding: "32px", display: "grid", gap: "16px" }}>
      <h1 style={{ margin: 0 }}>Issuer console shell failed</h1>
      <p style={{ margin: 0 }}>{props.error.message}</p>
      <button onClick={props.reset} style={{ width: "fit-content" }}>
        Retry
      </button>
    </main>
  );
}
