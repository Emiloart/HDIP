import type { ReactNode } from "react";

type AppShellProps = {
  eyebrow: string;
  title: string;
  description: string;
  children: ReactNode;
};

type SurfaceCardProps = {
  title: string;
  body: string;
  accent?: "neutral" | "positive";
};

export function AppShell(props: AppShellProps) {
  return (
    <main
      style={{
        maxWidth: "960px",
        margin: "0 auto",
        padding: "56px 24px 80px",
        display: "grid",
        gap: "24px",
      }}
    >
      <header style={{ display: "grid", gap: "12px" }}>
        <div
          style={{
            fontSize: "12px",
            fontWeight: 700,
            letterSpacing: "0.16em",
            textTransform: "uppercase",
            color: "#6b7280",
          }}
        >
          {props.eyebrow}
        </div>
        <h1 style={{ fontSize: "40px", lineHeight: 1.05, margin: 0 }}>{props.title}</h1>
        <p style={{ fontSize: "18px", lineHeight: 1.6, margin: 0, color: "#475569" }}>
          {props.description}
        </p>
      </header>
      <section
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
          gap: "16px",
        }}
      >
        {props.children}
      </section>
    </main>
  );
}

export function SurfaceCard(props: SurfaceCardProps) {
  const borderColor = props.accent === "positive" ? "#14532d" : "#cbd5e1";
  const background = props.accent === "positive" ? "#f0fdf4" : "#ffffff";

  return (
    <article
      style={{
        border: `1px solid ${borderColor}`,
        background,
        borderRadius: "18px",
        padding: "20px",
        minHeight: "160px",
        display: "grid",
        gap: "12px",
      }}
    >
      <h2 style={{ margin: 0, fontSize: "18px" }}>{props.title}</h2>
      <p style={{ margin: 0, lineHeight: 1.6, color: "#334155" }}>{props.body}</p>
    </article>
  );
}
